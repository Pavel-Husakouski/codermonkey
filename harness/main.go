package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runCmd(c *Ctx, argv []string, env []string) (string, error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = c.Root
	cmd.Env = append(os.Environ(), env...)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	return out.String(), err
}

func tail(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 300 {
		s = "..." + s[len(s)-300:]
	}
	return s
}

func runOne(c *Ctx, l Lang) (Result, string) {
	r := Result{Lang: l.Name}
	if _, err := exec.LookPath(l.Tool); err != nil && !(l.Name == "js" && c.Node != "node") {
		r.Status, r.Note = "skipped", l.Tool+" not found"
		return r, ""
	}
	if l.Src != "" {
		if _, err := os.Stat(filepath.Join(c.Root, c.src(l.Src))); err != nil {
			r.Status, r.Note = "skipped", l.Src+" not found"
			return r, ""
		}
	}
	r.Version = l.Version(c)
	if l.Prepare != nil {
		if err := l.Prepare(c); err != nil {
			r.Status, r.Note = "FAIL", "prepare: "+err.Error()
			return r, ""
		}
	}
	if l.Build != nil {
		for _, cmd := range l.Build(c) {
			if out, err := runCmd(c, cmd, nil); err != nil {
				r.Status, r.Note = "FAIL", "build: "+tail(out)
				return r, ""
			}
		}
	}
	argv := l.Run(c)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = c.Root
	cmd.Env = append(os.Environ(), "ITER="+itoa(c.Iter))
	pipe, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		r.Status, r.Note = "FAIL", "start: "+err.Error()
		return r, ""
	}
	var phase string
	var other strings.Builder
	sc := bufio.NewScanner(pipe)
	for sc.Scan() {
		line := sc.Text()
		if s, ph, ok := parseLine(line); ok {
			phase = ph
			r.Samples = append(r.Samples, s)
			fmt.Fprintf(os.Stderr, "\r  %s: %d/%d  ", l.Name, len(r.Samples), c.Iter)
		} else {
			other.WriteString(line + "\n")
		}
	}
	fmt.Fprintln(os.Stderr)
	if err := cmd.Wait(); err != nil {
		r.Status, r.Note = "FAIL", err.Error()+": "+tail(other.String())
		return r, phase
	}
	if len(r.Samples) == 0 {
		r.Status, r.Note = "FAIL", "no samples parsed: "+tail(other.String())
		return r, phase
	}
	r.Status = "ok"
	return r, phase
}

func main() {
	root := flag.String("root", ".", "project root")
	bench := flag.String("bench", "all", "floyd|lcs|quicksort|all (comma list ok)")
	iter := flag.Int("iter", 100, "iterations per run")
	lang := flag.String("lang", "all", "comma list of languages, or all")
	out := flag.String("out", "", "directory for <bench>.md (default: root)")
	node := flag.String("node", "node", "node binary")
	flag.Parse()

	absRoot, _ := filepath.Abs(*root)
	if *out == "" {
		*out = absRoot
	}
	var benches []string
	if *bench == "all" {
		for _, n := range []string{"floyd", "lcs", "quicksort"} {
			if st, err := os.Stat(filepath.Join(absRoot, n)); err == nil && st.IsDir() {
				benches = append(benches, n)
			}
		}
	} else {
		benches = strings.Split(*bench, ",")
	}
	want := map[string]bool{}
	if *lang != "all" {
		for _, l := range strings.Split(*lang, ",") {
			want[l] = true
		}
	}

	for _, b := range benches {
		c := &Ctx{Root: absRoot, Bench: b, Iter: *iter, Node: *node}
		var res []Result
		phase := "main"
		for _, l := range langs {
			if len(want) > 0 && !want[l.Name] {
				continue
			}
			fmt.Fprintf(os.Stderr, "== %s / %s\n", b, l.Name)
			r, ph := runOne(c, l)
			if ph != "" {
				phase = ph
			}
			if r.Status != "ok" {
				fmt.Fprintf(os.Stderr, "  %s: %s\n", r.Status, r.Note)
			}
			res = append(res, r)
			if err := writeReport(filepath.Join(*out, b+".md"), b, phase, *iter, res); err != nil {
				fmt.Fprintln(os.Stderr, "report:", err)
				os.Exit(1)
			}
		}
		fmt.Fprintf(os.Stderr, "wrote %s\n", filepath.Join(*out, b+".md"))
	}
}
