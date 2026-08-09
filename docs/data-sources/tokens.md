# mazevault_tokens

Lists managed external API tokens for the organization, including their expiry status.
Token values are never returned.

## Example Usage

```hcl
data "mazevault_tokens" "all" {}

output "expiring_tokens" {
  value = [for t in data.mazevault_tokens.all.tokens : t.name if t.days_until_expiry <= t.lead_time_days]
}
```

## Attribute Reference

- `tokens` — List of tokens, each with: `id`, `name`, `provider`, `status`, `storage_mode`,
  `expires_at`, `lead_time_days`, `is_expired`, `days_until_expiry`.
