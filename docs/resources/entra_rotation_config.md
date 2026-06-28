# mazevault_entra_rotation_config

Manages the automatic rotation configuration for a Microsoft Entra (Azure AD) application credential managed by MazeVault.

This resource controls when and how client secrets are automatically rotated: lead time before expiry, Key Vault write-back, Spring Boot actuator refresh, staged rotation soak windows, and post-rotation webhook or agent actions.

> **No hard-delete:** Destroying this resource sets `rotation_enabled = false` and removes it from Terraform state. The backend record is **not** deleted. To re-enable rotation, re-apply or re-import the resource.

## Example Usage

### Minimal — enable rotation with defaults

```hcl
resource "mazevault_entra_rotation_config" "app" {
  credential_id = var.entra_credential_id
}
```

### Write new secret to Key Vault after rotation

```hcl
resource "mazevault_entra_rotation_config" "app" {
  credential_id                = var.entra_credential_id
  rotation_enabled             = true
  rotation_days_before_expiry  = 30
  grace_period_days            = 7

  kv_integration_ids = [var.keyvault_integration_id]
  secret_name        = "my-app-client-secret"

  notification_emails = ["ops@example.com"]
}
```

### Staged rotation with Spring Boot refresh

```hcl
resource "mazevault_entra_rotation_config" "app" {
  credential_id                = var.entra_credential_id
  rotation_enabled             = true
  rotation_days_before_expiry  = 21
  grace_period_days            = 3

  staged_rotation_enabled = true
  soak_window_hours       = 24

  spring_endpoints = [
    "https://app1.internal/actuator/refresh",
    "https://app2.internal/actuator/refresh",
  ]

  post_rotation_actions {
    type       = "webhook"
    order      = 1
    on_failure = "notify"
    config = {
      url = "https://alerting.example.com/entra-rotated"
    }
  }
}
```

## Argument Reference

- `credential_id` - (Required, ForceNew) The UUID of the Entra credential to configure. Changing this value destroys and recreates the resource.
- `rotation_enabled` - (Optional) Whether automatic rotation is enabled. Defaults to `true`.
- `rotation_days_before_expiry` - (Optional) Number of days before credential expiry to start the rotation. Defaults to `30`.
- `grace_period_days` - (Optional) Number of days the old credential remains active after the new one is created. Defaults to `7`.
- `kv_integration_ids` - (Optional) IDs of Key Vault integrations that should receive the new credential value after rotation.
- `secret_name` - (Optional) Name of the secret in the target Key Vault. Required when `kv_integration_ids` is set.
- `spring_endpoints` - (Optional) URLs of Spring Boot Actuator `/actuator/refresh` endpoints to call after rotation.
- `webhook_urls` - (Optional) Webhook URLs to call after rotation. The payload contains the new credential metadata.
- `staged_rotation_enabled` - (Optional) Enable secondary-first staged rotation. Defaults to `false`.
- `soak_window_hours` - (Optional) Minimum hours to wait between secondary verification and primary promotion during staged rotation. Defaults to `48`.
- `post_rotation_actions` - (Optional) Ordered list of actions to execute after a successful rotation. See [Post-Rotation Actions](#post-rotation-actions) below.

### Post-Rotation Actions

Each `post_rotation_actions` block supports:

- `type` - (Required) Action type. Supported values: `spring_actuator_refresh`, `webhook`, `agent_command`, `agent_secret_sync`, `iis_recycle`.
- `config` - (Optional) String key/value map with action-specific configuration parameters.
- `order` - (Optional) Execution order. Lower numbers run first.
- `on_failure` - (Optional) Behaviour when this action fails: `continue`, `rollback`, or `notify`.
- `gateway_id` - (Optional) Pin this action to a specific gateway ID.
- `target_environment` - (Optional) Override the environment context used to resolve which gateway executes this action.

## Attribute Reference

In addition to all arguments above, the following computed attributes are exported:

- `id` - Equals `credential_id` (the backend has no separate rotation config UUID for Entra credentials).
- `last_rotated_at` - RFC 3339 timestamp of the last successful rotation.

## Import

Entra rotation configs can be imported using the **credential UUID**:

```
terraform import mazevault_entra_rotation_config.app <credential-uuid>
```

The import ID is the `credential_id`.
