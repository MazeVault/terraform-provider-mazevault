# mazevault_hsm_device

Manages an organization-level HSM/KMS device for centralized key protection.

Device connection configuration is encrypted at rest with the system master key.

## Example Usage

```hcl
resource "mazevault_hsm_device" "azure_kv_hsm" {
  organization_id = var.org_id
  name            = "Azure Managed HSM"
  provider_type   = "azure"
  priority        = 100

  environments = ["production"]

  config = {
    vault_url   = "https://example-mhsm.managedhsm.azure.net/"
    auth_method = "managed_identity"
  }
}
```

HashiCorp Vault Transit (keys generated and held by Vault; signing is remote):

```hcl
resource "mazevault_hsm_device" "vault_transit" {
  organization_id = var.org_id
  name            = "Vault Transit"
  provider_type   = "hashicorp_vault"
  priority        = 100

  environments = ["production"]

  config = {
    vault_url  = "https://vault.example.net:8200"
    token      = var.vault_token
    mount_path = "transit" # optional, defaults to "transit"
  }
}
```

## Argument Reference

- `organization_id` - (Required, ForceNew) ID of the organization that owns the device.
- `name` - (Required) Display name of the HSM/KMS device.
- `provider_type` - (Required) HSM/KMS provider type: `azure_managed_hsm`, `aws_cloudhsm`, `gcp_kms`, `pkcs11`, or `hashicorp_vault`.
- `config` - (Required, Sensitive) Provider-specific connection configuration map. Encrypted at rest.
- `description` - (Optional, Computed) Free-text description.
- `environments` - (Optional, Computed) Environments this device applies to.
- `applies_to` - (Optional, Computed) Resource kinds this device applies to.
- `priority` - (Optional, Computed) Routing priority; lower numbers are preferred. Defaults to `100`.

## Attribute Reference

In addition to all arguments above, the following computed attributes are exported:

- `id` - Unique identifier of the HSM device.
- `status` - Operational status of the device.
- `health_status` - Latest health-check status.
- `is_kek_device` - Whether this device holds the key-encryption key (KEK).

## Import

HSM devices can be imported by UUID:

```
terraform import mazevault_hsm_device.example <uuid>
```
