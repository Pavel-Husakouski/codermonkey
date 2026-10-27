using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Globalization;
using System.Threading;

// Same benchmark as Bench.cs, rewritten for zero-cost abstractions:
// array of structs, struct comparer and struct iterators as generic arguments,
// so every call is resolved statically and inlined (no delegates, no interface dispatch).
class Bench
{
    internal const int N = 1000 * 1000;

    static readonly long start = Stopwatch.GetTimestamp();

    internal static double time()
    {
        return Stopwatch.GetElapsedTime(start).TotalMilliseconds;
    }

    internal struct ById : IComparer<TestClass>
    {
        public int Compare(TestClass x, TestClass y)
        {
            return string.CompareOrdinal(x.Id, y.Id);
        }
    }

    internal static readonly ById byId = default;

    internal struct NewTestClass : IFunc<int, TestClass>
    {
        public TestClass Invoke(int x)
        {
            return new TestClass();
        }
    }

    internal static TestClass[] Init(int number)
    {
        var select = Select<RangeIter, int, TestClass, NewTestClass>(Range(0, number), default);
        return ToArray<SelectIter<RangeIter, int, TestClass, NewTestClass>, TestClass>(select);
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

    internal static void QuickSort<T, TCmp>(Span<T> vals, TCmp compare) where TCmp : struct, IComparer<T>
    {
        QuickSortImpl(vals, compare, 0, vals.Length - 1);
    }

    static void QuickSortImpl<T, TCmp>(Span<T> vals, TCmp compare, int left, int right) where TCmp : struct, IComparer<T>
    {
        if (right <= left)
            return;

        int i = left;
        int j = right;
        int mid = (int)((uint)(left + right) >> 1);
        T midVal = vals[mid];

        while (i <= j)
        {
            while (compare.Compare(vals[i], midVal) > 0)
            {
                i++;
            }

            while (compare.Compare(vals[j], midVal) < 0)
            {
                j--;
            }

            if (i <= j)
            {
                ref T a = ref vals[i];
                ref T b = ref vals[j];
                (a, b) = (b, a);

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

    internal readonly struct TestClass
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

    internal interface IFunc<T, U>
    {
        U Invoke(T x);
    }

    // Pull iterator with a size hint, so ToArray can allocate once.
    internal interface IIter<T>
    {
        int Count { get; }
        bool Next(out T x);
    }

    internal struct RangeIter : IIter<int>
    {
        int i;
        readonly int end;

        public RangeIter(int start, int end)
        {
            this.i = start;
            this.end = end;
        }

        public int Count => end - i;

        public bool Next(out int x)
        {
            x = i;
            if (i >= end)
                return false;
            i++;
            return true;
        }
    }

    internal struct SelectIter<TSrc, T, U, TFunc> : IIter<U>
        where TSrc : struct, IIter<T>
        where TFunc : struct, IFunc<T, U>
    {
        TSrc source;
        readonly TFunc projector;

        public SelectIter(TSrc source, TFunc projector)
        {
            this.source = source;
            this.projector = projector;
        }

        public int Count => source.Count;

        public bool Next(out U x)
        {
            if (!source.Next(out T t))
            {
                x = default!;
                return false;
            }
            x = projector.Invoke(t);
            return true;
        }
    }

    static T[] ToArray<TIter, T>(TIter iterator) where TIter : struct, IIter<T>
    {
        T[] arr = new T[iterator.Count];
        int n = 0;
        while (iterator.Next(out T x))
        {
            arr[n++] = x;
        }
        return arr;
    }

    static RangeIter Range(int start, int end)
    {
        return new RangeIter(start, end);
    }

    static SelectIter<TSrc, T, U, TFunc> Select<TSrc, T, U, TFunc>(TSrc iterator, TFunc projector)
        where TSrc : struct, IIter<T>
        where TFunc : struct, IFunc<T, U>
    {
        return new SelectIter<TSrc, T, U, TFunc>(iterator, projector);
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
