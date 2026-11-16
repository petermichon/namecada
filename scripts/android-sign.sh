#!/usr/bin/env bash
# Sign an Android release APK with the Namecada keystore.
#
# One-time keystore (never commit it):
#   mkdir -p ~/.config/namecada
#   keytool -genkeypair -keystore ~/.config/namecada/release.jks \
#     -alias namecada -keyalg RSA -keysize 4096 -validity 10000
#
# Then:
#   npx tauri android build --apk
#   npm run android:sign
#
# Passwords: NAMECADA_KEYSTORE_PASS / NAMECADA_KEY_PASS, or apksigner prompts.
# NAMECADA_KEYSTORE and NAMECADA_KEY_ALIAS override the defaults.
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=android-env.sh
. "$here/android-env.sh"

keystore="${NAMECADA_KEYSTORE:-$HOME/.config/namecada/release.jks}"
key_alias="${NAMECADA_KEY_ALIAS:-namecada}"
apk="${1:-}"
if [ -z "$apk" ]; then
	apk="$(ls -t src-tauri/gen/android/app/build/outputs/apk/*/*/*-unsigned.apk 2>/dev/null | head -1 || true)"
fi
if [ -z "$apk" ] || [ ! -f "$apk" ]; then
	echo "usage: npm run android:sign [path/to/app-release-unsigned.apk]" >&2
	exit 1
fi
if [ ! -f "$keystore" ]; then
	echo "keystore not found: $keystore (see docs/android.md)" >&2
	exit 1
fi

apksigner="$(ls -1 "$ANDROID_HOME"/build-tools/*/apksigner 2>/dev/null | sort -V | tail -1)"
if [ -z "$apksigner" ] || [ ! -x "$apksigner" ]; then
	echo "apksigner missing: install build-tools via sdkmanager" >&2
	exit 1
fi

signed="${apk%.apk}-signed.apk"
args=(sign --ks "$keystore" --ks-key-alias "$key_alias" --out "$signed")
[ -n "${NAMECADA_KEYSTORE_PASS:-}" ] && args+=(--ks-pass "env:NAMECADA_KEYSTORE_PASS")
[ -n "${NAMECADA_KEY_PASS:-}" ] && args+=(--key-pass "env:NAMECADA_KEY_PASS")

"$apksigner" "${args[@]}" "$apk"
"$apksigner" verify --print-certs "$signed" | head -3
echo "signed: $signed"
