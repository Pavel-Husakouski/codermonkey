#![allow(non_snake_case)]

use std::cell::Cell;
use std::time::{Duration, Instant};

const N: usize = 512;
const INF: i32 = 1000000000;
const EXPECTED: i64 = 12970375;

thread_local! {
    static START: Instant = Instant::now();
}

fn time() -> f64 {
    START.with(|s| s.elapsed().as_nanos() as f64 / 1e6)
}

// flat N*N matrix, LCG-generated (uint32 state wraps mod 2^32)
fn Init() -> Vec<i32> {
    let s = Cell::new(12345u32);
    let mut d = vec![0i32; N * N];
    for i in 0..N {
        for j in 0..N {
            s.set(s.get().wrapping_mul(1664525).wrapping_add(1013904223));
            let s = s.get();
            d[i * N + j] = if i == j {
                0
            } else if ((s >> 8) % 10) < 7 {
                INF
            } else {
                ((s >> 16) & 1023) as i32 + 1
            };
        }
    }
    d
}

fn Floyd(d: &mut Vec<i32>) {
    for k in 0..N {
        for i in 0..N {
            let dik = d[i * N + k];
            for j in 0..N {
                let t = dik + d[k * N + j];
                if t < d[i * N + j] {
                    d[i * N + j] = t;
                }
            }
        }
    }
}

fn Checksum(d: &[i32]) -> i64 {
    d.iter().map(|&x| x as i64).sum()
}

fn Main() {
    let mut watch = time();
    let mut d = Init();
    let init = time() - watch;

    watch = time();
    Floyd(&mut d);
    let floyd = time() - watch;

    if Checksum(&d) != EXPECTED {
        eprintln!("Error: checksum mismatch: {} != {}", Checksum(&d), EXPECTED);
        std::process::exit(1);
    }

    eprintln!("Init: {:.2} ms, Floyd: {:.2} ms", init, floyd);
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
