#if canImport(Glibc)
import Glibc
#elseif canImport(Darwin)
import Darwin
#endif

let N = 512
let INF: Int32 = 1_000_000_000
let EXPECTED: Int64 = 12970375

func nanos() -> UInt64 {
    var ts = timespec()
    clock_gettime(CLOCK_MONOTONIC, &ts)
    return UInt64(ts.tv_sec) &* 1_000_000_000 &+ UInt64(ts.tv_nsec)
}

let startNanos = nanos()

func time() -> Double {
    return Double(nanos() &- startNanos) / 1e6
}

func fixed2(_ x: Double) -> String {
    let c = Int((x * 100).rounded())
    let frac = c % 100
    return "\(c / 100)." + (frac < 10 ? "0" : "") + "\(frac)"
}

// flat N*N matrix, LCG-generated (mod 2^32)
func Init() -> [Int32] {
    var d = [Int32](repeating: 0, count: N * N)
    var s: UInt32 = 12345
    for i in 0..<N {
        for j in 0..<N {
            s = s &* 1664525 &+ 1013904223
            if i == j {
                d[i * N + j] = 0
            } else if (s >> 8) % 10 < 7 {
                d[i * N + j] = INF
            } else {
                d[i * N + j] = Int32((s >> 16) & 1023) + 1
            }
        }
    }
    return d
}

func Floyd(_ d: inout [Int32]) {
    for k in 0..<N {
        for i in 0..<N {
            let dik = d[i * N + k]
            for j in 0..<N {
                let t = dik + d[k * N + j]
                if t < d[i * N + j] {
                    d[i * N + j] = t
                }
            }
        }
    }
}

func Main() {
    var watch = time()
    var d = Init()
    let initMs = time() - watch

    watch = time()
    Floyd(&d)
    let floydMs = time() - watch

    var sum: Int64 = 0
    for v in d {
        sum += Int64(v)
    }
    if sum != EXPECTED {
        fputs("checksum mismatch: \(sum) != \(EXPECTED)\n", stderr)
        exit(1)
    }

    fputs("Init: \(fixed2(initMs)) ms, Floyd: \(fixed2(floydMs)) ms\n", stderr)
}

func delay(_ ms: Int) {
    usleep(UInt32(ms) * 1000)
}

var iterations = 100
// ITER env var overrides the iteration count (smoke testing only)
if let s = getenv("ITER"), let n = Int(String(cString: s)) {
    iterations = n
}
for _ in 0..<iterations {
    Main()
    delay(1000)
}
