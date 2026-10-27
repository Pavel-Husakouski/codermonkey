
const N = 1000*1000;

import { performance } from 'perf_hooks';

function time(): number {
    return performance.now();
}

const byId = (x: TestClass, y: TestClass): number => {
    const a = x.Id, b = y.Id;
    return a === b ? 0 : a > b ? 1 : -1;
};

function Init(number: number): TestClass[] {
    return ToArray(Select(Range(0, number), (_x: number) => new TestClass()));
}

function Sort(vals: TestClass[]): void {
    QuickSort(vals, byId);
}

function Main(number: number): void {
    let watch = time();
    const vals = Init(number);
    const init = time() - watch;

    watch = time();
    Sort(vals);
    const sort = time() - watch;

    console.warn(`Init: ${init.toFixed(2)} ms, Sort: ${sort.toFixed(2)} ms`);
}

function QuickSort(vals: TestClass[], compare: (x: TestClass, y: TestClass) => number): void {
    QuickSortImpl(vals, compare, 0, vals.length - 1);
}

function QuickSortImpl(vals: TestClass[], compare: (x: TestClass, y: TestClass) => number, left: number, right: number): void {
    if (right <= left)
        return;

    let i = left;
    let j = right;
    const mid = (left + right) >>> 1;
    const midVal = vals[mid];

    while (i <= j) {
        while (compare(vals[i], midVal) > 0) {
            i++;
        }

        while (compare(vals[j], midVal) < 0) {
            j--;
        }

        if (i <= j) {
            const tmp = vals[i];
            vals[i] = vals[j];
            vals[j] = tmp;

            i++;
            j--;
        }
    }

    if(left < i - 1) {
        QuickSortImpl(vals, compare, left, i - 1);
    }

    if(right > i) {
        QuickSortImpl(vals, compare, i, right);
    }
}

class TestClass {
    Id: string;
    Value: string;

    constructor() {
        this.Id = CreateGuid();
        this.Value = CreateGuid();
    }

    toString(): string {
        return this.Id;
    }
}

function ToArray<T>(iterator: Generator<T>): T[] {
    return [...iterator];
}

function* Range(start: number, end: number): Generator<number> {
    for(let i = start; i<end; i++) {
        yield i;
    }
}

function* Select<T, U>(iterator: Generator<T>, projector: (x: T) => U): Generator<U> {
    for(let x of iterator) {
        yield projector(x);
    }
}

function CreateGuid(): string {
    return String(Math.floor(Math.random() * 1e16));
}

function delay(time: number): Promise<void> {
    return new Promise<void>((resolve) => {
        setTimeout(resolve, time);
    });
}

const ITER = Number(process.env.ITER ?? "100");
let promise: Promise<void> = Promise.resolve();
for (let i = 0; i < ITER; i++) {
    promise = promise.then(() => Main(N)).then(() => delay(1000));
}
