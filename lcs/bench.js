'use strict';

const N = 4096;
const EXPECTED = 6647453;

const { performance } = require('perf_hooks');

function time(){
    return performance.now();
}

function Init() {
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

function Lcs(a, b, prev, cur) {
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

function Checksum(d) {
    let sum = 0;
    for(let i = 0; i<d.length; i++) {
        sum += d[i];
    }
    return sum;
}

function Main() {
    let watch = time();
    const { a, b, prev, cur } = Init();
    const init = time() - watch;

    global.gc?.();
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

function delay(time) {
    return new Promise((resolve)=>{
        setTimeout(resolve, time);
    });
}

for(let i = 0, promise = Promise.resolve(); i<100;i++) {
    promise = promise.then(()=>Main()).then(()=>delay(1000));
}
