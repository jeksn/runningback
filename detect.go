package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// Proc is a single process that is listening on at least one TCP port.
type Proc struct {
	PID     int    `json:"pid"`
	Pids    []int  `json:"pids,omitempty"` // set when identical workers were collapsed
	Command string `json:"command"`
	Label   string `json:"label"`
	Cwd     string `json:"cwd"`
	Dir     string `json:"projectDir,omitempty"` // project root (git checkout) for dev processes
	Project string `json:"project"`
	Kind    string `json:"kind"` // "dev", "service" or "system"
	Ports   []int  `json:"ports"`
	URL     string `json:"url,omitempty"`
}

// filter controls which kinds of processes are reported and which
// directories count as "my projects".
type filter struct {
	All   bool
	Roots []string
}

var serviceBins = map[string]bool{
	"mysqld": true, "mariadbd": true, "postgres": true, "redis-server": true,
	"mongod": true, "memcached": true, "nginx": true, "caddy": true, "httpd": true,
	"traefik": true, "vault": true,
}

var nonWebPorts = map[int]bool{
	3306: true, 33060: true, 5432: true, 6379: true, 11211: true, 27017: true,
}

func defaultRoots() []string {
	if v := os.Getenv("RB_ROOTS"); v != "" {
		var roots []string
		for _, r := range strings.Split(v, ":") {
			if r = strings.TrimSpace(r); r != "" {
				roots = append(roots, r)
			}
		}
		return roots
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	// Dev holds regular projects, Herd holds Laravel sites served by Herd.
	return []string{filepath.Join(home, "Dev"), filepath.Join(home, "Herd")}
}

// detect returns every process listening on a TCP port, sorted by kind
// (dev first) then project. System processes are only included when
// f.All is set.
func detect(f filter) ([]Proc, error) {
	out, err := commandOutput("lsof", "-nP", "-iTCP", "-sTCP:LISTEN", "-Fpcn")
	if err != nil && len(out) == 0 {
		return []Proc{}, nil
	}

	procs := map[int]*Proc{}
	order := []int{}
	var cur *Proc
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 2 {
			continue
		}
		tag, val := line[0], line[1:]
		switch tag {
		case 'p':
			pid, err := strconv.Atoi(val)
			if err != nil {
				continue
			}
			if p, ok := procs[pid]; ok {
				cur = p
			} else {
				cur = &Proc{PID: pid}
				procs[pid] = cur
				order = append(order, pid)
			}
		case 'n':
			if cur == nil {
				continue
			}
			if i := strings.LastIndex(val, ":"); i >= 0 {
				if port, err := strconv.Atoi(val[i+1:]); err == nil && !slices.Contains(cur.Ports, port) {
					cur.Ports = append(cur.Ports, port)
				}
			}
		}
	}
	if len(order) == 0 {
		return []Proc{}, nil
	}

	pidArgs := make([]string, len(order))
	for i, pid := range order {
		pidArgs[i] = strconv.Itoa(pid)
	}
	pidList := strings.Join(pidArgs, ",")

	// Working directories, one batched lsof call.
	cwdOut, _ := commandOutput("lsof", "-d", "cwd", "-Fn", "-a", "-p", pidList)
	curPid := 0
	for _, line := range strings.Split(string(cwdOut), "\n") {
		if len(line) < 2 {
			continue
		}
		tag, val := line[0], line[1:]
		switch tag {
		case 'p':
			curPid, _ = strconv.Atoi(val)
		case 'n':
			if p, ok := procs[curPid]; ok {
				p.Cwd = val
			}
		}
	}

	// Full command lines, one batched ps call.
	psOut, _ := commandOutput("ps", "-o", "pid=,command=", "-p", pidList)
	for _, line := range strings.Split(string(psOut), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		if p, ok := procs[pid]; ok {
			p.Command = strings.Join(fields[1:], " ")
		}
	}

	result := []Proc{}
	for _, pid := range order {
		p := procs[pid]
		p.Label = label(p.Command)
		p.Kind = classify(p, f.Roots)
		if p.Kind == "dev" {
			p.Dir = projectRoot(p.Cwd, f.Roots)
		}
		p.Project = projectName(p)
		sort.Ints(p.Ports)
		p.URL = url(p)
		if p.Kind == "system" && !f.All {
			continue
		}
		result = append(result, *p)
	}

	result = collapseWorkers(result)

	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		if a.Kind != b.Kind {
			return kindRank(a.Kind) < kindRank(b.Kind)
		}
		if a.Project != b.Project {
			return a.Project < b.Project
		}
		if ka, kb := groupKey(a), groupKey(b); ka != kb {
			return ka < kb
		}
		return a.PID < b.PID
	})
	return result, nil
}

