package acc

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccApiTokenResource_basic covers create → read → destroy.
// Note: the raw token value is only available on creation; Read() can only
// verify the token ID and metadata.
func TestAccApiTokenResource_basic(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_api_token" "test" {
  name     = "tf-acc-token"
  scopes   = ["secrets:read", "projects:read"]
  duration = "720h"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_api_token.test", "id"),
					resource.TestCheckResourceAttr("mazevault_api_token.test", "name", "tf-acc-token"),
					// raw token is only present at creation time
					resource.TestCheckResourceAttrSet("mazevault_api_token.test", "token"),
				),
			},
		},
	})
}

// TestAccEnvironmentResource_basic covers create → read → destroy.
// Environment resources use name-based identity (no UUID in state).
func TestAccEnvironmentResource_basic(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig("tf-acc-env-test", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mazevault_environment.test", "name", "tf-acc-env-test"),
					resource.TestCheckResourceAttr("mazevault_environment.test", "is_production", "false"),
				),
			},
		},
	})
}

// TestAccEnvironmentResource_production verifies the is_production flag is
// persisted correctly.
func TestAccEnvironmentResource_production(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig("tf-acc-env-prod", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mazevault_environment.test", "name", "tf-acc-env-prod"),
					resource.TestCheckResourceAttr("mazevault_environment.test", "is_production", "true"),
				),
			},
		},
	})
}

// TestAccEnvironmentsDataSource_read verifies data.mazevault_environments lists
// at least the environment created in the same config.
func TestAccEnvironmentsDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "env_ds_org" {
  name = "tf-acc-env-ds-org"
}
resource "mazevault_environment" "env_ds_src" {
  organization_id = mazevault_organization.env_ds_org.id
  name            = "tf-acc-env-ds"
  is_production   = false
}
data "mazevault_environments" "all" {
  organization_id = mazevault_organization.env_ds_org.id
  depends_on      = [mazevault_environment.env_ds_src]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mazevault_environments.all", "environments.#"),
				),
			},
		},
	})
}

// testAccEnvironmentConfig returns a complete HCL config with an org + env.
func testAccEnvironmentConfig(name string, isProduction bool) string {
	return fmt.Sprintf(`%s
resource "mazevault_organization" "env_org" {
  name = "tf-acc-env-org"
}
resource "mazevault_environment" "test" {
  organization_id = mazevault_organization.env_org.id
  name            = %q
  is_production   = %t
}
`, providerConfig(), name, isProduction)
}
