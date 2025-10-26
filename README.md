# Go - 1 Billion Rows Challenge

## Why

I am quite late to the party, pretty much everyone already solved this in all languages. But for me this is just one opportunity to get a bit deeper Go knowledge.

## How

Going to create a very basic implementation. Then using data from the profiler to address slowest parts of the code until I get good enough result on my machine.

My environment:

- Intel 1250U, NVME SSD, 32 Gb
- Ubuntu 22.04 under WSL2 (Windows 11)
- On battery in `Most power efficiency mode`
- Go 1.25.3

In profiler I normally looked at top 5 `go tool pprof profiler` then `top5`

## Initial Data

File for challenge was generated using original [1brc java project](https://github.com/gunnarmorling/1brc).

I used `Serkan Ozal` result as a target, his official result is `00:01.880`, but on my machine

```bash
time ./calculate_average_serkan-ozal.sh

real    0m3.926s
user    0m0.005s
sys     0m0.014s
```

so my machine is about two times slower.

I picked their result because this is the top one that does not use GraalVM, but JVM.

## My code

### Implementation 1

- Read file in 64 MB chunks
- Use `bytes.Index`, `bytes.Split`, `strconv.ParseFloat` and `map` to store results
- 4 threads

Result

```
real    1m40.605s
user    6m19.778s
sys     0m25.630s
```

Top 5 from profiler

```
Duration: 100.19s, Total samples = 391.71s (390.96%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top5
Showing nodes accounting for 112.79s, 28.79% of 391.71s total
Dropped 293 nodes (cum <= 1.96s)
Showing top 5 nodes out of 67
      flat  flat%   sum%        cum   cum%
    39.29s 10.03% 10.03%     39.52s 10.09%  strconv.readFloat
    20.06s  5.12% 15.15%     65.95s 16.84%  runtime.mapaccess2_faststr
    18.02s  4.60% 19.75%     18.02s  4.60%  runtime.nextFreeFast (inline)
    17.94s  4.58% 24.33%     33.35s  8.51%  runtime.mallocgcTiny
    17.48s  4.46% 28.79%     17.48s  4.46%  aeshashbody
(pprof)
```

Next step is to improve parsing float.

### Implementation 2

Replaced `ParseFloat` with custom impl. Runtime is down to 75 seconds, but custom parsing float is still the slowest operation

```
real    1m16.288s
user    4m42.292s
sys     0m23.970s
```

profiler

```
Duration: 75.31s, Total samples = 291.89s (387.58%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top5
Showing nodes accounting for 89.70s, 30.73% of 291.89s total
Dropped 273 nodes (cum <= 1.46s)
Showing top 5 nodes out of 67
      flat  flat%   sum%        cum   cum%
    25.17s  8.62%  8.62%     25.19s  8.63%  main.parseFloatCustom
    18.31s  6.27% 14.90%     61.84s 21.19%  runtime.mapaccess2_faststr
    16.28s  5.58% 20.47%     16.28s  5.58%  aeshashbody
    15.29s  5.24% 25.71%     15.29s  5.24%  runtime.nextFreeFast (inline)
    14.65s  5.02% 30.73%    101.98s 34.94%  bytes.genSplit
```
