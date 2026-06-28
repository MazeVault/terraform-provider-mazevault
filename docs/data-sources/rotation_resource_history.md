# mazevault_rotation_resource_history

Lists the rotation execution history for a specific resource.

Use this data source to audit past rotation attempts, diagnose failures, or monitor the cadence of automated rotations.

## Example Usage

### Fetch execution history for a secret

```hcl
data "mazevault_rotation_resource_history" "my_secret" {
  kind        = "secret"
  resource_id = var.secret_id
}

output "last_rotation_status" {
  value = length(data.mazevault_rotation_resource_history.my_secret.executions) > 0 ? (
    data.mazevault_rotation_resource_history.my_secret.executions[0].status
  ) : "never rotated"
}
```

### Check for failed rotations

```hcl
data "mazevault_rotation_resource_history" "cert_history" {
  kind        = "certificate"
  resource_id = var.cert_id
}

output "failed_executions" {
  value = [
    for e in data.mazevault_rotation_resource_history.cert_history.executions :
    { id = e.id, started_at = e.started_at, error = e.error }
    if e.status == "failed"
  ]
}
```

## Argument Reference

- `kind` - (Required) Resource kind. Supported values: `secret`, `certificate`, `entra_credential`, `ssh_key`.
- `resource_id` - (Required) UUID of the resource whose rotation history should be retrieved.

## Attribute Reference

- `executions` - Ordered list of rotation execution records (most-recent first). Each element contains:
  - `id` - Execution ID.
  - `config_id` - ID of the rotation config that triggered this execution.
  - `status` - Execution status: `pending`, `running`, `success`, `failed`, or `skipped`.
  - `started_at` - RFC 3339 start timestamp. Empty string if not yet started.
  - `completed_at` - RFC 3339 completion timestamp. Empty string if still running.
  - `error` - Error message if the execution failed. Empty string on success.
