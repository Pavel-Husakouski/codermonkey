package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Lang describes how to build and run one language implementation.
// All paths are relative to the project root (commands run with cwd=root).
type Lang struct {
	Name    string
	Tool    string                  // binary that must exist in PATH
	Version func(c *Ctx) string     // toolchain version
	Build   func(c *Ctx) [][]string // build commands (may be empty)
	Run     func(c *Ctx) []string   // run command
	Prepare func(c *Ctx) error      // optional step before Build
	Src     string                  // optional source file; skipped if the bench lacks it
}

type Ctx struct {
	Root  string
	Bench string
	Iter  int
	Node  string
}

func (c *Ctx) src(name string) string { return filepath.Join(c.Bench, name) }
func (c *Ctx) out(name string) string { return filepath.Join("build", c.Bench, name) }

func sh(name string, args ...string) string {
	b, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return "?"
	}
	return string(b)
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return s
}

var verRe = regexp.MustCompile(`\d+(\.\d+)+`)

func ver(name string, args ...string) string {
	if m := verRe.FindString(sh(name, args...)); m != "" {
		return m
	}
	return "?"
}

// jsIter makes a copy of bench.js with the iteration count patched (bench.js hardcodes 100).
func jsIter(c *Ctx) error {
	if c.Iter == 100 {
		return nil
	}
	b, err := os.ReadFile(filepath.Join(c.Root, c.src("bench.js")))
	if err != nil {
		return err
	}
	patched := strings.Replace(string(b), "i<100;i++", "i<"+itoa(c.Iter)+";i++", 1)
	if err := os.MkdirAll(filepath.Join(c.Root, "build", c.Bench), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.Root, c.out("bench_iter.js")), []byte(patched), 0o644)
}

func jsPath(c *Ctx) string {
	if c.Iter == 100 {
		return c.src("bench.js")
	}
	return c.out("bench_iter.js")
}

func mk(c *Ctx) [][]string { return [][]string{{"mkdir", "-p", filepath.Join("build", c.Bench)}} }

var langs = []Lang{
	{Name: "js", Tool: "node",
		Version: func(c *Ctx) string { return ver(c.Node, "-v") },
		Prepare: jsIter,
		Run:     func(c *Ctx) []string { return []string{c.Node, jsPath(c)} }},
	{Name: "bun", Tool: "bun",
		Version: func(c *Ctx) string { return ver("bun", "--version") },
		Prepare: jsIter,
		Run:     func(c *Ctx) []string { return []string{"bun", jsPath(c)} }},
	{Name: "deno", Tool: "deno",
		Version: func(c *Ctx) string { return ver("deno", "--version") },
		Run:     func(c *Ctx) []string { return []string{"deno", "run", "--allow-env", c.src("bench.ts")} }},
	{Name: "scriptc", Tool: "scriptc",
		Version: func(c *Ctx) string { return ver("scriptc", "--version") },
		Build: func(c *Ctx) [][]string {
			return append(mk(c), []string{"scriptc", "build", c.src("bench.sc.ts"), "-o", c.out("bench-scriptc")})
		},
		Run: func(c *Ctx) []string { return []string{c.out("bench-scriptc")} }},
	{Name: "go", Tool: "go",
		Version: func(c *Ctx) string { return ver("go", "version") },
		Build: func(c *Ctx) [][]string {
			return append(mk(c), []string{"go", "build", "-o", c.out("bench-go"), c.src("bench.go")})
		},
		Run: func(c *Ctx) []string { return []string{c.out("bench-go")} }},
	{Name: "java", Tool: "javac",
		Version: func(c *Ctx) string { return ver("java", "-version") },
		Build: func(c *Ctx) [][]string {
			return [][]string{{"mkdir", "-p", c.out("java")}, {"javac", "-d", c.out("java"), c.src("Bench.java")}}
		},
		Run: func(c *Ctx) []string { return []string{"java", "-cp", c.out("java"), "Bench"} }},
	{Name: "cs", Tool: "dotnet",
		Version: func(c *Ctx) string { return ver("dotnet", "--version") },
		Build: func(c *Ctx) [][]string {
			return [][]string{{"dotnet", "publish", "-c", "Release", c.src("Bench.cs"), "-o", c.out("cs"), "-v", "q", "--nologo"}}
		},
		Run: func(c *Ctx) []string { return []string{c.out("cs/Bench")} }},
	{Name: "cs-opt", Tool: "dotnet", Src: "BenchOpt.cs",
		Version: func(c *Ctx) string { return ver("dotnet", "--version") },
		Build: func(c *Ctx) [][]string {
			return [][]string{{"dotnet", "publish", "-c", "Release", c.src("BenchOpt.cs"), "-o", c.out("cs-opt"), "-v", "q", "--nologo"}}
		},
		Run: func(c *Ctx) []string { return []string{c.out("cs-opt/BenchOpt")} }},
	{Name: "c", Tool: "gcc",
		Version: func(c *Ctx) string { return ver("gcc", "-dumpfullversion") },
		Build: func(c *Ctx) [][]string {
			return append(mk(c), []string{"gcc", "-O2", "-std=gnu11", "-Wall", "-Wextra", "-o", c.out("bench-c"), c.src("bench.c"), "-lm"})
		},
		Run: func(c *Ctx) []string { return []string{c.out("bench-c")} }},
	{Name: "cpp", Tool: "g++",
		Version: func(c *Ctx) string { return ver("g++", "-dumpfullversion") },
		Build: func(c *Ctx) [][]string {
			return append(mk(c), []string{"g++", "-O2", "-std=c++20", "-Wall", "-Wextra", "-o", c.out("bench-cpp"), c.src("bench.cpp")})
		},
		Run: func(c *Ctx) []string { return []string{c.out("bench-cpp")} }},
	{Name: "rust", Tool: "rustc",
		Version: func(c *Ctx) string { return ver("rustc", "--version") },
		Build: func(c *Ctx) [][]string {
			return append(mk(c), []string{"rustc", "-C", "opt-level=3", "-o", c.out("bench-rust"), c.src("bench.rs")})
		},
		Run: func(c *Ctx) []string { return []string{c.out("bench-rust")} }},
	{Name: "zig", Tool: "zig",
		Version: func(c *Ctx) string { return ver("zig", "version") },
		Build: func(c *Ctx) [][]string {
			return append(mk(c), []string{"zig", "build-exe", c.src("bench.zig"), "-O", "ReleaseFast",
				"-femit-bin=" + c.out("bench-zig"), "--cache-dir", c.out(".zig-cache"), "--global-cache-dir", c.out(".zig-global-cache")})
		},
		Run: func(c *Ctx) []string { return []string{c.out("bench-zig")} }},
	{Name: "swift", Tool: "swiftc",
		Version: func(c *Ctx) string { return ver("swift", "--version") },
		Build: func(c *Ctx) [][]string {
			return append(mk(c), []string{"swiftc", "-O", "-module-cache-path", c.out(".swift-modcache"), "-o", c.out("bench-swift"), c.src("bench.swift")})
		},
		Run: func(c *Ctx) []string { return []string{c.out("bench-swift")} }},
}
