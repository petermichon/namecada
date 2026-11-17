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
(`apps/desktop/src-tauri/gen/android/gradlew --stop`) so builds do not reuse it.

## Build and run

Run the Tauri CLI from `apps/desktop`, which is the shell's project root (its
`src-tauri/` sits beside its `package.json`):

```sh
cd apps/desktop
npx tauri android init      # once, generates src-tauri/gen/android
npx tauri android build     # release artifacts
```

For development with hot reload, use the wrapper — it sources the environment,
refuses to start without an authorized device, and re-applies the reverse rule
that a replug drops:

```sh
npm run android:dev          # plain dev server on 5221
npm run android:dev:host     # through devhost, optional
```

Extra arguments go to `tauri android dev`, e.g.
`npm run android:dev -- --config '{"build":{…}}'`.

The phone can hold both builds at once: `bundle.android.debugApplicationIdSuffix`
in `tauri.conf.json` makes debug installs `com.namecada.app.debug`, and the dev
wrapper writes a debug-only `app_name` resource so the launcher shows
**Namecada Dev** while the release stays plain **Namecada** / `com.namecada.app`.
Neither build touches the other's data.

`devUrl` is `http://localhost:5221`. On a physical device the CLI picks the
host's default-route address (a VPN interface on this machine), which the phone
cannot reach, so the wrapper passes `--host 127.0.0.1` — that replaces the
`devUrl` host and exports `TAURI_DEV_HOST`, which `web/vite.config.ts` binds —
and reverses the port.

### Optional: one tunnel for every project

Like `dev:host` on the desktop, `android:dev:host` routes the WebView through
devhost instead of a per-project port: `tauri android dev --host` only accepts
an IP, so the phone loads `127.0.0.1:7357` and devhost maps that bare-IP host to
the project named by its `phone` key. Its hook runs the root `dev:host`, so the
same command also gives you the browser at `namecada.localhost:7357` — Vite on
5220 plus the Rust `serve` on 8080. It reverses `tcp:7357`, so the rule is
shared across projects.

The trade-off is that devhost serves one project at a time on the phone, and
8080 must be free: stop any other `npm run dev`/`dev:host` first.

### adb crash notifications

`android dev` polls `adb shell pidof …` while it runs. When a client resolves
`localhost` to `::1` it misses the IPv4 listener, starts a second server, and
that one aborts when it loses the 5037 bind — which surfaces as a DrKonqi
"fatal error" notification even with no core written. `scripts/android-env.sh`
pins `ADB_SERVER_SOCKET=tcp:127.0.0.1:5037`, and the loop stops.

Separately, platform-tools 37.0.1 aborts inside libadbmdns (Google issue
573133666, fixed in 37.0.2); the same file exports `ADB_MDNS=0` for that. Drop
it once `sdkmanager "platform-tools"` offers 37.0.2.

A full scan — 107 checked sources at 16 in flight — reaches completion in about
three seconds on the A015 (Android 14, WebView 148), so the phone keeps the whole
source set; the 30 s deadline only exists for stalled hosts.

The phone can hold both builds at once: `bundle.android.debugApplicationIdSuffix`
in `tauri.conf.json` makes debug installs `com.namecada.app.debug`, so
`android dev` never collides with the signed release and its data. The release
stays `com.namecada.app`.

## Icons

`apps/desktop/src-tauri/icons/app-icon-manifest.json` is the source: `app-icon.svg` for the
desktop/legacy icon, plus `android-bg.svg` and `android-fg.svg` so Android gets a
real adaptive icon (solid background, seal-only foreground inside the safe zone)
instead of a masked rounded square. Regenerate everything with:

```sh
npx tauri icon apps/desktop/src-tauri/icons/app-icon-manifest.json
```

The Android launcher mipmaps are written into the generated project, so a fresh
`android init` needs one `tauri icon` afterwards.

## Signing

Release APKs are unsigned. Create a keystore once, outside the repository:

```sh
mkdir -p ~/.config/namecada
keytool -genkeypair -keystore ~/.config/namecada/release.jks \
  -alias namecada -keyalg RSA -keysize 4096 -validity 10000
```

Then build and sign:

```sh
(cd apps/desktop && npx tauri android build --apk)
npm run android:sign
```

`android:sign` finds the newest `*-unsigned.apk` under
`apps/desktop/src-tauri/gen/android/`, signs it with
`$HOME/.config/namecada/release.jks`, writes `…-signed.apk` beside it and prints
the certificate. Passwords come from `NAMECADA_KEYSTORE_PASS` / `NAMECADA_KEY_PASS`
(or apksigner prompts); `NAMECADA_KEYSTORE` and `NAMECADA_KEY_ALIAS` override the
paths. `gen/android` stays gitignored, so nothing about the keystore or the
generated project is committed.

## Distribution

Tagged releases are built and signed by `.github/workflows/android.yml`, which
attaches a universal APK and one APK per ABI to the GitHub release for
Obtainium. Stores build from the same source: the Fastlane listing lives in
`fastlane/metadata/android/en-US/`. The tag, key and submission rules — and the
F-Droid recipe — are in [Distribution](distribution.md). The download page at `namecada.com`
lives in `apps/site/`: static files, an nginx image and a compose file for the
edge host.

That compose file follows the host's convention: CI publishes
`ghcr.io/petermichon/namecada-site`, and the opt-in `deploy` profile runs
freshdock, which pulls a new image and recreates the service once its
healthcheck passes. A page change ships by pushing to `main`, with no SSH and
no manual pull.
