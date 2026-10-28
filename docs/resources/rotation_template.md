# mazevault_rotation_template

Manages a reusable rotation policy template in the authenticated organization's catalog. Creating a template does not automatically apply it to rotation configs.

## Example Usage

```hcl
resource "mazevault_rotation_template" "quarterly" {
  name                   = "quarterly-rotation"
  description            = "Rotate every 90 days, renew 14 days before expiry"
  organization_id        = var.organization_id
  project_id             = mazevault_project.app.id  # optional: scope to one project
  is_default             = false
  rotation_interval_days = 90
  lead_time_days         = 14
  grace_period_days      = 7
  max_retry_attempts     = 3
  retry_delay_seconds    = 30
  timeout_minutes        = 30
  secret_custody_mode    = "local"
}
```

## Argument Reference

- `name` - (Required) Human-readable name for the template.
- `description` - (Optional) Optional description of what this template is for.
- `organization_id` - (Optional) Owning organization. Omit to use the authenticated organization; the server resolves and returns it (`Computed`).
- `project_id` - (Optional) Scope the template to a single project. Omit for an organization-wide template.
- `is_default` - (Optional) Whether this is the default template. Only one template per organization can be the default.
- `rotation_interval_days` - (Optional) How frequently (in days) the secret or certificate should be rotated. Must be a non-negative integer (`0` is a valid explicit value).
- `lead_time_days` - (Optional) Days before expiry to start the rotation. Applies to certificate renewal. Must be a non-negative integer (`0` is a valid explicit value).
- `grace_period_days` - (Optional) Days after rotation to keep the previous value active for rollback. Must be a non-negative integer (`0` is a valid explicit value).
- `max_retry_attempts` - (Optional) Maximum number of retry attempts on failure. Must be a non-negative integer (`0` is a valid explicit value).
- `retry_delay_seconds` - (Optional) Delay in seconds between retry attempts on failure. Must be a non-negative integer (`0` is a valid explicit value).
- `timeout_minutes` - (Optional) Maximum execution time in minutes before the rotation is considered timed out. Must be a non-negative integer (`0` is a valid explicit value).
- `secret_custody_mode` - (Optional) Secret custody mode: `local`, `transient`, or `delegated`.

## Attribute Reference

- `id` - The unique identifier of the rotation template.

!> **Catalog, not inheritance.** A rotation template is a reusable catalog entry of default settings. Creating or updating a template does **not** retroactively change rotation configs that were previously created from it.

## Import

Rotation templates can be imported using their UUID:

```
terraform import mazevault_rotation_template.quarterly <template-uuid>
```
