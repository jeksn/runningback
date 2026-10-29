package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func commandOutput(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

func usage() {
	fmt.Fprint(os.Stderr, `runningback — see what is running in your projects

Usage:
  runningback              interactive TUI (falls back to list when not a terminal)
  runningback tui           interactive TUI
  runningback list          one-shot list
  runningback list --json   machine-readable output (used by the menu bar app)
  runningback kill PID      send SIGTERM to a process

Flags (list / tui):
  --all          also include system processes
  --root DIR     projects root, repeatable (default $RB_ROOTS or ~/Dev)

TUI keys:
  up/down  select   o  open URL   x  kill process
  r  refresh       q  quit
`)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "runningback:", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]
	if len(args) == 0 {
		if isTTY() {
			return cmdTUI(nil)
		}
		return cmdList(nil)
	}
	switch args[0] {
	case "tui":
		return cmdTUI(args[1:])
	case "list":
		return cmdList(args[1:])
	case "kill":
		return cmdKill(args[1:])
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		os.Exit(2)
		return nil
	}
}

type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error { *m = append(*m, s); return nil }

func newFilterFlags(name string) (*flag.FlagSet, *filter) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	f := &filter{}
	fs.BoolVar(&f.All, "all", false, "also include system processes")
	fs.Var((*multiFlag)(&f.Roots), "root", "projects root directory (repeatable; default $RB_ROOTS or ~/Dev)")
	return fs, f
}

func cmdList(args []string) error {
	fs, f := newFilterFlags("list")
	jsonOut := fs.Bool("json", false, "output JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(f.Roots) == 0 {
		f.Roots = defaultRoots()
	}
	procs, err := detect(*f)
	if err != nil {
		return err
	}
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(procs)
	}
	printList(os.Stdout, procs, *f)
	return nil
}

func cmdTUI(args []string) error {
	fs, f := newFilterFlags("tui")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(f.Roots) == 0 {
		f.Roots = defaultRoots()
	}
	return runTUI(*f)
}

func cmdKill(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: runningback kill PID (get PIDs from runningback list)")
	}
	pid, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("not a PID: %s", args[0])
	}
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return err
	}
	fmt.Printf("sent SIGTERM to pid %d\n", pid)
	return nil
}
