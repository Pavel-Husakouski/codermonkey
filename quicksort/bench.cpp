#include <charconv>
#include <chrono>
#include <cmath>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <optional>
#include <string>
#include <thread>
#include <utility>
#include <vector>

constexpr int N = 1000 * 1000;

static const auto start = std::chrono::steady_clock::now();

double time() {
    return std::chrono::duration<double, std::milli>(std::chrono::steady_clock::now() - start).count();
}

class TestClass;
using Comparer = int (*)(const TestClass *, const TestClass *);

std::string CreateGuid();

class TestClass {
public:
    std::string Id;
    std::string Value;

    TestClass() : Id(CreateGuid()), Value(CreateGuid()) {}

    const std::string &toString() const {
        return Id;
    }
};

int byId(const TestClass *x, const TestClass *y) {
    const std::string &a = x->Id, &b = y->Id;
    int c = a.compare(b);
    return (c > 0) - (c < 0);
}

class Range {
public:
    Range(int64_t start, int64_t end) : i(start), end(end) {}

    bool next(int64_t *out) {
        if (i >= end)
            return false;
        *out = i++;
        return true;
    }

private:
    int64_t i, end;
};

class Select {
public:
    Select(Range *iterator, TestClass (*projector)(int64_t)) : source(iterator), projector(projector) {}

    bool next(std::optional<TestClass> *out) {
        int64_t x;
        if (!source->next(&x))
            return false;
        *out = projector(x);
        return true;
    }

private:
    Range *source;
    TestClass (*projector)(int64_t);
};

std::vector<TestClass> ToArray(Select &iterator) {
    std::vector<TestClass> arr;
    std::optional<TestClass> x;
    while (iterator.next(&x))
        arr.push_back(std::move(*x));
    return arr;
}

static TestClass newTestClass(int64_t) {
    return TestClass();
}

std::vector<TestClass> Init(int number) {
    Range range(0, number);
    Select select(&range, newTestClass);
    return ToArray(select);
}

void QuickSortImpl(TestClass *vals, Comparer compare, int left, int right) {
    if (right <= left)
        return;

    int i = left;
    int j = right;
    int mid = (int)(((unsigned)left + (unsigned)right) >> 1);
    TestClass midVal = vals[mid];

    while (i <= j) {
        while (compare(&vals[i], &midVal) > 0) {
            i++;
        }

        while (compare(&vals[j], &midVal) < 0) {
            j--;
        }

        if (i <= j) {
            std::swap(vals[i], vals[j]);

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

void QuickSort(std::vector<TestClass> &vals, Comparer compare) {
    QuickSortImpl(vals.data(), compare, 0, (int)vals.size() - 1);
}

void Sort(std::vector<TestClass> &vals) {
    QuickSort(vals, byId);
}

void Main(int number) {
    double watch = time();
    std::vector<TestClass> vals = Init(number);
    double init = time() - watch;

    // no gc()
    watch = time();
    Sort(vals);
    double sort = time() - watch;

    char line[128];
    snprintf(line, sizeof line, "Init: %.2f ms, Sort: %.2f ms\n", init, sort);
    fputs(line, stderr);
}

// xorshift128+ seeded from the clock, stands in for Math.random()
static uint64_t rng[2];

static void seedRandom() {
    uint64_t z = (uint64_t)std::chrono::system_clock::now().time_since_epoch().count();
    for (int k = 0; k < 2; k++) {
        z += 0x9E3779B97F4A7C15ull;
        uint64_t s = z;
        s = (s ^ (s >> 30)) * 0xBF58476D1CE4E5B9ull;
        s = (s ^ (s >> 27)) * 0x94D049BB133111EBull;
        rng[k] = s ^ (s >> 31);
    }
}

static double randomDouble() {
    uint64_t s1 = rng[0];
    const uint64_t s0 = rng[1];
    rng[0] = s0;
    s1 ^= s1 << 23;
    rng[1] = s1 ^ s0 ^ (s1 >> 17) ^ (s0 >> 26);
    return (double)((rng[1] + s0) >> 11) * (1.0 / 9007199254740992.0);
}

// String(v) for a whole double below 1e21
std::string numberToString(double v) {
    if (v < 9007199254740992.0)
        return std::to_string((unsigned long long)v);
    char buf[64];
    auto r = std::to_chars(buf, buf + sizeof buf, v, std::chars_format::scientific);
    *r.ptr = 0;
    char *e = buf;
    std::string digits;
    for (; e != r.ptr && *e != 'e'; e++) {
        if (*e >= '0' && *e <= '9')
            digits += *e;
    }
    int exp10 = atoi(e + 1);
    size_t total = (size_t)exp10 + 1;
    std::string s(total, '0');
    s.replace(0, digits.size() < total ? digits.size() : total, digits, 0, total);
    return s;
}

std::string CreateGuid() {
    return numberToString(std::floor(randomDouble() * 1e16));
}

void delay(int ms) {
    std::this_thread::sleep_for(std::chrono::milliseconds(ms));
}

int main() {
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
