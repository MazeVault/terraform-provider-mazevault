# mazevault_project_template

Manages an organization-level **Project Template** of category `secrets` or `keys` — the entries shown in the frontend *Project Templates* view. A template pre-configures how projects of that kind are created: protocol/provider defaults, the environments a project may use, and the variables prompted at project creation.

> Certificate-category templates additionally provision the issuing certificate template and are managed by [`mazevault_certificate_template`](certificate_template.md). Passing `category = "certificates"` here is rejected at plan time.

## Example Usage

```hcl
resource "mazevault_project_template" "secrets_manager" {
  organization_id = mazevault_organization.main.id
  name            = "Secrets — Azure Key Vault"
  description     = "Secret projects backed by an Azure Key Vault integration"
  category        = "secrets"
  subtype         = "secrets_manager"

  protocol_config = jsonencode({
    provider      = "azure_key_vault"
    rotation_days = 30
  })

  environment_rules = jsonencode({
    allow_all = false
    allowed   = ["dev", "itg", "stg", "tst", "pre", "prod"]
  })

  variable_schema = jsonencode([
    { name = "vault_url", type = "string", required = true },
  ])
}

resource "mazevault_project_template" "database_secrets" {
  organization_id = mazevault_organization.main.id
  name            = "Secrets — Database credentials"
  category        = "secrets"
  subtype         = "database"
}

resource "mazevault_project_template" "ssh_keys" {
  organization_id = mazevault_organization.main.id
  name            = "Keys — SSH"
  category        = "keys"
  subtype         = "ssh"
}
```

## Argument Reference

- `organization_id` - (Required, ForceNew) Organization that owns the template.
- `name` - (Required) Template name.
- `description` - (Optional) Human-readable description.
- `category` - (Required, ForceNew) `secrets` or `keys`.
- `subtype` - (Optional) Category-specific subtype. `secrets`: `secrets_manager`, `database`, `helm_chart`, `configuration`, `agent`, `custom`. `keys`: `aws_kms`, `azure_key_vault`, `hashicorp_vault`, `ssh`, `gpg`, `custom`. Defaults to `custom`.
- `version` - (Optional) Free-form version label.
- `protocol_config` - (Optional) JSON object with protocol/provider defaults. Defaults to `{}`.
- `environment_rules` - (Optional) JSON object restricting usable environments. Defaults to `{"allow_all": true}`.
- `variable_schema` - (Optional) JSON array of variables prompted when a project is created from the template. Defaults to `[]`.

JSON arguments are compared semantically; key order and whitespace do not produce drift.

## Attribute Reference

- `id` - Template UUID.

## Import

```
terraform import mazevault_project_template.secrets_manager <organization-uuid>/<template-uuid>
```