func groupKey(p Proc) string {
	if p.Dir != "" {
		return p.Dir
	}
	return p.Project
}

// projectRoot walks up from a process cwd to the directory of the
// project it belongs to: the closest ancestor containing a .git, or
// the directory directly under one of the roots. This makes a Vite
// server and a php server both resolve to the same project.
func projectRoot(cwd string, roots []string) string {
	if cwd == "" {
		return cwd
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		for _, root := range roots {
			if filepath.Dir(dir) == root {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return cwd
		}
		dir = parent
	}
}

// classify decides whether a process belongs to one of the project
// roots (dev), is a known local service, or is unrelated system noise.
func classify(p *Proc, roots []string) string {
	for _, root := range roots {
		if p.Cwd == root || strings.HasPrefix(p.Cwd, root+"/") {
			return "dev"
		}
	}
	if fields := strings.Fields(p.Command); len(fields) > 0 {
		// Process titles may carry a suffix, e.g. "nginx: worker process".
		if serviceBins[strings.TrimSuffix(filepath.Base(fields[0]), ":")] {
			return "service"
		}
	}
	return "system"
}

func projectName(p *Proc) string {
	if p.Kind == "dev" {
		if p.Dir != "" {
			return filepath.Base(p.Dir)
		}
		if p.Cwd != "" {
			return filepath.Base(p.Cwd)
		}
	}
	if fields := strings.Fields(p.Command); len(fields) > 0 {
		// Process titles may carry a suffix, e.g. "nginx: worker process".
		return strings.TrimSuffix(filepath.Base(fields[0]), ":")
	}
	return "pid " + strconv.Itoa(p.PID)
}

// collapseWorkers merges identical service processes (same binary and
// ports, e.g. a pool of nginx workers) into a single entry that lists
// every PID, so the list shows one row instead of eight.
func collapseWorkers(procs []Proc) []Proc {
	type key struct{ project, ports, cmd string }
	index := map[key]int{}
	out := make([]Proc, 0, len(procs))
	for _, p := range procs {
		if p.Kind == "dev" {
			out = append(out, p)
			continue
		}
		cmd := ""
		if fields := strings.Fields(p.Command); len(fields) > 0 {
			cmd = filepath.Base(fields[0])
		}
		k := key{p.Project, fmt.Sprint(p.Ports), cmd}
		if i, ok := index[k]; ok {
			out[i].Pids = append(out[i].Pids, p.PID)
			if p.PID < out[i].PID {
				out[i].PID = p.PID
				out[i].Command = p.Command
				out[i].Cwd = p.Cwd
				out[i].Label = p.Label
			}
			continue
		}
		p.Pids = []int{p.PID}
		index[k] = len(out)
		out = append(out, p)
	}
	for i := range out {
		sort.Ints(out[i].Pids)
	}
	return out
}

// url returns the address to open in a browser, preferring the first
// port that is not a known database port. Empty for non-dev processes.
func url(p *Proc) string {
	if p.Kind != "dev" {
		return ""
	}
	for _, port := range p.Ports {
		if !nonWebPorts[port] {
			return "http://localhost:" + strconv.Itoa(port)
		}
	}
	return ""
}

// label shortens a full command line into something readable:
// paths become basenames and flags are dropped.
func label(cmd string) string {
	fields := strings.Fields(cmd)
	var parts []string
	for i, f := range fields {
		if i > 0 && strings.HasPrefix(f, "-") {
			continue
		}
		if strings.Contains(f, string(filepath.Separator)) {
			f = filepath.Base(f)
		}
		parts = append(parts, f)
	}
	s := strings.Join(parts, " ")
	if len(s) > 60 {
		s = s[:57] + "..."
	}
	return s
}

func kindRank(kind string) int {
	switch kind {
	case "dev":
		return 0
	case "service":
		return 1
	}
	return 2
}

// relPath shows a project path relative to the root it was matched
// under, with ~ for the home directory.
func relPath(path string, roots []string) string {
	for _, root := range roots {
		if path == root || strings.HasPrefix(path, root+"/") {
			rel := strings.Trim(strings.TrimPrefix(path, root), "/")
			if rel == "" {
				return "~/" + filepath.Base(root)
			}
			return "~/" + filepath.Join(filepath.Base(root), rel)
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		path = strings.Replace(path, home, "~", 1)
	}
	return path
}
