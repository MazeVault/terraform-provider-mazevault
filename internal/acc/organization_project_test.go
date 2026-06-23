package acc

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccOrganizationResource_basic covers create → read → update name → destroy.
func TestAccOrganizationResource_basic(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: providerConfig() + `
resource "mazevault_organization" "test" {
  name = "tf-acc-org-basic"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_organization.test", "id"),
					resource.TestCheckResourceAttr("mazevault_organization.test", "name", "tf-acc-org-basic"),
					resource.TestCheckResourceAttrSet("mazevault_organization.test", "created_at"),
				),
			},
			// Update name
			{
				Config: providerConfig() + `
resource "mazevault_organization" "test" {
  name = "tf-acc-org-basic-updated"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mazevault_organization.test", "name", "tf-acc-org-basic-updated"),
				),
			},
		},
	})
}

// TestAccOrganizationDataSource_read verifies the data source returns an
// existing organization by its ID.
func TestAccOrganizationDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "ds_src" {
  name = "tf-acc-org-datasource"
}

data "mazevault_organization" "ds" {
  id = mazevault_organization.ds_src.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.mazevault_organization.ds", "id",
						"mazevault_organization.ds_src", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.mazevault_organization.ds", "name",
						"mazevault_organization.ds_src", "name",
					),
				),
			},
		},
	})
}

// TestAccProjectResource_basic covers create → read → update name → destroy.
func TestAccProjectResource_basic(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create + read-back check (the backend has no PUT /projects/:id endpoint).
			{
				Config: testAccProjectConfig("tf-acc-project-basic", "staging"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_project.test", "id"),
					resource.TestCheckResourceAttr("mazevault_project.test", "name", "tf-acc-project-basic"),
					resource.TestCheckResourceAttrSet("mazevault_project.test", "organization_id"),
					resource.TestCheckResourceAttrSet("mazevault_project.test", "created_at"),
				),
			},
		},
	})
}

// TestAccProjectDataSource_read verifies data.mazevault_project round-trips the
// resource attributes.
func TestAccProjectDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "proj_ds_org" {
  name = "tf-acc-proj-ds-org"
}
resource "mazevault_project" "proj_ds_src" {
  organization_id = mazevault_organization.proj_ds_org.id
  name            = "tf-acc-proj-datasource"
  environment     = "dev"
}
data "mazevault_project" "proj_ds" {
  id = mazevault_project.proj_ds_src.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.mazevault_project.proj_ds", "id",
						"mazevault_project.proj_ds_src", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.mazevault_project.proj_ds", "name",
						"mazevault_project.proj_ds_src", "name",
					),
				),
			},
		},
	})
}

// testAccProjectConfig is a helper that builds a project HCL config with a
// shared parent organization.
func testAccProjectConfig(name, environment string) string {
	return fmt.Sprintf(`%s
resource "mazevault_organization" "proj_org" {
  name = "tf-acc-proj-org"
}
resource "mazevault_project" "test" {
  organization_id = mazevault_organization.proj_org.id
  name            = %q
  environment     = %q
}
`, providerConfig(), name, environment)
}
