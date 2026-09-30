# Controlled insertion and recycled sprites

Run `go run ./examples/insertionqueue` from the module root. This composition
uses two generated font atlases, with different cell dimensions and opposite
character orders. No production assets are needed.

`scrolling.New` receives an `InsertionConfig`. Each token identifies its font
and explicit pen advance; the transport inserts at the viewport's right edge,
ramps between speeds and pauses on embedded commands. A paused message leaves
the sprite queue running. Font metrics retain source crops and bearings.

`motion.RecycledQueue` owns twelve records and their cached poses. Spacing,
depth step, population growth and wrap are independent parameters. `Recycle`
initializes a returning record; `Project` supplies this composition's sine
bounce, camera and size. `sprites.ImageSlots` owns its prepared selection and
draws the borrowed ball artwork in the chosen layer order. For this camera,
farther depth is drawn first.

The insertion program is finite. A viewport-width tail of blank tokens drains
the visible letters before `Reset` begins another passage. For continuous
message repetition with a specified gap, use `Repeat` on the regular scrolling
transport instead. Reset reuses the origin bank and does not reset other effects.

Capture with `-capture /tmp/dck-insertionqueue -frame 120 -frames 500` for ten
seconds at 50 Hz. Readback happens only during export. Interactive updates use
bounded records, poses and selection storage; drawing does not advance clocks
or create a stage-sized image.
