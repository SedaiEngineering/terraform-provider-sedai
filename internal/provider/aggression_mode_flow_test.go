package provider_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// fakeModeServer emulates the Sedai 1.148 settings endpoints that matter for
// aggression mode: one group, one account, and a Kubernetes, an ECS and a
// Lambda resource. A switch behaves as seen live (SED-24714): it sets the
// mode, derives isProd, toggles autonomousActionWithoutTraffic and clears the
// custom CPU floor to -1. Every write is logged, so tests can check the order
// of switches and saves.
type fakeModeServer struct {
	*httptest.Server
	mu        sync.Mutex
	group     map[string]interface{}
	account   map[string]interface{}
	resources map[string]map[string]interface{}
	writes    []string
	// failMode makes a switch fail: key "group:KUBE", "account:ECS" or
	// "resource:<id>"; value "false" (status OK, result false: nothing saved)
	// or "500" (server error, as a server older than 1.148 answers).
	failMode map[string]string
}

func kubeSection(mode string, isProd bool) map[string]interface{} {
	return map[string]interface{}{
		"availability":                   map[string]interface{}{"configMode": "DATA_PILOT"},
		"optimization":                   map[string]interface{}{"setting": map[string]interface{}{"configMode": "DATA_PILOT"}},
		"sedaiSync":                      map[string]interface{}{"enabled": false},
		"isProd":                         map[string]interface{}{"enabled": isProd},
		"optimizationMode":               map[string]interface{}{"mode": mode},
		"autonomousActionWithoutTraffic": map[string]interface{}{"enabled": !isProd},
		"enableVerticalScaling": map[string]interface{}{
			"enabled": true, "minPerContainerCpuInCores": 0.25, "minPerContainerMemoryInBytes": 1000,
		},
	}
}

func settingsBlob() map[string]interface{} {
	return map[string]interface{}{
		"appCommonSettings": map[string]interface{}{
			"availability": map[string]interface{}{"configMode": "DATA_PILOT"},
			"optimization": map[string]interface{}{"setting": map[string]interface{}{"configMode": "DATA_PILOT"}},
			"sedaiSync":    map[string]interface{}{"enabled": false},
		},
		"kubeSettings": kubeSection("PROD", true),
		"ecsSettings":  kubeSection("PROD", true),
	}
}

