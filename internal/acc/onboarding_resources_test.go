package acc

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccConfigTemplateResource_namingConvention covers create → read → update →
// destroy for a Configuration Management template of type naming_convention.
// Requires a backend with the config_management license feature.
func TestAccConfigTemplateResource_namingConvention(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "ct_org" {
  name = "tf-acc-ct-org"
}
resource "mazevault_config_template" "naming" {
  organization_id = mazevault_organization.ct_org.id
  name            = "tf-acc-naming"
  template_type   = "naming_convention"
  definition = jsonencode({
    patterns = {
      project   = "{company}-{env}-{project}"
      key_vault = "{company}{env}{project}kv{index}"
      secret    = "{project}-{purpose}"
    }
    non_prod_prefix       = "npr"
    non_prod_environments = ["dev", "itg", "stg", "tst", "pre"]
    variables = [
      { name = "company", required = true, max_length = 4, pattern = "^[a-z]{2,4}$" },
      { name = "env", required = true },
      { name = "project", required = true, max_length = 12 },
      { name = "purpose", required = false },
      { name = "index", required = false, default = "01" },
    ]
    resource_constraints = {
      key_vault = { max_length = 24, allowed_chars = "^[a-z0-9-]+$" }
    }
  })
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_config_template.naming", "id"),
					resource.TestCheckResourceAttr("mazevault_config_template.naming", "template_type", "naming_convention"),
					resource.TestCheckResourceAttr("mazevault_config_template.naming", "is_active", "true"),
					resource.TestCheckResourceAttr("mazevault_config_template.naming", "priority", "100"),
				),
			},
			{
				Config: providerConfig() + `
resource "mazevault_organization" "ct_org" {
  name = "tf-acc-ct-org"
}
resource "mazevault_config_template" "naming" {
  organization_id = mazevault_organization.ct_org.id
  name            = "tf-acc-naming-renamed"
  template_type   = "naming_convention"
  priority        = 10
  definition = jsonencode({
    patterns = {
      project = "{company}-{env}-{project}"
    }
    non_prod_prefix       = "npr"
    non_prod_environments = ["dev"]
    variables = [
      { name = "company", required = true },
      { name = "env", required = true },
      { name = "project", required = true },
    ]
    resource_constraints = {}
  })
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mazevault_config_template.naming", "name", "tf-acc-naming-renamed"),
					resource.TestCheckResourceAttr("mazevault_config_template.naming", "priority", "10"),
				),
			},
		},
	})
}

// TestAccConfigTemplateResource_secretPatternValidation proves that the backend
// validator result is surfaced as a plan/apply error instead of a generic 400.
func TestAccConfigTemplateResource_secretPatternValidation(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "ct_org2" {
  name = "tf-acc-ct-org2"
}
resource "mazevault_config_template" "bad" {
  organization_id = mazevault_organization.ct_org2.id
  name            = "tf-acc-bad-pattern"
  template_type   = "secret_pattern"
  definition = jsonencode({
    patterns   = [{ name = "x", type = "regex", regex = "(unclosed", severity = "high" }]
    min_length = 8
    max_length = 128
  })
}
`,
				ExpectError: regexpMustCompile(`definition rejected by backend validator`),
			},
			{
				Config: providerConfig() + `
resource "mazevault_organization" "ct_org2" {
  name = "tf-acc-ct-org2"
}
resource "mazevault_config_template" "good" {
  organization_id = mazevault_organization.ct_org2.id
  name            = "tf-acc-good-pattern"
  template_type   = "secret_pattern"
  definition = jsonencode({
    patterns = [{
      name     = "password-assignment"
      type     = "regex"
      regex    = "(?i)password\\s*="
      severity = "high"
    }]
    entropy_threshold = 3.5
    min_length        = 8
    max_length        = 128
  })
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_config_template.good", "id"),
				),
			},
		},
	})
}

// TestAccProjectTemplateResource_secrets covers create → update → destroy of a
// secrets-category Project Template.
func TestAccProjectTemplateResource_secrets(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "pt_org" {
  name = "tf-acc-pt-org"
}
resource "mazevault_project_template" "secrets" {
  organization_id = mazevault_organization.pt_org.id
  name            = "tf-acc-secrets-manager"
  category        = "secrets"
  subtype         = "secrets_manager"
  protocol_config = jsonencode({ provider = "azure_key_vault", rotation_days = 30 })
  variable_schema = jsonencode([{ name = "vault_url", type = "string", required = true }])
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("mazevault_project_template.secrets", "id"),
					resource.TestCheckResourceAttr("mazevault_project_template.secrets", "category", "secrets"),
					resource.TestCheckResourceAttr("mazevault_project_template.secrets", "subtype", "secrets_manager"),
					resource.TestCheckResourceAttr("mazevault_project_template.secrets", "environment_rules", `{"allow_all":true}`),
				),
			},
			{
				Config: providerConfig() + `
resource "mazevault_organization" "pt_org" {
  name = "tf-acc-pt-org"
}
resource "mazevault_project_template" "secrets" {
  organization_id   = mazevault_organization.pt_org.id
  name              = "tf-acc-secrets-manager-v2"
  description       = "updated"
  category          = "secrets"
  subtype           = "database"
  version           = "2"
  protocol_config   = jsonencode({ provider = "postgres" })
  environment_rules = jsonencode({ allow_all = false, allowed = ["dev", "prod"] })
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mazevault_project_template.secrets", "name", "tf-acc-secrets-manager-v2"),
					resource.TestCheckResourceAttr("mazevault_project_template.secrets", "subtype", "database"),
					resource.TestCheckResourceAttr("mazevault_project_template.secrets", "version", "2"),
				),
			},
		},
	})
}

