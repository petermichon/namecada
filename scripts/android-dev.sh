#!/usr/bin/env bash
# Run Namecada on a USB Android device with hot reload.
#
#   npm run android:dev            # plain dev server on 5221
#   npm run android:dev:host       # through devhost, one tunnel for all apps
#
# Extra arguments go to `tauri android dev`.
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=android-env.sh
. "$here/android-env.sh"

mode="plain"
if [ "${1:-}" = "--host" ]; then
	mode="host"
	shift
fi

if [ -z "${ANDROID_HOME:-}" ] || [ ! -d "$ANDROID_HOME" ]; then
	echo "Android SDK missing: follow docs/android.md first." >&2
	exit 1
fi

if ! adb devices | awk 'NR > 1 && $2 == "device" { found = 1 } END { exit !found }'; then
	echo "No authorized Android device: plug it in, enable USB debugging, accept the prompt." >&2
	exit 1
fi

port=5221
config=()
if [ "$mode" = "host" ]; then
	port=7357
	if ! (exec 3<>/dev/tcp/127.0.0.1/7357) 2>/dev/null; then
		echo "devhost is not listening on 7357: start it with \`dev-proxy\`." >&2
		exit 1
	fi
	config=(
		--config
		'{"build":{"devUrl":"http://namecada.localhost:7357","beforeDevCommand":"npm run dev:host"}}'
	)
fi

# The physical-device CLI picks the default-route address and loses the mapping
# on every replug, so pin loopback and apply the rule up front.
adb reverse "tcp:$port" "tcp:$port" >/dev/null

cd "$(dirname "$here")"
exec npx tauri android dev --host 127.0.0.1 "${config[@]}" "$@"
