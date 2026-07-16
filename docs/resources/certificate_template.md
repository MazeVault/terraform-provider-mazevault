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
| `challenge_type` | no | ACME challenge policy: `auto` (default), `dns-01`, `http-01`, `tls-alpn-01`. Ignored for non-ACME issuers. |

## Attribute Reference

| Attribute | Description |
|-----------|-------------|
| `id` | Unique identifier of the certificate template. |

## Import

Most `mazevault_certificate_template` resources can be imported by UUID:

```
terraform import mazevault_certificate_template.example <uuid>
```
