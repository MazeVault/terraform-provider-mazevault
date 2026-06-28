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
  required_providers {
    mazevault = {
      source  = "MazeVault/mazevault"
      version = "~> 1.0"
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
  environment     = "production"
}

resource "mazevault_secret" "db_password" {
  project_id  = mazevault_project.app.id
  key         = "DB_PASSWORD"
  value       = var.db_password
  environment = "production"
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

## Known Limitations

- `mazevault_role`: Roles cannot be deleted via API. Destroying this resource removes it from Terraform state only.
- `mazevault_ca`: Project CAs cannot be deleted once initialized. Destroying removes from state only.
- `mazevault_ca` `ocsp_url` / `crl_url`: These attributes are `Optional + Computed`. When set, the specified URLs are embedded in the Authority Information Access (AIA) and CRL Distribution Point (CDP) extensions of **all certificates issued by that CA** from that point forward. Changing them does not re-issue existing certificates; use the `POST /admin/certificates/backfill-aia` endpoint to retroactively update existing cert metadata.
- `mazevault_deployment`: No API endpoint exists to delete deployments. Destroying removes from state only.
- `mazevault_integration`: The `environment` argument is required and forces resource replacement on change (`RequiresReplace`). The backend `UpdateIntegration` endpoint does not accept environment changes; the resource must be recreated. Configurations created with older provider versions that omit `environment` must be updated before applying.
- `mazevault_rotation_config` **(BREAKING v2.0)**: The `environment`, `rotation_strategy`, `workflow_steps_json`, `scope`, and `grace_period_minutes` arguments have been removed.  Replace `ttl_hours` with `rotation_interval_days`.  Run `terraform state rm` and re-import if upgrading from v1.x.
- `mazevault_rotation_workflow`: The `scope` and `grace_period_minutes` arguments have been removed; the backend does not accept them on rotation configs.
- `mazevault_certificate_rotation_config`: No hard-delete endpoint exists.  Destroying this resource sets `enabled = false` and removes it from Terraform state.
- `mazevault_entra_rotation_config`: No hard-delete endpoint exists.  Destroying this resource sets `rotation_enabled = false` and removes it from Terraform state.
- `mazevault_secret_rotation_target`: The `config_json` argument must be a valid JSON object encoded as a string (use `jsonencode()` in Terraform). Supplying malformed JSON produces a diagnostic error at apply time and aborts the operation. There is no direct GET-by-ID endpoint; Read uses a list scan which is O(n) in the number of targets. Destroying a target that has already been deleted externally is idempotent and succeeds silently.
- `mazevault_sync_rule`: The backend does not expose a dedicated GET-by-ID endpoint for sync rules. Read() performs a list scan against `/api/v1/projects/:projectId/sync-rules` and locates the matching rule by UUID. This is O(n) in the number of rules per project. Destroy and Update use a separate path `/api/v1/projects/sync-rules/:id` that does not carry the project ID in the URL.
