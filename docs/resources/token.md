# mazevault_token

External API token (e.g. Jira, Signi) managed by MazeVault with a mandatory expiry date.
MazeVault stores the value encrypted (or offloaded in Orchestrator Mode), tracks the expiry, and
raises an incident before it expires so the token can be renewed manually in the external system.

Requires a **PoC or Enterprise** license (feature `token_management`).

## Example Usage

```hcl
terraform {
  required_version = ">= 1.11"
}

resource "mazevault_token" "jira" {
  name             = "jira-automation-token"
  token_provider   = "jira"
  description      = "CI automation token for Jira Cloud"
  value_wo         = var.jira_token
  value_wo_version = 1
  expires_at       = "2027-01-01T00:00:00Z"
  lead_time_days   = 30
}
```

`value_wo` is available only during apply and is not stored in Terraform plan or state.
Terraform cannot detect changes to a write-only value, so pair it with `value_wo_version`.
Incrementing the version replaces the token resource.

## Argument Reference

- `name` (Required) — Display name, unique within the project scope.
- `value_wo` (Optional, Sensitive, Write-only) — Recommended Terraform 1.11+ value input for
  newly created tokens. Must be paired with `value_wo_version`.
- `value_wo_version` (Optional) — Stored replacement trigger. Required with `value_wo`.
- `value` (Optional, Sensitive, Deprecated) — Legacy value input. Terraform stores it in state;
  changing or removing it forces replacement.
- `expires_at` (Required) — Mandatory RFC-3339 expiration date.
- `token_provider` (Optional) — Free-text provider label (e.g. `jira`, `signi`, `generic`).
- `description` (Optional) — Free-text description.
- `project_id` (Optional) — Project assignment; immutable (forces replacement).
- `lead_time_days` (Optional) — Days before expiry to raise an incident (defaults to 30).

## Attribute Reference

- `id`, `status`, `storage_mode`, `external_provider_type`, `is_expired`, `days_until_expiry`,
  `created_at`, `updated_at` — computed.

## Import

```
terraform import mazevault_token.example <uuid>
```

Import retrieves metadata only; MazeVault never returns the token value from the read endpoint.
An imported resource can manage metadata without configuring `value` or `value_wo`.

Do not transparently migrate an existing resource from legacy `value` to `value_wo`. Both changes
replace the token, and deleting a token also deletes its MazeVault deployment targets. Inventory
and recreate every target in a controlled maintenance window if replacement is intentional.
Changing HCL does not remove plaintext from historical Terraform state versions or backups; rotate
the external token and apply the retention procedure for the configured state backend.
