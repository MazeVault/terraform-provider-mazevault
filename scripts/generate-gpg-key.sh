#!/usr/bin/env bash
# generate-gpg-key.sh
#
# Generates an RSA-4096 GPG key for signing Terraform Provider releases.
# The public key fingerprint must be added to:
#   https://registry.terraform.io/settings/gpg-keys
#
# The private key (exported below) must be added as GitHub secret GPG_PRIVATE_KEY
# in the standalone terraform-provider-mazevault repository.
#
# Usage:
#   bash terraform-provider-mazevault/scripts/generate-gpg-key.sh

set -euo pipefail

EMAIL="${1:-team@mazevault.com}"
NAME="${2:-MazeVault Team}"

echo "==> Generating RSA-4096 GPG key for: $NAME <$EMAIL>"

# Generate key non-interactively (will prompt for passphrase once)
gpg --batch --gen-key <<EOF
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: $NAME
Name-Email: $EMAIL
Expire-Date: 3y
%ask-passphrase
%commit
EOF

echo ""
echo "==> Key generated. Exporting public key (add this to registry.terraform.io/settings/gpg-keys):"
gpg --armor --export "$EMAIL"

echo ""
echo "==> To export private key for GitHub secret GPG_PRIVATE_KEY:"
echo "    gpg --armor --export-secret-keys $EMAIL > gpg-private-key.asc"
echo "    gh secret set GPG_PRIVATE_KEY --repo MazeVault/terraform-provider-mazevault < gpg-private-key.asc"
echo "    shred -u gpg-private-key.asc   # delete the file securely after upload"

echo ""
echo "==> Key fingerprints:"
gpg --fingerprint "$EMAIL"
