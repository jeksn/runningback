package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"
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
  runningback kill PID|PORT  send SIGTERM to a process or port

Flags (list / tui):
  --all          also include system processes
  --root DIR     projects root, repeatable (default $RB_ROOTS or ~/Dev and ~/Herd)

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
		return fmt.Errorf("usage: runningback kill PID|PORT")
	}
	n, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("not a PID or port: %s", args[0])
	}
	// If something is listening on that port, kill it - the port is
	// what users see in the list. Otherwise treat the argument as a PID.
	f := filter{All: true, Roots: defaultRoots()}
	if procs, err := detect(f); err == nil {
		if pids := killTargets(procs, n); len(pids) > 0 {
			for _, pid := range pids {
				if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
					return err
				}
			}
			fmt.Printf("sent SIGTERM to port %d\n", n)
			return nil
		}
	}
	if err := syscall.Kill(n, syscall.SIGTERM); err != nil {
		return err
	}
	fmt.Printf("sent SIGTERM to pid %d\n", n)
	return nil
}

// killTargets collects every PID listening on the given port.
func killTargets(procs []Proc, port int) []int {
	var pids []int
	for _, p := range procs {
		if slices.Contains(p.Ports, port) {
			pids = append(pids, p.Pids...)
			if len(p.Pids) == 0 {
				pids = append(pids, p.PID)
			}
		}
	}
	return pids
}
