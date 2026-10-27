

const N = 4096;
const EXPECTED = 6647453;

import { performance } from 'perf_hooks';

function time(): number {
    return performance.now();
}

function Init(): { a: Uint8Array; b: Uint8Array; prev: Int32Array; cur: Int32Array } {
    const a = new Uint8Array(N);
    const b = new Uint8Array(N);
    let s = 12345;
    for(let i = 0; i<N; i++) {
        s = (Math.imul(s, 1664525) + 1013904223) >>> 0;
        a[i] = (s >>> 16) & 3;
    }
    for(let i = 0; i<N; i++) {
        s = (Math.imul(s, 1664525) + 1013904223) >>> 0;
        b[i] = (s >>> 16) & 3;
    }
    return { a, b, prev: new Int32Array(N + 1), cur: new Int32Array(N + 1) };
}

function Lcs(a: Uint8Array, b: Uint8Array, prev: Int32Array, cur: Int32Array): Int32Array {
    for(let i = 1; i<=N; i++) {
        const ai = a[i - 1];
        cur[0] = 0;
        for(let j = 1; j<=N; j++) {
            if (ai === b[j - 1]) {
                cur[j] = prev[j - 1] + 1;
            } else {
                const u = prev[j];
                const l = cur[j - 1];
                cur[j] = u > l ? u : l;
            }
        }
        const t = prev;
        prev = cur;
        cur = t;
    }
    return prev;
}

function Checksum(d: Int32Array): number {
    let sum = 0;
    for(let i = 0; i<d.length; i++) {
        sum += d[i];
    }
    return sum;
}

function Main(): void {
    let watch = time();
    const { a, b, prev, cur } = Init();
    const init = time() - watch;

    watch = time();
    const res = Lcs(a, b, prev, cur);
    const lcs = time() - watch;

    const sum = Checksum(res);
    if (sum !== EXPECTED) {
        console.error(`Checksum mismatch: ${sum}, expected ${EXPECTED}`);
        process.exit(1);
    }

    console.warn(`Init: ${init.toFixed(2)} ms, Lcs: ${lcs.toFixed(2)} ms`);
}

function delay(time: number): Promise<void> {
    return new Promise<void>((resolve) => {
        setTimeout(resolve, time);
    });
}

const ITER = Number(process.env.ITER ?? "100");
let promise: Promise<void> = Promise.resolve();
for (let i = 0; i < ITER; i++) {
    promise = promise.then(() => Main()).then(() => delay(1000));
}
