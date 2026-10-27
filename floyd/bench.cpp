#include <chrono>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <thread>
#include <vector>

constexpr int N = 512;
constexpr int32_t INF = 1000000000;
constexpr int64_t EXPECTED = 12970375;

static const auto start = std::chrono::steady_clock::now();

double time() {
    return std::chrono::duration<double, std::milli>(std::chrono::steady_clock::now() - start).count();
}

// flat N*N matrix from a 32-bit LCG: ~70% missing edges, the rest weights 1..1024
std::vector<int32_t> Init() {
    std::vector<int32_t> d(N * N);
    uint32_t s = 12345;
    for (int i = 0; i < N; i++) {
        for (int j = 0; j < N; j++) {
            s = s * 1664525u + 1013904223u;
            int32_t v;
            if (i == j)
                v = 0;
            else if (((s >> 8) % 10) < 7)
                v = INF;
            else
                v = (int32_t)((s >> 16) & 1023) + 1;
            d[i * N + j] = v;
        }
    }
    return d;
}

void Floyd(int32_t *d) {
    for (int k = 0; k < N; k++) {
        for (int i = 0; i < N; i++) {
            int32_t dik = d[i * N + k];
            for (int j = 0; j < N; j++) {
                int32_t t = dik + d[k * N + j];
                if (t < d[i * N + j])
                    d[i * N + j] = t;
            }
        }
    }
}

int64_t Checksum(const std::vector<int32_t> &d) {
    int64_t sum = 0;
    for (int32_t v : d)
        sum += v;
    return sum;
}

void Main() {
    double watch = time();
    std::vector<int32_t> d = Init();
    double init = time() - watch;

    watch = time();
    Floyd(d.data());
    double floyd = time() - watch;

    // checksum stays outside the timing
    int64_t sum = Checksum(d);
    if (sum != EXPECTED) {
        fprintf(stderr, "checksum mismatch: %lld != %lld\n", (long long)sum, (long long)EXPECTED);
        exit(1);
    }

    char line[128];
    snprintf(line, sizeof line, "Init: %.2f ms, Floyd: %.2f ms\n", init, floyd);
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
