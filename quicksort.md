# quicksort

- date: 2026-10-27 13:48
- cpu: AMD Ryzen 7 260 w/ Radeon 780M Graphics (16 threads)
- os: linux/amd64
- iterations: 10
- times in ms; sorted by Sort median

| # | lang | version | init median | Sort first | Sort min | Sort median | Sort mean | Sort p95 | x best | status |
|--:|------|---------|------------:|------:|------:|------:|------:|------:|------:|--------|
| 1 | deno | 2.9.6 | 219.47 | 880.89 | 748.60 | 844.94 | 876.42 | 1063.88 | 1.00x | ok |
| 2 | js | 26.12.0 | 232.03 | 959.96 | 786.63 | 849.64 | 848.70 | 959.96 | 1.01x | ok |
| 3 | bun | 1.4.0 | 159.09 | 1046.17 | 1034.27 | 1096.37 | 1110.39 | 1248.69 | 1.30x | ok |
