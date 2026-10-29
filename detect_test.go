package main

import (
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

func TestProjectName(t *testing.T) {
	p := Proc{Kind: "dev", Cwd: "/Users/jeksn/Herd/satsuma/public"}
	if got := projectName(&p); got != "satsuma" {
		t.Fatalf("projectName = %q, want satsuma", got)
	}
	p = Proc{Kind: "service", Command: "nginx: worker process"}
	if got := projectName(&p); got != "nginx" {
		t.Fatalf("projectName = %q, want nginx", got)
	}
}
