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
	lastKind := ""
	for _, p := range procs {
		if p.Kind != lastKind {
			if lastKind != "" {
				fmt.Fprintln(w)
			}
			lastKind = p.Kind
			fmt.Fprintln(w, kindTitle(p.Kind))
		}
		printProc(w, p, f, false)
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

func printProc(w io.Writer, p Proc, f filter, selected bool) {
	label := p.Label
	if p.Kind != "dev" {
		label = ""
	}
	fmt.Fprintf(w, "  %-20s %-28s %-36s %s\n",
		p.Project, detail(p), label, "pid "+fmt.Sprint(p.PID))
	if p.Kind == "dev" {
		fmt.Fprintf(w, "  %-20s %s\n", "", relPath(p, f.Roots))
	}
}
