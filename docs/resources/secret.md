# mazevault_secret

Encrypted secret stored in MazeVault with optional rotation and environment scoping.

## Example Usage

```hcl
terraform {
	required_version = ">= 1.11"

	required_providers {
		mazevault = {
			source = "MazeVault/mazevault"
		}
		random = {
			source = "hashicorp/random"
		}
	}
}

ephemeral "random_password" "db" {
	length           = 40
	special          = true
	override_special = "_-+=@#"
}

resource "mazevault_secret" "db_password" {
	project_id       = mazevault_project.app.id
	key              = "POSTGRES_PASSWORD"
	environment      = "production"
	value_wo         = ephemeral.random_password.db.result
	value_wo_version = 1
}
```

Use `value_wo` with Terraform 1.11+ ephemeral values when plaintext must not be
stored in Terraform plan or state. Increment `value_wo_version` when Terraform
should send a new write-only value to MazeVault.

Legacy configurations can still use `value`, but Terraform stores sensitive
arguments in state. Prefer `value_wo` or server-side `generate` for new secrets.

## Argument Reference

- `project_id` (string, required) — Parent MazeVault project ID.
- `key` (string, required) — Secret key/name.
- `environment` (string, required) — Environment scope.
- `value` (string, optional, sensitive) — Legacy plaintext input. Mutually exclusive with `value_wo` and `generate`; stored in Terraform state.
- `value_wo` (string, optional, sensitive, write-only) — Terraform 1.11+ write-only plaintext input. Mutually exclusive with `value` and `generate`; not stored in plan or state.
- `value_wo_version` (number, optional) — Required when `value_wo` is set. Increment to resend/update the write-only secret value.
- `generate` (block, optional) — Server-side value generation. Mutually exclusive with `value` and `value_wo`.
- `ttl_hours` (number, optional) — Time-to-live in hours.
- `metadata` (map of string, optional) — Key-value metadata.
- `rotation` (block, optional) — Rotation configuration.

Computed attributes include `id`, `version`, `status`, and `created_at`.

## Import

Most `mazevault_secret` resources can be imported by UUID:

```
terraform import mazevault_secret.example <uuid>
```
