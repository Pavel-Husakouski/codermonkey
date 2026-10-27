'use strict';

const N = 512;
const INF = 1000000000;
const EXPECTED = 12970375;

const { performance } = require('perf_hooks');

function time(){
    return performance.now();
}

function Init() {
    const d = new Int32Array(N * N);
    let s = 12345;
    for(let i = 0; i<N; i++) {
        for(let j = 0; j<N; j++) {
            s = (Math.imul(s, 1664525) + 1013904223) >>> 0;
            d[i * N + j] = i === j ? 0 : ((s >>> 8) % 10) < 7 ? INF : ((s >>> 16) & 1023) + 1;
        }
    }
    return d;
}

function Floyd(d) {
    for(let k = 0; k<N; k++) {
        for(let i = 0; i<N; i++) {
            const dik = d[i * N + k];
            for(let j = 0; j<N; j++) {
                const t = dik + d[k * N + j];
                if (t < d[i * N + j]) {
                    d[i * N + j] = t;
                }
            }
        }
    }
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
    const d = Init();
    const init = time() - watch;

    global.gc?.();
    watch = time();
    Floyd(d);
    const floyd = time() - watch;

    const sum = Checksum(d);
    if (sum !== EXPECTED) {
        console.error(`Checksum mismatch: ${sum}, expected ${EXPECTED}`);
        process.exit(1);
    }

    console.warn(`Init: ${init.toFixed(2)} ms, Floyd: ${floyd.toFixed(2)} ms`);
}

function delay(time) {
    return new Promise((resolve)=>{
        setTimeout(resolve, time);
    });
}

for(let i = 0, promise = Promise.resolve(); i<100;i++) {
    promise = promise.then(()=>Main()).then(()=>delay(1000));
}
