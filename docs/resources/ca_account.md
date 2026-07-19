# mazevault_ca_account

Connects a Certificate Authority account record or integration to a MazeVault
organization. Supported account types include external commercial CAs, ACME
providers such as Let's Encrypt / ZeroSSL, self-hosted providers such as
SmallStep, Microsoft ADCS, HashiCorp Vault, EST, Venafi, and internal CA account
references.

> Use `mazevault_ca` for normal lifecycle management of the organization's
> **internal** root CA. `mazevault_ca_account` can represent CA account records,
> including the internal CA account reference, but it does not initialize or
> configure internal root CA key material.

## Example Usage

### DigiCert (API key)

```hcl
resource "mazevault_ca_account" "digicert" {
  organization_id = mazevault_organization.acme.id
  name            = "DigiCert Production"
  provider_type   = "digicert"
  api_key         = var.digicert_api_key
  base_url        = "https://www.digicert.com"

  settings = {
    organization_id       = var.digicert_organization_id
    server_platform_id    = "45"
    dcv_method            = "dns_txt"
    certificate_dcv_scope = "base_domain"
  }
}
```

### Let's Encrypt (ACME)

```hcl
resource "mazevault_ca_account" "letsencrypt" {
  organization_id = mazevault_organization.acme.id
  name            = "Let's Encrypt"
  provider_type   = "letsencrypt"
  email           = "pki@example.com"
  # directory_url is defaulted for letsencrypt/zerossl; set it for custom ACME.
}
```

### ZeroSSL / custom ACME with External Account Binding (EAB)

```hcl
resource "mazevault_ca_account" "zerossl" {
  organization_id = mazevault_organization.acme.id
  name            = "ZeroSSL"
  provider_type   = "zerossl"
  email           = "pki@example.com"
  eab_kid         = var.zerossl_eab_kid
  eab_hmac_key    = var.zerossl_eab_hmac_key
}

resource "mazevault_ca_account" "custom_acme" {
  organization_id = mazevault_organization.acme.id
  name            = "Private ACME"
  provider_type   = "acme"
  email           = "pki@example.com"
  directory_url   = "https://acme.internal.example.com/directory"
}
```

## Argument Reference

| Argument | Required | Description |
|----------|----------|-------------|
| `organization_id` | yes | Organization that owns the CA account. Changing forces replacement. |
| `name` | yes | Display name for the CA account. |
| `provider_type` | yes | One of `digicert`, `sslmarket`, `rapidssl`, `sectigo`, `globalsign`, `entrust`, `letsencrypt`, `zerossl`, `acme`, `internal`, `adcs`, `vault`, `smallstep`, `est`, `venafi`. Values must be canonical lowercase with no surrounding whitespace. `rapidssl` uses the SSLMarket integration for backward compatibility; prefer `sslmarket` for new configurations. I.CA/PostSignum and PSD2 account creation are not exposed through Terraform until the backend integrations report a connectable status. Changing forces replacement. |
| `api_key` | no | API key/credential for API-style providers (DigiCert, SSLMarket, ZeroSSL API). Sensitive, write-only. |
| `base_url` | no | Base URL for the provider API (self-hosted CAs). |
| `email` | no | ACME account contact email. **Required for ACME providers** (`letsencrypt`, `zerossl`, `acme`). |
| `directory_url` | no | ACME directory URL. **Required for custom `acme`**; defaulted for `letsencrypt`/`zerossl`. |
| `eab_kid` | no | External Account Binding key ID for ACME providers that require it (e.g. ZeroSSL). |
| `eab_hmac_key` | no | External Account Binding HMAC key. Sensitive, write-only. |
| `settings` | no | Non-secret provider settings persisted in CA account metadata, such as DigiCert `organization_id`, `container_id`, `server_platform_id`, `dcv_method`, `certificate_dcv_scope`, or SSLMarket contact defaults. Do **not** put API keys or passwords here. |

## Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `id` | Unique identifier of the CA account. |
| `status` | Current connection status. |

> Credential fields (`api_key`, `email`, `directory_url`, `eab_kid`,
> `eab_hmac_key`) are **write-only**: the backend never returns them, so they are
> preserved from state on read and are not detected as drift.

> `settings` is also preserved from Terraform state on read because the provider
> keeps credentials and non-secret settings separate while the backend response
> does not expose provider metadata as a typed Terraform field.

## Import

```
terraform import mazevault_ca_account.example <organization_id>/<resource_id>
```
