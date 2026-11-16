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

if ! adb devices | awk 'NR > 1 && $2 == "device" { found = 1 } END { exit !found }'; then
	echo "No authorized Android device: plug it in, enable USB debugging, accept the prompt." >&2
	exit 1
fi

# The physical-device CLI never sets this up itself, and the mapping is lost on
# every replug.
adb reverse tcp:5221 tcp:5221 >/dev/null

cd "$(dirname "$here")"
exec npx tauri android dev --host 127.0.0.1 "$@"
