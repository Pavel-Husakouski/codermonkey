#if canImport(Glibc)
import Glibc
#elseif canImport(Darwin)
import Darwin
#endif

let N = 1000 * 1000

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

// bytewise utf8 compare (== strings.Compare); String < is ~2x slower here
let byId: (TestClass, TestClass) -> Int = { x, y in
    let a = x.Id, b = y.Id
    return a.utf8.withContiguousStorageIfAvailable { pa in
        b.utf8.withContiguousStorageIfAvailable { pb in
            let n = min(pa.count, pb.count)
            let c = memcmp(pa.baseAddress!, pb.baseAddress!, n)
            return c != 0 ? (c < 0 ? -1 : 1) : (pa.count == pb.count ? 0 : pa.count < pb.count ? -1 : 1)
        }!
    }!
}

func Init(_ number: Int) -> [TestClass] {
    return ToArray(Select(Range(0, number), { _ in TestClass() }))
}

func Sort(_ vals: inout [TestClass]) {
    QuickSort(&vals, byId)
}

func Main(_ number: Int) {
    var watch = time()
    var vals = Init(number)
    let initMs = time() - watch

    // gc()
    watch = time()
    Sort(&vals)
    let sortMs = time() - watch

    fputs("Init: \(fixed2(initMs)) ms, Sort: \(fixed2(sortMs)) ms\n", stderr)
    withExtendedLifetime(vals) {}  // array is released after the timed sections
}

func QuickSort(_ vals: inout [TestClass], _ compare: (TestClass, TestClass) -> Int) {
    QuickSortImpl(&vals, compare, 0, vals.count - 1)
}

func QuickSortImpl(_ vals: inout [TestClass], _ compare: (TestClass, TestClass) -> Int, _ left: Int, _ right: Int) {
    if right <= left {
        return
    }

    var i = left
    var j = right
    let mid = Int(bitPattern: UInt(bitPattern: left &+ right) &>> 1)
    let midVal = vals[mid]

    while i <= j {
        while compare(vals[i], midVal) > 0 {
            i += 1
        }

        while compare(vals[j], midVal) < 0 {
            j -= 1
        }

        if i <= j {
            let tmp = vals[i]
            vals[i] = vals[j]
            vals[j] = tmp

            i += 1
            j -= 1
        }
    }

    if left < i - 1 {
        QuickSortImpl(&vals, compare, left, i - 1)
    }

    if right > i {
        QuickSortImpl(&vals, compare, i, right)
    }
}

final class TestClass: CustomStringConvertible {
    let Id: String
    let Value: String

    init() {
        Id = CreateGuid()
        Value = CreateGuid()
    }

    var description: String {
        return Id
    }
}

func ToArray<I: IteratorProtocol>(_ iterator: I) -> [I.Element] {
    var it = iterator
    var result: [I.Element] = []
    while let x = it.next() {
        result.append(x)
    }
    return result
}

struct RangeIterator: IteratorProtocol {
    var i: Int
    let end: Int

    mutating func next() -> Int? {
        if i < end {
            let v = i
            i += 1
            return v
        }
        return nil
    }
}

func Range(_ start: Int, _ end: Int) -> RangeIterator {
    return RangeIterator(i: start, end: end)
}

struct SelectIterator<I: IteratorProtocol, U>: IteratorProtocol {
    var iterator: I
    let projector: (I.Element) -> U

    mutating func next() -> U? {
        if let x = iterator.next() {
            return projector(x)
        }
        return nil
    }
}

func Select<I: IteratorProtocol, U>(_ iterator: I, _ projector: @escaping (I.Element) -> U) -> SelectIterator<I, U> {
    return SelectIterator(iterator: iterator, projector: projector)
}

// xorshift128+, seeded from the clock (stands in for Math.random)
var rngS0: UInt64 = 0
var rngS1: UInt64 = 0

func seedRandom() {
    var ts = timespec()
    clock_gettime(CLOCK_REALTIME, &ts)
    var z = UInt64(ts.tv_sec) &* 1_000_000_000 &+ UInt64(ts.tv_nsec) ^ nanos()
    func splitmix() -> UInt64 {
        z = z &+ 0x9E37_79B9_7F4A_7C15
        var r = z
        r = (r ^ (r &>> 30)) &* 0xBF58_476D_1CE4_E5B9
        r = (r ^ (r &>> 27)) &* 0x94D0_49BB_1331_11EB
        return r ^ (r &>> 31)
    }
    rngS0 = splitmix()
    rngS1 = splitmix()
}

func random() -> Double {
    var s1 = rngS0
    let s0 = rngS1
    rngS0 = s0
    s1 ^= s1 &<< 23
    s1 ^= s1 &>> 17
    s1 ^= s0
    s1 ^= s0 &>> 26
    rngS1 = s1
    return Double((rngS0 &+ rngS1) &>> 11) * 0x1.0p-53
}

// String(v) of a whole double below 1e21, as JS prints it
func formatWhole(_ v: Double) -> String {
    if v < 9007199254740992.0 {
        return String(UInt64(v))
    }
    let d = v.description
    var mant = d
    var exp = 0
    if let e = d.firstIndex(where: { $0 == "e" || $0 == "E" }) {
        mant = String(d[d.startIndex..<e])
        exp = Int(d[d.index(after: e)...].replacingPlus())!
    }
    var intPart = mant
    var fracPart = ""
    if let dot = mant.firstIndex(of: ".") {
        intPart = String(mant[mant.startIndex..<dot])
        fracPart = String(mant[mant.index(after: dot)...])
    }
    var digits = intPart + fracPart
    let pointPos = intPart.count + exp
    if digits.count < pointPos {
        digits += String(repeating: "0", count: pointPos - digits.count)
    }
    var out = String(digits.prefix(pointPos))
    let rest = digits.dropFirst(pointPos)
    if rest.contains(where: { $0 != "0" }) {
        out += "." + rest
    }
    return out
}

extension Substring {
    func replacingPlus() -> String {
        return self.hasPrefix("+") ? String(self.dropFirst()) : String(self)
    }
}

func CreateGuid() -> String {
    return formatWhole((random() * 1e16).rounded(.down))
}

func delay(_ ms: Int) {
    usleep(UInt32(ms) * 1000)
}

seedRandom()
var iterations = 100
// ITER env var overrides the iteration count (smoke testing only)
if let s = getenv("ITER"), let n = Int(String(cString: s)) {
    iterations = n
}
for _ in 0..<iterations {
    Main(N)
    delay(1000)
}
