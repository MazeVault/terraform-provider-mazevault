# mazevault_certificate_rotation_target

Deployment target of a managed certificate: where the certificate (and key) is written on issuance and every renewal, and how the consuming service is reloaded and verified.

Target types: `agent_file`, `agent_keystore`, `agent_truststore`, `agent_container_file`, `kubernetes_secret`, `secret_manager`, `database`, `weblogic`.

## Example Usage

```hcl
# nginx on a host with a MazeVault agent: write PEM + key, reload nginx,
# verify the served certificate on :443 and roll back on mismatch.
resource "mazevault_certificate_rotation_target" "nginx" {
  certificate_id = mazevault_certificate.web.id
  target_type    = "agent_file"
  agent_id       = var.web_agent_id
  content_mode   = "full_chain_with_key"

  target_config_json = jsonencode({
    file = {
      cert_path        = "/etc/nginx/ssl/www.example.com.pem"
      key_path         = "/etc/nginx/ssl/www.example.com.key"
      file_permissions = "0640"
      owner_user       = "root"
      owner_group      = "nginx"
      reload = {
        type   = "nginx_reload"
        verify = { port = 443, sni = "www.example.com" }
      }
    }
  })
}

# Certificate bind-mounted into a container; reload the container with SIGHUP
# only inside the maintenance window (files are written immediately).
resource "mazevault_certificate_rotation_target" "dafos_web" {
  certificate_id = mazevault_certificate.dafos.id
  target_type    = "agent_container_file"
  agent_id       = var.dafos_agent_id

  target_config_json = jsonencode({
    container_file = {
      container      = "dafos-web"
      host_path      = "/opt/dafos/nginx/ssl"
      container_path = "/etc/nginx/ssl"
      file = {
        cert_path = "/opt/dafos/nginx/ssl/site.pem"
        key_path  = "/opt/dafos/nginx/ssl/site.key"
      }
      reload = { type = "docker_signal", container = "dafos-web", signal = "HUP", defer_to_window = true }
    }
  })

  deployment_window_json = jsonencode({
    days = ["Sat", "Sun"], from = "02:00", to = "04:00", time_zone = "Europe/Prague"
  })
}

# Issuer chain into the host's system trust anchors.
resource "mazevault_certificate_rotation_target" "anchors" {
  certificate_id = mazevault_certificate.internal_ca_leaf.id
  target_type    = "agent_truststore"
  agent_id       = var.app_agent_id
  content_mode   = "public_chain"

  target_config_json = jsonencode({
    trust_store = { store_type = "system" }
  })
}
```

## Argument Reference

- `certificate_id` (Required, ForceNew) — UUID of the managed certificate.
- `target_type` (Required, ForceNew) — see the list above.
- `agent_id` (Optional) — required for agent-backed types.
- `integration_id`, `database_integration_id` (Optional) — for secret-manager / Kubernetes / database targets.
- `content_mode` (Optional, default `full_chain_with_key`) — `public_only`, `public_chain`, `full_chain_with_key`, `key_only`, `custom`. For an OpenPGP key only `public_only` and `key_only` are accepted (see below).
- `target_config_json` (Required) — JSON object following the backend `TargetConfiguration` union. Reload strategy types: `none`, `systemd_reload`, `systemd_restart`, `nginx_reload`, `apache_graceful`, `haproxy_reload`, `docker_signal`, `docker_exec`, `docker_restart`, `command`; optional `verify = { host, port, sni, timeout_seconds }`; `defer_to_window = true` postpones the reload to the deployment window.
- `deployment_window_json` (Optional) — `{ days, from, to, time_zone }`; `from`/`to` are `HH:MM`, windows may wrap midnight.
- `enabled`, `sync_on_renewal` (Optional, default `true`), `priority` (Optional, default `0`).

## Attribute Reference

- `id` — UUID of the target.
- `last_sync_status` — outcome of the last deployment (`synced`, `deployed_reload_deferred`, `failed`).

## OpenPGP Keys

A certificate record of type `pgp` (an imported OpenPGP key) is deployed as **one ASCII-armored
block per target**: `content_mode = "public_only"` delivers the public key, `"key_only"` the
private key. The provider default `full_chain_with_key` is refused for such a record, so set
`content_mode` explicitly; use two targets when a host needs both blocks. Supported target
types are `agent_file`, `agent_container_file`, `kubernetes_secret` (`secret_type = "Opaque"`
with an explicit `cert_key` / `key_key` data key) and `secret_manager`. Keystore, trust store,
wallet and WebLogic targets hold X.509 certificates and are rejected, as is `reload.verify`.

```hcl
resource "mazevault_certificate_rotation_target" "partner_public_key" {
  certificate_id = var.partner_openpgp_key_id
  target_type    = "agent_file"
  agent_id       = var.app_agent_id
  content_mode   = "public_only"

  target_config_json = jsonencode({
    file = {
      cert_path        = "/etc/payments/keys/partner-public.asc"
      file_permissions = "0644"
    }
  })
}

resource "mazevault_certificate_rotation_target" "signing_private_key" {
  certificate_id = var.signing_openpgp_key_id
  target_type    = "agent_file"
  agent_id       = var.app_agent_id
  content_mode   = "key_only"

  target_config_json = jsonencode({
    file = {
      key_path         = "/etc/payments/keys/signing-private.asc"
      file_permissions = "0600"
    }
  })
}
```

A passphrase-protected private key is delivered protected; MazeVault never holds the passphrase.
