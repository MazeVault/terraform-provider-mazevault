# mazevault_hsm_devices

Lists all HSM/KMS devices configured for an organization.

## Example Usage

```hcl
data "mazevault_hsm_devices" "all" {
  organization_id = var.org_id
}

output "hsm_device_names" {
  value = [for d in data.mazevault_hsm_devices.all.devices : d.name]
}
```

## Argument Reference

- `organization_id` - (Required) The organization to list devices for.

## Attribute Reference

- `devices` - List of HSM devices. Each element exports:
  - `id` - Unique identifier of the device.
  - `name` - Display name.
  - `description` - Free-text description.
  - `provider_type` - HSM/KMS provider type (e.g. `aws`, `azure`, `gcp`, `pkcs11`).
  - `priority` - Routing priority.
  - `status` - Operational status.
  - `health_status` - Latest health-check status.
  - `is_kek_device` - Whether this device holds the key-encryption key (KEK).
