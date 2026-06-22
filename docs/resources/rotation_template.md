# mazevault_rotation_template

Manages a reusable rotation policy template. Templates can be applied to rotation configs to inherit consistent policy settings across secrets and certificates.

## Example Usage

```hcl
resource "mazevault_rotation_template" "quarterly" {
  name                   = "quarterly-rotation"
  description            = "Rotate every 90 days, renew 14 days before expiry"
  organization_id        = var.organization_id
  is_default             = false
  rotation_interval_days = 90
  lead_days              = 14
  retention_days         = 7
  max_retries            = 3
  timeout_minutes        = 30
}
```

## Argument Reference

- `name` - (Required) Human-readable name for the template.
- `description` - (Optional) Optional description of what this template is for.
- `organization_id` - (Optional) Scope the template to a specific organization. Omit to apply globally.
- `is_default` - (Optional) Whether this is the default template. Only one template per organization can be the default.
- `rotation_interval_days` - (Optional) How frequently (in days) the secret or certificate should be rotated.
- `lead_days` - (Optional) Days before expiry to start the rotation. Applies to certificate renewal.
- `retention_days` - (Optional) Days after rotation to keep the previous value active for rollback.
- `max_retries` - (Optional) Maximum number of retry attempts on failure.
- `timeout_minutes` - (Optional) Maximum execution time in minutes before the rotation is considered timed out.

## Attribute Reference

- `id` - The unique identifier of the rotation template.

## Import

Rotation templates can be imported using their UUID:

```
terraform import mazevault_rotation_template.quarterly <template-uuid>
```
