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

### Smallstep step-ca (JWK provisioner)

```hcl
resource "mazevault_ca_account" "smallstep" {
  organization_id = mazevault_organization.acme.id
  name            = "Shared step-ca"
  provider_type   = "smallstep"

  # Connection details are non-secret metadata.
  settings = {
    server_url       = "https://step-ca.pki.svc.cluster.local:9000"
    provisioner_name = "mazevault"
    provisioner_type = "JWK"
    root_fingerprint = var.step_ca_root_fingerprint
    default_ttl      = "24h"
  }

  # Secret material goes to the encrypted credentials map (write-only).
  credentials = {
    provisioner_password = var.step_ca_provisioner_password
  }
}
```

### Active Directory Certificate Services (web enrollment)

```hcl
resource "mazevault_ca_account" "adcs" {
  organization_id = mazevault_organization.acme.id
  name            = "Corp ADCS"
  provider_type   = "adcs"

  settings = {
    enrollment_url  = "https://ca.corp.example.com/certsrv"
    enrollment_mode = "web"
    auth_type       = "basic"
    template        = "WebServer"
    verify_ssl      = "true"
  }

  credentials = {
    username = var.adcs_enroll_user
    password = var.adcs_enroll_password
  }
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
| `credentials` | no | Generic map of provider-specific **secret** material, encrypted at rest and write-only. Required keys by provider: `smallstep` → `provisioner_password` (optionally `provisioner_key_json`); `adcs` → `username`, `password`; `vault` → `token`; `est` → `username`, `password`. Keys set here override the typed credential arguments. |
| `settings` | no | Non-secret provider settings persisted in CA account metadata (**plaintext**), such as DigiCert `organization_id`, `container_id`, `server_platform_id`, `dcv_method`, `certificate_dcv_scope`; Smallstep `server_url`, `provisioner_name`, `provisioner_type`, `root_fingerprint`, `ca_bundle_pem`, `default_ttl`, `insecure_skip_tls`; ADCS `enrollment_url`, `enrollment_mode` (`web`/`wcce`/`dcom`), `auth_type`, `template`, `verify_ssl`, `timeout_seconds`. SSLMarket: `mode` (`order` default, or `acme` for the SSLmarket ACME endpoint with external account binding — HTTP-01 only; requires `acme_product_code`, optional `acme_certificate_validity_days` 1–199 and `acme_organization_id`), `product_code`, `dv_auth_method` (`dns`/`email`/`file`), `order_wait_alert_hours`, `dns_solver_type` + `dns_solver_config` (lets MazeVault place DCV TXT records itself; without them a ticket with the record is opened in the organization ticketing integration), owner/authorised-contact/invoice defaults (`owner_*`, `auth_*`, `invoice_*`), default `tech_*` contact and `voucher`. Keys that look like secrets (`*password*`, `*secret*`, `*token*`, `api_key`, `*private_key*`, `*hmac*`) are **rejected at plan time** — put them in `credentials`. |

## Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `id` | Unique identifier of the CA account. |
| `status` | Current connection status. |

> Credential fields (`api_key`, `email`, `directory_url`, `eab_kid`,
> `eab_hmac_key`, `credentials`) are **write-only**: the backend never returns them, so they are
> preserved from state on read and are not detected as drift.

> `settings` is also preserved from Terraform state on read because the provider
> keeps credentials and non-secret settings separate while the backend response
> does not expose provider metadata as a typed Terraform field.

## Import

```
terraform import mazevault_ca_account.example <organization_id>/<resource_id>
```
