# mazevault_entra_integration_scope

Selects which Microsoft Entra ID application registrations a project-scoped Entra integration manages. In `selected_apps` mode only the listed applications are synchronised into the project and become eligible for client-secret lifecycle management (an application can be bound to one project at a time). After the scope is written the integration is synchronised from Microsoft Graph so that [`data.mazevault_entra_app_registration`](../data-sources/entra_app_registration.md) can resolve credential IDs for [`mazevault_entra_rotation_config`](entra_rotation_config.md).

> **Destroy semantics:** destroying this resource only removes it from Terraform state. The backend cannot hold a `selected_apps` scope without applications and the provider never widens the scope to `legacy_all` implicitly; destroy the `mazevault_integration` to drop all bindings. Entra ID itself is never modified by this resource.

## Example Usage

```hcl
resource "mazevault_integration" "entra" {
  project_id    = mazevault_project.identity.id
  name          = "Entra ID — tenant"
  type          = "identity_provider"
  provider_name = "entra_id"
  environment   = "prod"

  config = {
    tenant_id   = var.tenant_id
    client_id   = var.automation_client_id
    auth_method = "client_secret"
  }

  sensitive_config = {
    client_secret = var.automation_client_secret
  }
}

resource "mazevault_entra_integration_scope" "payments" {
  integration_id = mazevault_integration.entra.id

  selected_apps = [{
    app_id       = var.payments_app_client_id
    object_id    = var.payments_app_object_id
    display_name = "APP_PAYMENTS_PROD"
  }]
}

data "mazevault_entra_app_registration" "payments" {
  integration_id = mazevault_integration.entra.id
  app_id         = var.payments_app_client_id

  depends_on = [mazevault_entra_integration_scope.payments]
}

resource "mazevault_entra_rotation_config" "payments" {
  credential_id           = data.mazevault_entra_app_registration.payments.primary_credential_id
  rotation_enabled        = true
  staged_rotation_enabled = true

  post_rotation_actions {
    type  = "azure_keyvault"
    order = 1
    config = {
      integration_id = mazevault_integration.kv.id
      secret_name    = "payments-client-secret"
    }
  }
}
```

## Argument Reference

- `integration_id` - (Required, ForceNew) Project Entra integration whose scope is managed.
- `scope_mode` - (Optional) `selected_apps` (default) manages only `selected_apps`; `legacy_all` manages every application visible to the integration identity and ignores `selected_apps`.
- `selected_apps` - (Optional) List of applications to manage. Non-empty when `scope_mode = "selected_apps"`. Each entry supports:
  - `app_id` - (Required) Entra application (client) ID.
  - `object_id` - (Required) Entra object ID of the registration (or service principal, see `binding_kind`).
  - `display_name` - (Required) Display name used until the first synchronisation refreshes it.
  - `tenant_id` - (Optional) Tenant ID; defaults to the integration tenant.
  - `binding_kind` - (Optional) `app_registration` (default) or `service_principal`.
- `sync_after_apply` - (Optional) Synchronise applications and credentials from Microsoft Graph after create/update. Defaults to `true`.

## Attribute Reference

- `id` - Equals `integration_id`.
- `project_id` - Project that owns the integration.

## Import

```
terraform import mazevault_entra_integration_scope.payments <integration-uuid>
```
