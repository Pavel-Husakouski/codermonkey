const std = @import("std");

const N = 1000 * 1000;

// gpa: individual alloc/free of every object is the point, so no arena; init.gpa is
// smp_allocator in release builds and a leak-checking DebugAllocator in Debug.
var gpa: std.mem.Allocator = undefined;
var io: std.Io = undefined;
var start: std.Io.Timestamp = undefined;
var prng: std.Random.Xoshiro256 = undefined;

fn time() f64 {
    const ns = start.durationTo(std.Io.Timestamp.now(io, .awake)).nanoseconds;
    return @as(f64, @floatFromInt(ns)) / 1e6;
}

fn byId(x: *const TestClass, y: *const TestClass) i32 {
    const a = x.Id;
    const b = y.Id;
    return switch (std.mem.order(u8, a, b)) {
        .lt => -1,
        .eq => 0,
        .gt => 1,
    };
}

fn Init(number: i64) !std.ArrayList(*TestClass) {
    return ToArray(Select(Range(0, number), newTestClass));
}

fn Sort(vals: []*TestClass) void {
    QuickSort(vals, byId);
}

fn Main(number: i64) !void {
    var watch = time();
    var vals = try Init(number);
    const init = time() - watch;

    // runtime.GC()
    watch = time();
    Sort(vals.items);
    const sort = time() - watch;

    std.debug.print("Init: {d:.2} ms, Sort: {d:.2} ms\n", .{ init, sort });

    for (vals.items) |t| t.deinit();
    vals.deinit(gpa);
}

// runtime function pointer (not comptime) so the comparator call is not inlined away
fn QuickSort(vals: []*TestClass, compare: *const fn (*const TestClass, *const TestClass) i32) void {
    QuickSortImpl(vals, compare, 0, @as(isize, @intCast(vals.len)) - 1);
}

fn QuickSortImpl(vals: []*TestClass, compare: *const fn (*const TestClass, *const TestClass) i32, left: isize, right: isize) void {
    if (right <= left)
        return;

    var i = left;
    var j = right;
    const mid = (left + right) >> 1;
    const midVal = vals[@intCast(mid)];

    while (i <= j) {
        while (compare(vals[@intCast(i)], midVal) > 0) {
            i += 1;
        }

        while (compare(vals[@intCast(j)], midVal) < 0) {
            j -= 1;
        }

        if (i <= j) {
            const tmp = vals[@intCast(i)];
            vals[@intCast(i)] = vals[@intCast(j)];
            vals[@intCast(j)] = tmp;

            i += 1;
            j -= 1;
        }
    }

    if (left < i - 1) {
        QuickSortImpl(vals, compare, left, i - 1);
    }

    if (right > i) {
        QuickSortImpl(vals, compare, i, right);
    }
}

const TestClass = struct {
    Id: []u8,
    Value: []u8,

    fn init() !*TestClass {
        const t = try gpa.create(TestClass);
        t.Id = try CreateGuid();
        t.Value = try CreateGuid();
        return t;
    }

    fn deinit(t: *TestClass) void {
        gpa.free(t.Id);
        gpa.free(t.Value);
        gpa.destroy(t);
    }

    fn toString(t: *const TestClass) []const u8 {
        return t.Id;
    }
};

fn newTestClass(_: i64) !*TestClass {
    return TestClass.init();
}

fn ToArray(iterator: anytype) !std.ArrayList(*TestClass) {
    var it = iterator;
    var list: std.ArrayList(*TestClass) = .empty;
    while (try it.next()) |x| {
        try list.append(gpa, x);
    }
    return list;
}

const RangeIter = struct {
    i: i64,
    end: i64,

    fn next(self: *RangeIter) ?i64 {
        if (self.i >= self.end) return null;
        const v = self.i;
        self.i += 1;
        return v;
    }
};

fn Range(s: i64, end: i64) RangeIter {
    return .{ .i = s, .end = end };
}

const SelectIter = struct {
    iterator: RangeIter,
    projector: *const fn (i64) anyerror!*TestClass,

    fn next(self: *SelectIter) !?*TestClass {
        const x = self.iterator.next() orelse return null;
        return try self.projector(x);
    }
};

fn Select(iterator: RangeIter, projector: *const fn (i64) anyerror!*TestClass) SelectIter {
    return .{ .iterator = iterator, .projector = projector };
}

fn CreateGuid() ![]u8 {
    const v = @floor(prng.random().float(f64) * 1e16);
    var buf: [32]u8 = undefined;
    const s = try std.fmt.bufPrint(&buf, "{d}", .{v});
    return gpa.dupe(u8, s);
}

fn delay(ms: i64) void {
    io.sleep(.fromMilliseconds(ms), .awake) catch {};
}

pub fn main(init: std.process.Init) !void {
    gpa = init.gpa;
    io = init.io;
    start = std.Io.Timestamp.now(io, .awake);
    prng = .init(@bitCast(@as(i64, @truncate(std.Io.Timestamp.now(io, .real).nanoseconds))));

    var iterations: usize = 100;
    // ITER env var overrides the iteration count (smoke testing only)
    if (init.environ_map.get("ITER")) |s| {
        if (std.fmt.parseInt(usize, s, 10)) |n| iterations = n else |_| {}
    }
    var i: usize = 0;
    while (i < iterations) : (i += 1) {
        try Main(N);
        delay(1000);
    }
}
