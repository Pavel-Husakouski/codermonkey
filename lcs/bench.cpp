#include <algorithm>
#include <chrono>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <thread>
#include <utility>
#include <vector>

constexpr int N = 4096;
constexpr int64_t EXPECTED = 6647453;

static const auto start = std::chrono::steady_clock::now();

double time() {
    return std::chrono::duration<double, std::milli>(std::chrono::steady_clock::now() - start).count();
}

struct State {
    std::vector<uint8_t> a, b;
    std::vector<int32_t> prev, cur;
};

// two sequences over a 4-symbol alphabet from a 32-bit LCG, plus zeroed rolling rows
State Init() {
    State st{std::vector<uint8_t>(N), std::vector<uint8_t>(N), std::vector<int32_t>(N + 1), std::vector<int32_t>(N + 1)};
    uint32_t s = 12345;
    for (int i = 0; i < N; i++) {
        s = s * 1664525u + 1013904223u;
        st.a[i] = (s >> 16) & 3;
    }
    for (int i = 0; i < N; i++) {
        s = s * 1664525u + 1013904223u;
        st.b[i] = (s >> 16) & 3;
    }
    return st;
}

void Lcs(State &st) {
    std::vector<int32_t> *prev = &st.prev, *cur = &st.cur;
    for (int i = 1; i <= N; i++) {
        (*cur)[0] = 0;
        for (int j = 1; j <= N; j++) {
            if (st.a[i - 1] == st.b[j - 1])
                (*cur)[j] = (*prev)[j - 1] + 1;
            else
                (*cur)[j] = std::max((*prev)[j], (*cur)[j - 1]);
        }
        std::swap(prev, cur);
    }
    // keep the last row in st.prev
    if (prev != &st.prev)
        st.prev.swap(st.cur);
}

int64_t Checksum(const std::vector<int32_t> &row) {
    int64_t sum = 0;
    for (int32_t v : row)
        sum += v;
    return sum;
}

void Main() {
    double watch = time();
    State st = Init();
    double init = time() - watch;

    watch = time();
    Lcs(st);
    double lcs = time() - watch;

    // checksum stays outside the timing
    int64_t sum = Checksum(st.prev);
    if (sum != EXPECTED) {
        fprintf(stderr, "checksum mismatch: %lld != %lld\n", (long long)sum, (long long)EXPECTED);
        exit(1);
    }

    char line[128];
    snprintf(line, sizeof line, "Init: %.2f ms, Lcs: %.2f ms\n", init, lcs);
    fputs(line, stderr);
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
    for (int i = 0; i < iterations; i++) {
        Main();
        delay(1000);
    }
    return 0;
}
