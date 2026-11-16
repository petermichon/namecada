# Android setup

The phone app is the same Tauri shell and Rust engine as the desktop build, so
only the platform toolchain is new. Pinned versions match Tauri 2.12.5's
template: `compileSdk 36`, AGP 9.3.1, Gradle 9.6.1.

## SDK, NDK and Rust target

```sh
mkdir -p "$HOME/Android/Sdk/cmdline-tools"
curl -fL https://dl.google.com/android/repository/commandlinetools-linux-13114758_latest.zip -o /tmp/cmdline-tools.zip
unzip -q /tmp/cmdline-tools.zip -d "$HOME/Android/Sdk/cmdline-tools"
mv "$HOME/Android/Sdk/cmdline-tools/cmdline-tools" "$HOME/Android/Sdk/cmdline-tools/latest"
rm /tmp/cmdline-tools.zip
export ANDROID_HOME="$HOME/Android/Sdk"
export PATH="$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools:$PATH"
yes | sdkmanager --licenses
sdkmanager "platform-tools" "platforms;android-36" "build-tools;36.0.0" "ndk;30.0.16248370"
rustup target add aarch64-linux-android
cargo install cargo-ndk
```

Only `aarch64-linux-android` is needed for a phone; add
`x86_64-linux-android` only for an emulator.

## Environment

For the shell that runs Tauri:

```sh
source scripts/android-env.sh
```

That sets `ANDROID_HOME`, `NDK_HOME` (from the installed NDK) and the SDK
`PATH` entries. To make it permanent, put the same exports in
`~/.config/environment.d/android.conf` (Fedora/KDE) or your shell profile.

## Check the install

Gradle may pull extra SDK pieces on first build (it installed platform 37 even
though the setup pins 36); let it, and re-run `sdkmanager --list_installed` to
see the final set.

```sh
npx tauri info                # sanity-check the desktop environment
sdkmanager --list_installed   # the authoritative SDK/NDK set
cargo ndk --version
rustup target list --installed | grep android
adb devices
```

`tauri info` does not probe Java. The template needs a **JDK** (with `javac`),
not a JRE: Fedora's `java-25-openjdk` is runtime-only, so Gradle fails with
`does not provide the required capabilities: [JAVA_COMPILER]`. Fedora 44 ships
no JDK 21 or 17, so install the matching devel package:

```sh
sudo dnf install -y java-25-openjdk-devel
```

`scripts/android-env.sh` picks the first JDK under `/usr/lib/jvm` that has
`javac` (25, then 21, then 17). After a JDK change, stop the old Gradle daemon
(`src-tauri/gen/android/gradlew --stop`) so builds do not reuse it.

## Build and run

```sh
npx tauri android init      # once, generates src-tauri/gen/android
npx tauri android dev       # USB device with debugging enabled
npx tauri android build     # release artifacts
```

`devUrl` is `http://localhost:5221`. On a physical device the CLI picks the
host's default-route address (a VPN interface on this machine), which the phone
cannot reach, and it only wires `adb reverse` for emulators. Use loopback plus a
reverse mapping instead:

```sh
source scripts/android-env.sh
adb reverse tcp:5221 tcp:5221
npx tauri android dev --host 127.0.0.1
```

`--host` replaces the `devUrl` host and exports `TAURI_DEV_HOST`, which
`web/vite.config.ts` binds. Re-add the reverse rule after every replug.

A hostname proxy such as `devhost` (`*.localhost:7357`, one tunnel for every
project) does not fit: the CLI rejects a name (`invalid IP address syntax` for
`--host`) and always rewrites the `devUrl` host, so the phone cannot load
`namecada.localhost:7357`. A bare `127.0.0.1` is the only host that works.

A full scan — 107 checked sources at 16 in flight — reaches completion in about
three seconds on the A015 (Android 14, WebView 148), so the phone keeps the whole
source set; the 30 s deadline only exists for stalled hosts.

## Keystore

Debug builds use an auto-generated debug keystore. A release keystore must stay
out of the repo; `src-tauri/gen/android` is generated and its commit policy is
still open in `TODO.md`.
