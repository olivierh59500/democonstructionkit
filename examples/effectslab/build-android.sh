#!/bin/sh
set -eu

# Build only. Installation and device selection remain explicit caller actions.
lab_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
kit_dir=$(CDPATH= cd -- "$lab_dir/../.." && pwd)
sdk_path=${ANDROID_HOME:-${ANDROID_SDK_ROOT:-/opt/homebrew/share/android-commandlinetools}}
jdk_path=${JAVA_HOME:-/opt/homebrew/opt/openjdk@17}
ndk_path=${ANDROID_NDK_HOME:-$sdk_path/ndk/28.2.13676358}
if [ ! -f "$sdk_path/platforms/android-36/android.jar" ]; then
    echo "Set ANDROID_HOME to an Android SDK containing platform 36." >&2
    exit 1
fi
if [ ! -x "$jdk_path/bin/java" ] || [ ! -d "$ndk_path" ]; then
    echo "Set JAVA_HOME to Java 17 and ANDROID_NDK_HOME to an installed NDK." >&2
    exit 1
fi
export JAVA_HOME="$jdk_path" ANDROID_HOME="$sdk_path" ANDROID_SDK_ROOT="$sdk_path" ANDROID_NDK_HOME="$ndk_path"
export PATH="$jdk_path/bin:$PATH"
export CGO_LDFLAGS="${CGO_LDFLAGS:--Wl,-z,max-page-size=16384 -Wl,-z,common-page-size=16384}"
export GOWORK=off
cd "$kit_dir"
mkdir -p "$lab_dir/android/app/libs"
GOOS=android GOARCH=arm64 go list -m -tags=android all >/dev/null
go run github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@v2.9.11 bind \
    -target android/arm64 -androidapi 23 -javapkg com.olivierh.dckeffects \
    -o "$lab_dir/android/app/libs/libdckeffects.aar" ./examples/effectslab/mobile
"$lab_dir/android/gradlew" -p "$lab_dir/android" --console=plain :app:assembleDebug
echo "APK: $lab_dir/android/app/build/outputs/apk/debug/app-debug.apk"
