const std = @import("std");

const N = 4096;
const EXPECTED: i64 = 6647453;

var io: std.Io = undefined;
var start: std.Io.Timestamp = undefined;
// two sequences and two rolling DP rows, reused across iterations
var a: [N]u8 = undefined;
var b: [N]u8 = undefined;
var rowA: [N + 1]i32 = undefined;
var rowB: [N + 1]i32 = undefined;
var prev: *[N + 1]i32 = undefined;
var cur: *[N + 1]i32 = undefined;

fn time() f64 {
    const ns = start.durationTo(std.Io.Timestamp.now(io, .awake)).nanoseconds;
    return @as(f64, @floatFromInt(ns)) / 1e6;
}

fn Init() void {
    var s: u32 = 12345;
    for (0..N) |i| {
        s = s *% 1664525 +% 1013904223;
        a[i] = @intCast((s >> 16) & 3);
    }
    for (0..N) |i| {
        s = s *% 1664525 +% 1013904223;
        b[i] = @intCast((s >> 16) & 3);
    }
    @memset(&rowA, 0);
    @memset(&rowB, 0);
    prev = &rowA;
    cur = &rowB;
}

fn Lcs() void {
    for (1..N + 1) |i| {
        cur[0] = 0;
        for (1..N + 1) |j| {
            if (a[i - 1] == b[j - 1]) {
                cur[j] = prev[j - 1] + 1;
            } else {
                cur[j] = @max(prev[j], cur[j - 1]);
            }
        }
        const t = prev;
        prev = cur;
        cur = t;
    }
}

fn Checksum() i64 {
    var sum: i64 = 0;
    for (prev) |v| sum += v;
    return sum;
}

fn Main() void {
    var watch = time();
    Init();
    const init = time() - watch;

    watch = time();
    Lcs();
    const lcs = time() - watch;

    const sum = Checksum();
    if (sum != EXPECTED) {
        std.debug.print("checksum mismatch: got {d}, want {d}\n", .{ sum, EXPECTED });
        std.process.exit(1);
    }

    std.debug.print("Init: {d:.2} ms, Lcs: {d:.2} ms\n", .{ init, lcs });
}

fn delay(ms: i64) void {
    io.sleep(.fromMilliseconds(ms), .awake) catch {};
}

pub fn main(init: std.process.Init) !void {
    io = init.io;
    start = std.Io.Timestamp.now(io, .awake);

    var iterations: usize = 100;
    // ITER env var overrides the iteration count (smoke testing only)
    if (init.environ_map.get("ITER")) |s| {
        if (std.fmt.parseInt(usize, s, 10)) |n| iterations = n else |_| {}
    }
    var i: usize = 0;
    while (i < iterations) : (i += 1) {
        Main();
        delay(1000);
    }
}
