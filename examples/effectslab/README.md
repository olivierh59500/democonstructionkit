# DCK Effects Lab

A self-contained Go/Ebitengine composition with a generated bitmap font and logo.
The same scene runs on desktop and Android, with no external images or music.

From the module root:

```sh
go run ./examples/effectslab
go run ./examples/effectslab -eco
```

The logo uses `composite.CellWarp`. A single scroller follows a sampled sine,
explicit spline coordinates, or a user-defined curve. Paths change every eight
seconds until one is selected manually. `scrolling.Window` and `DrawAt` maintain
continuous text movement. `kit.Layers` fades the live content in, and a two-pass
`kit.Pipeline` adds its water reflection followed by a temporary magnifier.
The lens appears from second four through ten of each fourteen-second cycle.

Press **P** to select a path, **W** for water, **L** for the lens, or **Q** for the
resolution profile. The four bottom areas provide the same controls by mouse or
touch. Quality changes allocate replacement surfaces only when requested.

## Bounded measurements and captures

```sh
go run ./examples/effectslab -frames 900 -profile examples/effectslab/output/high.json
go run ./examples/effectslab -eco -frames 900 -profile examples/effectslab/output/eco.json
go run ./examples/effectslab -frames 360 -capture-frame 330 -capture examples/effectslab/output/lens.png
```

The default profile renders at **640×360**, with one source pixel per reflection
strip. Economy mode renders at **320×180**, with four source pixels per strip.
Both use a fixed 60 Hz simulation clock and the same effect timings. Quality is
fixed during bounded runs to keep the reports comparable. Capture and profiling
should be separate runs: PNG encoding affects allocation totals and frame rate.

JSON records mean/max CPU wall time around `Update` and `Draw` submission, actual
Ebitengine FPS/TPS, measured callback rates, Go memory statistics and logical RGBA
surface sizes. It excludes the first 60 update ticks (or one quarter of short
runs) from timing. These numbers **do not measure GPU execution, power use, total
process memory, physical texture allocation or battery drain**. Rendering causes
no explicit GPU readback; `-capture` reads one frame only.

## Android

Install Java 17, Android SDK platform 36 and an Android NDK, then build:

```sh
./examples/effectslab/build-android.sh
```

Set `JAVA_HOME`, `ANDROID_HOME` and `ANDROID_NDK_HOME` if they are not at the
script's Homebrew defaults. The output is
`examples/effectslab/android/app/build/outputs/apk/debug/app-debug.apk`.
The script only builds; it never selects or installs on a connected device.

After explicitly installing this APK, the following launch examples select a
bounded profile. `frames` is a long integer extra:

```sh
adb shell am force-stop com.olivierh.dckeffects
adb shell am start -n com.olivierh.dckeffects/.MainActivity --el frames 900 --es profile high.json
adb shell run-as com.olivierh.dckeffects cat files/high.json

adb shell am force-stop com.olivierh.dckeffects
adb shell am start -n com.olivierh.dckeffects/.MainActivity --ez eco true --el frames 900 --es profile eco.json
adb shell run-as com.olivierh.dckeffects cat files/eco.json
```

Wait for the fifteen-second run to finish before reading its report. The finished
image remains visible; effect updates stop. A normal launcher start is interactive.
Generated AARs, APKs, Gradle state and `output/` are local build artifacts.

Displays refreshing faster than 60 Hz reuse the last rendered scene between
simulation ticks. The report distinguishes Draw callbacks from `rendered_scenes`.

## Validation

```sh
go test -race ./examples/effectslab/...
go vet ./examples/effectslab/...
```

The test and desktop commands need an available native graphics session.
