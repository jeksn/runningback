package main

import (
	"strings"
	"testing"
)

func TestFrameRendersWithoutPanic(t *testing.T) {
	procs := []Proc{
		{PID: 2588, Project: "barnvagn", Kind: "dev", Label: "node astro.mjs dev",
			Cwd: "/Users/jeksn/Dev/work-projects/barnvagn", Ports: []int{4322}, URL: "http://localhost:4322"},
		{PID: 30613, Project: "mysqld", Kind: "service", Ports: []int{3306, 33060}},
	}
	f := filter{Roots: []string{"/Users/jeksn/Dev"}}
	for sel := -1; sel <= len(procs); sel++ {
		out := frame(procs, sel, "", f)
		if !strings.Contains(out, "barnvagn") || !strings.Contains(out, "mysqld") {
			t.Fatalf("frame missing projects for sel=%d:\n%s", sel, out)
		}
	}
	empty := frame(nil, 0, "sent SIGTERM", f)
	if !strings.Contains(empty, "0 running") || !strings.Contains(empty, "SIGTERM") {
		t.Fatalf("unexpected empty frame:\n%s", empty)
	}
}

func TestLabel(t *testing.T) {
	got := label("/opt/homebrew/bin/node /Users/jeksn/Dev/x/node_modules/astro/bin/astro.mjs dev")
	want := "node astro.mjs dev"
	if got != want {
		t.Fatalf("label = %q, want %q", got, want)
	}
	if got := label("mysqld --port=3306"); got != "mysqld" {
		t.Fatalf("label = %q, want %q", got, "mysqld")
	}
}

func TestRelPath(t *testing.T) {
	p := Proc{Cwd: "/Users/jeksn/Dev/work-projects/barnvagn"}
	f := filter{Roots: []string{"/Users/jeksn/Dev"}}
	if got := relPath(p, f.Roots); got != "~/Dev/work-projects/barnvagn" {
		t.Fatalf("relPath = %q", got)
	}
}