func newFakeModeServer() *fakeModeServer {
	f := &fakeModeServer{
		group:     settingsBlob(),
		account:   settingsBlob(),
		resources: map[string]map[string]interface{}{},
		failMode:  map[string]string{},
	}
	for id, typ := range map[string]string{
		"kube-1": "KubernetesAppResourceSettingDetail",
		"ecs-1":  "ECSAppResourceSettingDetail",
	} {
		s := kubeSection("PROD", true)
		s["type"] = typ
		f.resources[id] = s
	}
	f.resources["lambda-1"] = map[string]interface{}{
		"type":         "ServerlessFunctionResourceSettingDetail",
		"availability": map[string]interface{}{"configMode": "DATA_PILOT"},
		"optimization": map[string]interface{}{"setting": map[string]interface{}{"configMode": "DATA_PILOT"}},
		"sedaiSync":    map[string]interface{}{"enabled": false},
	}

	const base = "/api/settingsV2/topology/configs/"
	sectionKey := map[string]string{"KUBE": "kubeSettings", "ECS": "ecsSettings"}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/settingsV2/topology/group/init", func(w http.ResponseWriter, r *http.Request) {
		writeOK(w, true)
	})
	mux.HandleFunc(base, func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		p := strings.TrimPrefix(r.URL.Path, base)
		q := r.URL.Query()
		switch {
		case r.Method == http.MethodGet && p == "group":
			writeOK(w, map[string]interface{}{"setting": f.group})
		case r.Method == http.MethodGet && p == "account":
			writeOK(w, map[string]interface{}{"attributes": map[string]interface{}{"setting": f.account}})
		case r.Method == http.MethodGet && p == "resource":
			s, ok := f.resources[q.Get("resourceId")]
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeOK(w, map[string]interface{}{"attributes": map[string]interface{}{"setting": s}})

		case r.Method == http.MethodPost && (p == "group/mode" || p == "account/mode"):
			scope := strings.TrimSuffix(p, "/mode")
			rt, mode := q.Get("resourceType"), q.Get("mode")
			if f.fail(w, scope+":"+rt, "mode:"+scope+":"+rt+":"+mode) {
				return
			}
			blob := f.group
			if scope == "account" {
				blob = f.account
			}
			applySwitch(blob[sectionKey[rt]].(map[string]interface{}), mode)
			f.writes = append(f.writes, "mode:"+scope+":"+rt+":"+mode)
			writeOK(w, true)
		case r.Method == http.MethodPost && p == "resource/mode":
			id, mode := q.Get("resourceId"), q.Get("mode")
			if f.fail(w, "resource:"+id, "mode:resource:"+id+":"+mode) {
				return
			}
			applySwitch(f.resources[id], mode)
			f.writes = append(f.writes, "mode:resource:"+id+":"+mode)
			writeOK(w, true)

		case r.Method == http.MethodPost && p == "resource":
			var body map[string]interface{}
			readJSON(r, &body)
			id := body["resourceId"].(string)
			f.resources[id] = body["setting"].(map[string]interface{})
			f.writes = append(f.writes, "save:resource:"+id)
			writeOK(w, true)
		case r.Method == http.MethodPost && strings.HasPrefix(p, "group/"):
			readJSON(r, &f.group)
			f.writes = append(f.writes, "save:group")
			writeOK(w, true)
		case r.Method == http.MethodPost && strings.HasPrefix(p, "account/"):
			readJSON(r, &f.account)
			f.writes = append(f.writes, "save:account")
			writeOK(w, true)
		default:
			http.NotFound(w, r)
		}
	})
	f.Server = httptest.NewServer(mux)
	return f
}

// fail answers a switch with the configured failure, logs it, and reports
// whether it did. Called with f.mu held.
func (f *fakeModeServer) fail(w http.ResponseWriter, key, logEntry string) bool {
	switch f.failMode[key] {
	case "false":
		f.writes = append(f.writes, "failed-"+logEntry)
		writeOK(w, false)
		return true
	case "500":
		f.writes = append(f.writes, "failed-"+logEntry)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "ERROR", "message": "Internal Server Error"})
		return true
	}
	return false
}

func applySwitch(section map[string]interface{}, mode string) {
	section["optimizationMode"] = map[string]interface{}{"mode": mode}
	section["isProd"] = map[string]interface{}{"enabled": mode == "PROD"}
	section["autonomousActionWithoutTraffic"] = map[string]interface{}{"enabled": mode != "PROD"}
	section["enableVerticalScaling"].(map[string]interface{})["minPerContainerCpuInCores"] = -1.0
}

// setFailures replaces the switch failures; nil clears them.
func (f *fakeModeServer) setFailures(m map[string]string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if m == nil {
		m = map[string]string{}
	}
	f.failMode = m
}

// switchOutsideTerraform changes a mode the way the UI would.
func (f *fakeModeServer) switchOutsideTerraform(scope, rt, mode string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	blob := f.group
	if scope == "account" {
		blob = f.account
	}
	applySwitch(blob[map[string]string{"KUBE": "kubeSettings", "ECS": "ecsSettings"}[rt]].(map[string]interface{}), mode)
}

// takeWrites returns the writes logged since the last call.
func (f *fakeModeServer) takeWrites() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := f.writes
	f.writes = nil
	return out
}

// kubeValue reads a nested value from a group or account kubeSettings section.
func (f *fakeModeServer) kubeValue(scope string, keys ...string) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	blob := f.group
	if scope == "account" {
		blob = f.account
	}
	var v interface{} = blob["kubeSettings"]
	for _, k := range keys {
		v = v.(map[string]interface{})[k]
	}
	return v
}

func writeOK(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "OK", "result": result})
}

func readJSON(r *http.Request, v interface{}) {
	b, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(b, v)
}

