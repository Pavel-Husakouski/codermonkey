using System;
using System.Diagnostics;
using System.Globalization;
using System.Threading;

class Bench
{
    internal const int N = 512;
    internal const int INF = 1000000000;
    internal const long EXPECTED = 12970375;

    static readonly long start = Stopwatch.GetTimestamp();

    internal static double time()
    {
        return Stopwatch.GetElapsedTime(start).TotalMilliseconds;
    }

    internal static int[] Init()
    {
        int[] d = new int[N * N];
        uint s = 12345;
        unchecked
        {
            for (int i = 0; i < N; i++)
            {
                for (int j = 0; j < N; j++)
                {
                    s = s * 1664525 + 1013904223;
                    int v;
                    if (i == j)
                        v = 0;
                    else if (((s >> 8) % 10) < 7)
                        v = INF;
                    else
                        v = (int)((s >> 16) & 1023) + 1;
                    d[i * N + j] = v;
                }
            }
        }
        return d;
    }

    internal static void Floyd(int[] d)
    {
        for (int k = 0; k < N; k++)
        {
            for (int i = 0; i < N; i++)
            {
                int dik = d[i * N + k];
                for (int j = 0; j < N; j++)
                {
                    int t = dik + d[k * N + j];
                    if (t < d[i * N + j])
                        d[i * N + j] = t;
                }
            }
        }
    }

    internal static long Checksum(int[] d)
    {
        long sum = 0;
        for (int i = 0; i < N * N; i++)
        {
            sum += d[i];
        }
        return sum;
    }

    internal static void Main(int number)
    {
        double watch = time();
        int[] d = Init();
        double init = time() - watch;

        // GC.Collect();
        watch = time();
        Floyd(d);
        double floyd = time() - watch;

        Console.Error.WriteLine($"Init: {init:F2} ms, Floyd: {floyd:F2} ms");

        long sum = Checksum(d);
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
