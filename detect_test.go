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

func TestRole(t *testing.T) {
	cases := map[string]string{
		// Real command lines seen on this machine.
		"/Users/jeksn/Library/Application Support/Herd/bin/php85 -S 127.0.0.1:8000 /Users/jeksn/Herd/satsuma/vendor/laravel/framework/src/Illuminate/Foundation/Console/../resources/server.php": "server",
		"/Users/jeksn/.vite-plus/js_runtime/node/24.22.1/bin/node /Users/jeksn/Herd/satsuma/node_modules/.pnpm/@voidzero-dev+vite-plus-core@1.4.0/node_modules/@voidzero-dev/vite-plus-core/dist/vite/bin/vite.js dev": "vite",
		"node /Users/jeksn/Herd/satsuma/node_modules/.bin/../.pnpm/vite-plus@1.4.0/node_modules/vite-plus/bin/vp dev": "vite",
		"/opt/homebrew/Cellar/node/26.8.2/bin/node /Users/jeksn/Dev/work-projects/barnvagn/node_modules/astro/bin/astro.mjs dev": "astro",
		// Other common patterns.
		"php artisan serve":                                       "server",
		"php artisan reverb:start --port=6001":                   "reverb",
		"/Users/jeksn/Herd/bin/php85 artisan queue:work --once":  "queue",
		"php artisan horizon":                                     "horizon",
		"node server.js":                                          "server",
	}
	for cmd, want := range cases {
		if got := role(cmd); got != want {
			t.Errorf("role(%q) = %q, want %q", cmd, got, want)
		}
	}
}

func TestDefaultRootsDiscovery(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("RB_ROOTS", "")

	// A project folder and a Herd-style folder both hold git checkouts.
	mkdirAll(t, home+"/Dev/runningback/.git")
	mkdirAll(t, home+"/Herd/satsuma/.git")
	// Media dirs and dotdirs are ignored even if they contain one.
	mkdirAll(t, home+"/Movies/cloned-repo/.git")
	mkdirAll(t, home+"/.cache/cloned-repo/.git")
	// A plain folder without checkouts is not a root.
	mkdirAll(t, home+"/Documents/invoices")

	roots := defaultRoots()
	want := []string{home + "/Dev", home + "/Herd"}
	if !reflect.DeepEqual(roots, want) {
		t.Fatalf("defaultRoots = %v, want %v", roots, want)
	}

	// The config file beats discovery and expands ~.
	cfg := home + "/.config/runningback"
	mkdirAll(t, cfg)
	os.WriteFile(cfg+"/roots", []byte("# my roots\n~/Work\n"+home+"/Sites\n"), 0o644)
	roots = defaultRoots()
	if !reflect.DeepEqual(roots, []string{home + "/Work", home + "/Sites"}) {
		t.Fatalf("config roots = %v", roots)
	}

	// The environment beats both.
	t.Setenv("RB_ROOTS", home+"/Custom")
	if roots = defaultRoots(); !reflect.DeepEqual(roots, []string{home + "/Custom"}) {
		t.Fatalf("env roots = %v", roots)
	}
}

func TestKillTargets(t *testing.T) {
	procs := []Proc{
		{PID: 2588, Kind: "dev", Ports: []int{4322}},
		{PID: 16321, Pids: []int{16321, 16322}, Kind: "service", Ports: []int{80, 443}},
	}
	if got := killTargets(procs, 4322); !reflect.DeepEqual(got, []int{2588}) {
		t.Fatalf("killTargets(4322) = %v, want [2588]", got)
	}
	if got := killTargets(procs, 80); !reflect.DeepEqual(got, []int{16321, 16322}) {
		t.Fatalf("killTargets(80) = %v, want [16321 16322]", got)
	}
	if got := killTargets(procs, 9999); got != nil {
		t.Fatalf("killTargets(9999) = %v, want nil", got)
	}
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}