func expectWrites(f *fakeModeServer, want ...string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if got := f.takeWrites(); !reflect.DeepEqual(got, want) {
			return fmt.Errorf("writes: got %v, want %v", got, want)
		}
		return nil
	}
}

// expectWritesBefore checks the writes of the previous step from a PreConfig,
// for steps that end in ExpectError (their Check doesn't run).
func expectWritesBefore(t *testing.T, f *fakeModeServer, want ...string) {
	t.Helper()
	if got := f.takeWrites(); !reflect.DeepEqual(got, want) {
		t.Errorf("writes of the previous step: got %v, want %v", got, want)
	}
}

func useFakeServer(t *testing.T) *fakeModeServer {
	f := newFakeModeServer()
	t.Cleanup(f.Close)
	t.Setenv("SEDAI_BASE_URL", f.URL)
	t.Setenv("SEDAI_API_TOKEN", "fake-token")
	return f
}

func groupSettingsConfig(kube, ecs string) string {
	return fmt.Sprintf(`
resource "sedai_group_settings" "test" {
  group_id          = "grp-1"
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  kube_app_settings {
%s
  }
  ecs_app_settings {
%s
  }
}`, kube, ecs)
}

func accountSettingsConfig(kube, ecs string) string {
	return fmt.Sprintf(`
resource "sedai_account_settings" "test" {
  account_id        = "acct-1"
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  kube_app_settings {
%s
  }
  ecs_app_settings {
%s
  }
}`, kube, ecs)
}

// TestSettingsAggressionModeFlow_Group checks, with real Terraform runs,
// that a mode change switches before saving, that a guardrail set in the
// same block survives the switch's reset with no drift afterwards, that an
// unchanged mode is not switched again, and the plan-time refusals.
func TestSettingsAggressionModeFlow_Group(t *testing.T) {
	f := useFakeServer(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Switch kube to TURBO and set the CPU floor the switch resets.
			{
				Config: groupSettingsConfig(`
    aggression_mode                = "TURBO"
    vertical_scaling_min_cpu_cores = 0.5`, `
    aggression_mode = "EXPEDITE"`),
				Check: resource.ComposeTestCheckFunc(
					expectWrites(f, "mode:group:KUBE:TURBO", "mode:group:ECS:EXPEDITE", "save:group"),
					resource.TestCheckResourceAttr("sedai_group_settings.test", "kube_app_settings.aggression_mode", "TURBO"),
					resource.TestCheckResourceAttr("sedai_group_settings.test", "ecs_app_settings.aggression_mode", "EXPEDITE"),
					func(*terraform.State) error {
						if got := f.kubeValue("group", "enableVerticalScaling", "minPerContainerCpuInCores"); got != 0.5 {
							return fmt.Errorf("CPU floor after apply: got %v, want 0.5 (saved after the switch)", got)
						}
						if got := f.kubeValue("group", "optimizationMode", "mode"); got != "TURBO" {
							return fmt.Errorf("stored mode: got %v, want TURBO", got)
						}
						return nil
					},
				),
			},
			// Same modes, new guardrail: no switch, only a save.
			{
				Config: groupSettingsConfig(`
    aggression_mode                = "TURBO"
    vertical_scaling_min_cpu_cores = 0.75`, `
    aggression_mode = "EXPEDITE"`),
				Check: expectWrites(f, "save:group"),
			},
			// Back to PROD, with an is_prod that agrees.
			{
				Config: groupSettingsConfig(`
    aggression_mode                = "PROD"
    is_prod                        = true
    vertical_scaling_min_cpu_cores = 0.75`, `
    aggression_mode = "EXPEDITE"`),
				Check: expectWrites(f, "mode:group:KUBE:PROD", "save:group"),
			},
			// is_prod = true with TURBO is refused at plan time.
			{
				Config: groupSettingsConfig(`
    aggression_mode = "TURBO"
    is_prod         = true`, ``),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`is_prod = true requires aggression_mode = "PROD"`),
			},
			// TURBO on ECS is refused at plan time.
			{
				Config:      groupSettingsConfig(``, `    aggression_mode = "TURBO"`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`TURBO is not available for ECS`),
			},
			// A valid last config, so the post-test destroy can run.
			{
				Config: groupSettingsConfig(`
    aggression_mode                = "PROD"
    is_prod                        = true
    vertical_scaling_min_cpu_cores = 0.75`, `
    aggression_mode = "EXPEDITE"`),
			},
		},
	})
}

