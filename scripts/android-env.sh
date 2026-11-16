# Source this file, do not execute it: . scripts/android-env.sh
#
# Points the shell at the Android SDK installed by docs/android.md.

export ANDROID_HOME="${ANDROID_HOME:-$HOME/Android/Sdk}"
export PATH="$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools:$PATH"

ndk="$(ls -1 "$ANDROID_HOME/ndk" 2>/dev/null | tail -1)"
if [ -n "$ndk" ]; then
	export NDK_HOME="$ANDROID_HOME/ndk/$ndk"
fi

# Gradle needs a JDK with javac; Fedora's java-25-openjdk is runtime-only, so
# install java-25-openjdk-devel (Fedora 44 ships no JDK 21 or 17).
for jdk in java-25-openjdk java-21-openjdk java-17-openjdk; do
	if [ -x "/usr/lib/jvm/$jdk/bin/javac" ]; then
		export JAVA_HOME="/usr/lib/jvm/$jdk"
		break
	fi
done
