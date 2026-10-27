#define _GNU_SOURCE
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

#define N 4096
#define EXPECTED 6647453LL

double time_ms(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (double)ts.tv_sec * 1e3 + (double)ts.tv_nsec / 1e6;
}

// two sequences over a 4-symbol alphabet from a 32-bit LCG
void Init(uint8_t *a, uint8_t *b) {
    uint32_t s = 12345;
    for (int i = 0; i < N; i++) {
        s = s * 1664525u + 1013904223u;
        a[i] = (s >> 16) & 3;
    }
    for (int i = 0; i < N; i++) {
        s = s * 1664525u + 1013904223u;
        b[i] = (s >> 16) & 3;
    }
}

// two rolling rows, returns the row holding the last result
int32_t *Lcs(const uint8_t *a, const uint8_t *b, int32_t *prev, int32_t *cur) {
    for (int i = 1; i <= N; i++) {
        cur[0] = 0;
        for (int j = 1; j <= N; j++) {
            if (a[i - 1] == b[j - 1])
                cur[j] = prev[j - 1] + 1;
            else
                cur[j] = prev[j] > cur[j - 1] ? prev[j] : cur[j - 1];
        }
        int32_t *t = prev;
        prev = cur;
        cur = t;
    }
    return prev;
}

int64_t Checksum(const int32_t *row) {
    int64_t sum = 0;
    for (int j = 0; j <= N; j++) {
        sum += row[j];
    }
    return sum;
}

void Main(void) {
    double watch = time_ms();
    uint8_t *a = malloc(N);
    uint8_t *b = malloc(N);
    int32_t *prev = calloc(N + 1, sizeof *prev);
    int32_t *cur = calloc(N + 1, sizeof *cur);
    Init(a, b);
    double init = time_ms() - watch;

    watch = time_ms();
    int32_t *res = Lcs(a, b, prev, cur);
    double lcs = time_ms() - watch;

    fprintf(stderr, "Init: %.2f ms, Lcs: %.2f ms\n", init, lcs);

    // outside the timing
    int64_t sum = Checksum(res);
    if (sum != EXPECTED) {
        fprintf(stderr, "checksum mismatch: %lld != %lld\n", (long long)sum, EXPECTED);
        exit(1);
    }
    free(a);
    free(b);
    free(prev);
    free(cur);
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
