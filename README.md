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

## Configuration

- `RB_ROOTS` — colon-separated list of project roots. Default: `~/Dev` and `~/Herd` (Herd-served Laravel sites).
  Or pass `--root DIR` (repeatable) to `list`/`tui`.
- `RB_BIN` — path to the `runningback` binary, for the menu bar app.

## Layout

    main.go, detect.go, list.go, tui.go   Go CLI + TUI (stdlib only)
    main_test.go                          unit tests
    menubar/                              SwiftUI menu bar app (SwiftPM)
