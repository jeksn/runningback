package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	keyUp    = []byte{27, 91, 65}
	keyDown  = []byte{27, 91, 66}
	ansiDim  = "\033[2m"
	ansiBold = "\033[1m"
	ansiRev  = "\033[7m"
	ansiWarn = "\033[33m"
	ansiURL  = "\033[36m"
	ansiOff  = "\033[0m"
)

func isTTY() bool {
	st, err := os.Stdin.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

func stty(args ...string) (string, error) {
	cmd := exec.Command("stty", args...)
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

type winsize struct{ Row, Col, X, Y uint16 }

func termWidth() int {
	ws := winsize{}
	// TIOCGWINSZ on darwin
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), 0x40087468, uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.Col == 0 {
		return 100
	}
	return int(ws.Col)
}

func detail(p Proc) string {
	if p.URL != "" {
		return p.URL
	}
	parts := make([]string, len(p.Ports))
	for i, port := range p.Ports {
		parts[i] = fmt.Sprint(port)
	}
	return "ports " + strings.Join(parts, ",")
}

func trunc(s string, w int) string {
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	return string(r[:w])
}

func pad(s string, w int) string {
	r := []rune(s)
	if len(r) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(r))
}

func clampSel(sel, n int) int {
	if n == 0 {
		return 0
	}
	if sel < 0 {
		return 0
	}
	if sel >= n {
		return n - 1
	}
	return sel
}

// bannerArt is the figlet wordmark shown in the TUI header on
// roomy terminals. One line contains backticks, so the constant
// is built with string concatenation.
const bannerArt = `                           _             __               __
   _______  ______  ____  (_)___  ____ _/ /_  ____ ______/ /__
  / ___/ / / / __ \/ __ \/ / __ \/ __ ` + "`" + `/ __ \/ __ ` + "`" + `/ ___/ //_/
 / /  / /_/ / / / / / / / / / / / /_/ / /_/ / /_/ / /__/ ,<
/_/   \__,_/_/ /_/_/ /_/_/_/ /_/\__, /_.___/\__,_/\___/_/|_|
                               /____/`
// breakdown summarizes the running list as "1 project · 4 services",
// counting dev processes by project checkout rather than by process.
func breakdown(procs []Proc) string {
	projects := map[string]bool{}
	services := 0
	for _, p := range procs {
		if p.Kind == "dev" {
			key := p.Dir
			if key == "" {
				key = p.Cwd
			}
			projects[key] = true
		} else {
			services++
		}
	}
	var parts []string
	if n := len(projects); n > 0 {
		if n == 1 {
			parts = append(parts, "1 project")
		} else {
			parts = append(parts, fmt.Sprintf("%d projects", n))
		}
	}
	if services > 0 {
		if services == 1 {
			parts = append(parts, "1 service")
		} else {
			parts = append(parts, fmt.Sprintf("%d services", services))
		}
	}
	if len(parts) == 0 {
		return "nothing running"
	}
	return strings.Join(parts, " · ")
}

func termHeight() int {
	ws := winsize{}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), 0x40087468, uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.Row == 0 {
		return 40
	}
	return int(ws.Row)
}

