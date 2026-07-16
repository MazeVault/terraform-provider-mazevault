# mazevault_ca

Manages the single internal root Certificate Authority (CA) of an organization.

Each organization has **at most one** internal CA. Certificate templates and
issued certificates chain to it. All parameters are immutable — changing any of
them replaces the CA. Destroying the resource **soft-deletes** the internal CA
(via `DELETE /organizations/:id/internal-ca`); the delete is **rejected with
HTTP 409** while active (non-revoked, non-expired) certificates were issued by
it, so revoke or let those certificates expire first.

## Example Usage

```hcl
resource "mazevault_ca" "org_root" {
  organization_id = var.org_id
  name            = "Example Organization Internal CA"
  valid_years     = 10
  key_size        = 4096
}

# Issue a certificate from this internal CA
resource "mazevault_certificate" "svc" {
  common_name                = "svc.internal.example.com"
  project_id                 = mazevault_project.pki.id
  template_id                = mazevault_certificate_template.web_tls.certificate_template_id
  organization_ca_account_id = mazevault_ca.org_root.account_id
}
```

## Argument Reference

- `organization_id` - (Required, ForceNew) ID of the organization to initialize the internal CA in.
- `name` - (Required, ForceNew) Common name / display name of the internal root CA.
- `valid_years` - (Required, ForceNew) Number of years the CA certificate is valid.
- `key_size` - (Required, ForceNew) Key size for the CA. For RSA use `2048` or `4096`; for ECDSA use `256` (P-256), `384` (P-384) or `521` (P-521).
- `key_algorithm` - (Optional, Computed, ForceNew) Key algorithm for the CA: `RSA` (default) or `ECDSA`. Must be paired with a compatible `key_size`.

## Attribute Reference

In addition to all arguments above, the following computed attributes are exported:

- `id` - Unique identifier of the internal CA (the certificate authority).
- `account_id` - ID of the internal CA **account** (`OrganizationCAAccount`). Pass this as `organization_ca_account_id` on `mazevault_certificate` to issue certificates from this internal CA.
- `type` - CA type as returned by the API.
- `status` - Current operational status of the CA.
- `not_after` - ISO-8601 timestamp until which the CA certificate is valid.
- `serial_number` - Serial number of the CA certificate.
- `ocsp_url` - OCSP responder URL published for certificates issued by this CA.
- `crl_url` - CRL Distribution Point (CDP) URL published for certificates issued by this CA.

## Import

The internal CA is a singleton per organization, so it is imported by organization ID:

```
terraform import mazevault_ca.org_root <organization_id>
```

