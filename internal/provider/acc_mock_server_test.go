package provider_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// mockSedaiServer is a lightweight HTTP server that emulates the Sedai API
// for unit-level error-path testing without hitting a real backend.
type mockSedaiServer struct {
	server *httptest.Server
	mu     sync.Mutex

	// EOFOnAccountCreateN: close the connection (simulate EOF) on the Nth POST to
	// /api/site/accounts. 0 = never close.
	EOFOnAccountCreateN int
	// AccountCreateCallCount tracks how many times the account create endpoint was hit.
	AccountCreateCallCount int
	// CreatedAccounts stores accounts as returned to callers.
	CreatedAccounts []map[string]interface{}
	// CreatedGroups stores groups returned to callers.
	CreatedGroups []map[string]interface{}
	// ReturnHTTPCode overrides the HTTP status for ALL non-EOF responses when non-zero.
	ReturnHTTPCode int
	// ReturnBody overrides the response body when non-empty.
	ReturnBody string
}

// newMockServer starts a mock Sedai API server and returns it.
// Call Close() when done.
func newMockServer() *mockSedaiServer {
	m := &mockSedaiServer{}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/site/accounts", func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()

		if r.Method == http.MethodPost {
			m.AccountCreateCallCount++
			n := m.AccountCreateCallCount

			if m.EOFOnAccountCreateN > 0 && n == m.EOFOnAccountCreateN {
				// Simulate EOF by hijacking and closing the connection
				hj, ok := w.(http.Hijacker)
				if ok {
					conn, _, _ := hj.Hijack()
					conn.Close()
					return
				}
			}

			if m.ReturnHTTPCode != 0 {
				m.writeError(w, m.ReturnHTTPCode)
				return
			}

			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)

			name, _ := req["name"].(string)
			id := fmt.Sprintf("mock-acct-%03d", n)
			acct := map[string]interface{}{
				"id":             id,
				"name":           name,
				"cloud_provider": req["cloudProvider"],
			}
			m.CreatedAccounts = append(m.CreatedAccounts, acct)
			m.writeJSON(w, http.StatusOK, map[string]interface{}{
				"status": "OK",
				"result": acct,
			})
			return
		}

		if r.Method == http.MethodGet {
			m.writeJSON(w, http.StatusOK, map[string]interface{}{
				"status": "OK",
				"result": m.CreatedAccounts,
			})
			return
		}

		http.NotFound(w, r)
	})

	mux.HandleFunc("/api/site/accounts/", func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()

		id := strings.TrimPrefix(r.URL.Path, "/api/site/accounts/")

		if r.Method == http.MethodDelete {
			m.writeJSON(w, http.StatusOK, map[string]interface{}{"status": "OK"})
			return
		}

		if r.Method == http.MethodGet {
			for _, acct := range m.CreatedAccounts {
				if acct["id"] == id {
					m.writeJSON(w, http.StatusOK, map[string]interface{}{
						"status": "OK",
						"result": acct,
					})
					return
				}
			}
			m.writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"status": "ERROR",
				"error":  "account not found",
			})
			return
		}

		http.NotFound(w, r)
	})

	mux.HandleFunc("/api/userProfile", func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.writeJSON(w, http.StatusOK, map[string]interface{}{
			"displayName": "test-user",
			"userId":      "user-1",
		})
	})

	mux.HandleFunc("/api/sedaigroup/create/", func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()

		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}

		if m.ReturnHTTPCode != 0 {
			m.writeError(w, m.ReturnHTTPCode)
			return
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		_ = json.Unmarshal(body, &req)

		id := fmt.Sprintf("mock-grp-%03d", len(m.CreatedGroups)+1)
		grp := map[string]interface{}{
			"id":   id,
			"name": req["name"],
		}
		m.CreatedGroups = append(m.CreatedGroups, grp)
		m.writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "OK",
			"result": grp,
		})
	})

	mux.HandleFunc("/api/sedaigroup/groups", func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "OK",
			"result": m.CreatedGroups,
		})
	})

	mux.HandleFunc("/api/sedaigroup/allGroupDefinitions/", func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()

		id := strings.TrimPrefix(r.URL.Path, "/api/sedaigroup/allGroupDefinitions/")
		for _, grp := range m.CreatedGroups {
			if grp["id"] == id {
				m.writeJSON(w, http.StatusOK, map[string]interface{}{
					"status": "OK",
					"result": grp,
				})
				return
			}
		}
		m.writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"status": "ERROR",
			"error":  "group not found",
		})
	})

	mux.HandleFunc("/api/graphql", func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{
				"deleteGroupDefinitionByGroupId": true,
			},
		})
	})

	mux.HandleFunc("/api/settingsV2/topology/configs/group", func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()

		if r.Method == http.MethodPost {
			m.writeJSON(w, http.StatusOK, map[string]interface{}{"status": "OK"})
			return
		}
		if r.Method == http.MethodGet {
			m.writeJSON(w, http.StatusOK, map[string]interface{}{
				"status": "OK",
				"result": map[string]interface{}{
					"availabilityMode": "DATA_PILOT",
					"optimizationMode": "DATA_PILOT",
					"sedaiSyncEnabled": false,
				},
			})
			return
		}
		http.NotFound(w, r)
	})

	// Catch-all for endpoints not explicitly handled
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()

		if m.ReturnHTTPCode != 0 {
			m.writeError(w, m.ReturnHTTPCode)
			return
		}
		m.writeJSON(w, http.StatusOK, map[string]interface{}{
			"status": "OK",
			"result": nil,
		})
	})

	m.server = httptest.NewServer(mux)
	return m
}

func (m *mockSedaiServer) URL() string {
	return m.server.URL
}

func (m *mockSedaiServer) Close() {
	m.server.Close()
}

func (m *mockSedaiServer) writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (m *mockSedaiServer) writeError(w http.ResponseWriter, code int) {
	msg := http.StatusText(code)
	m.writeJSON(w, code, map[string]interface{}{
		"status": "ERROR",
		"error":  msg,
	})
}
