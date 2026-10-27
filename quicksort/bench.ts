const N = 1000*1000;

// ITER env var overrides the iteration count (default 100)
const ITER = Number(Deno.env.get("ITER") ?? "100");

type Comparator<T> = (x: T, y: T) => number;

function time(): number {
    return performance.now();
}

const byId: Comparator<TestClass> = (x, y) => {
    const a = x.Id, b = y.Id;
    return a === b ? 0 : a > b ? 1 : -1;
};

function Init(number: number): TestClass[] {
    return ToArray(Select(Range(0, number), () => new TestClass()));
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

function QuickSort<T>(vals: T[], compare: Comparator<T>): void {
    QuickSortImpl(vals, compare, 0, vals.length - 1);
}

function QuickSortImpl<T> (vals: T[], compare: Comparator<T>, left: number, right: number): void {
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

function ToArray<T>(iterator: Iterable<T>): T[] {
    return [...iterator];
}

function* Range(start: number, end: number): Generator<number> {
    for(let i = start; i<end; i++) {
        yield i;
    }
}

function* Select<T, U>(iterator: Iterable<T>, projector: (x: T) => U): Generator<U> {
    for(const x of iterator) {
        yield projector(x);
    }
}

function CreateGuid(): string {
    return String(Math.floor(Math.random() * 1e16));
}

function delay(time: number): Promise<void> {
    return new Promise((resolve)=>{
        setTimeout(resolve, time);
    });
}

for(let i = 0, promise = Promise.resolve(); i<ITER;i++) {
    promise = promise.then(()=>Main(N)).then(()=>delay(1000));
}
