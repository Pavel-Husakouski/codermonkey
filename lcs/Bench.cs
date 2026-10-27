using System;
using System.Diagnostics;
using System.Globalization;
using System.Threading;

class Bench
{
    internal const int N = 4096;
    internal const long EXPECTED = 6647453;

    static readonly long start = Stopwatch.GetTimestamp();

    internal static double time()
    {
        return Stopwatch.GetElapsedTime(start).TotalMilliseconds;
    }

    internal static void Init(byte[] a, byte[] b)
    {
        uint s = 12345;
        unchecked
        {
            for (int i = 0; i < N; i++)
            {
                s = s * 1664525 + 1013904223;
                a[i] = (byte)((s >> 16) & 3);
            }
            for (int i = 0; i < N; i++)
            {
                s = s * 1664525 + 1013904223;
                b[i] = (byte)((s >> 16) & 3);
            }
        }
    }

    // two rolling rows; returns the final prev row
    internal static int[] Lcs(byte[] a, byte[] b, int[] prev, int[] cur)
    {
        for (int i = 1; i <= N; i++)
        {
            cur[0] = 0;
            byte ai = a[i - 1];
            for (int j = 1; j <= N; j++)
            {
                if (ai == b[j - 1])
                    cur[j] = prev[j - 1] + 1;
                else
                    cur[j] = Math.Max(prev[j], cur[j - 1]);
            }
            int[] t = prev;
            prev = cur;
            cur = t;
        }
        return prev;
    }

    internal static long Checksum(int[] row)
    {
        long sum = 0;
        for (int j = 0; j <= N; j++)
        {
            sum += row[j];
        }
        return sum;
    }

    internal static void Main(int number)
    {
        double watch = time();
        byte[] a = new byte[N];
        byte[] b = new byte[N];
        Init(a, b);
        int[] prev = new int[N + 1];
        int[] cur = new int[N + 1];
        double init = time() - watch;

        // GC.Collect();
        watch = time();
        int[] res = Lcs(a, b, prev, cur);
        double lcs = time() - watch;

        Console.Error.WriteLine($"Init: {init:F2} ms, Lcs: {lcs:F2} ms");

        long sum = Checksum(res);
        if (sum != EXPECTED)
        {
            Console.Error.WriteLine($"Checksum mismatch: {sum} != {EXPECTED}");
            Environment.Exit(1);
        }
    }

    static void delay(int time)
    {
        Thread.Sleep(time);
    }

    static void Main()
    {
        CultureInfo.DefaultThreadCurrentCulture = CultureInfo.InvariantCulture;
        CultureInfo.CurrentCulture = CultureInfo.InvariantCulture;

        int iterations = 100;
        // ITER env var overrides the iteration count (smoke testing only)
        string? s = Environment.GetEnvironmentVariable("ITER");
        if (!string.IsNullOrEmpty(s) && int.TryParse(s, out int n))
        {
            iterations = n;
        }
        for (int i = 0; i < iterations; i++)
        {
            Main(N);
            delay(1000);
        }
    }
}
