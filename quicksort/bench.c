#define _GNU_SOURCE
#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#define N (1000 * 1000)

// whole double below 1e17: at most 17 digits plus NUL
#define STR_CAP 18

typedef struct TestClass {
    char Id[STR_CAP];
    char Value[STR_CAP];
} TestClass;

typedef int (*Comparer)(const TestClass *, const TestClass *);

typedef struct Iter {
    int (*next)(struct Iter *self, void **out);
    struct Iter *source;
    void *(*projector)(void *);
    int64_t i, end;
} Iter;

double time_ms(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (double)ts.tv_sec * 1e3 + (double)ts.tv_nsec / 1e6;
}

int byId(const TestClass *x, const TestClass *y) {
    const char *a = x->Id, *b = y->Id;
    return strcmp(a, b);
}

void CreateGuid(char *out);
TestClass *NewTestClass(void);
Iter *Range(int64_t start, int64_t end);
Iter *Select(Iter *iterator, void *(*projector)(void *));
TestClass **ToArray(Iter *iterator, size_t *len);
void QuickSort(TestClass **vals, size_t len, Comparer compare);
void QuickSortImpl(TestClass **vals, Comparer compare, int left, int right);

static void *newTestClassProjector(void *x) {
    (void)x;
    return NewTestClass();
}

TestClass **Init(int number, size_t *len) {
    Iter *range = Range(0, number);
    Iter *select = Select(range, newTestClassProjector);
    TestClass **vals = ToArray(select, len);
    free(select);
    free(range);
    return vals;
}

void Sort(TestClass **vals, size_t len) {
    QuickSort(vals, len, byId);
}

void Main(int number) {
    size_t len;
    double watch = time_ms();
    TestClass **vals = Init(number, &len);
    double init = time_ms() - watch;

    // no gc()
    watch = time_ms();
    Sort(vals, len);
    double sort = time_ms() - watch;

    fprintf(stderr, "Init: %.2f ms, Sort: %.2f ms\n", init, sort);

    for (size_t k = 0; k < len; k++) {
        free(vals[k]);
    }
    free(vals);
}

void QuickSort(TestClass **vals, size_t len, Comparer compare) {
    QuickSortImpl(vals, compare, 0, (int)len - 1);
}

void QuickSortImpl(TestClass **vals, Comparer compare, int left, int right) {
    if (right <= left)
        return;

    int i = left;
    int j = right;
    int mid = (int)(((unsigned)left + (unsigned)right) >> 1);
    TestClass *midVal = vals[mid];

    while (i <= j) {
        while (compare(vals[i], midVal) > 0) {
            i++;
        }

        while (compare(vals[j], midVal) < 0) {
            j--;
        }

        if (i <= j) {
            TestClass *tmp = vals[i];
            vals[i] = vals[j];
            vals[j] = tmp;

            i++;
            j--;
        }
    }

    if (left < i - 1) {
        QuickSortImpl(vals, compare, left, i - 1);
    }

    if (right > i) {
        QuickSortImpl(vals, compare, i, right);
    }
}

TestClass *NewTestClass(void) {
    TestClass *t = malloc(sizeof *t);
    CreateGuid(t->Id);
    CreateGuid(t->Value);
    return t;
}

const char *TestClass_toString(const TestClass *t) {
    return t->Id;
}

TestClass **ToArray(Iter *iterator, size_t *len) {
    size_t n = 0, cap = 16;
    TestClass **arr = malloc(cap * sizeof *arr);
    void *x;
    while (iterator->next(iterator, &x)) {
        if (n == cap) {
            cap *= 2;
            arr = realloc(arr, cap * sizeof *arr);
        }
        arr[n++] = x;
    }
    *len = n;
    return arr;
}

static int rangeNext(Iter *self, void **out) {
    if (self->i >= self->end)
        return 0;
    *out = (void *)(intptr_t)self->i++;
    return 1;
}

Iter *Range(int64_t start, int64_t end) {
    Iter *it = calloc(1, sizeof *it);
    it->next = rangeNext;
    it->i = start;
    it->end = end;
    return it;
}

static int selectNext(Iter *self, void **out) {
    void *x;
    if (!self->source->next(self->source, &x))
        return 0;
    *out = self->projector(x);
    return 1;
}

Iter *Select(Iter *iterator, void *(*projector)(void *)) {
    Iter *it = calloc(1, sizeof *it);
    it->next = selectNext;
    it->source = iterator;
    it->projector = projector;
    return it;
}

// xorshift128+ seeded from the clock, stands in for Math.random()
static uint64_t rng[2];

static void seedRandom(void) {
    struct timespec ts;
    clock_gettime(CLOCK_REALTIME, &ts);
    uint64_t z = (uint64_t)ts.tv_sec * 1000000007ull + (uint64_t)ts.tv_nsec;
    for (int k = 0; k < 2; k++) {
        z += 0x9E3779B97F4A7C15ull;
        uint64_t s = z;
        s = (s ^ (s >> 30)) * 0xBF58476D1CE4E5B9ull;
        s = (s ^ (s >> 27)) * 0x94D049BB133111EBull;
        rng[k] = s ^ (s >> 31);
    }
}

static double randomDouble(void) {
    uint64_t s1 = rng[0];
    const uint64_t s0 = rng[1];
    rng[0] = s0;
    s1 ^= s1 << 23;
    rng[1] = s1 ^ s0 ^ (s1 >> 17) ^ (s0 >> 26);
    return (double)((rng[1] + s0) >> 11) * (1.0 / 9007199254740992.0);
}

// String(v) for a whole double below 1e21
void numberToString(double v, char *out) {
    if (v < 9007199254740992.0) {
        snprintf(out, STR_CAP, "%llu", (unsigned long long)v);
        return;
    }
    char e[64];
    for (int p = 1; p <= 17; p++) {
        snprintf(e, sizeof e, "%.*e", p - 1, v);
        if (strtod(e, NULL) == v)
            break;
    }
    char digits[32];
    int nd = 0;
    char *c = e;
    for (; *c != 'e'; c++) {
        if (*c >= '0' && *c <= '9')
            digits[nd++] = *c;
    }
    int exp10 = atoi(c + 1);
    int total = exp10 + 1;
    if (total >= STR_CAP)
        abort();
    memset(out, '0', (size_t)total);
    memcpy(out, digits, (size_t)(nd < total ? nd : total));
    out[total] = '\0';
}

void CreateGuid(char *out) {
    numberToString(floor(randomDouble() * 1e16), out);
}

void delay(int ms) {
    struct timespec ts = {ms / 1000, (long)(ms % 1000) * 1000000L};
    nanosleep(&ts, NULL);
}

int main(void) {
    int iterations = 100;
    // ITER env var overrides the iteration count (smoke testing only)
    const char *s = getenv("ITER");
    if (s && *s) {
        char *end;
        long n = strtol(s, &end, 10);
        if (*end == '\0')
            iterations = (int)n;
    }
    seedRandom();
    for (int i = 0; i < iterations; i++) {
        Main(N);
        delay(1000);
    }
    return 0;
}
