#!/usr/bin/env bash
# extract-to-standalone-repo.sh
#
# Extracts terraform-provider-mazevault from the maze-core monorepo into
# a standalone public GitHub repository ready for Terraform Registry publication.
#
# Prerequisites:
#   - gh CLI installed and authenticated (gh auth login)
#   - GPG key generated (see scripts/generate-gpg-key.sh)
#   - Run from the maze-core repo root
#
# Usage:
#   bash terraform-provider-mazevault/scripts/extract-to-standalone-repo.sh [--dry-run]

set -euo pipefail

DRY_RUN=false
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=true

PROVIDER_PREFIX="maze-core/terraform-provider-mazevault"
TARGET_REPO="MazeVault/terraform-provider-mazevault"
SDK_MODULE_OLD="github.com/MazeVault/maze-core/sdks/go"
SDK_MODULE_NEW="github.com/MazeVault/mazevault-go-sdk"

run() {
  if $DRY_RUN; then
    echo "[DRY-RUN] $*"
  else
    echo "+ $*"
    "$@"
  fi
}

echo "==> Step 1: Split provider subtree into a local branch"
run git subtree split \
  --prefix=terraform-provider-mazevault \
  -b provider-split

echo "==> Step 2: Create standalone GitHub repo (public)"
run gh repo create "$TARGET_REPO" \
  --public \
  --description "Terraform Provider for MazeVault — manage secrets, certificates, and RBAC as code" \
  --homepage "https://registry.terraform.io/providers/MazeVault/mazevault" \
  --clone=false

echo "==> Step 3: Push split branch to new repo"
REMOTE_URL="https://github.com/${TARGET_REPO}.git"
run git remote add provider-standalone "$REMOTE_URL" 2>/dev/null || true
run git push provider-standalone provider-split:main

echo "==> Step 4 (manual): Update go.mod in the new repo"
echo "    Replace: replace ${SDK_MODULE_OLD} => ../../sdks/go"
echo "    Remove the replace directive entirely and either:"
echo "      a) Vendor the SDK into the provider repo: cp -r ../../sdks/go internal/mazevault-sdk"
echo "      b) Publish SDK as ${SDK_MODULE_NEW} and update all imports"

echo "==> Step 5 (manual): Add GitHub Actions secrets to $TARGET_REPO"
echo "    gh secret set GPG_PRIVATE_KEY --repo $TARGET_REPO < gpg-private-key.asc"
echo "    gh secret set PASSPHRASE      --repo $TARGET_REPO"

echo "==> Step 6 (manual): Push initial version tag"
echo "    git tag v1.0.0 && git push $REMOTE_URL v1.0.0"

echo "==> Step 7 (manual): Register on Terraform Registry"
echo "    https://registry.terraform.io/publish/provider"
echo "    Select: MazeVault/terraform-provider-mazevault"

echo ""
echo "Done. Check --dry-run output above and proceed manually for steps 4-7."
