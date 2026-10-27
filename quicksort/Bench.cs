using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Globalization;
using System.Threading;

class Bench
{
    internal const int N = 1000 * 1000;

    static readonly long start = Stopwatch.GetTimestamp();

    internal static double time()
    {
        return Stopwatch.GetElapsedTime(start).TotalMilliseconds;
    }

    internal static readonly Comparison<TestClass> byId = (x, y) =>
    {
        string a = x.Id, b = y.Id;
        return string.CompareOrdinal(a, b);
    };

    internal static TestClass[] Init(int number)
    {
        return ToArray(Select(Range(0, number), x => new TestClass()));
    }

    internal static void Sort(TestClass[] vals)
    {
        QuickSort(vals, byId);
    }

    internal static void Main(int number)
    {
        double watch = time();
        TestClass[] vals = Init(number);
        double init = time() - watch;

        // GC.Collect();
        watch = time();
        Sort(vals);
        double sort = time() - watch;

        Console.Error.WriteLine($"Init: {init:F2} ms, Sort: {sort:F2} ms");
    }

    internal static void QuickSort(TestClass[] vals, Comparison<TestClass> compare)
    {
        QuickSortImpl(vals, compare, 0, vals.Length - 1);
    }

    static void QuickSortImpl(TestClass[] vals, Comparison<TestClass> compare, int left, int right)
    {
        if (right <= left)
            return;

        int i = left;
        int j = right;
        int mid = (int)((uint)(left + right) >> 1);
        TestClass midVal = vals[mid];

        while (i <= j)
        {
            while (compare(vals[i], midVal) > 0)
            {
                i++;
            }

            while (compare(vals[j], midVal) < 0)
            {
                j--;
            }

            if (i <= j)
            {
                TestClass tmp = vals[i];
                vals[i] = vals[j];
                vals[j] = tmp;

                i++;
                j--;
            }
        }

        if (left < i - 1)
        {
            QuickSortImpl(vals, compare, left, i - 1);
        }

        if (right > i)
        {
            QuickSortImpl(vals, compare, i, right);
        }
    }

    internal class TestClass
    {
        public readonly string Id;
        public readonly string Value;

        public TestClass()
        {
            this.Id = CreateGuid();
            this.Value = CreateGuid();
        }

        public override string ToString()
        {
            return this.Id;
        }
    }

    static T[] ToArray<T>(IEnumerable<T> iterator)
    {
        return [.. iterator];
    }

    static IEnumerable<int> Range(int start, int end)
    {
        for (int i = start; i < end; i++)
        {
            yield return i;
        }
    }

    static IEnumerable<U> Select<T, U>(IEnumerable<T> iterator, Func<T, U> projector)
    {
        foreach (T x in iterator)
        {
            yield return projector(x);
        }
    }

    static string CreateGuid()
    {
        return Format(Math.Floor(Random.Shared.NextDouble() * 1e16));
    }

    // String(v) of JS for a whole double below 1e21: shortest digits, zero padded.
    internal static string Format(double v)
    {
        if (v < 9007199254740992.0)
        {
            return ((long)v).ToString(CultureInfo.InvariantCulture);
        }

        string r = v.ToString("R", CultureInfo.InvariantCulture);
        int e = r.IndexOf('E');
        if (e < 0)
            return r;

        string mantissa = r.Substring(0, e);
        int exp = int.Parse(r.Substring(e + 1), CultureInfo.InvariantCulture);
        int dot = mantissa.IndexOf('.');
        string digits = mantissa.Replace(".", "");
        int frac = dot < 0 ? 0 : mantissa.Length - dot - 1;
        int zeros = exp - frac;
        return zeros > 0 ? digits + new string('0', zeros) : digits;
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