// TestAccProjectTemplateResource_rejectsCertificates documents that the
// certificates category is reserved for mazevault_certificate_template.
func TestAccProjectTemplateResource_rejectsCertificates(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "pt_org2" {
  name = "tf-acc-pt-org2"
}
resource "mazevault_project_template" "certs" {
  organization_id = mazevault_organization.pt_org2.id
  name            = "tf-acc-should-fail"
  category        = "certificates"
}
`,
				ExpectError: regexpMustCompile(`managed by mazevault_certificate_template`),
			},
		},
	})
}

// TestAccCAAccountResource_smallstepCredentials verifies that the Smallstep
// provisioner password travels in the encrypted credentials map and that
// secret-like keys are refused in plaintext settings. Connection status is not
// asserted because no step-ca is reachable from the acceptance backend.
func TestAccCAAccountResource_smallstepCredentials(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_organization" "ss_org" {
  name = "tf-acc-ss-org"
}
resource "mazevault_ca_account" "bad" {
  organization_id = mazevault_organization.ss_org.id
  name            = "tf-acc-smallstep-bad"
  provider_type   = "smallstep"
  settings = {
    server_url           = "https://step-ca.invalid:9000"
    provisioner_password = "must-not-be-here"
  }
}
`,
				ExpectError: regexpMustCompile(`Secret material in settings`),
			},
			{
				Config: providerConfig() + `
resource "mazevault_organization" "ss_org" {
  name = "tf-acc-ss-org"
}
resource "mazevault_ca_account" "missing" {
  organization_id = mazevault_organization.ss_org.id
  name            = "tf-acc-smallstep-missing"
  provider_type   = "smallstep"
  settings = {
    server_url       = "https://step-ca.invalid:9000"
    provisioner_name = "mazevault"
  }
}
`,
				ExpectError: regexpMustCompile(`provisioner_password`),
			},
		},
	})
}

// TestAccCertificateResource_keyStorageModeValidation covers the plan-time
// custody contract of mazevault_certificate without issuing anything.
func TestAccCertificateResource_keyStorageModeValidation(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "mazevault_certificate" "bad_mode" {
  common_name      = "bad.example.invalid"
  key_storage_mode = "cloud"
}
`,
				ExpectError: regexpMustCompile(`Invalid key_storage_mode`),
			},
			{
				Config: providerConfig() + `
resource "mazevault_certificate" "bad_combo" {
  common_name        = "bad.example.invalid"
  key_integration_id = "00000000-0000-0000-0000-000000000000"
}
`,
				ExpectError: regexpMustCompile(`Requires key_storage_mode = "external"`),
			},
		},
	})
}
