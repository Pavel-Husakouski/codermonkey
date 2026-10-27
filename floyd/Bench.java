public class Bench {
    static final int N = 512;
    static final int INF = 1000000000;
    static final long EXPECTED = 12970375L;

    static final long start = System.nanoTime();

    static double time() {
        return (System.nanoTime() - start) / 1e6;
    }

    static int[] Init() {
        int[] d = new int[N * N];
        int s = 12345;
        for (int i = 0; i < N; i++) {
            for (int j = 0; j < N; j++) {
                s = s * 1664525 + 1013904223;
                int v;
                if (i == j) v = 0;
                else if (Integer.remainderUnsigned(s >>> 8, 10) < 7) v = INF;
                else v = ((s >>> 16) & 1023) + 1;
                d[i * N + j] = v;
            }
        }
        return d;
    }

    static void Floyd(int[] d) {
        for (int k = 0; k < N; k++) {
            for (int i = 0; i < N; i++) {
                int dik = d[i * N + k];
                for (int j = 0; j < N; j++) {
                    int t = dik + d[k * N + j];
                    if (t < d[i * N + j]) d[i * N + j] = t;
                }
            }
        }
    }

    static long Checksum(int[] d) {
        long sum = 0;
        for (int i = 0; i < d.length; i++) {
            sum += d[i];
        }
        return sum;
    }

    static void Main() {
        double watch = time();
        int[] d = Init();
        double init = time() - watch;

        watch = time();
        Floyd(d);
        double floyd = time() - watch;

        long sum = Checksum(d);
        if (sum != EXPECTED) {
            System.err.printf("Checksum mismatch: %d != %d%n", sum, EXPECTED);
            System.exit(1);
        }

        System.err.printf("Init: %.2f ms, Floyd: %.2f ms%n", init, floyd);
    }

    static void delay(long time) {
        try {
            Thread.sleep(time);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    public static void main(String[] args) {
        int iterations = 100;
        // ITER env var overrides the iteration count (smoke testing only)
        String s = System.getenv("ITER");
        if (s != null && !s.isEmpty()) {
            try {
                iterations = Integer.parseInt(s);
            } catch (NumberFormatException e) {
            }
        }
        for (int i = 0; i < iterations; i++) {
            Main();
            delay(1000);
        }
    }
}
