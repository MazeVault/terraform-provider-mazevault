package acc

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ─── mazevault_certificate_rotation_config ────────────────────────────────────

// TestAccCertificateRotationConfigResource_basic covers create → read → update → destroy
// for a certificate rotation config.
//
// Requires MAZEVAULT_TEST_CERT_ID to point to an existing certificate in the
// connected MazeVault instance.  If the variable is absent the test is skipped.
func TestAccCertificateRotationConfigResource_basic(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)
	certID := os.Getenv("MAZEVAULT_TEST_CERT_ID")
	if certID == "" {
		t.Skip("MAZEVAULT_TEST_CERT_ID not set; skipping certificate rotation config acceptance test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with minimal settings.
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "mazevault_certificate_rotation_config" "test" {
  certificate_id    = %q
  enabled           = true
  renewal_lead_days = 21
}
`, certID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_certificate_rotation_config.test", "id"),
					resource.TestCheckResourceAttr("mazevault_certificate_rotation_config.test", "certificate_id", certID),
					resource.TestCheckResourceAttr("mazevault_certificate_rotation_config.test", "enabled", "true"),
					resource.TestCheckResourceAttr("mazevault_certificate_rotation_config.test", "renewal_lead_days", "21"),
				),
			},
			// Update renewal lead days and add notification email.
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "mazevault_certificate_rotation_config" "test" {
  certificate_id      = %q
  enabled             = true
  renewal_lead_days   = 14
  max_retry_attempts  = 5
  notification_emails = ["ops@example.com"]
}
`, certID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mazevault_certificate_rotation_config.test", "renewal_lead_days", "14"),
					resource.TestCheckResourceAttr("mazevault_certificate_rotation_config.test", "max_retry_attempts", "5"),
					resource.TestCheckResourceAttr("mazevault_certificate_rotation_config.test", "notification_emails.#", "1"),
					resource.TestCheckResourceAttr("mazevault_certificate_rotation_config.test", "notification_emails.0", "ops@example.com"),
				),
			},
			// Disable rotation (simulates destroy then re-enable).
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "mazevault_certificate_rotation_config" "test" {
  certificate_id    = %q
  enabled           = false
  renewal_lead_days = 14
}
`, certID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mazevault_certificate_rotation_config.test", "enabled", "false"),
				),
			},
		},
	})
}

// ─── mazevault_entra_rotation_config ─────────────────────────────────────────

// TestAccEntraRotationConfigResource_basic covers create → read → update → destroy
// for an Entra credential rotation config.
//
// Requires MAZEVAULT_TEST_ENTRA_CREDENTIAL_ID to point to an existing Entra
// credential in the connected MazeVault instance.  If the variable is absent
// the test is skipped.
func TestAccEntraRotationConfigResource_basic(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)
	credID := os.Getenv("MAZEVAULT_TEST_ENTRA_CREDENTIAL_ID")
	if credID == "" {
		t.Skip("MAZEVAULT_TEST_ENTRA_CREDENTIAL_ID not set; skipping entra rotation config acceptance test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with minimal settings.
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "mazevault_entra_rotation_config" "test" {
  credential_id                = %q
  rotation_enabled             = true
  rotation_days_before_expiry  = 30
  grace_period_days            = 7
}
`, credID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_entra_rotation_config.test", "id"),
					resource.TestCheckResourceAttr("mazevault_entra_rotation_config.test", "credential_id", credID),
					resource.TestCheckResourceAttr("mazevault_entra_rotation_config.test", "rotation_enabled", "true"),
					resource.TestCheckResourceAttr("mazevault_entra_rotation_config.test", "rotation_days_before_expiry", "30"),
					resource.TestCheckResourceAttr("mazevault_entra_rotation_config.test", "grace_period_days", "7"),
				),
			},
			// Update — shorten lead time and add staged rotation.
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "mazevault_entra_rotation_config" "test" {
  credential_id                = %q
  rotation_enabled             = true
  rotation_days_before_expiry  = 21
  grace_period_days            = 3
  staged_rotation_enabled      = true
  soak_window_hours            = 24
}
`, credID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mazevault_entra_rotation_config.test", "rotation_days_before_expiry", "21"),
					resource.TestCheckResourceAttr("mazevault_entra_rotation_config.test", "grace_period_days", "3"),
					resource.TestCheckResourceAttr("mazevault_entra_rotation_config.test", "staged_rotation_enabled", "true"),
					resource.TestCheckResourceAttr("mazevault_entra_rotation_config.test", "soak_window_hours", "24"),
				),
			},
			// Disable rotation (simulates destroy).
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "mazevault_entra_rotation_config" "test" {
  credential_id    = %q
  rotation_enabled = false
}
`, credID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mazevault_entra_rotation_config.test", "rotation_enabled", "false"),
				),
			},
		},
	})
}

