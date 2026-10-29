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

func frame(procs []Proc, sel int, status string, f filter) string {
	sel = clampSel(sel, len(procs))
	var b strings.Builder
	w := termWidth()
	b.WriteString("\033[H\033[2J")
	fmt.Fprintf(&b, "%s runningback%s %s%d running%s\r\n\r\n",
		ansiBold, ansiOff, ansiDim, len(procs), ansiOff)

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
			row = fmt.Sprintf("%s%s%-9s%s %-24s %s%-28s%s %s pid %d",
				marker, ansiBold, p.Role, ansiOff, trunc(p.Label, 24), ansiURL, detail(p), ansiOff, ansiDim, p.PID)
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
	fmt.Fprintf(&b, "%s r refresh · o open · x kill · q quit%s\r\n", ansiDim, ansiOff)
	if status != "" {
		fmt.Fprintf(&b, "%s %s%s", ansiWarn, status, ansiOff)
	}
	return b.String()
}

func render(procs []Proc, sel int, status string, f filter) {
	fmt.Fprint(os.Stdout, frame(procs, sel, status, f))
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
