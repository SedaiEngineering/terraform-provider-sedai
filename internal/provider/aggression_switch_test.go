package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/impl"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// switchAPI is a minimal fake of the settings endpoints for the switch
// helpers: the current mode of a group and of a resource, and a log of
// every request.
type switchAPI struct {
	mu        sync.Mutex
	requests  []string
	groupMode string // "" = the group has no settings
	resMode   string // "" = the resource has no settings
}

func (a *switchAPI) start(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		defer a.mu.Unlock()
		a.requests = append(a.requests, r.Method+" "+r.URL.Path)
		section := func(mode string) map[string]interface{} {
			return map[string]interface{}{
				"isProd":           map[string]interface{}{"enabled": mode == "PROD"},
				"optimizationMode": map[string]interface{}{"mode": mode},
			}
		}
		var body map[string]interface{}
		switch r.URL.Path {
		case "/api/settingsV2/topology/configs/group":
			if a.groupMode == "" {
				body = map[string]interface{}{"status": "ERROR", "message": "group not found"}
			} else {
				body = map[string]interface{}{"status": "OK", "result": map[string]interface{}{
					"setting": map[string]interface{}{"kubeSettings": section(a.groupMode)},
				}}
			}
		case "/api/settingsV2/topology/configs/resource":
			if a.resMode == "" {
				body = map[string]interface{}{"status": "ERROR", "message": "resource not found"}
			} else {
				s := section(a.resMode)
				s["type"] = "KubernetesAppResourceSettingDetail"
				body = map[string]interface{}{"status": "OK", "result": map[string]interface{}{
					"attributes": map[string]interface{}{"setting": s},
				}}
			}
		default: // the /mode switch endpoints apply the mode, as Sedai does
			mode := r.URL.Query().Get("mode")
			if strings.HasPrefix(r.URL.Path, "/api/settingsV2/topology/configs/group/") {
				a.groupMode = mode
			} else {
				a.resMode = mode
			}
			body = map[string]interface{}{"status": "OK", "result": true}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	impl.SetConfig(srv.URL, "fake-token")
}

func (a *switchAPI) posts() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	var out []string
	for _, r := range a.requests {
		if strings.HasPrefix(r, "POST ") {
			out = append(out, r)
		}
	}
	return out
}

func (a *switchAPI) count() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.requests)
}

func groupPlan(mode string) groupSettingsResourceModel {
	m := basetypes.NewStringNull()
	if mode != "" {
		m = basetypes.NewStringValue(mode)
	}
	return groupSettingsResourceModel{
		GroupID:         basetypes.NewStringValue("grp-1"),
		KubeAppSettings: &kubeAppSettingsModel{AggressionMode: m},
	}
}

func resourcePlan(mode string) resourceSettingsModel {
	m := basetypes.NewStringNull()
	if mode != "" {
		m = basetypes.NewStringValue(mode)
	}
	return resourceSettingsModel{ResourceID: basetypes.NewStringValue("kube-1"), AggressionMode: m}
}

func switchGroupModes(p groupSettingsResourceModel) error {
	_, err := switchGroupAggressionModes(context.Background(), p)
	return err
}

func switchResourceMode(p resourceSettingsModel) error {
	_, err := switchResourceAggressionMode(context.Background(), p)
	return err
}

func TestSettingsSwitchGroupAggressionModes(t *testing.T) {
	t.Run("not managed: no request at all", func(t *testing.T) {
		a := &switchAPI{groupMode: "PROD"}
		a.start(t)
		if err := switchGroupModes(groupPlan("")); err != nil || a.count() != 0 {
			t.Errorf("want no error and no request, got %v, %d requests", err, a.count())
		}
	})
	t.Run("already in place: read only, no switch", func(t *testing.T) {
		a := &switchAPI{groupMode: "TURBO"}
		a.start(t)
		if err := switchGroupModes(groupPlan("TURBO")); err != nil || len(a.posts()) != 0 {
			t.Errorf("want no error and no POST, got %v, %v", err, a.posts())
		}
	})
	t.Run("different: one switch", func(t *testing.T) {
		a := &switchAPI{groupMode: "PROD"}
		a.start(t)
		err := switchGroupModes(groupPlan("TURBO"))
		if p := a.posts(); err != nil || len(p) != 1 || p[0] != "POST /api/settingsV2/topology/configs/group/mode" {
			t.Errorf("want one group switch, got %v, %v", err, p)
		}
	})
	t.Run("no settings: clear error, no switch", func(t *testing.T) {
		a := &switchAPI{}
		a.start(t)
		err := switchGroupModes(groupPlan("TURBO"))
		if err == nil || !strings.Contains(err.Error(), "has no settings yet") || len(a.posts()) != 0 {
			t.Errorf("want a no-settings error and no POST, got %v, %v", err, a.posts())
		}
	})
}

func TestSettingsSwitchResourceAggressionMode(t *testing.T) {
	t.Run("not managed: no request at all", func(t *testing.T) {
		a := &switchAPI{resMode: "PROD"}
		a.start(t)
		if err := switchResourceMode(resourcePlan("")); err != nil || a.count() != 0 {
			t.Errorf("want no error and no request, got %v, %d requests", err, a.count())
		}
	})
	t.Run("already in place: no switch", func(t *testing.T) {
		a := &switchAPI{resMode: "EXPEDITE"}
		a.start(t)
		if err := switchResourceMode(resourcePlan("EXPEDITE")); err != nil || len(a.posts()) != 0 {
			t.Errorf("want no error and no POST, got %v, %v", err, a.posts())
		}
	})
	t.Run("different: one switch", func(t *testing.T) {
		a := &switchAPI{resMode: "PROD"}
		a.start(t)
		err := switchResourceMode(resourcePlan("EXPEDITE"))
		if p := a.posts(); err != nil || len(p) != 1 || p[0] != "POST /api/settingsV2/topology/configs/resource/mode" {
			t.Errorf("want one resource switch, got %v, %v", err, p)
		}
	})
	t.Run("read fails: wrapped error, no switch", func(t *testing.T) {
		srv := httptest.NewServer(http.NotFoundHandler())
		srv.Close() // connection refused
		impl.SetConfig(srv.URL, "fake-token")
		err := switchResourceMode(resourcePlan("TURBO"))
		if err == nil || !strings.Contains(err.Error(), "could not read the current aggression mode of resource kube-1") {
			t.Errorf("want a wrapped read error, got %v", err)
		}
	})
}

func TestSettingsWithSwitchHint_KeepsTheCause(t *testing.T) {
	cause := errors.New("HTTP 500")
	err := withSwitchHint(cause)
	if !errors.Is(err, cause) || !strings.HasPrefix(err.Error(), "HTTP 500. ") || !strings.Contains(err.Error(), "1.148") {
		t.Errorf("want the cause first, still wrapped, then the hint; got %v", err)
	}
}
