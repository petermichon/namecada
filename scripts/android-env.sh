# Source this file, do not execute it: . scripts/android-env.sh
#
# Points the shell at the Android SDK installed by docs/android.md.

export ANDROID_HOME="${ANDROID_HOME:-$HOME/Android/Sdk}"
export PATH="$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools:$PATH"

ndk="$(ls -1 "$ANDROID_HOME/ndk" 2>/dev/null | tail -1)"
if [ -n "$ndk" ]; then
	export NDK_HOME="$ANDROID_HOME/ndk/$ndk"
fi

# Uncomment when AGP rejects the distro JDK.
# export JAVA_HOME=/usr/lib/jvm/java-21-openjdk
