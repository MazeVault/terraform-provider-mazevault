# mazevault_rotation_resources

Lists all rotation resources registered in MazeVault, optionally filtered by resource kind and/or environment scope.

Use this data source to enumerate every managed resource that participates in automatic rotation, check their current status, or feed their IDs into other resources and modules.

## Example Usage

### List all rotation resources

```hcl
data "mazevault_rotation_resources" "all" {}

output "rotation_resource_count" {
  value = length(data.mazevault_rotation_resources.all.resources)
}
```

### Filter by kind

```hcl
data "mazevault_rotation_resources" "secrets" {
  kind = "secret"
}
```

### Filter by kind and environment scope

```hcl
data "mazevault_rotation_resources" "prod_certs" {
  kind              = "certificate"
  environment_scope = "production"
}

output "overdue_certs" {
  value = [
    for r in data.mazevault_rotation_resources.prod_certs.resources :
    r.display_name if r.status_summary == "overdue"
  ]
}
```

## Argument Reference

- `kind` - (Optional) Filter by resource kind. Supported values: `secret`, `certificate`, `entra_credential`, `ssh_key`. Omit to return all kinds.
- `environment_scope` - (Optional) Filter by environment scope (e.g. `staging`, `production`). Omit to return all environments.

## Attribute Reference

- `resources` - List of matching rotation resources. Each element contains:
  - `id` - Rotation resource record ID.
  - `resource_kind` - Kind of the underlying resource (`secret`, `certificate`, etc.).
  - `resource_id` - UUID of the underlying resource.
  - `project_id` - Project this resource belongs to.
  - `environment_scope` - Environment scope of the resource.
  - `display_name` - Human-readable name.
  - `enabled` - Whether automatic rotation is enabled.
  - `manual_only` - Whether the resource only supports manual rotation triggers.
  - `status_summary` - Short status string (e.g. `ok`, `overdue`, `error`, `pending`).
  - `next_due_at` - RFC 3339 timestamp of the next scheduled rotation. Empty string if not scheduled.
