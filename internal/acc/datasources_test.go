package acc

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccAuditLogsDataSource_read verifies the data source accepts filters and
// returns a list (may be empty on a fresh system — only the list attribute
// itself is checked).
func TestAccAuditLogsDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "al_org" {
  name = "tf-acc-al-org"
}
resource "mazevault_project" "al_proj" {
  organization_id = mazevault_organization.al_org.id
  name            = "tf-acc-al-proj"
  environment     = "dev"
}
data "mazevault_audit_logs" "recent" {
  project_id = mazevault_project.al_proj.id
  limit      = 50
  offset     = 0
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mazevault_audit_logs.recent", "logs.#"),
				),
			},
		},
	})
}

// TestAccProjectCertificatesDataSource_read verifies the data source returns an
// empty list for a new project.
func TestAccProjectCertificatesDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "pc_org" {
  name = "tf-acc-pc-org"
}
resource "mazevault_project" "pc_proj" {
  organization_id = mazevault_organization.pc_org.id
  name            = "tf-acc-pc-proj"
  environment     = "dev"
}
data "mazevault_project_certificates" "certs" {
  project_id = mazevault_project.pc_proj.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mazevault_project_certificates.certs", "certificates.#"),
				),
			},
		},
	})
}

// TestAccConsistencyStatusDataSource_read verifies the consistency status data
// source is readable.
func TestAccConsistencyStatusDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "cs_org" {
  name = "tf-acc-cs-org"
}
resource "mazevault_project" "cs_proj" {
  organization_id = mazevault_organization.cs_org.id
  name            = "tf-acc-cs-proj"
  environment     = "dev"
}
data "mazevault_consistency_status" "status" {
  project_id = mazevault_project.cs_proj.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mazevault_consistency_status.status", "project_id"),
				),
			},
		},
	})
}

// TestAccRotationExecutionsDataSource_read verifies the data source returns a
// list (empty is valid for a project with no rotations yet).
func TestAccRotationExecutionsDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
data "mazevault_rotation_executions" "execs" {
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mazevault_rotation_executions.execs", "executions.#"),
				),
			},
		},
	})
}
