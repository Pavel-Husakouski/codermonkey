# codermonkey

A set of microbenchmarks that implement the same small programs in many languages, to compare how compilers and runtimes perform on the same work.

Each benchmark uses the same algorithm, the same data layout and the same deterministic input in every language. Only the language and its toolchain change. A small Go harness builds and runs every implementation, collects timings and writes a Markdown report for each benchmark.

## Disclaimer

These are microbenchmarks. They measure a few hot loops and one allocation-heavy workload, not real application performance. Results depend a lot on the compiler version, the flags, the CPU and how idiomatic each implementation is. Treat them as a comparison of these programs, not a ranking of the languages.

## Benchmarks

| bench | what it does | what it stresses |
|-------|--------------|------------------|
| [`floyd`](floyd/) | Floyd–Warshall all-pairs shortest paths on a 512×512 `int32` matrix filled by an LCG (~70% of edges missing) | tight triple loop, flat array indexing, bounds checks, vectorization |
| [`lcs`](lcs/) | Longest common subsequence of two 4096-element sequences over a 4-letter alphabet, with two rolling `int32` rows | branchy inner loop, data dependencies between iterations |
| [`quicksort`](quicksort/) | Builds 1,000,000 objects with random numeric-string `Id`/`Value` through `Range` → `Select` → `ToArray` iterators, then sorts them by `Id` with a hand-written quicksort and a comparer callback | allocation, GC, string creation and comparison, iterator abstractions, indirect calls |

`floyd` and `lcs` check a checksum after every iteration and exit with an error on a mismatch, so a faster but wrong implementation can't win. `quicksort` uses non-deterministic random IDs and has no checksum.

## Languages

| key | language / runtime | source | build |
|-----|--------------------|--------|-------|
| `js` | Node.js | `bench.js` | — |
| `bun` | Bun | `bench.js` | — |
| `deno` | Deno | `bench.ts` | — |
| `scriptc` | scriptc (TypeScript compiled to native) | `bench.sc.ts` | `scriptc build` |
| `go` | Go | `bench.go` | `go build` |
| `java` | Java | `Bench.java` | `javac`, run on `java` |
| `cs` | C# / .NET | `Bench.cs` | `dotnet publish -c Release` |
| `cs-opt` | C# / .NET, zero-cost abstractions (struct comparer and struct iterators as generic arguments) | `BenchOpt.cs` (quicksort only) | `dotnet publish -c Release` |
| `c` | C (GCC) | `bench.c` | `gcc -O2 -std=gnu11` |
| `cpp` | C++ (G++) | `bench.cpp` | `g++ -O2 -std=c++20` |
| `rust` | Rust | `bench.rs` | `rustc -C opt-level=3` |
| `zig` | Zig | `bench.zig` | `zig build-exe -O ReleaseFast` |
| `swift` | Swift | `bench.swift` | `swiftc -O` |

The harness skips a language if its toolchain is not on `PATH`, or if the benchmark has no source file for it.

## Methodology

- Every program runs `Main()` in a loop (100 iterations by default) and sleeps 1 s between iterations, so the GC and the JIT can settle.
- Each iteration times two phases separately with a monotonic clock:
  - **init**: build the input data
  - **main phase** (`Floyd`, `Lcs`, `Sort`): the work being measured
- Each iteration prints one line to stderr:
  ```
  Init: 0.94 ms, Floyd: 41.16 ms
  ```
- The `ITER` environment variable sets the iteration count. For Node and Bun, the harness writes a patched copy of `bench.js` instead.
- Builds and runs happen one after another, never in parallel, so implementations don't compete for the CPU.

### Report columns

`<bench>.md` has one row per language, sorted by the median of the main phase:

| column | meaning |
|--------|---------|
| init median | median time of the init phase |
| first | first iteration (shows JIT warm-up and cold caches) |
| min / median / mean / p95 | statistics over all iterations |
| x best | median divided by the fastest median |
| status | `ok`, `FAIL` (build, run or checksum error) or `skipped` |

## Usage

Requirements: [`just`](https://github.com/casey/just), Go (for the harness), and the toolchains for whichever languages you want to run.

```sh
just                                        # list recipes
just run                                    # all benchmarks, all languages, 100 iterations
just iter=10 run                            # fewer iterations
just bench=floyd run                        # one benchmark
just bench=floyd,lcs lang=go,rust iter=20 run
just node=/path/to/node run                 # use a specific Node binary
just clean                                  # remove build/
```

Build output goes to `build/<bench>/`. Reports are written to `<bench>.md` in the project root, and rewritten after each language finishes, so a partial run still leaves a usable report.

## Results

Latest reports:

- [floyd.md](floyd.md)
- [lcs.md](lcs.md)
- [quicksort.md](quicksort.md)

Each report lists the date, CPU, OS and iteration count it was produced with. Numbers from different machines are not comparable.

## Adding things

**A new language:** add a source file to each benchmark directory, then add a `Lang` entry in [`harness/runners.go`](harness/runners.go). The entry gives the tool name, version command, build commands and run command. Set `Src` if only some benchmarks have an implementation.

**A new benchmark:** create a directory `<name>/` with one implementation per language, following the rules below. Then add the name to the default list in [`harness/main.go`](harness/main.go) and to the `bench` comment in the `justfile`. Every implementation must:

- use the same algorithm, sizes and deterministic input as the others
- time init and the main phase separately
- print `Init: <ms> ms, <Phase>: <ms> ms` to stderr each iteration
- honour `ITER` and sleep 1 s between iterations
- check a checksum and exit non-zero on a mismatch, when the output is deterministic
