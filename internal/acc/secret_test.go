package acc

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccSecretResource_basic covers create → read → update value → destroy.
func TestAccSecretResource_basic(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccSecretConfig("DB_PASSWORD", "initial-value", "staging"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_secret.test", "id"),
					resource.TestCheckResourceAttr("mazevault_secret.test", "key", "DB_PASSWORD"),
					resource.TestCheckResourceAttr("mazevault_secret.test", "value", "initial-value"),
					resource.TestCheckResourceAttr("mazevault_secret.test", "environment", "staging"),
					resource.TestCheckResourceAttrSet("mazevault_secret.test", "version"),
					resource.TestCheckResourceAttrSet("mazevault_secret.test", "status"),
					resource.TestCheckResourceAttrSet("mazevault_secret.test", "created_at"),
				),
			},
			// Update value
			{
				Config: testAccSecretConfig("DB_PASSWORD", "updated-value", "staging"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mazevault_secret.test", "value", "updated-value"),
				),
			},
		},
	})
}

// TestAccSecretResource_withMetadata covers the optional metadata map.
func TestAccSecretResource_withMetadata(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "sec_meta_org" {
  name = "tf-acc-sec-meta-org"
}
resource "mazevault_project" "sec_meta_proj" {
  organization_id = mazevault_organization.sec_meta_org.id
  name            = "tf-acc-sec-meta-proj"
  environment     = "dev"
}
resource "mazevault_secret" "test" {
  project_id  = mazevault_project.sec_meta_proj.id
  key         = "API_KEY"
  value       = "secret123"
  environment = "dev"
  metadata = {
    team    = "platform"
    service = "payments"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_secret.test", "id"),
					resource.TestCheckResourceAttr("mazevault_secret.test", "metadata.team", "platform"),
					resource.TestCheckResourceAttr("mazevault_secret.test", "metadata.service", "payments"),
				),
			},
		},
	})
}

// TestAccSecretResource_withTTL verifies that ttl_hours is persisted correctly.
func TestAccSecretResource_withTTL(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "sec_ttl_org" {
  name = "tf-acc-sec-ttl-org"
}
resource "mazevault_project" "sec_ttl_proj" {
  organization_id = mazevault_organization.sec_ttl_org.id
  name            = "tf-acc-sec-ttl-proj"
  environment     = "dev"
}
resource "mazevault_secret" "test" {
  project_id  = mazevault_project.sec_ttl_proj.id
  key         = "TEMP_TOKEN"
  value       = "ephemeral"
  environment = "dev"
  ttl_hours   = 72
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_secret.test", "id"),
					resource.TestCheckResourceAttr("mazevault_secret.test", "ttl_hours", "72"),
				),
			},
		},
	})
}

// TestAccSecretDataSource_read verifies data.mazevault_secret returns the
// correct secret attributes.
func TestAccSecretDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "sec_ds_org" {
  name = "tf-acc-sec-ds-org"
}
resource "mazevault_project" "sec_ds_proj" {
  organization_id = mazevault_organization.sec_ds_org.id
  name            = "tf-acc-sec-ds-proj"
  environment     = "dev"
}
resource "mazevault_secret" "src" {
  project_id  = mazevault_project.sec_ds_proj.id
  key         = "DATASOURCE_KEY"
  value       = "datasource-value"
  environment = "dev"
}
data "mazevault_secret" "ds" {
  id = mazevault_secret.src.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.mazevault_secret.ds", "id",
						"mazevault_secret.src", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.mazevault_secret.ds", "key",
						"mazevault_secret.src", "key",
					),
				),
			},
		},
	})
}

// testAccSecretConfig returns a complete HCL config with org + project + secret.
func testAccSecretConfig(key, value, environment string) string {
	return fmt.Sprintf(`%s
resource "mazevault_organization" "sec_org" {
  name = "tf-acc-sec-org"
}
resource "mazevault_project" "sec_proj" {
  organization_id = mazevault_organization.sec_org.id
  name            = "tf-acc-sec-proj"
  environment     = %q
}
resource "mazevault_secret" "test" {
  project_id  = mazevault_project.sec_proj.id
  key         = %q
  value       = %q
  environment = %q
}
`, providerConfig(), environment, key, value, environment)
}