func frame(procs []Proc, sel int, status string, f filter) string {
	sel = clampSel(sel, len(procs))
	var b strings.Builder
	w := termWidth()
	b.WriteString("\033[H\033[2J")

	artLines := strings.Split(bannerArt, "\n")
	maxArt := 0
	for _, line := range artLines {
		if n := len([]rune(line)); n > maxArt {
			maxArt = n
		}
	}
	if termHeight() >= 24 && w >= maxArt+2 {
		for _, line := range artLines {
			fmt.Fprintf(&b, "%s%s%s\r\n", ansiURL, trunc(line, w-1), ansiOff)
		}
		fmt.Fprintf(&b, "\r\n %s%s%s\r\n\r\n", ansiDim, breakdown(procs), ansiOff)
	} else {
		fmt.Fprintf(&b, "%s runningback%s %s%s%s\r\n\r\n",
			ansiBold, ansiOff, ansiDim, breakdown(procs), ansiOff)
	}

	lastKind, lastDir := "", ""
	for i, p := range procs {
		if p.Kind != lastKind {
			lastKind = p.Kind
			lastDir = ""
			fmt.Fprintf(&b, "\r\n%s %s%s\r\n", ansiBold, kindTitle(p.Kind), ansiOff)
		}
		marker := "  "
		if i == sel {
			marker = "▶ "
		}
		var row string
		if p.Kind == "dev" {
			if p.Dir != lastDir {
				lastDir = p.Dir
				fmt.Fprintf(&b, "\r\n %s%-22s%s %s%s%s\r\n",
					ansiBold, p.Project, ansiOff, ansiDim, relPath(p.Dir, f.Roots), ansiOff)
			}
			age := ""
			if p.Age != "" {
				age = " up " + p.Age
			}
			row = fmt.Sprintf("%s%s%-9s%s %-24s %s%-28s%s %s pid %-6d%s%s",
				marker, ansiBold, p.Role, ansiOff, trunc(p.Label, 24), ansiURL, detail(p),
				ansiOff, ansiDim, p.PID, age, ansiOff)
		} else {
			row = fmt.Sprintf("%s%-20s %s%-28s%s %s%s pid %d",
				marker, p.Project, ansiURL, detail(p), ansiOff, ansiDim, trunc(p.Label, 34), p.PID)
		}
		row = trunc(row, w-1)
		if i == sel {
			fmt.Fprintf(&b, "%s%s%s\r\n", ansiRev, pad(row, w-1), ansiOff)
		} else {
			fmt.Fprintf(&b, "%s\r\n", row)
		}
	}

	b.WriteString("\r\n")
	if len(procs) > 0 {
		p := procs[sel]
		if p.Kind == "dev" {
			fmt.Fprintf(&b, "%s %s%s\r\n", ansiDim, relPath(p.Cwd, f.Roots), ansiOff)
		}
	}
	fmt.Fprintf(&b, "%s r refresh · o open · c copy · x kill · q quit%s\r\n", ansiDim, ansiOff)
	if status != "" {
		fmt.Fprintf(&b, "%s %s%s", ansiWarn, status, ansiOff)
	}
	return b.String()
}

func render(procs []Proc, sel int, status string, f filter) {
	fmt.Fprint(os.Stdout, frame(procs, sel, status, f))
}

func copyToClipboard(text string) string {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		return err.Error()
	}
	return "copied " + text
}

func runTUI(f filter) error {
	if !isTTY() {
		procs, err := detect(f)
		if err != nil {
			return err
		}
		printList(os.Stdout, procs, f)
		return nil
	}

	saved, err := stty("-g")
	if err != nil {
		procs, derr := detect(f)
		if derr != nil {
			return derr
		}
		printList(os.Stdout, procs, f)
		return nil
	}
	if _, err := stty("raw", "-echo"); err != nil {
		return err
	}
	defer stty(saved)

	procs, err := detect(f)
	if err != nil {
		return err
	}
	sel, status := 0, ""

	keys := make(chan []byte, 16)
	go func() {
		buf := make([]byte, 64)
		for {
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				b := make([]byte, n)
				copy(b, buf[:n])
				keys <- b
			}
			if err != nil {
				close(keys)
				return
			}
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")
	render(procs, sel, status, f)

	clamp := func() {
		if len(procs) == 0 {
			sel = 0
			return
		}
		if sel < 0 {
			sel = 0
		}
		if sel >= len(procs) {
			sel = len(procs) - 1
		}
	}

	for {
		select {
		case <-ticker.C:
			if p, err := detect(f); err == nil {
				procs = p
				clamp()
				render(procs, sel, status, f)
			}
		case b, ok := <-keys:
			if !ok {
				return nil
			}
			byteIs := func(c byte) bool { return len(b) == 1 && b[0] == c }
			switch {
			case byteIs('q') || byteIs(3):
				return nil
			case bytes.Equal(b, keyUp) || byteIs('k'):
				sel--
			case bytes.Equal(b, keyDown) || byteIs('j'):
				sel++
			case byteIs('r'):
				if p, err := detect(f); err == nil {
					procs = p
					clamp()
					status = "refreshed"
				}
			case byteIs('o') || byteIs('\r') || byteIs('\n'):
				if len(procs) > 0 && procs[sel].URL != "" {
					if err := exec.Command("open", procs[sel].URL).Start(); err == nil {
						status = "opened " + procs[sel].URL
					} else {
						status = err.Error()
					}
				}
			case byteIs('c'):
				if len(procs) > 0 && procs[sel].URL != "" {
					status = copyToClipboard(procs[sel].URL)
				}
			case byteIs('x'):
				if len(procs) > 0 {
					p := procs[sel]
					if err := syscall.Kill(p.PID, syscall.SIGTERM); err != nil {
						status = err.Error()
					} else {
						status = "sent SIGTERM to " + p.Project + " (pid " + fmt.Sprint(p.PID) + ")"
						time.Sleep(300 * time.Millisecond)
						if np, err := detect(f); err == nil {
							procs = np
							clamp()
						}
					}
				}
			}
			clamp()
			render(procs, sel, status, f)
		}
	}
}
