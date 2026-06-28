# mazevault_project_rotation_configs

Lists all rotation configurations associated with a specific project.

Use this data source to enumerate every secret rotation config within a project — for example to audit rotation coverage, generate compliance reports, or pass config IDs to automation pipelines.

## Example Usage

### List all rotation configs in a project

```hcl
data "mazevault_project_rotation_configs" "my_project" {
  project_id = var.project_id
}

output "rotation_config_ids" {
  value = [for c in data.mazevault_project_rotation_configs.my_project.configs : c.id]
}
```

### Find secrets with rotation disabled

```hcl
data "mazevault_project_rotation_configs" "my_project" {
  project_id = var.project_id
}

output "rotation_disabled_secrets" {
  value = [
    for c in data.mazevault_project_rotation_configs.my_project.configs :
    c.secret_id if !c.enabled
  ]
}
```

### Find overdue rotations

```hcl
data "mazevault_project_rotation_configs" "my_project" {
  project_id = var.project_id
}

output "overdue_configs" {
  value = [
    for c in data.mazevault_project_rotation_configs.my_project.configs :
    { id = c.id, secret_id = c.secret_id, status = c.status }
    if c.status == "overdue"
  ]
}
```

## Argument Reference

- `project_id` - (Required) UUID of the project whose rotation configs should be listed.

## Attribute Reference

- `configs` - List of rotation configs belonging to the project. Each element contains:
  - `id` - Rotation config ID.
  - `secret_id` - ID of the secret this config applies to.
  - `enabled` - Whether automatic rotation is enabled.
  - `schedule` - Cron schedule expression. Empty string if not set.
  - `rotation_interval_days` - Rotation interval in days. `0` if not configured.
  - `status` - Current rotation status (e.g. `ok`, `overdue`, `error`, `pending`).
  - `last_rotated_at` - RFC 3339 timestamp of the last successful rotation. Empty string if never rotated.
  - `next_rotation_at` - RFC 3339 timestamp of the next scheduled rotation. Empty string if not scheduled.
