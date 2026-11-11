# mazevault_certificate

Read a certificate by ID, including its effective issuer chain.

## Example Usage

```hcl
data "mazevault_certificate" "issuing_ca" {
  id = var.issuing_ca_certificate_id
}

# Deploy the certificate with the same issuer chain MazeVault exports and installs.
resource "local_file" "issuing_ca_fullchain" {
  content  = data.mazevault_certificate.issuing_ca.certificate_pem
  filename = "${path.module}/issuing-ca-fullchain.pem"
}

output "issuing_ca_chain_source" {
  value = data.mazevault_certificate.issuing_ca.chain_source # stored | inventory | none
}
```

## Argument Reference

- `id` (Required) — Certificate ID.

## Attribute Reference

- `common_name` — Certificate Common Name (primary user ID of an OpenPGP key).
- `serial_number` — Certificate serial number (decimal); key ID of an OpenPGP key. A serial number is unique per issuer only — use `thumbprint_sha256` to identify a certificate.
- `certificate_pem` — PEM bundle: the certificate followed by its **effective issuer chain**. For an OpenPGP key: the ASCII-armored public key.
- `chain_pem` — Issuer chain only, nearest issuer first and root last; empty when no issuer is known.
- `chain_source` — Provenance of the chain: `stored` (captured at import or issuance), `inventory` (no chain was stored, issuers were resolved from verified CA certificates in the same organization), or `none`.
- `chain_complete` — `true` when the chain ends at a self-signed root certificate.
- `status` — `active`, `revoked` or `expired`.
- `expires_at` — Expiry date (ISO-8601). An OpenPGP key that never expires reports `9999-12-31T00:00:00Z`.
- `certificate_type` — Certificate subtype (for example `ssl_tls`, `client_auth`), or `pgp` for an imported OpenPGP key.
- `thumbprint_sha256` — SHA-256 fingerprint in lowercase hex, the identity MazeVault de-duplicates on. For an OpenPGP key: SHA-256 of its primary public key.

## OpenPGP Keys

OpenPGP keys imported into MazeVault live in the certificate inventory with
`certificate_type = "pgp"`. They have no issuer chain (`chain_source = "none"`), cannot be
issued or renewed through the provider, and are deployed with
[`mazevault_certificate_rotation_target`](../resources/certificate_rotation_target.md) using
`content_mode = "public_only"` or `"key_only"`.

```hcl
data "mazevault_certificate" "partner_key" {
  id = var.partner_openpgp_key_id
}

output "partner_public_key" {
  value = data.mazevault_certificate.partner_key.certificate_pem # -----BEGIN PGP PUBLIC KEY BLOCK-----
}
```

## Chain Semantics

The chain returned here is identical to what `GET /certificates/{id}/export?include_chain=true`
produces and what agents deploy. A chain stored with the certificate is returned unchanged; only
certificates **without** a stored chain are completed from the organization inventory, and a
candidate issuer is accepted solely when it cryptographically verifies the certificate signature.

See the [Terraform Registry](https://registry.terraform.io/providers/MazeVault/mazevault/latest/docs/data-sources/certificate) for the published reference.
