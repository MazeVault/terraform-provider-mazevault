# mazevault_certificate

Issues a certificate through MazeVault via a CSR submission.

The certificate is signed by the organization's internal CA or an external CA
account, depending on how issuance is resolved (see below). When `csr_pem` is
omitted, MazeVault generates the key pair and CSR **server-side** — no external
key/CSR tooling is required. Destroying this resource **revokes** the
certificate.

## Example Usage

### Issue from the organization internal CA (server-side key generation)

```hcl
resource "mazevault_ca" "org_root" {
  organization_id = var.org_id
  name            = "Example Internal CA"
  valid_years     = 10
  key_size        = 4096
}

resource "mazevault_certificate_template" "web_tls" {
  project_id      = mazevault_project.pki.id
  name            = "Web TLS"
  type            = "ssl_tls"
  validity_period = "8760h"
  key_usage       = ["digitalSignature", "keyEncipherment"]
}

resource "mazevault_certificate" "app_tls" {
  common_name                = "app.internal.example.com"
  project_id                 = mazevault_project.pki.id
  template_id                = mazevault_certificate_template.web_tls.certificate_template_id
  organization_ca_account_id = mazevault_ca.org_root.account_id
  key_size                   = 2048
}
```

### Issue from an externally generated key (upload a CSR)

```hcl
resource "mazevault_certificate" "external_key" {
  common_name = "api.example.com"
  template_id = mazevault_certificate_template.web_tls.certificate_template_id
  csr_pem     = file("${path.module}/api.example.com.csr")
}
```

## Argument Reference

- `common_name` - (Required, ForceNew) Common Name (CN) for the certificate.
- `project_id` - (Optional, Computed, ForceNew) ID of the project to assign the issued certificate to.
- `template_id` - (Optional, ForceNew) ID of the certificate template to apply. Use the **certificate** template id (`mazevault_certificate_template.<name>.certificate_template_id`).
- `csr_pem` - (Optional, ForceNew) PEM-encoded Certificate Signing Request. When omitted, MazeVault generates the key pair and CSR server-side.
- `ttl` - (Optional) Requested certificate TTL (e.g. `"8760h"`). May be constrained by the template.
- `key_size` - (Optional) Key size in bits for server-side key generation (used when `csr_pem` is omitted).
- `organization_ca_account_id` - (Optional, Computed) ID of the organization CA account to sign with, overriding the template's CA resolution. Set to `mazevault_ca.<name>.account_id` to issue from the organization internal CA, or to an external `mazevault_ca_account` id. When omitted, the template's CA — or the active internal CA fallback — is used.
- `dcv_method` - (Optional, ForceNew) Domain validation method for commercial CAs (`email`, `dns`, `file`). Overrides the template; empty inherits the template, then the CA account.
- `ca_product_id` - (Optional, ForceNew) CA account product to order (commercial CAs). Overrides the template product.
- `order_years` - (Optional, ForceNew) Subscription length in years for multi-year commercial CA orders (SSLMarket, 1–6). Length of the paid order, not of the certificate.
- `requested_validity_days` - (Optional, ForceNew) Requested certificate validity in days for CAs that honour it.
- `tech_contact` - (Optional, ForceNew) Map with the technical contact forwarded with the order (`title`, `firstname`, `lastname`, `email`, `tel`). Overrides the template contact.
- `voucher` - (Optional, ForceNew, Sensitive) Voucher code applied to the commercial CA order. Inherited from the CA account when empty.

Settings are layered: the request wins, the template fills in what the request leaves empty, and the CA account supplies the business defaults (owner, authorised contact, invoice). The resolved settings are validated against the account's product catalog **before** anything is ordered (unsupported validation method, too many SANs, wildcard on a non-wildcard product, subscription longer than the product allows), so a request the CA would reject fails at `terraform apply` — never as a paid order.

### Issue through SSLMarket with per-request overrides

```hcl
resource "mazevault_certificate" "shop_tls" {
  common_name = "shop.example.cz"
  project_id  = mazevault_project.pki.id
  template_id = mazevault_certificate_template.sslmarket_dv.certificate_template_id

  dcv_method  = "dns"   # MazeVault places the _dnsauth TXT record itself or opens a ticket with it
  order_years = 3
  tech_contact = {
    firstname = "Petr"
    lastname  = "Svoboda"
    email     = "petr.svoboda@example.cz"
  }
}
```

Commercial CA orders are asynchronous: `terraform apply` returns once the CA accepted the order and the resource points at the **pending** certificate (`status = "pending"`, empty `certificate_pem`). MazeVault validates the domain and downloads the certificate when the CA issues it; read the material later with `data "mazevault_certificate"`.

### CA resolution order

The signing CA is resolved in this order:

1. `organization_ca_account_id` (explicit override), if set.
2. The CA account configured on the certificate template, if any.
3. Fallback to the organization's active internal CA.

## Attribute Reference

In addition to all arguments above, the following computed attributes are exported:

- `id` - Unique identifier of the issued certificate.
- `serial_number` - X.509 serial number of the issued certificate.
- `certificate_pem` - PEM-encoded certificate chain.
- `private_key_pem` - (Sensitive) PEM-encoded private key. Only present when MazeVault generates the key pair server-side; `null` otherwise.
- `status` - Certificate status (`active`, `revoked`, `expired`).
- `expires_at` - ISO-8601 expiry timestamp.

## Import

Certificates can be imported by UUID:

```
terraform import mazevault_certificate.example <uuid>
```
