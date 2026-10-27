#if canImport(Glibc)
import Glibc
#elseif canImport(Darwin)
import Darwin
#endif

let N = 4096
let EXPECTED: Int64 = 6647453

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

// two LCG-generated (mod 2^32) strings over a 4-symbol alphabet, plus the two DP rows
func Init() -> ([UInt8], [UInt8], [Int32], [Int32]) {
    var a = [UInt8](repeating: 0, count: N)
    var b = [UInt8](repeating: 0, count: N)
    var s: UInt32 = 12345
    for i in 0..<N {
        s = s &* 1664525 &+ 1013904223
        a[i] = UInt8((s >> 16) & 3)
    }
    for i in 0..<N {
        s = s &* 1664525 &+ 1013904223
        b[i] = UInt8((s >> 16) & 3)
    }
    return (a, b, [Int32](repeating: 0, count: N + 1), [Int32](repeating: 0, count: N + 1))
}

func Lcs(_ a: [UInt8], _ b: [UInt8], _ prev: inout [Int32], _ cur: inout [Int32]) {
    for i in 1...N {
        cur[0] = 0
        for j in 1...N {
            if a[i - 1] == b[j - 1] {
                cur[j] = prev[j - 1] + 1
            } else {
                cur[j] = max(prev[j], cur[j - 1])
            }
        }
        swap(&prev, &cur)
    }
}

func Main() {
    var watch = time()
    let (a, b, p, c) = Init()
    var prev = p
    var cur = c
    let initMs = time() - watch

    watch = time()
    Lcs(a, b, &prev, &cur)
    let lcsMs = time() - watch

    var sum: Int64 = 0
    for v in prev {
        sum += Int64(v)
    }
    if sum != EXPECTED {
        fputs("checksum mismatch: \(sum) != \(EXPECTED)\n", stderr)
        exit(1)
    }

    fputs("Init: \(fixed2(initMs)) ms, Lcs: \(fixed2(lcsMs)) ms\n", stderr)
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
