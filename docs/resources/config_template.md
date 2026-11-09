# mazevault_config_template

Manages an organization-level **Configuration Management template**: a typed policy document that MazeVault enforces or applies across projects — naming conventions, secret detection patterns, secret naming policies, discovery policies, exclusion rules and resource visibility policies.

The document in `definition` is validated by the backend for the selected `template_type` **before** anything is persisted; validation findings are reported as plan/apply errors with the offending field.

> **Requirements:** the `config_management` license feature and the `config:write` permission. `resource_visibility_policy` additionally requires visibility-policy administrator rights.
>
> **Breaking change:** earlier provider versions documented `content`/`format` arguments for this resource. They never matched the backend contract (`POST /organizations/:id/config-templates` requires `template_type` + `definition`) and have been replaced. Agent file-rendering templates are a different feature and are not managed by this resource.

## Example Usage

### Naming convention

```hcl
resource "mazevault_config_template" "naming" {
  organization_id = mazevault_organization.main.id
  name            = "Corporate naming convention"
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
```

### Secret pattern and exclusion rule

```hcl
resource "mazevault_config_template" "password_pattern" {
  organization_id = mazevault_organization.main.id
  name            = "Password assignments"
  template_type   = "secret_pattern"
  priority        = 10

  definition = jsonencode({
    patterns = [{
      name     = "password-assignment"
      type     = "regex"
      regex    = "(?i)password\\s*="
      severity = "high" # critical | high | medium | low
    }]
    entropy_threshold = 3.5 # 0–8
    min_length        = 8
    max_length        = 128
  })
}

resource "mazevault_config_template" "exclude_vendor" {
  organization_id = mazevault_organization.main.id
  name            = "Ignore vendored code"
  template_type   = "exclusion_rule"

  definition = jsonencode({
    rules = [{
      name    = "vendor"
      type    = "path" # path | filename | content | extension
      pattern = "vendor/"
      reason  = "third-party code"
    }]
    excluded_extensions = [".min.js"]
  })
}
```

### Discovery policy

```hcl
resource "mazevault_config_template" "discovery" {
  organization_id = mazevault_organization.main.id
  name            = "Default discovery"
  template_type   = "discovery_policy"

  definition = jsonencode({
    scan_paths           = ["/etc", "/opt/apps"]
    scan_depth           = 5
    file_types           = ["yaml", "json", "env"]
    auto_adopt           = false
    confidence_threshold = 0.8
    schedule             = "0 * * * *"
  })
}
```

### Resource visibility policy (admin only)

```hcl
resource "mazevault_config_template" "hide_prod_vaults" {
  organization_id = mazevault_organization.main.id
  name            = "Production vaults admin-only"
  template_type   = "resource_visibility_policy"

  definition = jsonencode({
    resource_rules = [{
      id            = "rule-prod-kv"
      resource_type = "key_vault"
      action        = "admin_only"
      match_mode    = "exact"
      field         = "name"
      exact_values  = ["prod-vault"]
    }]
  })
}
```

## Argument Reference

- `organization_id` - (Required, ForceNew) Organization that owns the template.
- `name` - (Required) Template name.
- `description` - (Optional) Human-readable description.
- `template_type` - (Required, ForceNew) One of `naming_convention`, `secret_pattern`, `secret_naming_policy`, `discovery_policy`, `exclusion_rule`, `resource_visibility_policy`, `parser`.
- `definition` - (Required) JSON object with the type-specific document; build it with `jsonencode()`. Key order and whitespace differences are not reported as drift.
- `is_active` - (Optional) Whether the template is enforced. Defaults to `true`.
- `is_default` - (Optional) Marks the template as the organization default of its type. Defaults to `false`.
- `priority` - (Optional) Evaluation priority, lower runs first. Defaults to `100`.

## Attribute Reference

- `id` - Template UUID.

## Import

```
terraform import mazevault_config_template.naming <organization-uuid>/<template-uuid>
```
