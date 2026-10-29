package main

import (
	"os"
	"reflect"
	"testing"
)

func TestCollapseWorkers(t *testing.T) {
	procs := []Proc{
		{PID: 2588, Project: "barnvagn", Kind: "dev", Ports: []int{4322}},
		{PID: 16321, Project: "nginx", Kind: "service", Command: "nginx: worker process", Ports: []int{60, 80, 443}},
		{PID: 16322, Project: "nginx", Kind: "service", Command: "nginx: worker process", Ports: []int{60, 80, 443}},
		{PID: 16323, Project: "nginx", Kind: "service", Command: "nginx: worker process", Ports: []int{60, 80, 443}},
		{PID: 30613, Project: "mysqld", Kind: "service", Command: "mysqld --port=3306", Ports: []int{3306}},
	}
	got := collapseWorkers(procs)
	want := []Proc{
		{PID: 2588, Project: "barnvagn", Kind: "dev", Ports: []int{4322}},
		{PID: 16321, Pids: []int{16321, 16322, 16323}, Project: "nginx", Kind: "service", Command: "nginx: worker process", Ports: []int{60, 80, 443}},
		{PID: 30613, Pids: []int{30613}, Project: "mysqld", Kind: "service", Command: "mysqld --port=3306", Ports: []int{3306}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collapseWorkers = %+v\nwant %+v", got, want)
	}
}

func TestProjectRoot(t *testing.T) {
	root := t.TempDir()

	// Git project: a server started from a subdirectory resolves to the checkout.
	proj := root + "/satsuma"
	mkdirAll(t, proj+"/.git")
	mkdirAll(t, proj+"/public")
	if got := projectRoot(proj+"/public", []string{root}); got != proj {
		t.Fatalf("projectRoot = %q, want %q", got, proj)
	}

	// Non-git project: the directory directly under the root wins.
	plain := root + "/plain"
	mkdirAll(t, plain+"/nested")
	if got := projectRoot(plain+"/nested", []string{root}); got != plain {
		t.Fatalf("projectRoot = %q, want %q", got, plain)
	}
}

func TestProjectName(t *testing.T) {
	p := Proc{Kind: "dev", Dir: "/Users/jeksn/Herd/satsuma", Cwd: "/Users/jeksn/Herd/satsuma/public"}
	if got := projectName(&p); got != "satsuma" {
		t.Fatalf("projectName = %q, want satsuma", got)
	}
	p = Proc{Kind: "service", Command: "nginx: worker process"}
	if got := projectName(&p); got != "nginx" {
		t.Fatalf("projectName = %q, want nginx", got)
	}
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}
