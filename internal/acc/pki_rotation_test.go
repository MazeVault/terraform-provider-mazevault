package acc

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccRenewalPolicyResource_basic covers create → read → update → destroy.
func TestAccRenewalPolicyResource_basic(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: providerConfig() + `
resource "mazevault_organization" "rp_org" {
  name = "tf-acc-rp-org"
}
resource "mazevault_renewal_policy" "test" {
  organization_id  = mazevault_organization.rp_org.id
  name             = "tf-acc-renewal-policy"
  lead_days        = 30
  key_reuse_enabled = false
  auto_approve     = false
  notify_emails    = "ops@example.com"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_renewal_policy.test", "id"),
					resource.TestCheckResourceAttr("mazevault_renewal_policy.test", "name", "tf-acc-renewal-policy"),
					resource.TestCheckResourceAttr("mazevault_renewal_policy.test", "lead_days", "30"),
					resource.TestCheckResourceAttr("mazevault_renewal_policy.test", "key_reuse_enabled", "false"),
					resource.TestCheckResourceAttr("mazevault_renewal_policy.test", "auto_approve", "false"),
					resource.TestCheckResourceAttr("mazevault_renewal_policy.test", "notify_emails", "ops@example.com"),
				),
			},
			// Update lead_days and auto_approve
			{
				Config: providerConfig() + `
resource "mazevault_organization" "rp_org" {
  name = "tf-acc-rp-org"
}
resource "mazevault_renewal_policy" "test" {
  organization_id  = mazevault_organization.rp_org.id
  name             = "tf-acc-renewal-policy-updated"
  lead_days        = 45
  key_reuse_enabled = false
  auto_approve     = true
  notify_emails    = "ops@example.com"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mazevault_renewal_policy.test", "name", "tf-acc-renewal-policy-updated"),
					resource.TestCheckResourceAttr("mazevault_renewal_policy.test", "lead_days", "45"),
					resource.TestCheckResourceAttr("mazevault_renewal_policy.test", "auto_approve", "true"),
				),
			},
		},
	})
}

// TestAccRotationConfigResource_basic covers create → read → destroy for a
// rotation config tied to a secret.
func TestAccRotationConfigResource_basic(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "rc_org" {
  name = "tf-acc-rc-org"
}
resource "mazevault_project" "rc_proj" {
  organization_id = mazevault_organization.rc_org.id
  name            = "tf-acc-rc-proj"
  environment     = "staging"
}
resource "mazevault_secret" "rc_sec" {
  project_id  = mazevault_project.rc_proj.id
  key         = "RC_SECRET"
  value       = "rotate-me"
  environment = "staging"
}
resource "mazevault_rotation_config" "test" {
  secret_id              = mazevault_secret.rc_sec.id
  rotation_interval_days = 30
  enabled                = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_rotation_config.test", "id"),
					resource.TestCheckResourceAttr("mazevault_rotation_config.test", "rotation_interval_days", "30"),
					resource.TestCheckResourceAttr("mazevault_rotation_config.test", "enabled", "true"),
				),
			},
		},
	})
}

// TestAccRenewalQueueDataSource_read verifies data.mazevault_renewal_queue
// returns a list (may be empty on a fresh environment).
func TestAccRenewalQueueDataSource_read(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "rq_org" {
  name = "tf-acc-rq-org"
}
data "mazevault_renewal_queue" "all" {
  organization_id = mazevault_organization.rq_org.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mazevault_renewal_queue.all", "items.#"),
				),
			},
		},
	})
}
