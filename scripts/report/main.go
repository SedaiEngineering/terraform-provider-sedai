package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"sort"
	"strings"
	"time"
)

// testEvent matches the shape emitted by `go test -json`.
type testEvent struct {
	Action  string  `json:"Action"`
	Test    string  `json:"Test"`
	Package string  `json:"Package"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"`
}

type testResult struct {
	ID       string
	Name     string
	Status   string // PASS, FAIL, SKIP, NOT_RUN
	Duration float64
	Output   []string
}

var manifestNames = map[string]string{
	// PROV
	"PROV-001": "Provider configures from env vars",
	"PROV-002": "Provider base_url override",
	"PROV-003": "Provider missing base_url → error",
	"PROV-004": "Provider missing api_token → error",
	"PROV-005": "Provider invalid base_url → error",
	"PROV-006": "Provider 401 response → auth error",
	"PROV-007": "Provider timeout → error",
	"PROV-008": "Provider version attribute",
	// ACCT
	"ACCT-001": "Create AWS account (role+external_id)",
	"ACCT-002": "Create AWS account (access_key)",
	"ACCT-003": "Partial managed_services, 0-change re-plan",
	"ACCT-004": "All AWS managed services",
	"ACCT-005": "Update managed_services",
	"ACCT-006": "Destroy is idempotent",
	"ACCT-007": "Import account",
	"ACCT-008": "Import then 0-change plan",
	"ACCT-009": "Create GCP account (service_account_json)",
	"ACCT-010": "Create Azure account",
	"ACCT-011": "Create K8s account",
	"ACCT-012": "Computed fields populated after create",
	"ACCT-013": "Name is immutable (requires replace)",
	"ACCT-014": "cloud_provider is immutable",
	"ACCT-015": "integration_type is immutable",
	"ACCT-016": "region update in-place",
	"ACCT-017": "Account read after external delete",
	"ACCT-018": "Account with all optional fields",
	"ACCT-019": "Account without optional fields",
	"ACCT-020": "Account re-plan 0 changes",
	"ACCT-021": "Validator: missing cloud_provider",
	"ACCT-022": "Validator: invalid cloud_provider",
	"ACCT-023": "Validator: invalid integration_type",
	"ACCT-024": "Validator: role without external_id",
	"ACCT-025": "Validator: external_id without role",
	"ACCT-026": "Validator: access_key without secret_key",
	"ACCT-027": "Validator: secret_key without access_key",
	"ACCT-028": "Validator: access_key + role conflict",
	"ACCT-029": "Validator: access_key + service_account conflict",
	"ACCT-030": "Validator: cluster_url required for K8s",
	"ACCT-031": "Validator: project_id required for GCP",
	"ACCT-032": "Validator: tenant_id required for Azure",
	"ACCT-033": "Validator: invalid managed_service",
	"ACCT-034": "Validator: empty name",
	"ACCT-035": "Validator: name too long",
	"ACCT-036": "Create account with tags",
	"ACCT-037": "Update tags",
	"ACCT-038": "Remove tags",
	"ACCT-039": "Empty tags list",
	"ACCT-040": "Tag key-value special characters",
	"ACCT-041": "Large managed_services list",
	"ACCT-042": "Duplicate managed_service entries",
	"ACCT-043": "managed_services order stability",
	"ACCT-044": "12 concurrent accounts",
	"ACCT-045": "12 accounts 0-change re-plan",
	// GRP
	"GRP-001": "Create group",
	"GRP-002": "Create group with description",
	"GRP-003": "Create group with resource_types",
	"GRP-004": "Update group description",
	"GRP-005": "Add resource_type to group",
	"GRP-006": "Remove resource_type from group",
	"GRP-007": "Destroy group",
	"GRP-008": "Import group by ID",
	"GRP-009": "Import then 0-change plan",
	"GRP-010": "Group read after external delete",
	"GRP-011": "Group name immutable",
	"GRP-012": "Group re-plan 0 changes",
	"GRP-013": "Group with tags",
	"GRP-014": "Group tag update",
	"GRP-015": "Group empty name → error",
	"GRP-016": "Group name too long → error",
	"GRP-017": "Duplicate group name",
	"GRP-018": "Group with all resource types",
	"GRP-019": "Group resource_types order stability",
	"GRP-020": "Group with 50 resource_types",
	"GRP-021": "Create 12 groups concurrently",
	"GRP-022": "Group update re-plan 0 changes",
	"GRP-023": "Group + settings together",
	"GRP-024": "Group delete with settings attached",
	"GRP-025": "Group with CloudWatch MP",
	"GRP-026": "lambda_count not stored in state (no spurious drift)",
	"GRP-027": "ec2_count not stored in state",
	"GRP-028": "azure_vm_count not stored in state",
	"GRP-029": "rds_count not stored in state",
	"GRP-030": "ecs_count not stored in state",
	"GRP-031": "resource count re-plan 0 changes after discovery",
	"GRP-032": "Group KUBERNETES_DAEMONSET alias",
	"GRP-033": "Group list field empty",
	"GRP-034": "Group lifecycle ignore_changes no inconsistent result",
	"GRP-035": "Group settings survive group update",
	// GSET
	"GSET-001": "Create group_settings DATA_PILOT/DATA_PILOT",
	"GSET-002": "Create group_settings CO_PILOT/DATA_PILOT",
	"GSET-003": "Create group_settings AUTO/AUTO",
	"GSET-004": "Update availability_mode",
	"GSET-005": "Update optimization_mode",
	"GSET-006": "P0: sedai_sync_enabled omitted → no false→null drift",
	"GSET-007": "P0: sedai_sync_enabled omitted default is false",
	"GSET-008": "P0: sedai_sync_enabled=true persists",
	"GSET-009": "P0: sedai_sync_enabled toggle false→true→false",
	"GSET-010": "P0: sedai_sync_enabled re-plan 0 changes",
	"GSET-011": "P0: sedai_sync_enabled no drift after 3 plans",
	"GSET-012": "Destroy group_settings",
	"GSET-013": "Import group_settings",
	"GSET-014": "Import then 0-change plan",
	"GSET-015": "group_id immutable",
	"GSET-016": "kube_app_settings block",
	"GSET-017": "bucket_settings block",
	"GSET-018": "app_settings block",
	"GSET-019": "container_app_settings block",
	"GSET-020": "ecs_app_settings block",
	"GSET-021": "serverless_settings block",
	"GSET-022": "volume_settings block",
	"GSET-023": "Multiple settings blocks",
	"GSET-024": "Remove settings block",
	"GSET-025": "Settings block update",
	"GSET-026": "Validator: invalid availability_mode",
	"GSET-027": "Validator: invalid optimization_mode",
	"GSET-028": "Validator: AUTO+kube_app_settings",
	"GSET-029": "Partial spec 0-change re-plan",
	"GSET-030": "Settings survive concurrent group ops",
	"GSET-031": "kube_app_settings target_cpu field",
	"GSET-032": "serverless_settings concurrency field",
	"GSET-033": "app_settings target_memory field",
	"GSET-034": "settings block null fields = unmanaged",
	"GSET-035": "group_settings read after external delete",
	// ASET
	"ASET-001": "Create account_settings DATA_PILOT",
	"ASET-002": "Create account_settings CO_PILOT",
	"ASET-003": "Create account_settings AUTO",
	"ASET-004": "P0: sedai_sync_enabled omitted → no drift",
	"ASET-005": "P0: sedai_sync_enabled=true persists",
	"ASET-006": "P0: toggle sync false→true→false",
	"ASET-007": "Update availability_mode",
	"ASET-008": "Update optimization_mode",
	"ASET-009": "Destroy account_settings",
	"ASET-010": "Import account_settings",
	"ASET-011": "Import then 0-change plan",
	"ASET-012": "account_settings kube_app_settings block",
	"ASET-013": "account_settings serverless_settings block",
	"ASET-014": "account_settings bucket_settings block",
	"ASET-015": "account_settings 0-change re-plan",
	"ASET-016": "Validator: AUTO+app_settings",
	"ASET-017": "Validator: invalid mode",
	"ASET-018": "account_id immutable (requires replace)",
	"ASET-019": "Partial spec 0-change re-plan",
	"ASET-020": "Settings survive account update",
	// PRI
	"PRI-001": "Single group priority=1",
	"PRI-002": "Multiple groups priorities 1,2,3",
	"PRI-003": "Priority 1 = highest",
	"PRI-004": "Update priorities (reorder)",
	"PRI-005": "Validator: priority=0 rejected",
	"PRI-006": "Priority=1 valid",
	"PRI-007": "Import by composite ID",
	"PRI-008": "Destroy = no-op",
	"PRI-009": "Priority re-plan 0 changes",
	"PRI-010": "Add group to existing priority resource",
	"PRI-011": "Remove group from priority resource",
	"PRI-012": "Group deleted externally → removed from state",
	// RSET
	"RSET-001": "Create resource_settings for EC2",
	"RSET-002": "Create resource_settings for Lambda",
	"RSET-003": "Update resource_settings",
	"RSET-004": "Destroy resource_settings",
	"RSET-005": "Import resource_settings",
	"RSET-006": "resource_settings 0-change re-plan",
	"RSET-007": "P0: resource_settings sedai_sync_enabled no drift",
	"RSET-008": "resource_settings kube_app_settings",
	"RSET-009": "resource_settings partial spec",
	"RSET-010": "resource_settings group_id+resource_type immutable",
	// MP
	"MP-001": "Create CloudWatch use_account_credentials",
	"MP-002": "Create CloudWatch access_key+secret_key",
	"MP-003": "Create CloudWatch role+external_id",
	"MP-004": "Update CloudWatch dimensions, 0-change re-plan",
	"MP-005": "Destroy CloudWatch",
	"MP-006": "Import CloudWatch",
	"MP-007": "CloudWatch read after external delete",
	"MP-008": "Create Datadog",
	"MP-009": "Datadog update api_key",
	"MP-010": "Destroy Datadog",
	"MP-011": "Datadog api_key sensitive",
	"MP-012": "Datadog 0-change re-plan",
	"MP-013": "Create GKE monitoring provider",
	"MP-014": "GKE import",
	"MP-015": "GKE 0-change re-plan",
	"MP-016": "Create New Relic",
	"MP-017": "New Relic import",
	"MP-018": "New Relic 0-change re-plan",
	"MP-019": "FP provider no-auth",
	"MP-020": "FP provider JWT",
	"MP-021": "FP provider client_credentials",
	"MP-022": "FP provider drift test",
	"MP-023": "Azure monitoring provider",
	"MP-024": "VM monitoring provider",
	"MP-025": "BigQuery monitoring provider",
	"MP-026": "CloudWatch with lb_dimensions",
	"MP-027": "CloudWatch with app_dimensions",
	"MP-028": "CloudWatch with instance_dimensions",
	"MP-029": "Multiple MPs per account",
	"MP-030": "P0: EOF recovery on create",
	"MP-031": "Duplicate MP → backend rejects",
	"MP-032": "MP account_id immutable",
	"MP-033": "MP name immutable",
	"MP-034": "MP missing required field",
	"MP-035": "MP invalid integration_type",
	// DS
	"DS-001": "data.sedai_account lookup by name",
	"DS-002": "data.sedai_account not found → error",
	"DS-003": "data.sedai_account multiple matches → error",
	"DS-004": "data.sedai_account all fields present",
	"DS-005": "data.sedai_groups returns list",
	"DS-006": "data.sedai_group lookup by name",
	"DS-007": "data.sedai_group not found → error",
	"DS-008": "data.sedai_groups empty list",
	"DS-009": "data.sedai_group_settings consistency",
	"DS-010": "data.sedai_group_settings all blocks",
	"DS-011": "data.sedai_account_settings all fields",
	"DS-012": "data.sedai_resource_settings",
	"DS-013": "data.sedai_group_priority read",
	"DS-014": "data source after resource update consistency",
	"DS-015": "data source read-only (plan shows no changes)",
	// DEP
	"DEP-001": "Dependency chain account→group→settings→CW",
	"DEP-002": "Dependents created after account ready",
	"DEP-003": "Group settings blocked on group creation",
	"DEP-004": "MP blocked on account creation",
	"DEP-005": "Destroy order: CW→settings→group→account",
	"DEP-006": "Circular dependency prevented",
	"DEP-007": "Data source depends_on pattern",
	"DEP-008": "Cross-resource reference ID propagation",
	"DEP-009": "Partial destroy preserves dependents",
	"DEP-010": "Full Diligent scenario",
	// DRIFT
	"DRIFT-001": "P0: GSET sedai_sync_enabled omitted no drift",
	"DRIFT-002": "P0: ASET sedai_sync_enabled omitted no drift",
	"DRIFT-003": "P0: RSET sedai_sync_enabled omitted no drift",
	"DRIFT-004": "P0: GRP resource counts not in state",
	"DRIFT-005": "Account re-plan 0 changes",
	"DRIFT-006": "Group re-plan 0 changes",
	"DRIFT-007": "Group settings re-plan 0 changes",
	"DRIFT-008": "Account settings re-plan 0 changes",
	"DRIFT-009": "Group priority re-plan 0 changes",
	"DRIFT-010": "CloudWatch MP re-plan 0 changes",
	"DRIFT-011": "Datadog MP re-plan 0 changes",
	"DRIFT-012": "Full stack re-plan 0 changes",
	"DRIFT-013": "Backend update doesn't cause drift",
	"DRIFT-014": "sedai_sync_enabled 3x re-plan stability",
	"DRIFT-015": "settings blocks partial spec no drift",
	"DRIFT-016": "list field order stability",
	"DRIFT-017": "optional block absent no drift",
	"DRIFT-018": "computed field no drift",
	"DRIFT-019": "sensitive field no drift",
	"DRIFT-020": "multi-resource stack 3x re-plan",
	// SCALE
	"SCALE-001": "12 accounts in one apply pass",
	"SCALE-002": "P0: Full 48-resource Diligent stack",
	"SCALE-003": "P0: 48-resource re-plan 0 changes",
	"SCALE-004": "Partial apply → re-apply completes",
	"SCALE-005": "100-resource apply",
	"SCALE-006": "500-resource plan (plan only)",
	"SCALE-007": "Concurrent creates no conflict",
	"SCALE-008": "Concurrent deletes no conflict",
	"SCALE-009": "Large account+groups throughput",
	"SCALE-010": "Scale re-plan 0 changes",
	// ERR
	"ERR-001": "P0: EOF → recovery not found → clear error",
	"ERR-002": "P0: EOF → recovery finds account → use ID",
	"ERR-003": "P0: Group EOF → recovery → use ID",
	"ERR-004": "P0: CloudWatch EOF recovery",
	"ERR-005": "P0: EOF on settings → recovery",
	"ERR-006": "HTTP 401 → clear auth error",
	"ERR-007": "HTTP 403 → clear error",
	"ERR-008": "HTTP 500 → error surfaced",
	"ERR-009": "Network timeout → clear error",
	"ERR-010": "Retry on transient 503",
	"ERR-011": "Invalid JSON response → clear error",
	"ERR-012": "Empty response body → clear error",
	"ERR-013": "Delete not found → success",
	"ERR-014": "Read not found → removed from state",
	"ERR-015": "Apply fails → re-apply succeeds",
	// IMPORT
	"IMPORT-001": "Import sedai_account",
	"IMPORT-002": "Import sedai_group",
	"IMPORT-003": "Import sedai_group_settings",
	"IMPORT-004": "Import sedai_account_settings",
	"IMPORT-005": "Import sedai_group_priority",
	"IMPORT-006": "Import sedai_cloudwatch_monitoring_provider",
	"IMPORT-007": "Import sedai_datadog_monitoring_provider",
	"IMPORT-008": "Import sedai_newrelic_monitoring_provider",
	"IMPORT-009": "Import sedai_fp_monitoring_provider",
	"IMPORT-010": "Import sedai_gke_monitoring_provider",
	"IMPORT-011": "Import non-existent ID → clear error",
	"IMPORT-012": "After import, 0-change plan",
	"IMPORT-013": "Import wrong resource type → error",
	"IMPORT-014": "Full import-then-manage workflow",
	// MIGR
	"MIGR-001": "v1→v2 provider upgrade state compat",
	"MIGR-002": "v2 apply over v1 state",
	"MIGR-003": "v1 resource counts removed in v2",
	"MIGR-004": "CDKTF v0.21.0 compatibility",
	"MIGR-005": "State migration script",
	"MIGR-006": "Post-migration 0-change plan",
	"MIGR-007": "Rollback to v1 compatibility",
	"MIGR-008": "New resources in v2",
}

