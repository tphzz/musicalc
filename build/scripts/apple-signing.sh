#!/usr/bin/env bash
# Shared access to the runner's persistent Apple signing keychain.
unlock_signing_keychain() {
  signing_keychain="$HOME/Library/Keychains/login.keychain-db"
  local password_file="$HOME/.config/developer-id-signing/login-keychain-password"
  [[ -f "$password_file" ]] || { echo "Missing runner signing password file: $password_file" >&2; return 1; }
  [[ "$(stat -f '%u' "$password_file")" == "$(id -u)" && "$(stat -f '%Lp' "$password_file")" == "600" ]] ||
    { echo "Signing password file must be owned by the runner and mode 600" >&2; return 1; }
  local password
  password="$(cat "$password_file")"
  [[ -n "$password" && "$password" != *$'\n'* && "$password" != *$'\r'* ]] ||
    { echo "Signing password file must contain one nonempty line" >&2; return 1; }
  [[ -z "${GITHUB_ACTIONS:-}" ]] || printf '::add-mask::%s\n' "$password"
  security unlock-keychain -p "$password" "$signing_keychain" || return
  # Cover the 90-minute Apple jobs without disabling timeout or sleep locking.
  security set-keychain-settings -lut 7200 "$signing_keychain" || return
  security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k "$password" "$signing_keychain" >/dev/null || return
  unset password
  export signing_keychain
}

select_developer_identity() {
  security find-identity -v -p codesigning "$signing_keychain" |
    python3 -c 'import re,sys
matches=set(re.findall(r"([A-Fa-f0-9]{40})\s+\"Developer ID Application: [^\"]+ \([A-Z0-9]{10}\)\"",sys.stdin.read()))
if len(matches)!=1: sys.exit("Expected exactly one valid Developer ID Application identity")
print(next(iter(matches)))'
}
