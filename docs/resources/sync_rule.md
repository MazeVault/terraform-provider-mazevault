# mazevault_sync_rule

Manages a synchronization rule between MazeVault and an external secret provider (Azure Key Vault, GitHub Actions, GitLab, etc.).

## Example Usage

```hcl
resource "mazevault_sync_rule" "akv_pull" {
  name              = "prod-akv-pull"
  project_id        = var.project_id
  integration_id    = var.akv_integration_id
  environment       = "production"
  path_prefix       = "secrets/myapp/"
  key_transform     = "app-{{environment}}-{{key}}"
  conflict_strategy = "mazevault_wins"
  sync_direction    = "pull"
  sync_mode         = "incremental"
}
```

## Argument Reference

- `name` - (Required) Human-readable name for the sync rule.
- `project_id` - (Required) The project this sync rule belongs to.
- `integration_id` - (Required) The integration (Azure KV, GitHub Actions, etc.) used for synchronization.
- `environment` - (Optional) The MazeVault environment to synchronize secrets into (e.g. `production`, `staging`).
- `path_prefix` - (Optional) Path prefix in the external system (e.g. `/secret/data/myapp` or `secrets/APP_`).
- `key_transform` - (Optional) Template for transforming external key names. Use `{{key}}` as a placeholder (e.g. `app-{{environment}}-{{key}}`).
- `conflict_strategy` - (Optional) How to resolve conflicts: `manual_resolution`, `mazevault_wins`, `external_wins`, `most_recent_wins`. Defaults to `manual_resolution`.
- `sync_direction` - (Optional) Direction of synchronization: `pull` (external → MazeVault), `push` (MazeVault → external), `bidirectional`. Defaults to `pull`.
- `sync_mode` - (Optional) Sync granularity: `incremental` (changes only) or `full_sync` (reconcile all). Defaults to `incremental`.

## Attribute Reference

- `id` - The unique identifier of the sync rule.

## Import

Sync rules can be imported using their UUID:

```
terraform import mazevault_sync_rule.akv_pull <sync-rule-uuid>
```
