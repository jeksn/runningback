package main

import (
	"strings"
	"testing"
)

func TestFrameRendersWithoutPanic(t *testing.T) {
	procs := []Proc{
		{PID: 2588, Project: "satsuma", Dir: "/Users/jeksn/Herd/satsuma", Kind: "dev", Label: "node vite.js dev",
			Cwd: "/Users/jeksn/Herd/satsuma", Ports: []int{5173}, URL: "http://localhost:5173"},
		{PID: 68501, Project: "satsuma", Dir: "/Users/jeksn/Herd/satsuma", Kind: "dev", Label: "php85 server.php",
			Cwd: "/Users/jeksn/Herd/satsuma/public", Ports: []int{8000}, URL: "http://localhost:8000"},
		{PID: 30613, Project: "mysqld", Kind: "service", Ports: []int{3306, 33060}},
	}
	f := filter{Roots: []string{"/Users/jeksn/Dev", "/Users/jeksn/Herd"}}
	for sel := -1; sel <= len(procs); sel++ {
		out := frame(procs, sel, "", f)
		// Helmet banner on tall terminals, exactly one group header,
		// one row per process.
		if !strings.Contains(out, "⣿") ||
			strings.Count(out, "\x1b[1msatsuma") != 1 ||
			!strings.Contains(out, "vite.js dev") || !strings.Contains(out, "php85") {
			t.Fatalf("frame wrong grouping for sel=%d:\n%s", sel, out)
		}
	}
	empty := frame(nil, 0, "sent SIGTERM", f)
	if !strings.Contains(empty, "nothing running") || !strings.Contains(empty, "SIGTERM") {
		t.Fatalf("unexpected empty frame:\n%s", empty)
	}
}

func TestBreakdown(t *testing.T) {
	procs := []Proc{
		{Kind: "dev", Dir: "/a/satsuma"},
		{Kind: "dev", Dir: "/a/satsuma"},
		{Kind: "dev", Dir: "/a/barnvagn"},
		{Kind: "service"},
		{Kind: "service"},
		{Kind: "service"},
		{Kind: "service"},
	}
	if got := breakdown(procs); got != "2 projects · 4 services" {
		t.Fatalf("breakdown = %q, want %q", got, "2 projects · 4 services")
	}
	if got := breakdown(procs[:2]); got != "1 project · 0 services" && got != "1 project" {
		t.Fatalf("breakdown = %q", got)
	}
	if got := breakdown(nil); got != "nothing running" {
		t.Fatalf("breakdown = %q", got)
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
	f := []string{"/Users/jeksn/Dev"}
	if got := relPath("/Users/jeksn/Dev/work-projects/barnvagn", f); got != "~/Dev/work-projects/barnvagn" {
		t.Fatalf("relPath = %q", got)
	}
}
