const N = 4096;
const EXPECTED = 6647453;

// ITER env var overrides the iteration count (default 100)
const ITER = Number(Deno.env.get("ITER") ?? "100");

function time(): number {
    return performance.now();
}

let a: Uint8Array;
let b: Uint8Array;
let prev: Int32Array;
let cur: Int32Array;

function Init(): void {
    a = new Uint8Array(N);
    b = new Uint8Array(N);
    prev = new Int32Array(N + 1);
    cur = new Int32Array(N + 1);
    let s = 12345;

    for (let i = 0; i < N; i++) {
        s = (Math.imul(s, 1664525) + 1013904223) >>> 0;
        a[i] = (s >>> 16) & 3;
    }

    for (let i = 0; i < N; i++) {
        s = (Math.imul(s, 1664525) + 1013904223) >>> 0;
        b[i] = (s >>> 16) & 3;
    }
}

function Lcs(): void {
    for (let i = 1; i <= N; i++) {
        const ai = a[i - 1];
        cur[0] = 0;

        for (let j = 1; j <= N; j++) {
            if (ai === b[j - 1]) {
                cur[j] = prev[j - 1] + 1;
            } else if (prev[j] > cur[j - 1]) {
                cur[j] = prev[j];
            } else {
                cur[j] = cur[j - 1];
            }
        }

        const t = prev;
        prev = cur;
        cur = t;
    }
}

function Checksum(): number {
    let sum = 0;

    for (let j = 0; j <= N; j++) {
        sum += prev[j];
    }

    return sum;
}

function Main(): void {
    let watch = time();
    Init();
    const init = time() - watch;

    watch = time();
    Lcs();
    const lcs = time() - watch;

    const sum = Checksum();
    if (sum !== EXPECTED) {
        console.error(`Checksum mismatch: ${sum} != ${EXPECTED}`);
        Deno.exit(1);
    }

    console.warn(`Init: ${init.toFixed(2)} ms, Lcs: ${lcs.toFixed(2)} ms`);
}

function delay(time: number): Promise<void> {
    return new Promise((resolve)=>{
        setTimeout(resolve, time);
    });
}

for(let i = 0, promise = Promise.resolve(); i<ITER;i++) {
    promise = promise.then(()=>Main()).then(()=>delay(1000));
}
