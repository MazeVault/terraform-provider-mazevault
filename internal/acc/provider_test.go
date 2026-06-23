package acc

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccProvider_configure verifies that the provider block is accepted by
// the Terraform plugin framework when valid credentials are supplied.
func TestAccProvider_configure(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a minimal organization to exercise the full provider
				// Configure() path without needing a pre-existing resource ID.
				Config: providerConfig() + `
resource "mazevault_organization" "smoke" {
  name = "tf-acc-smoke-provider"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_organization.smoke", "id"),
					resource.TestCheckResourceAttr("mazevault_organization.smoke", "name", "tf-acc-smoke-provider"),
				),
			},
		},
	})
}

/*








































}	})		},			},				),					resource.TestCheckResourceAttrSet("mazevault_integration.ado", "id"),				Check: resource.ComposeTestCheckFunc(				`,				  }				    azure_variable_group_id = env("ADO_GROUP_ID")				    azure_mode    = "variable_group"				    azure_pat     = env("ADO_PAT")				    azure_project = env("ADO_PROJECT")				    azure_org     = env("ADO_ORG")				    provider   = "azure_devops"				    type       = "gitops"				    name       = "tf-ado-test"				    project_id = "${env("MAZEVAULT_PROJECT_ID")}" 				  resource "mazevault_integration" "ado" {				  }				    api_token  = env("MAZEVAULT_API_TOKEN")				    server_url = env("MAZEVAULT_SERVER_URL")				  provider "mazevault" {				Config: `			{		Steps: []resource.TestStep{		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,	resource.Test(t, resource.TestCase{	testAccPreCheck(t)func TestAccIntegrationResource_basic(t *testing.T) {}	}		t.Skip("MAZEVAULT_API_TOKEN not set; skipping acceptance tests")	if v := os.Getenv("MAZEVAULT_API_TOKEN"); v == "" {	}		t.Skip("MAZEVAULT_SERVER_URL not set; skipping acceptance tests")	if v := os.Getenv("MAZEVAULT_SERVER_URL"); v == "" {func testAccPreCheck(t *testing.T) {
*/