var categoryOrder = []string{
	"PROV", "ACCT", "GRP", "GSET", "ASET", "PRI", "RSET",
	"MP", "DS", "DEP", "DRIFT", "SCALE", "ERR", "IMPORT", "MIGR",
}

func extractManifestID(testName string) string {
	// Test names are like "TestAccGSET/GSET-006" or "TestAccERR/ERR-001"
	// Extract the manifest ID from the subtest portion
	parts := strings.Split(testName, "/")
	for _, part := range parts {
		for _, cat := range categoryOrder {
			if strings.HasPrefix(part, cat+"-") {
				return part
			}
		}
	}
	return ""
}

func categoryOf(id string) string {
	parts := strings.SplitN(id, "-", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return "OTHER"
}

func main() {
	input := os.Stdin
	if len(os.Args) > 1 {
		f, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening %s: %v\n", os.Args[1], err)
			os.Exit(1)
		}
		defer f.Close()
		input = f
	}

	results := map[string]*testResult{}
	output := map[string][]string{}

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Bytes()
		var ev testEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}
		if ev.Test == "" {
			continue
		}

		id := extractManifestID(ev.Test)
		if id == "" {
			continue
		}

		switch ev.Action {
		case "pass":
			r := getOrCreate(results, id)
			r.Status = "PASS"
			r.Duration = ev.Elapsed
		case "fail":
			r := getOrCreate(results, id)
			r.Status = "FAIL"
			r.Duration = ev.Elapsed
		case "skip":
			r := getOrCreate(results, id)
			r.Status = "SKIP"
		case "output":
			output[id] = append(output[id], ev.Output)
		}
	}

	// Merge output into results
	for id, lines := range output {
		if r, ok := results[id]; ok {
			r.Output = lines
		}
	}

	// Pre-populate all known manifest IDs as NOT_RUN
	for id, name := range manifestNames {
		if _, ok := results[id]; !ok {
			results[id] = &testResult{ID: id, Name: name, Status: "NOT_RUN"}
		} else if results[id].Name == "" {
			results[id].Name = name
		}
	}

	// Count stats
	var total, pass, fail, skip, notRun int
	for _, r := range results {
		total++
		switch r.Status {
		case "PASS":
			pass++
		case "FAIL":
			fail++
		case "SKIP":
			skip++
		case "NOT_RUN":
			notRun++
		}
	}

	runAt := time.Now().Format("2006-01-02 15:04:05 MST")

	fmt.Print(renderHTML(results, runAt, total, pass, fail, skip, notRun))
}

