# mazevault_ca_account

Connects an **external** Certificate Authority account to a MazeVault organization
(DigiCert, SSLMarket, Sectigo, ACME providers such as Let's Encrypt / ZeroSSL,
SmallStep, Microsoft ADCS, HashiCorp Vault, …).

> The organization's **internal** CA is not managed by this resource. It is
> created and managed in the MazeVault UI under **Organization Settings →
> Project Settings → Certificate Authorities → Internal Certificate Authority**.

## Example Usage

### DigiCert (API key)

```hcl
resource "mazevault_ca_account" "digicert" {
  organization_id = mazevault_organization.acme.id
  name            = "DigiCert Production"
  provider_type   = "digicert"
  api_key         = var.digicert_api_key
  base_url        = "https://www.digicert.com"
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
| `provider_type` | yes | One of `digicert`, `sslmarket`, `sectigo`, `globalsign`, `entrust`, `letsencrypt`, `zerossl`, `acme`, `smallstep`, `adcs`, `vault`. Changing forces replacement. |
| `api_key` | no | API key/credential for API-style providers (DigiCert, SSLMarket, ZeroSSL API). Sensitive, write-only. |
| `base_url` | no | Base URL for the provider API (self-hosted CAs). |
| `email` | no | ACME account contact email. **Required for ACME providers** (`letsencrypt`, `zerossl`, `acme`). |
| `directory_url` | no | ACME directory URL. **Required for custom `acme`**; defaulted for `letsencrypt`/`zerossl`. |
| `eab_kid` | no | External Account Binding key ID for ACME providers that require it (e.g. ZeroSSL). |
| `eab_hmac_key` | no | External Account Binding HMAC key. Sensitive, write-only. |

## Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `id` | Unique identifier of the CA account. |
| `status` | Current connection status. |

> Credential fields (`api_key`, `email`, `directory_url`, `eab_kid`,
> `eab_hmac_key`) are **write-only**: the backend never returns them, so they are
> preserved from state on read and are not detected as drift.

## Import

```
terraform import mazevault_ca_account.example <organization_id>/<resource_id>
```