// ─── mazevault_rotation_resources datasource ─────────────────────────────────

// TestAccRotationResourcesDataSource_read verifies the data source returns a
// list attribute (may be empty on a fresh environment).
func TestAccRotationResourcesDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Unfiltered — must return a list (even empty).
			{
				Config: providerConfig() + `
data "mazevault_rotation_resources" "all" {}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mazevault_rotation_resources.all", "resources.#"),
				),
			},
			// Filter by kind=secret — must return a list attribute.
			{
				Config: providerConfig() + `
data "mazevault_rotation_resources" "secrets" {
  kind = "secret"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mazevault_rotation_resources.secrets", "resources.#"),
				),
			},
		},
	})
}

// ─── mazevault_rotation_resource_history datasource ──────────────────────────

// TestAccRotationResourceHistoryDataSource_read verifies the data source returns
// a list for a known resource.
//
// Requires MAZEVAULT_TEST_ROTATION_RESOURCE_KIND and
// MAZEVAULT_TEST_ROTATION_RESOURCE_ID to point to a resource that has
// rotation history.  If either variable is absent the test is skipped.
func TestAccRotationResourceHistoryDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)
	kind := os.Getenv("MAZEVAULT_TEST_ROTATION_RESOURCE_KIND")
	resID := os.Getenv("MAZEVAULT_TEST_ROTATION_RESOURCE_ID")
	if kind == "" || resID == "" {
		t.Skip("MAZEVAULT_TEST_ROTATION_RESOURCE_KIND / MAZEVAULT_TEST_ROTATION_RESOURCE_ID not set; skipping")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
data "mazevault_rotation_resource_history" "hist" {
  kind        = %q
  resource_id = %q
}
`, kind, resID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mazevault_rotation_resource_history.hist", "executions.#"),
				),
			},
		},
	})
}

// ─── mazevault_project_rotation_configs datasource ───────────────────────────

// TestAccProjectRotationConfigsDataSource_read verifies the data source returns
// a list for a project (may be empty on a fresh environment).
func TestAccProjectRotationConfigsDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "prc_org" {
  name = "tf-acc-prc-org"
}
resource "mazevault_project" "prc_proj" {
  organization_id = mazevault_organization.prc_org.id
  name            = "tf-acc-prc-proj"
  environment     = "dev"
}
resource "mazevault_secret" "prc_sec" {
  project_id  = mazevault_project.prc_proj.id
  key         = "PRC_SECRET"
  value       = "test-value"
  environment = "dev"
}
resource "mazevault_rotation_config" "prc_cfg" {
  secret_id              = mazevault_secret.prc_sec.id
  rotation_interval_days = 30
  enabled                = true
}
data "mazevault_project_rotation_configs" "prc" {
  project_id = mazevault_project.prc_proj.id
  depends_on = [mazevault_rotation_config.prc_cfg]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mazevault_project_rotation_configs.prc", "configs.#"),
					// After creating the rotation_config above there must be at least one.
					resource.TestCheckResourceAttrSet("data.mazevault_project_rotation_configs.prc", "configs.0.id"),
					resource.TestCheckResourceAttrSet("data.mazevault_project_rotation_configs.prc", "configs.0.secret_id"),
					resource.TestCheckResourceAttr("data.mazevault_project_rotation_configs.prc", "configs.0.enabled", "true"),
					resource.TestCheckResourceAttr("data.mazevault_project_rotation_configs.prc", "configs.0.rotation_interval_days", "30"),
				),
			},
		},
	})
}
