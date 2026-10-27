#![allow(non_snake_case)]

use std::cell::Cell;
use std::fmt;
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

const N: i32 = 1000 * 1000;

thread_local! {
    static START: Instant = Instant::now();
    // xorshift128+ state, seeded from the system time (splitmix64)
    static RNG: Cell<[u64; 2]> = Cell::new(seed());
}

fn seed() -> [u64; 2] {
    let mut z = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_nanos() as u64;
    let mut next = || {
        z = z.wrapping_add(0x9E3779B97F4A7C15);
        let mut x = z;
        x = (x ^ (x >> 30)).wrapping_mul(0xBF58476D1CE4E5B9);
        x = (x ^ (x >> 27)).wrapping_mul(0x94D049BB133111EB);
        x ^ (x >> 31)
    };
    [next(), next()]
}

fn random() -> f64 {
    RNG.with(|r| {
        let [mut s1, s0] = r.get();
        s1 ^= s1 << 23;
        s1 ^= s1 >> 17;
        s1 ^= s0 ^ (s0 >> 26);
        r.set([s0, s1]);
        ((s0.wrapping_add(s1)) >> 11) as f64 / 9007199254740992.0
    })
}

fn time() -> f64 {
    START.with(|s| s.elapsed().as_nanos() as f64 / 1e6)
}

fn byId(x: &TestClass, y: &TestClass) -> i32 {
    let a = &x.Id;
    let b = &y.Id;
    a.cmp(b) as i32
}

fn Init(number: i32) -> Vec<Box<TestClass>> {
    ToArray(Select(Range(0, number), |_x| Box::new(TestClass::new())))
}

fn Sort(vals: &mut Vec<Box<TestClass>>) {
    QuickSort(vals, byId);
}

fn Main(number: i32) {
    let mut watch = time();
    let mut vals = Init(number);
    let init = time() - watch;

    // gc()
    watch = time();
    Sort(&mut vals);
    let sort = time() - watch;

    eprintln!("Init: {:.2} ms, Sort: {:.2} ms", init, sort);
}

fn QuickSort(vals: &mut Vec<Box<TestClass>>, compare: fn(&TestClass, &TestClass) -> i32) {
    QuickSortImpl(vals, compare, 0, vals.len() as isize - 1);
}

fn QuickSortImpl(
    vals: &mut Vec<Box<TestClass>>,
    compare: fn(&TestClass, &TestClass) -> i32,
    left: isize,
    right: isize,
) {
    if right <= left {
        return;
    }

    let mut i = left;
    let mut j = right;
    let mid = (left + right) >> 1;
    // swapping Boxes moves pointers, not heap objects, so a raw pointer stays the captured pivot
    let midVal: *const TestClass = &*vals[mid as usize];

    while i <= j {
        while compare(&vals[i as usize], unsafe { &*midVal }) > 0 {
            i += 1;
        }

        while compare(&vals[j as usize], unsafe { &*midVal }) < 0 {
            j -= 1;
        }

        if i <= j {
            vals.swap(i as usize, j as usize);

            i += 1;
            j -= 1;
        }
    }

    if left < i - 1 {
        QuickSortImpl(vals, compare, left, i - 1);
    }

    if right > i {
        QuickSortImpl(vals, compare, i, right);
    }
}

#[allow(dead_code)]
struct TestClass {
    Id: String,
    Value: String,
}

impl TestClass {
    fn new() -> TestClass {
        let Id = CreateGuid();
        let Value = CreateGuid();
        TestClass { Id, Value }
    }
}

impl fmt::Display for TestClass {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        f.write_str(&self.Id)
    }
}

fn ToArray<T>(iterator: impl Iterator<Item = T>) -> Vec<T> {
    iterator.collect()
}

struct RangeIter {
    i: i32,
    end: i32,
}

impl Iterator for RangeIter {
    type Item = i32;
    fn next(&mut self) -> Option<i32> {
        if self.i < self.end {
            let v = self.i;
            self.i += 1;
            Some(v)
        } else {
            None
        }
    }
}

fn Range(start: i32, end: i32) -> RangeIter {
    RangeIter { i: start, end }
}

struct SelectIter<I, F> {
    iterator: I,
    projector: F,
}

impl<I: Iterator, U, F: FnMut(I::Item) -> U> Iterator for SelectIter<I, F> {
    type Item = U;
    fn next(&mut self) -> Option<U> {
        let x = self.iterator.next()?;
        Some((self.projector)(x))
    }
}

fn Select<I: Iterator, U, F: FnMut(I::Item) -> U>(iterator: I, projector: F) -> SelectIter<I, F> {
    SelectIter { iterator, projector }
}

fn CreateGuid() -> String {
    format!("{}", (random() * 1e16).floor())
}

fn delay(ms: u64) {
    std::thread::sleep(Duration::from_millis(ms));
}

fn main() {
    let mut iterations = 100;
    // ITER env var overrides the iteration count (smoke testing only)
    if let Ok(s) = std::env::var("ITER") {
        if let Ok(n) = s.parse::<i32>() {
            iterations = n;
        }
    }
    for _ in 0..iterations {
        Main(N);
        delay(1000);
    }
}
