# mazevault_certificate_template

Reusable certificate issuance policy (key type, validity, key usage, ACME
challenge policy) for a project.

## Example Usage

```hcl
resource "mazevault_certificate_template" "web_tls" {
  project_id      = mazevault_project.backend.id
  name            = "Web TLS 1Y"
  type            = "tls_server"
  validity_period = "8760h"
  key_usage       = ["digitalSignature", "keyEncipherment"]

  # ACME challenge policy — only meaningful when the issuing CA account is an
  # ACME provider (Let's Encrypt, ZeroSSL, custom ACME). One of:
  #   auto        wildcard → DNS-01, otherwise HTTP-01 (default)
  #   dns-01      DNS-01 for all names (requires a DNS solver on the CA account)
  #   http-01     HTTP-01 (cannot validate wildcard names)
  #   tls-alpn-01 reserved; not yet supported
  challenge_type = "dns-01"

  # Domain validation method for external commercial CAs that require one
  # (e.g. SSLMarket: email | dns | file). Empty uses the provider/account
  # default. Ignored by internal and ACME issuers.
  # dcv_method = "dns"
}

# Commercial CA (SSLMarket) template: issuing account, product, subscription
# length and the technical contact forwarded with every order.
resource "mazevault_certificate_template" "sslmarket_dv" {
  project_id      = mazevault_project.backend.id
  name            = "Public TLS (SSLMarket)"
  type            = "ssl_tls"
  validity_period = "4800h" # 200 days — public TLS maximum

  ca_account_id = mazevault_ca_account.sslmarket.id
  ca_product_id = "4f1c…"  # id from the synced product catalog
  dcv_method    = "dns"
  order_years   = 2        # paid order length; the CA issues follow-up certificates within it
  tech_contact = {
    firstname = "Jana"
    lastname  = "Nováková"
    email     = "pki@example.cz"
    tel       = "+420.123456789"
  }
}
```

## Argument Reference

| Argument | Required | Description |
|----------|----------|-------------|
| `project_id` | yes | Project that owns the template. Changing forces replacement. |
| `name` | yes | Template name. |
| `type` | yes | Template type (e.g. `tls_server`, `tls_client`, `code_signing`). |
| `validity_period` | yes | Duration string for certificate validity (e.g. `8760h`). |
| `key_usage` | no | List of X.509 key usage extensions. |
| `challenge_type` | no | ACME challenge policy: `auto` (default), `dns-01`, `http-01`, `tls-alpn-01`. Ignored for non-ACME issuers. SSLMarket accounts in ACME mode validate via `http-01` only. |
| `dcv_method` | no | Domain control validation method for external commercial CAs that require it (e.g. SSLMarket: `email`, `dns`, `file`). Empty uses the provider/account default. Ignored by internal and ACME issuers. |
| `ca_account_id` | no | Organization CA account that issues certificates from this template. Required for automatic renewal. |
| `ca_product_id` | no | Product from the CA account catalog (commercial CAs). Determines validation level, wildcard/SAN limits and the maximum subscription length. |
| `order_years` | no | Subscription length in years for multi-year commercial CA orders (SSLMarket, 1–6). Length of the paid order, **not** of the certificate; within the subscription the CA issues follow-up certificates itself and MazeVault adopts them without placing a new order. Empty means 1 year. |
| `tech_contact` | no | Map with the technical contact forwarded with commercial CA orders (`title`, `firstname`, `lastname`, `email`, `tel`). Requests may override it; empty means the CA account default. |

## Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `id` | Unique identifier of the certificate template. |

## Import

Most `mazevault_certificate_template` resources can be imported by UUID:

```
terraform import mazevault_certificate_template.example <uuid>
```
