# MazeVault Provider

The **MazeVault** Terraform provider allows you to manage your [MazeVault](https://mazevault.io) resources as infrastructure-as-code.

## Authentication

The provider supports two authentication methods:

### API Token (recommended for CI/CD)

```hcl
provider "mazevault" {
  server_url = "https://your-mazevault.example.com"
  api_token  = var.mazevault_token  # or MAZEVAULT_API_TOKEN env var
}
```

### Client Credentials (OAuth2)

```hcl
provider "mazevault" {
  server_url    = "https://your-mazevault.example.com"
  client_id     = var.client_id      # or MAZEVAULT_CLIENT_ID
  client_secret = var.client_secret  # or MAZEVAULT_CLIENT_SECRET
}
```

## Environment Variables

| Variable                  | Description                        |
|---------------------------|------------------------------------|
| `MAZEVAULT_SERVER_URL`    | Base URL of the MazeVault server   |
| `MAZEVAULT_API_TOKEN`     | API token for authentication       |
| `MAZEVAULT_CLIENT_ID`     | OAuth2 client ID                   |
| `MAZEVAULT_CLIENT_SECRET` | OAuth2 client secret               |

## Example Usage

```hcl
terraform {
  required_version = ">= 1.11"

  required_providers {
    mazevault = {
      source  = "MazeVault/mazevault"
      version = "~> 1.0"
    }
    random = {
      source = "hashicorp/random"
    }
  }
}

provider "mazevault" {
  server_url = "https://vault.example.com"
  api_token  = var.mazevault_token
}

resource "mazevault_project" "app" {
  organization_id = var.org_id
  name            = "my-application"
  type            = "secret"
}

ephemeral "random_password" "db_password" {
  length           = 40
  special          = true
  override_special = "_-+=@#"
}

resource "mazevault_secret" "db_password" {
  project_id  = mazevault_project.app.id
  key         = "DB_PASSWORD"
  environment = "production"

  # Terraform 1.11+ write-only argument. The password is not stored in plan or state.
  value_wo         = ephemeral.random_password.db_password.result
  value_wo_version = 1
}

resource "mazevault_secret" "api_key" {
  project_id  = mazevault_project.app.id
  key         = "API_KEY"
  environment = "production"

  # Generate the value server-side (never enters Terraform state in Orchestrator Mode).
  generate {
    length = 48
  }
}
```

## Resources

- [mazevault_organization](resources/organization.md)
- [mazevault_project](resources/project.md)
- [mazevault_project_settings](resources/project_settings.md)
- [mazevault_secret](resources/secret.md)
- [mazevault_secret_link](resources/secret_link.md)
- [mazevault_shared_secret](resources/shared_secret.md)
- [mazevault_service_identity](resources/service_identity.md)
- [mazevault_api_token](resources/api_token.md)
- [mazevault_ca](resources/ca.md)
- [mazevault_ca_account](resources/ca_account.md)
- [mazevault_certificate](resources/certificate.md)
- [mazevault_certificate_template](resources/certificate_template.md)
- [mazevault_config_template](resources/config_template.md)
- [mazevault_role](resources/role.md)
- [mazevault_group_mapping](resources/group_mapping.md)
- [mazevault_user](resources/user.md)
- [mazevault_user_role](resources/user_role.md)
- [mazevault_integration](resources/integration.md)
- [mazevault_integration_group](resources/integration_group.md)
- [mazevault_consistency_group](resources/consistency_group.md)
- [mazevault_keytab](resources/keytab.md)
- [mazevault_identity_provider](resources/identity_provider.md)
- [mazevault_environment](resources/environment.md)
- [mazevault_renewal_policy](resources/renewal_policy.md)
- [mazevault_approval_policy](resources/approval_policy.md)
- [mazevault_rotation_config](resources/rotation_config.md)
- [mazevault_rotation_workflow](resources/rotation_workflow.md)
- [mazevault_rotation_template](resources/rotation_template.md)
- [mazevault_secret_rotation_target](resources/secret_rotation_target.md)
- [mazevault_sync_rule](resources/sync_rule.md)
- [mazevault_deployment](resources/deployment.md)
- [mazevault_certificate_rotation_config](resources/certificate_rotation_config.md)
- [mazevault_entra_rotation_config](resources/entra_rotation_config.md)
- [mazevault_hsm_device](resources/hsm_device.md)

## Data Sources

- [mazevault_organization](data-sources/organization.md)
- [mazevault_project](data-sources/project.md)
- [mazevault_secret](data-sources/secret.md)
- [mazevault_certificate](data-sources/certificate.md)
- [mazevault_project_certificates](data-sources/project_certificates.md)
- [mazevault_project_cas](data-sources/project_cas.md)
- [mazevault_project_certificate_templates](data-sources/project_certificate_templates.md)
- [mazevault_project_csrs](data-sources/project_csrs.md)
- [mazevault_environments](data-sources/environments.md)
- [mazevault_ca_accounts](data-sources/ca_accounts.md)
- [mazevault_users](data-sources/users.md)
- [mazevault_roles](data-sources/roles.md)
- [mazevault_integrations](data-sources/integrations.md)
- [mazevault_audit_logs](data-sources/audit_logs.md)
- [mazevault_rotation_executions](data-sources/rotation_executions.md)
- [mazevault_renewal_queue](data-sources/renewal_queue.md)
- [mazevault_consistency_status](data-sources/consistency_status.md)
- [mazevault_rotation_resources](data-sources/rotation_resources.md)
- [mazevault_rotation_resource_history](data-sources/rotation_resource_history.md)
- [mazevault_project_rotation_configs](data-sources/project_rotation_configs.md)
- [mazevault_hsm_devices](data-sources/hsm_devices.md)

## Known Limitations

- `mazevault_role`: Roles cannot be deleted via API. Destroying this resource removes it from Terraform state only.
- `mazevault_ca`: Manages the **single internal root CA of an organization** (`organization_id`), created via `POST /organizations/:id/internal-ca/initialize`. All parameters are immutable — changing any of them replaces the CA. Destroying the resource soft-deletes the internal CA via `DELETE /organizations/:id/internal-ca`; the request is **rejected with HTTP 409** while active (non-revoked, non-expired) certificates were issued by it. The `ocsp_url` / `crl_url` attributes are read-only (`Computed`) and reflect the URLs published for certificates issued by the CA. The computed `account_id` exposes the internal CA account id — pass it as `organization_ca_account_id` on `mazevault_certificate` to issue from this internal CA. `key_algorithm` supports `RSA` (default) and `ECDSA` — pair it with a compatible `key_size` (RSA: 2048/4096; ECDSA: 256/384/521).
- `mazevault_certificate`: Issues a certificate via `POST /certificates/csr`. When `csr_pem` is omitted the key pair and CSR are generated **server-side** (no external tooling / `terraform init` needed); supply `csr_pem` to use an externally generated key. `project_id` assigns the issued certificate to a project. `organization_ca_account_id` pins the signing CA account (e.g. `mazevault_ca.account_id` for the org internal CA); when omitted, the template's CA — or the active internal CA fallback — is used. Destroying the resource **revokes** the certificate.
- `mazevault_ca_account`: Secret credentials (`api_key`, EAB HMAC key) remain write-only. Non-secret provider policy/configuration belongs in `settings` and is persisted as CA account metadata, for example DigiCert `organization_id`, `server_platform_id`, `dcv_method`, and `certificate_dcv_scope`. Do not place API keys or passwords in `settings`. The `provider_type` argument must use a documented canonical lowercase value with no surrounding whitespace; supported values include external providers, self-hosted providers, and `internal`. Use `mazevault_ca` for normal lifecycle management of the organization's internal root CA.
- `mazevault_deployment`: No API endpoint exists to delete deployments. Destroying removes from state only.
- `mazevault_integration`: The `environment` argument is required and forces resource replacement on change (`RequiresReplace`). The backend `UpdateIntegration` endpoint does not accept environment changes; the resource must be recreated. Configurations created with older provider versions that omit `environment` must be updated before applying. Beyond the typed `azure_*` (Azure DevOps) fields, the resource accepts a generic `config` map (non-secret) and `sensitive_config` map (secret, write-only) plus `interactive_auth_method`, so any integration type (Azure Key Vault, Entra, …) can be configured. `sensitive_config` is not refreshed from the server (write-only).
- `mazevault_rotation_config` **(BREAKING v2.0)**: The `environment`, `rotation_strategy`, `workflow_steps_json`, `scope`, and `grace_period_minutes` arguments have been removed.  Replace `ttl_hours` with `rotation_interval_days`.  Run `terraform state rm` and re-import if upgrading from v1.x.
- `mazevault_rotation_workflow`: The `scope` and `grace_period_minutes` arguments have been removed; the backend does not accept them on rotation configs.
- `mazevault_certificate_rotation_config`: No hard-delete endpoint exists.  Destroying this resource sets `enabled = false` and removes it from Terraform state. Certificate delivery on renewal (e.g. writing the renewed certificate into an Azure Key Vault) is expressed as `post_rotation_actions` entries — for Key Vault use `type = "azure_keyvault"` with `config = { integration_id = <kv integration id>, secret_name = "…" }`.
- `mazevault_entra_rotation_config`: No hard-delete endpoint exists.  Destroying this resource sets `rotation_enabled = false` and removes it from Terraform state. `post_rotation_actions` is the single source of truth for post-rotation delivery — if no blocks are defined, nothing is delivered. Key Vault write-back and Spring refresh are expressed as `post_rotation_actions` entries (`type = "azure_keyvault"` / `"spring_actuator_refresh"`), not via separate top-level fields. The `kv_integration_ids`, `secret_name`, `spring_endpoints`, and `webhook_urls` attributes are **read-only computed mirrors** derived from `post_rotation_actions`.
- `mazevault_project`: The `type` argument (`secret` or `certificate`) is immutable after creation (`RequiresReplace`); changing it would orphan type-specific resources. Only the `name` is updatable in place (via `PUT /projects/:id`). The former `environment` argument has been removed — projects have per-environment settings, not a single scalar environment.
- `mazevault_certificate_template`: This resource manages a **certificate-category Project Template**. It is created at the organization level (resolved from `project_id`) via `POST /organizations/:id/templates` with category `certificates`, so it appears in the frontend **Project Templates** view. MazeVault automatically creates (or reuses) the underlying certificate template and exposes its ID as the computed `certificate_template_id`. `validity_period` (a Go duration such as `8760h`) is converted to whole days for the template; `type` maps to the certificate subtype (e.g. `ssl_tls`). `id` is the project-template ID.
- `mazevault_environment`: The `name` and `slug` are immutable after creation (`RequiresReplace`). Only `is_production` and `incident_auto_escalation` are updatable in place; the backend PATCH rejects a name-only update.
- `mazevault_secret`: Exactly one of `value`, `value_wo`, or `generate` must be set. `value_wo` requires Terraform 1.11+ and is intended for Terraform ephemeral values; pair it with `value_wo_version` and increment the version to resend a new value. Legacy `value` remains supported but is stored in Terraform state as a sensitive value. With `generate`, the value is produced server-side (crypto/rand, DB/URL-safe charset) and is never returned by the create call — in Orchestrator Mode it is offloaded to the external provider and stored as `names_only`. In standard mode the value is retrievable via the read API.
- `mazevault_secret_rotation_target`: The `config_json` argument must be a valid JSON object encoded as a string (use `jsonencode()` in Terraform). Supplying malformed JSON produces a diagnostic error at apply time and aborts the operation. There is no direct GET-by-ID endpoint; Read uses a list scan which is O(n) in the number of targets. Destroying a target that has already been deleted externally is idempotent and succeeds silently.
- `mazevault_sync_rule`: The backend does not expose a dedicated GET-by-ID endpoint for sync rules. Read() performs a list scan against `/api/v1/projects/:projectId/sync-rules` and locates the matching rule by UUID. This is O(n) in the number of rules per project. Destroy and Update use a separate path `/api/v1/projects/sync-rules/:id` that does not carry the project ID in the URL.
