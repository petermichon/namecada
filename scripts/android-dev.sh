#!/usr/bin/env bash
# Run Namecada on a USB Android device with hot reload.
#
#   npm run android:dev
#   npm run android:dev -- --config '{"build":{...}}'
#
# Extra arguments go to `tauri android dev`.
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=android-env.sh
. "$here/android-env.sh"

if [ -z "${ANDROID_HOME:-}" ] || [ ! -d "$ANDROID_HOME" ]; then
	echo "Android SDK missing: follow docs/android.md first." >&2
	exit 1
fi

# adb aborts a server that loses the 5037 bind race, and each abort becomes a
# desktop crash notification. Keep the dev session from collecting them; adb
# still works, it just no longer dumps core.
ulimit -c 0 2>/dev/null || true

if ! adb devices | awk 'NR > 1 && $2 == "device" { found = 1 } END { exit !found }'; then
	echo "No authorized Android device: plug it in, enable USB debugging, accept the prompt." >&2
	exit 1
fi

# The physical-device CLI never sets this up itself, and the mapping is lost on
# every replug. One tunnel serves every app: devhost routes the bare-IP host it
# sees to the project in its "phone" field.
if ! (exec 3<>/dev/tcp/127.0.0.1/7357) 2>/dev/null; then
	echo "devhost is not listening on 7357: start it with \`dev-proxy\`." >&2
	exit 1
fi
adb reverse tcp:7357 tcp:7357 >/dev/null

cd "$(dirname "$here")"
exec npx tauri android dev --host 127.0.0.1 "$@"
