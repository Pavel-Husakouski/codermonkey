#define _GNU_SOURCE
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

#define N 512
#define INF 1000000000
#define EXPECTED 12970375LL

double time_ms(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (double)ts.tv_sec * 1e3 + (double)ts.tv_nsec / 1e6;
}

// flat N*N matrix from a 32-bit LCG, INF marks a missing edge
void Init(int32_t *d) {
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

int64_t Checksum(const int32_t *d) {
    int64_t sum = 0;
    for (int k = 0; k < N * N; k++) {
        sum += d[k];
    }
    return sum;
}

void Main(void) {
    int32_t *d = malloc((size_t)N * N * sizeof *d);
    double watch = time_ms();
    Init(d);
    double init = time_ms() - watch;

    watch = time_ms();
    Floyd(d);
    double floyd = time_ms() - watch;

    fprintf(stderr, "Init: %.2f ms, Floyd: %.2f ms\n", init, floyd);

    // outside the timing
    int64_t sum = Checksum(d);
    if (sum != EXPECTED) {
        fprintf(stderr, "checksum mismatch: %lld != %lld\n", (long long)sum, EXPECTED);
        exit(1);
    }
    free(d);
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
    for (int i = 0; i < iterations; i++) {
        Main();
        delay(1000);
    }
    return 0;
}
