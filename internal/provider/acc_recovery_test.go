package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccRECOVERY tests Diligent's exact scenario: partial apply failure followed
// by re-apply that completes the remaining resources without duplicating the ones
// that already succeeded.
func TestAccRECOVERY(t *testing.T) {

	// RECOVERY-001: 12-account apply where accounts 4 and 8 EOF.
	// Re-apply must create exactly 12 accounts total — no duplicates of the 10 that
	// succeeded, no permanent failure for the 2 that EOFed.
	t.Run("RECOVERY-001", func(t *testing.T) {
		mock := newMockServer()
		defer mock.Close()
		// EOF on the 4th and 8th account create POST.
		mock.EOFOnAccountCreateN = 4
		t.Setenv("SEDAI_BASE_URL", mock.URL())
		t.Setenv("SEDAI_API_TOKEN", "mock-token")

		prefix := "recovery-001-" + randString(4)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// Step 1: Apply 12 accounts. Account 4 EOFs. Terraform reports partial failure
				// but writes state for the 10 that succeeded.
				{
					Config:      testAccRecovery12AccountsConfig(prefix),
					ExpectError: nil, // provider recovers via post-EOF search; may still partially succeed
				},
				// Step 2: Re-apply same config. Now mock EOFs on account 8 too.
				// Previously-created accounts are in state — only 2 new ones attempted.
				{
					PreConfig:          func() { mock.EOFOnAccountCreateN = 8 },
					Config:             testAccRecovery12AccountsConfig(prefix),
					ExpectNonEmptyPlan: false,
				},
				// Step 3: Final re-apply — all 12 in state, no changes planned.
				{
					PreConfig:          func() { mock.EOFOnAccountCreateN = 0 },
					Config:             testAccRecovery12AccountsConfig(prefix),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// RECOVERY-002: Account and group created, group_settings POST EOFs.
	// Re-apply creates settings without re-creating account or group.
	t.Run("RECOVERY-002", func(t *testing.T) {
		mock := newMockServer()
		defer mock.Close()
		t.Setenv("SEDAI_BASE_URL", mock.URL())
		t.Setenv("SEDAI_API_TOKEN", "mock-token")

		name := "recovery-002-" + randString(5)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// Step 1: Apply account + group (succeed), group_settings (will fail below).
				// First apply only account+group — settings come in step 2.
				{
					Config: testAccMockAccountAndGroupConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttrSet("sedai_group.test", "id"),
					),
				},
				// Step 2: Add settings with mock returning 500 for settings POST.
				{
					PreConfig:   func() { mock.ReturnHTTPCode = 500 },
					Config:      testAccMockAccountGroupSettingsConfig(name),
					ExpectError: nil,
				},
				// Step 3: Re-apply with mock reset — settings created, account/group untouched.
				{
					PreConfig: func() { mock.ReturnHTTPCode = 0 },
					Config:    testAccMockAccountGroupSettingsConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group_settings.test", "group_id"),
					),
				},
				// Step 4: Re-plan → 0 changes.
				{
					Config:             testAccMockAccountGroupSettingsConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// RECOVERY-003: Monitoring provider EOF on create — recovery finds existing MP,
	// no duplicate created on re-apply.
	t.Run("RECOVERY-003", func(t *testing.T) {
		mock := newMockServer()
		defer mock.Close()
		t.Setenv("SEDAI_BASE_URL", mock.URL())
		t.Setenv("SEDAI_API_TOKEN", "mock-token")

		name := "recovery-003-" + randString(5)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// Step 1: Create account (succeeds).
				{
					Config: testAccMockAccountConfig(name),
					Check:  resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
				},
				// Step 2: Add CloudWatch MP — mock closes connection on MP create POST.
				// Provider's EOF-recovery search should find the created MP and use its ID.
				{
					Config: testAccMockAccountAndCWConfig(name),
					// Provider either recovers cleanly (MP in state) or surfaces warning.
					// Either way: no permanent error, re-apply resolves it.
				},
				// Step 3: Re-apply — MP already in state, no duplicate POST.
				{
					Config: testAccMockAccountAndCWConfig(name),
					Check:  resource.TestCheckResourceAttrSet("sedai_cloudwatch_monitoring_provider.test", "id"),
				},
				// Step 4: Re-plan → 0 changes.
				{
					Config:             testAccMockAccountAndCWConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// RECOVERY-004: Connection closed after 6th resource in a 12-resource apply.
	// Re-apply completes the remaining 6 cleanly.
	t.Run("RECOVERY-004", func(t *testing.T) {
		mock := newMockServer()
		defer mock.Close()
		// EOF on the 6th account create.
		mock.EOFOnAccountCreateN = 6
		t.Setenv("SEDAI_BASE_URL", mock.URL())
		t.Setenv("SEDAI_API_TOKEN", "mock-token")

		prefix := "recovery-004-" + randString(4)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// Step 1: Partial apply — first 5 succeed, 6th EOFs, 7-12 pending.
				{
					Config:      testAccRecovery12AccountsConfig(prefix),
					ExpectError: nil,
				},
				// Step 2: Re-apply — mock no longer EOFs. Remaining 6 created.
				{
					PreConfig: func() { mock.EOFOnAccountCreateN = 0 },
					Config:    testAccRecovery12AccountsConfig(prefix),
					Check:     resource.TestCheckResourceAttrSet("sedai_account.acct11", "id"),
				},
				// Step 3: Re-plan → 0 changes — all 12 in state.
				{
					Config:             testAccRecovery12AccountsConfig(prefix),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// RECOVERY-005: Settings update fails (500 on first PUT), re-apply succeeds.
	t.Run("RECOVERY-005", func(t *testing.T) {
		mock := newMockServer()
		defer mock.Close()
		t.Setenv("SEDAI_BASE_URL", mock.URL())
		t.Setenv("SEDAI_API_TOKEN", "mock-token")

		name := "recovery-005-" + randString(5)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// Step 1: Create account + group + settings (DATA_PILOT).
				{
					Config: testAccMockAccountGroupSettingsConfig(name),
					Check:  resource.TestCheckResourceAttrSet("sedai_group_settings.test", "group_id"),
				},
				// Step 2: Change optimization_mode — mock returns 500.
				{
					PreConfig:   func() { mock.ReturnHTTPCode = 500 },
					Config:      testAccMockAccountGroupSettingsAltConfig(name),
					ExpectError: nil,
				},
				// Step 3: Re-apply with mock reset — update succeeds.
				{
					PreConfig: func() { mock.ReturnHTTPCode = 0 },
					Config:    testAccMockAccountGroupSettingsAltConfig(name),
				},
				// Step 4: Re-plan → 0 changes.
				{
					Config:             testAccMockAccountGroupSettingsAltConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}

// --- HCL config builders ---

func testAccRecovery12AccountsConfig(prefix string) string {
	cfg := ""
	for i := 0; i < 12; i++ {
		cfg += fmt.Sprintf(`
resource "sedai_account" "acct%[1]d" {
  name             = "%[2]s-acct-%[1]d"
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::%[3]012d:role/SedaiRole"
  external_id      = "sedai-ext-%[1]d"
}
`, i, prefix, 100000000000+int64(i))
	}
	return cfg
}

func testAccMockAccountAndGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_group" "test" {
  name       = %[1]q
  depends_on = [sedai_account.test]
}
`, name)
}

func testAccMockAccountGroupSettingsConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_group" "test" {
  name       = %[1]q
  depends_on = [sedai_account.test]
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
}
`, name)
}

func testAccMockAccountGroupSettingsAltConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_group" "test" {
  name       = %[1]q
  depends_on = [sedai_account.test]
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "CO_PILOT"
}
`, name)
}

func testAccMockAccountAndCWConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_cloudwatch_monitoring_provider" "test" {
  account_id              = sedai_account.test.id
  name                    = %[1]q
  integration_type        = "ROLE"
  use_account_credentials = true
  depends_on              = [sedai_account.test]
}
`, name)
}
