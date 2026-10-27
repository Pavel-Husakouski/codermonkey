public class Bench {
    static final int N = 4096;
    static final long EXPECTED = 6647453L;

    static final long start = System.nanoTime();

    static double time() {
        return (System.nanoTime() - start) / 1e6;
    }

    static byte[] Gen(int[] state) {
        byte[] a = new byte[N];
        int s = state[0];
        for (int i = 0; i < N; i++) {
            s = s * 1664525 + 1013904223;
            a[i] = (byte) ((s >>> 16) & 3);
        }
        state[0] = s;
        return a;
    }

    static int[] Lcs(byte[] a, byte[] b, int[] prev, int[] cur) {
        for (int i = 1; i <= N; i++) {
            cur[0] = 0;
            for (int j = 1; j <= N; j++) {
                if (a[i - 1] == b[j - 1]) cur[j] = prev[j - 1] + 1;
                else cur[j] = Math.max(prev[j], cur[j - 1]);
            }
            int[] t = prev;
            prev = cur;
            cur = t;
        }
        return prev;
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
        int[] state = {12345};
        byte[] a = Gen(state);
        byte[] b = Gen(state);
        int[] prev = new int[N + 1];
        int[] cur = new int[N + 1];
        double init = time() - watch;

        watch = time();
        int[] last = Lcs(a, b, prev, cur);
        double lcs = time() - watch;

        long sum = Checksum(last);
        if (sum != EXPECTED) {
            System.err.printf("Checksum mismatch: %d != %d%n", sum, EXPECTED);
            System.exit(1);
        }

        System.err.printf("Init: %.2f ms, Lcs: %.2f ms%n", init, lcs);
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
