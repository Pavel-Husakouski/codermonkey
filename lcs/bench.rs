#![allow(non_snake_case)]

use std::cell::Cell;
use std::time::{Duration, Instant};

const N: usize = 4096;
const EXPECTED: i64 = 6647453;

thread_local! {
    static START: Instant = Instant::now();
}

fn time() -> f64 {
    START.with(|s| s.elapsed().as_nanos() as f64 / 1e6)
}

// two LCG-generated strings over a 4-symbol alphabet (uint32 state wraps mod 2^32)
// plus two zeroed rolling rows
fn Init() -> (Vec<u8>, Vec<u8>, Vec<i32>, Vec<i32>) {
    let s = Cell::new(12345u32);
    let mut a = vec![0u8; N];
    let mut b = vec![0u8; N];
    for x in a.iter_mut().chain(b.iter_mut()) {
        s.set(s.get().wrapping_mul(1664525).wrapping_add(1013904223));
        *x = ((s.get() >> 16) & 3) as u8;
    }
    (a, b, vec![0i32; N + 1], vec![0i32; N + 1])
}

// serial inner loop: cur[j] depends on cur[j-1]
fn Lcs(a: &[u8], b: &[u8], mut prev: Vec<i32>, mut cur: Vec<i32>) -> Vec<i32> {
    for i in 1..=N {
        cur[0] = 0;
        for j in 1..=N {
            if a[i - 1] == b[j - 1] {
                cur[j] = prev[j - 1] + 1;
            } else {
                cur[j] = prev[j].max(cur[j - 1]);
            }
        }
        std::mem::swap(&mut prev, &mut cur);
    }
    prev
}

fn Checksum(p: &[i32]) -> i64 {
    p.iter().map(|&x| x as i64).sum()
}

fn Main() {
    let mut watch = time();
    let (a, b, prev, cur) = Init();
    let init = time() - watch;

    watch = time();
    let prev = Lcs(&a, &b, prev, cur);
    let lcs = time() - watch;

    if Checksum(&prev) != EXPECTED {
        eprintln!("Error: checksum mismatch: {} != {}", Checksum(&prev), EXPECTED);
        std::process::exit(1);
    }

    eprintln!("Init: {:.2} ms, Lcs: {:.2} ms", init, lcs);
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
        Main();
        delay(1000);
    }
}
