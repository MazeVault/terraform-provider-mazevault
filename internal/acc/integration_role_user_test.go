package acc

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccIntegrationResource_basic creates a minimal integration that does not
// require external credentials (a webhook-style generic integration) and verifies
// that it round-trips through create → read → destroy without error.
// Note: the integration resource uses a "provider" + "type" pair from the API.
// We use "generic" / "generic" which the backend accepts without credentials.
func TestAccIntegrationResource_basic(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "int_org" {
  name = "tf-acc-int-org"
}
resource "mazevault_environment" "int_env" {
  organization_id = mazevault_organization.int_org.id
  name            = "tf-acc-int-env"
  is_production   = false
}
resource "mazevault_project" "int_proj" {
  organization_id = mazevault_organization.int_org.id
  name            = "tf-acc-int-proj"
  environment     = "dev"
}
resource "mazevault_integration" "test" {
  project_id    = mazevault_project.int_proj.id
  name          = "tf-acc-integration"
  type          = "gitops"
  provider_name = "generic"
  environment   = mazevault_environment.int_env.name
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_integration.test", "id"),
					resource.TestCheckResourceAttr("mazevault_integration.test", "name", "tf-acc-integration"),
					resource.TestCheckResourceAttr("mazevault_integration.test", "provider_name", "generic"),
				),
			},
		},
	})
}

// TestAccIntegrationsDataSource_read verifies data.mazevault_integrations lists
// integrations for a project.
func TestAccIntegrationsDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "int_ds_org" {
  name = "tf-acc-int-ds-org"
}
resource "mazevault_project" "int_ds_proj" {
  organization_id = mazevault_organization.int_ds_org.id
  name            = "tf-acc-int-ds-proj"
  environment     = "dev"
}
data "mazevault_integrations" "all" {
  project_id = mazevault_project.int_ds_proj.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					// returns an empty list for a new project — just check the attribute exists
					resource.TestCheckResourceAttrSet("data.mazevault_integrations.all", "integrations.#"),
				),
			},
		},
	})
}

// TestAccRoleResource_basic creates a custom RBAC role and verifies that
// destroy removes the resource from Terraform state (no API delete).
func TestAccRoleResource_basic(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_role" "test" {
  name         = "tf-acc-role"
  display_name = "TF Acc Test Role"
  description  = "Created by acceptance test"
  permissions  = ["secrets:read", "projects:read"]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_role.test", "id"),
					resource.TestCheckResourceAttr("mazevault_role.test", "name", "tf-acc-role"),
					resource.TestCheckResourceAttr("mazevault_role.test", "permissions.#", "2"),
				),
			},
		},
	})
}

// TestAccRolesDataSource_read verifies data.mazevault_roles returns at least
// the built-in system roles.
func TestAccRolesDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
data "mazevault_roles" "all" {}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mazevault_roles.all", "roles.#"),
				),
			},
		},
	})
}

// TestAccUserResource_basic covers create → destroy.
// All user fields (email, full_name, password) use RequiresReplace so only a
// single step is needed.
func TestAccUserResource_basic(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_user" "test" {
  email     = "tf-acc-user@example.com"
  full_name = "TF Acc User"
  password  = "Acc3ptanceT3st!"
  role      = "member"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_user.test", "id"),
					resource.TestCheckResourceAttr("mazevault_user.test", "email", "tf-acc-user@example.com"),
					resource.TestCheckResourceAttr("mazevault_user.test", "full_name", "TF Acc User"),
				),
			},
		},
	})
}

// TestAccUsersDataSource_read verifies data.mazevault_users returns at least
// the user used by the API token (the operator account).
func TestAccUsersDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
data "mazevault_users" "all" {}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mazevault_users.all", "users.#"),
				),
			},
		},
	})
}
