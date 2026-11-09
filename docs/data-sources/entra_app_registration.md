# mazevault_entra_app_registration (Data Source)

Looks up a Microsoft Entra ID application registration that has been synchronised into a project Entra integration and exposes its MazeVault identifiers and credentials. Use it to obtain the `credential_id` required by [`mazevault_entra_rotation_config`](../resources/entra_rotation_config.md) after [`mazevault_entra_integration_scope`](../resources/entra_integration_scope.md) has selected and synchronised the application.

Secret values are never exposed.

## Example Usage

```hcl
data "mazevault_entra_app_registration" "payments" {
  integration_id = mazevault_integration.entra.id
  app_id         = var.payments_app_client_id

  depends_on = [mazevault_entra_integration_scope.payments]
}

output "payments_primary_credential" {
  value = data.mazevault_entra_app_registration.payments.primary_credential_id
}
```

## Argument Reference

- `integration_id` - (Required) Project Entra integration ID.
- `app_id` - (Required) Entra application (client) ID.

## Attribute Reference

- `id` - MazeVault app registration UUID.
- `object_id` - Entra object ID.
- `tenant_id` - Entra tenant ID.
- `project_id` - Owning project.
- `display_name` - Display name from Entra ID.
- `managed_by_mazevault` - Whether MazeVault owns the credential lifecycle.
- `sync_status` - Last synchronisation status.
- `primary_credential_id` - ID of the active primary client-secret credential (empty when none).
- `secondary_credential_id` - ID of the most recently issued active non-primary client-secret credential (empty when none).
- `credentials` - All credentials, newest `end_date` first. Each entry exposes `id`, `key_id`, `display_name`, `credential_type` (`client_secret` or `certificate`), `is_primary`, `status`, `lifecycle_state`, `rotation_enabled`, `start_date`, `end_date`.

The data source fails when the application is not present in the integration — select it with `mazevault_entra_integration_scope` and make sure a synchronisation has run.
