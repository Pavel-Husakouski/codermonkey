'use strict';

const N = 1000*1000;

const { performance } = require('perf_hooks');

function time(){
    return performance.now();
}

const byId = (x, y) => {
    const a = x.Id, b = y.Id;
    return a === b ? 0 : a > b ? 1 : -1;
};

function Init(number) {
    return ToArray(Select(Range(0, number), x => new TestClass()));
}

function Sort(vals) {
    QuickSort(vals, byId);
}

function Main(number) {
    let watch = time();
    const vals = Init(number);
    const init = time() - watch;

    global.gc?.();
    watch = time();
    Sort(vals);
    const sort = time() - watch;

    console.warn(`Init: ${init.toFixed(2)} ms, Sort: ${sort.toFixed(2)} ms`);
}

function QuickSort(vals, compare){
    QuickSortImpl(vals, compare, 0, vals.length - 1);
}

function QuickSortImpl (vals, compare, left, right) {
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
    constructor() {
        this.Id = CreateGuid();
        this.Value = CreateGuid();
    }

    toString() {
        return this.Id;
    }
}

function ToArray(iterator) {
    return [...iterator];
}

function* Range(start, end) {
    for(let i = start; i<end; i++) {
        yield i;
    }
}

function* Select(iterator, projector) {
    for(let x of iterator) {
        yield projector(x);
    }
}

function CreateGuid() {
    return String(Math.floor(Math.random() * 1e16));
}

function delay(time) {
    return new Promise((resolve)=>{
        setTimeout(resolve, time);
    });
}

for(let i = 0, promise = Promise.resolve(); i<100;i++) {
    promise = promise.then(()=>Main(N)).then(()=>delay(1000));
}