func getOrCreate(m map[string]*testResult, id string) *testResult {
	if r, ok := m[id]; ok {
		return r
	}
	name := manifestNames[id]
	if name == "" {
		name = id
	}
	r := &testResult{ID: id, Name: name, Status: "NOT_RUN"}
	m[id] = r
	return r
}

func renderHTML(results map[string]*testResult, runAt string, total, pass, fail, skip, notRun int) string {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Sedai TF Provider — Test Results</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: 'Courier New', Courier, monospace; background: #0d1117; color: #e6edf3; padding: 24px; }
h1 { color: #58a6ff; font-size: 1.4em; margin-bottom: 4px; }
.meta { color: #8b949e; font-size: 0.85em; margin-bottom: 20px; }
.summary { display: flex; gap: 16px; margin-bottom: 28px; flex-wrap: wrap; }
.badge { padding: 6px 14px; border-radius: 6px; font-size: 0.9em; font-weight: bold; }
.badge-total { background: #21262d; color: #e6edf3; }
.badge-pass  { background: #1a4731; color: #3fb950; }
.badge-fail  { background: #5e1c1c; color: #f85149; }
.badge-skip  { background: #2d2a1a; color: #d29922; }
.badge-notrun{ background: #21262d; color: #8b949e; }
.category { margin-bottom: 32px; }
.category h2 { color: #8b949e; font-size: 0.9em; text-transform: uppercase; letter-spacing: 2px;
               padding: 6px 0; border-bottom: 1px solid #21262d; margin-bottom: 8px; }
table { width: 100%; border-collapse: collapse; }
th { text-align: left; color: #8b949e; font-size: 0.8em; padding: 6px 8px;
     border-bottom: 1px solid #21262d; text-transform: uppercase; letter-spacing: 1px; }
td { padding: 6px 8px; vertical-align: top; border-bottom: 1px solid #161b22; font-size: 0.85em; }
.id-cell { background: #1a2332; color: #79c0ff; font-weight: bold; white-space: nowrap; width: 110px; }
.name-cell { color: #e6edf3; }
.dur-cell { color: #8b949e; white-space: nowrap; width: 80px; }
.status-pass  { color: #3fb950; font-weight: bold; }
.status-fail  { color: #f85149; font-weight: bold; }
.status-skip  { color: #d29922; }
.status-notrun{ color: #484f58; }
details { margin-top: 6px; }
summary { cursor: pointer; color: #f85149; font-size: 0.8em; }
.output-block { background: #161b22; border: 1px solid #30363d; border-radius: 4px;
                padding: 8px; margin-top: 4px; font-size: 0.78em; color: #e6edf3;
                white-space: pre-wrap; max-height: 300px; overflow-y: auto; }
</style>
</head>
<body>
`)

	fmt.Fprintf(&sb, "<h1>Sedai TF Provider &mdash; Test Results</h1>\n")
	fmt.Fprintf(&sb, `<div class="meta">Run at %s</div>`+"\n", html.EscapeString(runAt))

	fmt.Fprintf(&sb, `<div class="summary">`)
	fmt.Fprintf(&sb, `<span class="badge badge-total">Total: %d</span>`, total)
	fmt.Fprintf(&sb, `<span class="badge badge-pass">Pass: %d</span>`, pass)
	fmt.Fprintf(&sb, `<span class="badge badge-fail">Fail: %d</span>`, fail)
	fmt.Fprintf(&sb, `<span class="badge badge-skip">Skip: %d</span>`, skip)
	fmt.Fprintf(&sb, `<span class="badge badge-notrun">Not Run: %d</span>`, notRun)
	fmt.Fprintf(&sb, "</div>\n")

	for _, cat := range categoryOrder {
		// Collect IDs for this category, sorted
		var catIDs []string
		for id := range results {
			if categoryOf(id) == cat {
				catIDs = append(catIDs, id)
			}
		}
		if len(catIDs) == 0 {
			continue
		}
		sort.Strings(catIDs)

		fmt.Fprintf(&sb, `<div class="category"><h2>%s</h2><table>`, cat)
		fmt.Fprintf(&sb, "<tr><th>ID</th><th>Test Name</th><th>Status</th><th>Duration</th></tr>\n")

		for _, id := range catIDs {
			r := results[id]
			statusClass := map[string]string{
				"PASS":    "status-pass",
				"FAIL":    "status-fail",
				"SKIP":    "status-skip",
				"NOT_RUN": "status-notrun",
			}[r.Status]

			dur := ""
			if r.Duration > 0 {
				dur = fmt.Sprintf("%.2fs", r.Duration)
			}

			fmt.Fprintf(&sb, "<tr>")
			fmt.Fprintf(&sb, `<td class="id-cell">%s</td>`, html.EscapeString(id))
			fmt.Fprintf(&sb, `<td class="name-cell">%s`, html.EscapeString(r.Name))

			if r.Status == "FAIL" && len(r.Output) > 0 {
				fmt.Fprintf(&sb, `<details><summary>Show output</summary><div class="output-block">`)
				for _, line := range r.Output {
					fmt.Fprintf(&sb, "%s", html.EscapeString(line))
				}
				fmt.Fprintf(&sb, "</div></details>")
			}

			fmt.Fprintf(&sb, "</td>")
			fmt.Fprintf(&sb, `<td class="%s">%s</td>`, statusClass, r.Status)
			fmt.Fprintf(&sb, `<td class="dur-cell">%s</td>`, dur)
			fmt.Fprintf(&sb, "</tr>\n")
		}

		fmt.Fprintf(&sb, "</table></div>\n")
	}

	sb.WriteString("</body></html>\n")
	return sb.String()
}
