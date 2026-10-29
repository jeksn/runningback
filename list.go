package main

import (
	"fmt"
	"io"
	"strings"
)

func printList(w io.Writer, procs []Proc, f filter) {
	if len(procs) == 0 {
		fmt.Fprintln(w, "nothing running under "+strings.Join(f.Roots, ", "))
		return
	}
	lastKind, lastDir := "", ""
	for _, p := range procs {
		if p.Kind != lastKind {
			if lastKind != "" {
				fmt.Fprintln(w)
			}
			lastKind = p.Kind
			lastDir = ""
			fmt.Fprintln(w, kindTitle(p.Kind))
		}
		if p.Kind == "dev" {
			if p.Dir != lastDir {
				if lastDir != "" {
					fmt.Fprintln(w)
				}
				lastDir = p.Dir
				fmt.Fprintf(w, "  %-22s %s\n", p.Project, relPath(p.Dir, f.Roots))
			}
			fmt.Fprintf(w, "    %-26s %-30s pid %d\n", trunc(p.Label, 26), detail(p), p.PID)
			continue
		}
		fmt.Fprintf(w, "  %-22s %-30s pid %d\n", p.Project, detail(p), p.PID)
	}
}

func kindTitle(kind string) string {
	switch kind {
	case "dev":
		return "projects"
	case "service":
		return "services"
	default:
		return "system"
	}
}
