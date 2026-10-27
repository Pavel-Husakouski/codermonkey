import java.math.BigDecimal;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.Iterator;
import java.util.NoSuchElementException;
import java.util.concurrent.ThreadLocalRandom;
import java.util.function.Function;

public class Bench {
    static final int N = 1000 * 1000;

    static final long start = System.nanoTime();

    static double time() {
        return (System.nanoTime() - start) / 1e6;
    }

    static final Comparator<TestClass> byId = (x, y) -> {
        String a = x.Id, b = y.Id;
        return a.compareTo(b);
    };

    static TestClass[] Init(int number) {
        return ToArray(Select(Range(0, number), x -> new TestClass()), new TestClass[0]);
    }

    static void Sort(TestClass[] vals) {
        QuickSort(vals, byId);
    }

    static void Main(int number) {
        double watch = time();
        TestClass[] vals = Init(number);
        double init = time() - watch;

        // System.gc();
        watch = time();
        Sort(vals);
        double sort = time() - watch;

        System.err.printf("Init: %.2f ms, Sort: %.2f ms%n", init, sort);
    }

    static void QuickSort(TestClass[] vals, Comparator<TestClass> compare) {
        QuickSortImpl(vals, compare, 0, vals.length - 1);
    }

    static void QuickSortImpl(TestClass[] vals, Comparator<TestClass> compare, int left, int right) {
        if (right <= left)
            return;

        int i = left;
        int j = right;
        int mid = (left + right) >>> 1;
        TestClass midVal = vals[mid];

        while (i <= j) {
            while (compare.compare(vals[i], midVal) > 0) {
                i++;
            }

            while (compare.compare(vals[j], midVal) < 0) {
                j--;
            }

            if (i <= j) {
                TestClass tmp = vals[i];
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

    static class TestClass {
        final String Id;
        final String Value;

        TestClass() {
            this.Id = CreateGuid();
            this.Value = CreateGuid();
        }

        @Override
        public String toString() {
            return this.Id;
        }
    }

    static <T> T[] ToArray(Iterator<T> iterator, T[] proto) {
        ArrayList<T> list = new ArrayList<>();
        while (iterator.hasNext()) {
            list.add(iterator.next());
        }
        return list.toArray(proto);
    }

    static Iterator<Integer> Range(int start, int end) {
        return new Iterator<>() {
            int i = start;

            public boolean hasNext() {
                return i < end;
            }

            public Integer next() {
                if (i >= end) throw new NoSuchElementException();
                return i++;
            }
        };
    }

    static <T, U> Iterator<U> Select(Iterator<T> iterator, Function<T, U> projector) {
        return new Iterator<>() {
            public boolean hasNext() {
                return iterator.hasNext();
            }

            public U next() {
                return projector.apply(iterator.next());
            }
        };
    }

    static String CreateGuid() {
        return format(Math.floor(ThreadLocalRandom.current().nextDouble() * 1e16));
    }

    // String(v) of JS for a whole double below 1e21: shortest digits, zero padded.
    static String format(double v) {
        if (v < 9007199254740992.0) {
            return Long.toString((long) v);
        }
        return new BigDecimal(Double.toString(v)).stripTrailingZeros().toPlainString();
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
            Main(N);
            delay(1000);
        }
    }
}
