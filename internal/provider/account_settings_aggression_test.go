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
	"time"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/impl"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// fakeAccountServer answers "no settings" for the first notReady reads of an
// account, then settings with a PROD Kubernetes section, and logs switches.
func fakeAccountServer(t *testing.T, notReady int) (*httptest.Server, func() (reads int, switches []string)) {
	var mu sync.Mutex
	reads := 0
	mode := "PROD"
	var switches []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/settingsV2/topology/configs/account":
			reads++
			if reads <= notReady {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "ERROR", "message": "settings not available"})
				return
			}
			kube := map[string]interface{}{
				"isProd":           map[string]interface{}{"enabled": mode == "PROD"},
				"optimizationMode": map[string]interface{}{"mode": mode},
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "OK", "result": map[string]interface{}{
				"attributes": map[string]interface{}{"setting": map[string]interface{}{
					"appCommonSettings": map[string]interface{}{}, // the SDK parses sections only when present
					"kubeSettings":      kube,
				}},
			}})
		case "/api/settingsV2/topology/configs/account/mode":
			switches = append(switches, r.URL.Query().Get("resourceType")+":"+r.URL.Query().Get("mode"))
			mode = r.URL.Query().Get("mode")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "OK", "result": true})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	impl.SetConfig(srv.URL, "fake-token")
	return srv, func() (int, []string) {
		mu.Lock()
		defer mu.Unlock()
		return reads, switches
	}
}

func turboAccountPlan() accountSettingsResourceModel {
	return accountSettingsResourceModel{
		AccountID:       basetypes.NewStringValue("acct-1"),
		KubeAppSettings: &kubeAppSettingsModel{AggressionMode: basetypes.NewStringValue("TURBO")},
	}
}

func fastAccountRetries(t *testing.T) {
	orig := accountSettingsRetryDelay
	accountSettingsRetryDelay = func(int) time.Duration { return time.Millisecond }
	t.Cleanup(func() { accountSettingsRetryDelay = orig })
}

// A new account's settings can appear after sedai_account's soft 30s poll;
// the switch must wait for them like the save does, not fail the apply.
func TestSettingsSwitchAccountAggressionModes_WaitsForSettings(t *testing.T) {
	fastAccountRetries(t)
	_, stats := fakeAccountServer(t, 2)

	if _, err := switchAccountAggressionModes(context.Background(), turboAccountPlan()); err != nil {
		t.Fatalf("want the switch to succeed once settings exist, got %v", err)
	}
	reads, switches := stats()
	// 3 reads until settings exist, one KUBE:TURBO switch, then 3 matching reads.
	if reads != 6 || len(switches) != 1 || switches[0] != "KUBE:TURBO" {
		t.Errorf("want 3 reads, one KUBE:TURBO switch, 3 reads, got %d reads, switches %v", reads, switches)
	}
}

func TestSettingsSwitchAccountAggressionModes_GivesUp(t *testing.T) {
	fastAccountRetries(t)
	_, stats := fakeAccountServer(t, 100)

	_, err := switchAccountAggressionModes(context.Background(), turboAccountPlan())
	if err == nil || !strings.Contains(err.Error(), "still has no settings") {
		t.Fatalf("want a clear give-up error, got %v", err)
	}
	if reads, switches := stats(); reads != 6 || len(switches) != 0 {
		t.Errorf("want 1 read + 5 retries and no switch, got %d reads, switches %v", reads, switches)
	}
}

func TestSettingsSwitchAccountAggressionModes_StopsWhenCancelled(t *testing.T) {
	_, stats := fakeAccountServer(t, 100) // real 2s backoff: cancellation must cut it short
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	_, err := switchAccountAggressionModes(ctx, turboAccountPlan())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
	if time.Since(start) > time.Second {
		t.Errorf("cancellation should not wait for the backoff, took %v", time.Since(start))
	}
	if _, switches := stats(); len(switches) != 0 {
		t.Errorf("want no switch, got %v", switches)
	}
}