// TestSettingsAggressionModeFlow_Account checks the account path end to end:
// both switches run before the save, the next plan is empty, and destroy
// leaves the aggression mode alone.
func TestSettingsAggressionModeFlow_Account(t *testing.T) {
	f := useFakeServer(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			for _, w := range f.takeWrites() {
				if strings.HasPrefix(w, "mode:") {
					return fmt.Errorf("destroy must not switch the aggression mode, got %q", w)
				}
			}
			if got := f.kubeValue("account", "optimizationMode", "mode"); got != "TURBO" {
				return fmt.Errorf("destroy must leave the aggression mode: got %v, want TURBO", got)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: accountSettingsConfig(`
    aggression_mode                = "TURBO"
    is_prod                        = false
    vertical_scaling_min_cpu_cores = 0.5`, `
    aggression_mode = "EXPEDITE"`),
				Check: resource.ComposeTestCheckFunc(
					expectWrites(f, "mode:account:KUBE:TURBO", "mode:account:ECS:EXPEDITE", "save:account"),
					resource.TestCheckResourceAttr("sedai_account_settings.test", "kube_app_settings.aggression_mode", "TURBO"),
					resource.TestCheckResourceAttr("sedai_account_settings.test", "ecs_app_settings.aggression_mode", "EXPEDITE"),
					func(*terraform.State) error {
						// is_prod = false is saved after the switch and keeps TURBO.
						if got := f.kubeValue("account", "optimizationMode", "mode"); got != "TURBO" {
							return fmt.Errorf("stored mode: got %v, want TURBO", got)
						}
						if got := f.kubeValue("account", "enableVerticalScaling", "minPerContainerCpuInCores"); got != 0.5 {
							return fmt.Errorf("CPU floor: got %v, want 0.5", got)
						}
						return nil
					},
				),
			},
		},
	})
}

// TestSettingsAggressionModeFlow_DriftAndImport checks that a mode changed
// outside Terraform (in the UI) shows in the plan and is switched back by the
// next apply, and that import reads the mode.
func TestSettingsAggressionModeFlow_DriftAndImport(t *testing.T) {
	f := useFakeServer(t)
	cfg := groupSettingsConfig(`    aggression_mode = "TURBO"`, ``)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check:  expectWrites(f, "mode:group:KUBE:TURBO", "save:group"),
			},
			// Someone switches the group to EXPEDITE in the UI: the plan shows it.
			{
				PreConfig:          func() { f.switchOutsideTerraform("group", "KUBE", "EXPEDITE") },
				Config:             cfg,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			// The next apply switches it back, then saves.
			{
				Config: cfg,
				Check:  expectWrites(f, "mode:group:KUBE:TURBO", "save:group"),
			},
			{
				ResourceName:  "sedai_group_settings.test",
				ImportState:   true,
				ImportStateId: "grp-1",
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("want 1 imported state, got %d", len(states))
					}
					a := states[0].Attributes
					if a["kube_app_settings.aggression_mode"] != "TURBO" || a["ecs_app_settings.aggression_mode"] != "PROD" {
						return fmt.Errorf("imported modes: kube %q, ecs %q; want TURBO, PROD",
							a["kube_app_settings.aggression_mode"], a["ecs_app_settings.aggression_mode"])
					}
					return nil
				},
			},
		},
	})
}

