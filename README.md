# runningback

See which of your projects has dev servers running, without jumping into
each project. One Go binary does the detection, an interactive TUI shows
it in your terminal, and a small SwiftUI menu bar app (Mole-style) keeps
a live count in the macOS menu bar.

    runningback              interactive TUI
    runningback list         one-shot list
    runningback list --json  machine-readable output (used by the menu bar app)
    runningback kill PID     send SIGTERM to a process

## How it works

`lsof` finds every process listening on a TCP port, a batched `lsof -p ... -d cwd`
looks up each process's working directory, and `ps` gets the full command line.
Anything whose cwd is under a projects root (default `~/Dev` and `~/Herd`) is shown as a
project with its URL; known service binaries (mysqld, postgres, redis-server,
...) are shown as services; everything else is system noise, hidden unless you
pass `--all`.

Dev servers are grouped by project (the nearest git checkout above the process
cwd) and each process gets a role inferred from its command line — `server`
(artisan serve / php -S), `vite`, `reverb`, `queue`, `horizon`, `astro`, `next`,
or the script name as a fallback.

## TUI keys

    up/down or j/k   select
    o or enter       open the selected URL in your browser
    x                kill the selected process (SIGTERM)
    r                refresh
    q                quit

The view also auto-refreshes every 5 seconds.

## Install

    make install        # builds and installs the Go binary (opt/homebrew/bin or ~/.local/bin)
    make install-bar    # also builds and installs RunningBackBar.app into /Applications

The menu bar app shells out to the installed `runningback` binary for
`list --json`. Point it at a specific binary with the `RB_BIN` environment
variable.

## Distribution

`make release` builds distributable artifacts for Apple Silicon macOS:

- `build/runningback-darwin-arm64.tar.gz` — the standalone CLI
- `build/RunningBackBar.zip` — the standalone menu bar app; the Go binary is
  bundled in `Contents/Resources/runningback`, so the app has no external
  dependencies

Ways to hand this to other people, easiest first:

1. **Zip on GitHub Releases.** Push the repo, tag a version (`git tag v0.1.0`),
   and the included workflow (`.github/workflows/release.yml`) builds both
   artifacts and attaches them to the release. Recipients unzip
   `RunningBackBar.zip`, drag it to /Applications, done. For the terminal CLI
   they untar the tarball into `~/.local/bin`.
2. **Homebrew tap.** Add a `runningback` formula (install the tarball binary)
   and a `runningback-bar` cask (install the app zip) in a tap repo, so others
   can `brew install`. Both formulas are small because the artifacts are
   self-contained.
3. **Signed and notarized.** The app is currently unsigned. Anyone who
   downloads it needs to right-click → Open the first time, or run
   `xattr -dr com.apple.quarantine RunningBackBar.app`. With an Apple
   Developer account ($99/yr) you can sign with a Developer ID certificate and
   notarize with `notarytool`, after which it opens normally for everyone.

Building from source needs Go 1.25+ and the Xcode Command Line Tools
(`xcode-select --install`), then `make install install-bar`.

## Configuration

- `RB_ROOTS` — colon-separated list of project roots. Default: `~/Dev` and `~/Herd` (Herd-served Laravel sites).
  Or pass `--root DIR` (repeatable) to `list`/`tui`.
- `RB_BIN` — path to the `runningback` binary, for the menu bar app.

## Layout

    main.go, detect.go, list.go, tui.go   Go CLI + TUI (stdlib only)
    main_test.go                          unit tests
    menubar/                              SwiftUI menu bar app (SwiftPM)
