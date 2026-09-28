# DCK Effects Lab

A self-contained Go/Ebitengine composition with a generated bitmap font and logo.
The same scene runs on desktop and Android, with no external images or music.

From the module root:

```sh
go run ./examples/effectslab
go run ./examples/effectslab -eco
go run ./examples/effectslab -composed
```

The logo uses `composite.CellWarp`. A single scroller follows a sampled sine,
explicit spline coordinates, or a user-defined curve. Paths change every eight
seconds until one is selected manually. The common `scrolling.New` transport
keeps the message loop continuous, with endpoint extrapolation and a pixel
viewport so partial letters enter and leave smoothly. `kit.Layers` fades the
live content in, and a two-pass
`kit.Pipeline` adds its water reflection followed by a temporary magnifier.
The lens appears from second one through eleven of each fourteen-second cycle.

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

`-composed` runs the embedded
[`composed-effects.json`](../authoring/projects/composed-effects.json) through
the same authoring compiler on desktop and Android. CRT and a water reflection
process one text layer; a timed magnifier then processes the complete scene:

```sh
go run ./examples/effectslab -composed -frames 900 -profile /tmp/dck-composed.json
```

The saved lens window repeats every nine seconds and remains active for seven,
so it returns throughout ordinary playback instead of ending after one cue.

The project declares 3.52 MiB of reusable post-processing RGBA surfaces at
640×360. A Pixel 10a run presented 744 measured intervals with p95 16.737 ms,
maximum 16.956 ms and none above 20 ms. The bounded report recorded 59.97
updates/s, 59.90 draws/s, 32.2 µs mean Update CPU and 1.12 ms mean Draw
submission CPU. One process snapshot showed 205,304 KiB PSS, 107,056 KiB
graphics memory and thermal status 0. These are short-window measurements,
not peak memory or battery consumption.

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

adb shell am start -S -n com.olivierh.dckeffects/.MainActivity \
  --ez composed true --el frames 900 --es profile composed.json
adb shell run-as com.olivierh.dckeffects cat files/composed.json
```

Wait for the fifteen-second run to finish before reading its report. Android
saves the bounded report and then resumes ordinary animation. A normal launcher
start is interactive.
Generated AARs, APKs, Gradle state and `output/` are local build artifacts.

Displays refreshing faster than 60 Hz reuse the last rendered scene between
simulation ticks. The report distinguishes Draw callbacks from `rendered_scenes`.

## Validation

```sh
go test -race ./examples/effectslab/...
go vet ./examples/effectslab/...
```

The test and desktop commands need an available native graphics session.
The progressive path renderer was also inspected and accepted on Pixel 10a.
A 744-interval sample had p95 16.742 ms, maximum 16.878 ms and none above
20 ms; a 1,800-tick profile measured 59.99 updates/s, 59.92 draws/s, 12.9 µs
mean Update CPU and 1.68 ms mean Draw submission CPU. Thermal status was 0.