// TestSettingsAggressionModeFlow_DataSources checks that the data sources
// return the mode, and null for a resource that is not Kubernetes or ECS.
func TestSettingsAggressionModeFlow_DataSources(t *testing.T) {
	f := useFakeServer(t)
	f.switchOutsideTerraform("group", "KUBE", "TURBO")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "sedai_group_settings" "g"    { group_id = "grp-1" }
data "sedai_account_settings" "a"  { account_id = "acct-1" }
data "sedai_resource_settings" "k" { resource_id = "kube-1" }
data "sedai_resource_settings" "l" { resource_id = "lambda-1" }`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.sedai_group_settings.g", "kube_app_settings.aggression_mode", "TURBO"),
					resource.TestCheckResourceAttr("data.sedai_group_settings.g", "ecs_app_settings.aggression_mode", "PROD"),
					resource.TestCheckResourceAttr("data.sedai_account_settings.a", "kube_app_settings.aggression_mode", "PROD"),
					resource.TestCheckResourceAttr("data.sedai_resource_settings.k", "aggression_mode", "PROD"),
					resource.TestCheckNoResourceAttr("data.sedai_resource_settings.l", "aggression_mode"),
				),
			},
		},
	})
}

// TestSettingsAggressionModeFlow_SwitchFailures checks that a failed switch
// stops the apply before the save, keeps the server's reason, and that an
// apply after a partial failure only does what is left.
func TestSettingsAggressionModeFlow_SwitchFailures(t *testing.T) {
	f := useFakeServer(t)
	both := groupSettingsConfig(`    aggression_mode = "TURBO"`, `    aggression_mode = "EXPEDITE"`)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// status OK with result false: nothing was saved, so it is an error.
			{
				PreConfig:   func() { f.setFailures(map[string]string{"group:KUBE": "false"}) },
				Config:      both,
				ExpectError: regexp.MustCompile(`was not changed`),
			},
			// A server error (as from Sedai older than 1.148) gets the hint.
			{
				PreConfig: func() {
					expectWritesBefore(t, f, "failed-mode:group:KUBE:TURBO") // no save after a failed switch
					f.setFailures(map[string]string{"group:KUBE": "500"})
				},
				Config:      both,
				ExpectError: regexp.MustCompile(`(?s)Sedai 1\.148.*admin-role`),
			},
			// Kubernetes switches, ECS fails: the apply stops before the save.
			{
				PreConfig: func() {
					expectWritesBefore(t, f, "failed-mode:group:KUBE:TURBO")
					f.setFailures(map[string]string{"group:ECS": "false"})
				},
				Config:      both,
				ExpectError: regexp.MustCompile(`was not changed`),
			},
			// Re-apply after the fix: only ECS is switched; Kubernetes is already TURBO.
			{
				PreConfig: func() {
					expectWritesBefore(t, f, "mode:group:KUBE:TURBO", "failed-mode:group:ECS:EXPEDITE")
					f.setFailures(nil)
				},
				Config: both,
				Check:  expectWrites(f, "mode:group:ECS:EXPEDITE", "save:group"),
			},
		},
	})
}

func resourceSettingsConfig(id, mode string) string {
	return fmt.Sprintf(`
resource "sedai_resource_settings" "test" {
  resource_id       = %q
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  aggression_mode   = %q
}`, id, mode)
}

// TestSettingsAggressionModeFlow_Resource checks the resource-level switch
// and that the SDK refuses TURBO on an ECS resource before sending it.
func TestSettingsAggressionModeFlow_Resource(t *testing.T) {
	f := useFakeServer(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: resourceSettingsConfig("kube-1", "TURBO"),
				Check: resource.ComposeTestCheckFunc(
					expectWrites(f, "mode:resource:kube-1:TURBO", "save:resource:kube-1"),
					resource.TestCheckResourceAttr("sedai_resource_settings.test", "aggression_mode", "TURBO"),
				),
			},
		},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      resourceSettingsConfig("ecs-1", "TURBO"),
				ExpectError: regexp.MustCompile(`(?i)turbo`),
			},
		},
	})
	for _, w := range f.takeWrites() {
		if strings.Contains(w, "ecs-1") {
			t.Errorf("TURBO on ECS must not reach the server, got write %q", w)
		}
	}
}
