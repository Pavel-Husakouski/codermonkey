const std = @import("std");

const N = 512;
const INF: i32 = 1000000000;
const EXPECTED: i64 = 12970375;

var io: std.Io = undefined;
var start: std.Io.Timestamp = undefined;
// one flat N*N matrix, reused across iterations
var d: [N * N]i32 = undefined;

fn time() f64 {
    const ns = start.durationTo(std.Io.Timestamp.now(io, .awake)).nanoseconds;
    return @as(f64, @floatFromInt(ns)) / 1e6;
}

fn Init() void {
    var s: u32 = 12345;
    for (0..N) |i| {
        for (0..N) |j| {
            s = s *% 1664525 +% 1013904223;
            if (i == j) {
                d[i * N + j] = 0;
            } else if ((s >> 8) % 10 < 7) {
                d[i * N + j] = INF;
            } else {
                d[i * N + j] = @as(i32, @intCast((s >> 16) & 1023)) + 1;
            }
        }
    }
}

fn Floyd() void {
    for (0..N) |k| {
        for (0..N) |i| {
            const dik = d[i * N + k];
            for (0..N) |j| {
                const t = dik + d[k * N + j];
                if (t < d[i * N + j]) d[i * N + j] = t;
            }
        }
    }
}

fn Checksum() i64 {
    var sum: i64 = 0;
    for (d) |v| sum += v;
    return sum;
}

fn Main() void {
    var watch = time();
    Init();
    const init = time() - watch;

    watch = time();
    Floyd();
    const floyd = time() - watch;

    const sum = Checksum();
    if (sum != EXPECTED) {
        std.debug.print("checksum mismatch: got {d}, want {d}\n", .{ sum, EXPECTED });
        std.process.exit(1);
    }

    std.debug.print("Init: {d:.2} ms, Floyd: {d:.2} ms\n", .{ init, floyd });
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
