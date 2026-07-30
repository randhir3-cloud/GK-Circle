#!/usr/bin/env bash
# Run ON the NUC once (interactive): bash scripts/setup-nuc-github-deploy-key.sh
# Creates a passphrase-free deploy key for git@github.com:randhir3-cloud/GK-Circle.git

set -euo pipefail

KEY="$HOME/.ssh/id_ed25519_gkcircle_github"
CONFIG="$HOME/.ssh/config"

echo "== GK Circle NUC GitHub deploy key setup =="

if [[ -f "$KEY" ]]; then
  echo "Key already exists: $KEY"
else
  echo "Generating deploy key (no passphrase)..."
  ssh-keygen -t ed25519 -C "nuc-gkcircle-deploy" -f "$KEY" -N ""
fi

mkdir -p "$HOME/.ssh"
chmod 700 "$HOME/.ssh"
chmod 600 "$KEY"
chmod 644 "${KEY}.pub"

# Ensure github.com uses deploy key only (not your personal id_ed25519 with passphrase)
if grep -q 'Host github.com' "$CONFIG" 2>/dev/null; then
  echo ""
  echo "NOTE: ~/.ssh/config already has a Host github.com block."
  echo "Ensure it contains:"
  echo ""
  cat <<'EOF'
Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_ed25519_gkcircle_github
    IdentitiesOnly yes
EOF
else
  echo "Appending github.com block to ~/.ssh/config ..."
  cat >>"$CONFIG" <<'EOF'

# GK Circle — GitHub deploy (NUC)
Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_ed25519_gkcircle_github
    IdentitiesOnly yes
EOF
  chmod 600 "$CONFIG"
fi

echo ""
echo "=== Add this deploy key to GitHub ==="
echo "Repo: randhir3-cloud/GK-Circle -> Settings -> Deploy keys -> Add deploy key"
echo "Title: nuc-gkcircle-deploy"
echo "Allow write access: OFF (read-only is enough for git pull)"
echo ""
cat "${KEY}.pub"
echo ""
echo "Then on the NUC:"
echo "  cd ~/apps/gkcircle"
echo "  git remote set-url origin git@github.com:randhir3-cloud/GK-Circle.git"
echo "  ssh -T git@github.com"
echo "  git pull"
