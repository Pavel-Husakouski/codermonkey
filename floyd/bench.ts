const N = 512;
const INF = 1000000000;
const EXPECTED = 12970375;

// ITER env var overrides the iteration count (default 100)
const ITER = Number(Deno.env.get("ITER") ?? "100");

function time(): number {
    return performance.now();
}

function Init(): Int32Array {
    const d = new Int32Array(N * N);
    let s = 12345;

    for (let i = 0; i < N; i++) {
        for (let j = 0; j < N; j++) {
            s = (Math.imul(s, 1664525) + 1013904223) >>> 0;

            if (i === j) {
                d[i * N + j] = 0;
            } else if (((s >>> 8) % 10) < 7) {
                d[i * N + j] = INF;
            } else {
                d[i * N + j] = ((s >>> 16) & 1023) + 1;
            }
        }
    }

    return d;
}

function Floyd(d: Int32Array): void {
    for (let k = 0; k < N; k++) {
        for (let i = 0; i < N; i++) {
            const dik = d[i * N + k];

            for (let j = 0; j < N; j++) {
                const t = dik + d[k * N + j];

                if (t < d[i * N + j]) {
                    d[i * N + j] = t;
                }
            }
        }
    }
}

function Checksum(d: Int32Array): number {
    let sum = 0;

    for (let i = 0; i < d.length; i++) {
        sum += d[i];
    }

    return sum;
}

function Main(): void {
    let watch = time();
    const d = Init();
    const init = time() - watch;

    watch = time();
    Floyd(d);
    const floyd = time() - watch;

    const sum = Checksum(d);
    if (sum !== EXPECTED) {
        console.error(`Checksum mismatch: ${sum} != ${EXPECTED}`);
        Deno.exit(1);
    }

    console.warn(`Init: ${init.toFixed(2)} ms, Floyd: ${floyd.toFixed(2)} ms`);
}

function delay(time: number): Promise<void> {
    return new Promise((resolve)=>{
        setTimeout(resolve, time);
    });
}

for(let i = 0, promise = Promise.resolve(); i<ITER;i++) {
    promise = promise.then(()=>Main()).then(()=>delay(1000));
}
