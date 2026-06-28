# mazevault_certificate_rotation_config

Manages the automatic renewal configuration for a certificate managed by MazeVault.

This resource controls when and how a certificate is automatically renewed: lead time before expiry, retry behaviour, post-renewal notifications, and actions to run after a successful renewal.

> **No hard-delete:** Destroying this resource disables automatic rotation (`enabled = false`) and removes it from Terraform state. The backend record is **not** deleted. To re-enable rotation, re-apply or re-import the resource.

## Example Usage

### Minimal — enable renewal with default settings

```hcl
resource "mazevault_certificate_rotation_config" "web_tls" {
  certificate_id = var.web_tls_cert_id
}
```

### Full configuration with notifications and a post-renewal webhook

```hcl
resource "mazevault_certificate_rotation_config" "web_tls" {
  certificate_id     = var.web_tls_cert_id
  enabled            = true
  renewal_lead_days  = 21
  max_retry_attempts = 5
  retry_delay_seconds = 120
  timeout_minutes    = 15
  notification_emails = ["ops@example.com", "pki@example.com"]

  post_rotation_actions {
    type       = "webhook"
    order      = 1
    on_failure = "notify"
    config = {
      url     = "https://hooks.example.com/cert-renewed"
      method  = "POST"
    }
  }

  post_rotation_actions {
    type              = "agent_command"
    order             = 2
    on_failure        = "continue"
    target_environment = "production"
    config = {
      command = "nginx -s reload"
    }
  }
}
```

### Cron-scheduled renewal

```hcl
resource "mazevault_certificate_rotation_config" "internal_ca" {
  certificate_id = var.internal_ca_cert_id
  enabled        = true
  schedule       = "0 2 * * 0"  # Every Sunday at 02:00 UTC
}
```

## Argument Reference

- `certificate_id` - (Required, ForceNew) The UUID of the certificate to configure. Changing this value destroys and recreates the resource.
- `enabled` - (Optional) Whether automatic renewal is enabled. Defaults to `true`.
- `schedule` - (Optional) Cron expression for scheduled renewal (e.g. `"0 3 * * *"`). When set, takes precedence over `renewal_lead_days`-based trigger timing.
- `renewal_lead_days` - (Optional) Number of days before certificate expiry to start the renewal process. Defaults to `30`.
- `max_retry_attempts` - (Optional) Maximum number of retry attempts on failure. Defaults to `3`.
- `retry_delay_seconds` - (Optional) Delay in seconds between retry attempts. Defaults to `60`.
- `timeout_minutes` - (Optional) Maximum execution time in minutes before the renewal is considered timed out. Defaults to `10`.
- `notification_emails` - (Optional) List of e-mail addresses to notify after a renewal event.
- `post_rotation_actions` - (Optional) Ordered list of actions to execute after a successful renewal. See [Post-Rotation Actions](#post-rotation-actions) below.

### Post-Rotation Actions

Each `post_rotation_actions` block supports:

- `type` - (Required) Action type. Supported values: `spring_actuator_refresh`, `webhook`, `agent_command`, `shell_script`, `azure_keyvault`, `kubernetes_secret`, `iis_recycle`.
- `config` - (Optional) String key/value map with action-specific configuration parameters.
- `order` - (Optional) Execution order. Lower numbers run first.
- `on_failure` - (Optional) Behaviour when this action fails: `continue`, `rollback`, or `notify`.
- `gateway_id` - (Optional) Pin this action to a specific gateway ID instead of using environment-based routing.
- `target_environment` - (Optional) Override the environment context used to resolve which gateway executes this action.

## Attribute Reference

In addition to all arguments above, the following computed attributes are exported:

- `id` - The unique identifier of the rotation config record.
- `config_initialized` - Whether the rotation config has been fully initialised by the backend.
- `last_rotated_at` - RFC 3339 timestamp of the last successful renewal.
- `next_rotation_at` - RFC 3339 timestamp of the next scheduled renewal.

## Import

Certificate rotation configs can be imported using the **certificate UUID**:

```
terraform import mazevault_certificate_rotation_config.web_tls <certificate-uuid>
```

The import ID is the `certificate_id`, **not** the rotation config record `id`.
