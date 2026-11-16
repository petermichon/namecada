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
npx tauri android build     # release artifacts
```

For development with hot reload, use the wrapper — it sources the environment,
refuses to start without an authorized device, and re-applies the reverse rule
that a replug drops:

```sh
npm run android:dev          # extra args go to `tauri android dev`
```

`devUrl` is `http://localhost:5221` on desktop. On Android the CLI picks the
host's default-route address (a VPN interface on this machine), which the phone
cannot reach, so the wrapper passes `--host 127.0.0.1`, which replaces the
`devUrl` host and exports `TAURI_DEV_HOST`; `web/vite.config.ts` binds it.

### One tunnel for every project

`tauri android dev --host` only accepts an IP, so the phone cannot load a
hostname. `devhost` therefore routes the bare-IP traffic it receives over the
tunnel to the project named by its `phone` key, and
`src-tauri/tauri.android.conf.json` points the Android `devUrl` at that proxy
port and starts the web server only:

```json
{ "build": { "devUrl": "http://namecada.localhost:7357",
             "beforeDevCommand": "npm run dev:host -w @namecada/web" } }
```

`npm run android:dev` then needs `dev-proxy` running (the script checks) and
reverses `tcp:7357` instead of per-project ports. The desktop keeps
`namecada.localhost:7357`; the phone sees `127.0.0.1:7357`, so devhost serves
one project at a time — set `phone` to whichever is being tested.

The CLI also forwards the dev port itself once it starts building; the wrapper
applies the rule up front so the first page load cannot race it.

### adb crash notifications

Two adb bugs surface as fatal-error notifications while `android dev` polls the
device:

- 37.0.1 aborts inside libadbmdns (Google issue 573133666, fixed in 37.0.2);
  `scripts/android-env.sh` exports `ADB_MDNS=0` to avoid it.
- Any version aborts a server that loses the 5037 bind race; `scripts/android-dev.sh`
  sets `ulimit -c 0`, so a lost race cannot raise a desktop crash notification.

Neither affects the device connection. Upgrade with `sdkmanager "platform-tools"`
when 37.0.2 is public, then drop the mDNS workaround.

A full scan — 107 checked sources at 16 in flight — reaches completion in about
three seconds on the A015 (Android 14, WebView 148), so the phone keeps the whole
source set; the 30 s deadline only exists for stalled hosts.

## Keystore

Debug builds use an auto-generated debug keystore. A release keystore must stay
out of the repo; `src-tauri/gen/android` is generated and its commit policy is
still open in `TODO.md`.
