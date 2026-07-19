# mazevault_secret

Read a secret value from MazeVault (sensitive output).

This data source returns secret material and Terraform will store the sensitive
result in state. Do not use it for values that must remain completely outside
Terraform state; use `mazevault_secret.value_wo` or server-side `generate`
instead.

## Example Usage

```hcl
data "mazevault_secret" "example" {
  # Refer to the Terraform Registry for required arguments.
}
```

## Argument Reference

See the [Terraform Registry](https://registry.terraform.io/providers/MazeVault/mazevault/latest/docs/data-sources/secret) for the full argument and attribute reference.
