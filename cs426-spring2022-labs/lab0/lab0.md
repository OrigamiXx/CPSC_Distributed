# [Initial version] CS426 Lab 0: Introduction to Go and multi-threaded programming

## Overview
The goal of this lab is to familiarize you with the Go programming language, multi-threaded programming, and the concepts and practices of concurrency and parallelization. Have fun!

## Logistics
**Policies**
- Lab 0 is meant to be an **individual** assignment. Please see the [Collaboration Policy](#collaboration-policy) for details.
- We will help you strategize how to debug but WE WILL NOT DEBUG YOUR CODE FOR YOU.
- Please keep and submit a time log of time spent and major challenges you've encountered. This may be familiar to you if you've taken CS323. See [Time logging](#time-logging) for details.

- Questions? post to Canvas or email the teaching staff at cs426ta@cs.yale.edu.
  - Richard Yang (yry@cs.yale.edu)
  - Xiao Shi (xiao.shi@aya.yale.edu)
  - Scott Pruett (spruett345@gmail.com) (unavailable until Jan 30th)

**Submission deadline: 23:59 ET Wednesday Feb 9, 2022**

**Submission logistics** TBA.

Your submission for this lab should include the following files:
```
discussions.txt
profile_striped_num_cpu_times_two.png
string_set.go
striped_string_set.go
time.log
```

## Preparation
1. Install Go (or run Go on a Zoo machine) following [this guide]().
2. Complete the [Tour of Go](https://go.dev/tour/list). Pay special attention to the concurrency module. Use the exercises in the tutorial as checkpoints of your Go familiarity, but we won't as you to submit your solutions or grade them. That said, skip them at your own peril as all of the labs this semster will be in Go.
3. Check out the list of [Go tips and FAQs](#go-tips-and-faqs) as you set up your development workflow.

## Part A. Thread-safe add-only string set
Build a data structure `LockedStringSet` which implements the `StringSet` interface (in `string_set.go`). Use `sync.RWMutex` to ensure that the data structure is thread-safe.

You can now run the provided simple unittest and possibly add your own. `-v` turns on the verbose flag which provides more information about the test run. `-race` turns on the [Go race detector](https://go.dev/doc/articles/race_detector), which is a tool that detects data races.
```
go test -run 'TestLocked/locked/(simple|concurrent)' -race -v
```
(You can run the test for `PredRange` after you've completed it in Part C.)

## Part B. Lock striping

**Benchmarks** and **Profilers** are important tools to evaluate performance of one's code, protocol, or systems. Single-machine, small-scale benchmarks are often called _micro-benchmarks_. Profilers collect the runtime metrics of a program (often a benchmark) to guide performance optimizations. The most common profiles include CPU profiles and memory profiles. Benchmarking and profiling can be somewhat of a minefield since it's easy to be measuring the wrong thing or drawing incorrect conclusions, e.g., sometimes one's benchmark setup take more CPU cycles to run than the core logic of interest. We will only get a taste of it in this lab.

The following command runs a selected set of benchmarks. Specifically, `-bench=LockedStringSet/adds$` selects benchmarks whose names match `LockedStringSet` (can also be a regex) and whose sub-benchmark names match `adds$`. `-run=^$` prevents tests in the same file from running such that the profiling results only contains traces while the benchmark(s) are running and not the tests.
```
go test -v -bench=LockedStringSet/adds$ -run=^$

goos: darwin
goarch: amd64
pkg: cs426.yale.edu/lab0
cpu: VirtualApple @ 2.50GHz
BenchmarkLockedStringSet
BenchmarkLockedStringSet/adds
BenchmarkLockedStringSet/adds-10                 8497728               136.6 ns/op
PASS
ok      cs426.yale.edu/lab0     1.534s
```

This command outputs the latency for the operation (`136.6 ns/op`) as well as the number of iterations run (`8497728`) to obtain this stat. The `-10` at the end of the benchmark name is the number of cores used for this benchmark--this output was generated on a 10-core M1. depending on the machine, this will differ.

When you are benchmarking, it's convenient to run multiple benchmarks in a single run. `go test` treats `/` in `-bench` or `-run` as delimiter of test/benchmark names and sub-test/sub-benchmark names, but treat the `/` delimited compoenents as regexes. You can read more on sub-tests and sub-benchmarks [here](https://pkg.go.dev/testing#hdr-Subtests_and_Sub_benchmarks).

In contrast, when you are profiling, limit to running a single benchmark, otherwise your profiling result will show traces for all benchmarks that are run during the `go test` invocation. The following command collects the CPU profile, and outputs the result into `profile.out`.
```
go test -v -bench=LockedStringSet/adds$ -run=^$ -cpuprofile profile.out
```

To visualize the profile, run `go tool pprof profile.out`. type `top` to view the most time consuming functions, and `png` to generate a graph (you need to install [graphviz](https://graphviz.org/download/)). The lab0 repo includes one such example graph.
```
$  go tool pprof profile.out

Type: cpu
Time: ... (EST)
Duration: 1.50s, Total samples = 4.42s (294.13%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 3800ms, 85.97% of 4420ms total
Dropped 25 nodes (cum <= 22.10ms)
Showing top 10 nodes out of 58
      flat  flat%   sum%        cum   cum%
    2290ms 51.81% 51.81%     2290ms 51.81%  runtime.usleep
     740ms 16.74% 68.55%     1560ms 35.29%  sync.(*Mutex).lockSlow
     160ms  3.62% 72.17%      160ms  3.62%  runtime.pthread_cond_wait
     150ms  3.39% 75.57%      240ms  5.43%  runtime.mapaccess2_faststr
     150ms  3.39% 78.96%     1400ms 31.67%  sync.(*RWMutex).Unlock
      80ms  1.81% 80.77%     1720ms 38.91%  runtime.lock2
      60ms  1.36% 82.13%       60ms  1.36%  runtime.asmcgocall
      60ms  1.36% 83.48%       60ms  1.36%  runtime.cansemacquire (inline)
      60ms  1.36% 84.84%       60ms  1.36%  runtime.pthread_cond_signal
      50ms  1.13% 85.97%       50ms  1.13%  runtime.tophash (inline)
(pprof) png
Generating report in profile001.png
```

As you can see in the sample output (`profile001.png`), the single RWMutex has become the bottleneck on the data structure (`runtime.usleep` is part of the RWMutex locking and waiting strategy to aoid thrashing or busy wait). "Lock striping" is a common technique to get around lock contention of a coarse-grained lock.

The idea is to split the data structure into "stripes" (or "buckets", or "shards"), each protected by its own finer-grained lock. Many queries can then be either parallelized or served from a single stripe. Typically, stripes are divided based on a good hash of the key or id (in our case, the string).

**B1.** Implement `StripedStringSet` which offers the same API as `StringSet` but uses lock striping. Add a factory function that constructs a `StripedStringSet` with a given stripe count. (Starter code in `striped_string_set.go`.)

You may find it helpful to review the following references to understand the technique, but you may not directly use or translate the code:
* [Lock striping in Java](https://www.baeldung.com/java-lock-stripping)
* [Concurrent map in Go](https://github.com/orcaman/concurrent-map/blob/master/concurrent_map.go)

**B2.** Keeping count. There are several strategies to keep count in this striped data structure:
1. iterate through the entire set across all stripes each time a count is needed;
2. keep a per-stripe counter and add up counters from every stripe when queried;
3. keep a global (i.e., one single) counter that gets updated when a string is added; (See [atomic counters](https://gobyexample.com/atomic-counters).)

What are the advantages and disadvantages of each of these approaches? Can you think of query patterns where one works better than another?

Include your thoughts (1~2 paragraphs) in a plain text file `discussions.txt` under a heading `B2`.

**ExtraCredit1.** Suppose we start with a `StripedStringSet` with x unique strings. Goroutine/thread 0 issues a `Count()` call, while threads 1 through N issues `Add()` calls with distinct strings. What values might the `Count()` in thread 0 return? Why? Does it matter which counting strategy (#1 through 3 above) we use? What about `LockedStringSet`? In light of this behavior, how might you define "correctness" for the method `Count()`? Include your thoughts in `discussions.txt` under a heading `ExtraCredit1`.

## Part C. Channels, goroutines, and parallelization
**C1.** Implement the API `PredRange(begin, end, pattern)` to return all strings matching a particular pattern within a range `[begin, end)` lexicographically. Please implement this function for both `LockedStringSet` and `StripedStringSet`.

For example, if the string set `s` contains `{"barabc", "bazdef", "fooabc", "tusabc", "zyxabc"}`, calling `s.PredRange("barabc", "zyxabc", "abc")` should return `["barabc", "fooabc", "tusabc"]`.

You may use [`regexp.Match`](https://pkg.go.dev/regexp).

**C2.** Parallelize your implementation of `PredRange` for `StripedStringSet` by spinning up a goroutine for each stripe and aggregate the results in the end.

**C3.** Pick _one_ of the `adds+counts` or `scans:parallel` sub-benchmarks. Run the provided subbenchmark and see the difference in performance between `LockedStringSet` and `StripedStringSet` with 2 stripes. What do you observe? Include the results in `discussions.txt` under a heading `C3`.

**C4.** Use the same subbenchmark as C3. Generate a graph visualization of a profile for `StripedStringSet` with `stripeCount == NumCPU() * 2`. Name this `profile_striped_num_cpu_times_two.png`.

**ExtraCredit2.** Discuss the effect of the parameter stripeCount on the performance (compared to `LockedStringSet`). What do you notice? Why? What's the optimal stripeCount (feel free to try other numbers and include the result in the discussion)? Include your thoughts in `discussions.txt` under a heading `ExtraCredit2`.

# End of Lab 0
---

# Go Tips and FAQs
 - After the Tour of Go, use https://go.dev/doc/effective_go and https://gobyexample.com/
 - Use these commands:
    - `go fmt`
    - `go mod tidy` cleans up the module dependencies.
    - `go test -race ...` turns on the Go race detector.
    - `go vet` examines Go source code and reports suspicious constructs, such as Printf calls whose arguments do not align with the format string. Vet uses heuristics that do not guarantee all reports are genuine problems, but it can find errors not caught by the compilers.

# Time logging

Source: from Prof. Stan Eisenstat's CS223/323 courses. Obtained via Prof. James Glenn [here](https://zoo.cs.yale.edu/classes/cs223/f2020/Projects/log.html).

Each lab submission must contain a complete log `time.log`. Your log file should be a plain text file of the general form (that below is mostly fictitious):

```
ESTIMATE of time to complete assignment: 10 hours

      Time     Time
Date  Started  Spent Work completed
----  -------  ----  --------------
8/01  10:15pm  0:45  read assignment and played several games to help me
                     understand the rules.
8/02   9:00am  2:20  wrote functions for determining whether a roll is
                     three of a kind, four of a kind, and all the other
                     lower categories
8/04   4:45pm  1:15  wrote code to create the graph for the components
8/05   7:05pm  2:00  discovered and corrected two logical errors; code now
                     passes all tests except where choice is Yahtzee
8/07  11:00am  1:35  finished debugging; program passes all public tests
               ----
               7:55  TOTAL time spent

I discussed my solution with: Petey Salovey, Biddy Martin, and Biff Linnane
(and watched four episodes of Futurama).

Debugging the graph construction was difficult because the size of the
graph made it impossible to check by hand.  Using asserts helped
tremendously, as did counting the incoming and outgoing edges for
each vertex.  The other major problem was my use of two different variables
in the same function called _score and score.  The last bug ended up being
using one in place of the other; I now realize the danger of having two
variables with names varying only in punctuation -- since they both sound
the same when reading the code back in my head it was not obvious when
I was using the wrong one.
```

Your log MUST contain:
 - your estimate of the time required (made prior to writing any code),
 - the total time you actually spent on the assignment,
 - the names of all others (but not members of the teaching staff) with whom you discussed the assignment for more than 10 minutes, and
 - a brief discussion (100 words MINIMUM) of the major conceptual and coding difficulties that you encountered in developing and debugging the program (and there will always be some).

The estimated and total times should reflect time outside of class.  Submissions
with missing or incomplete logs will be subject to a penalty of 5-10% of the
total grade, and omitting the names of collaborators is a violation of the
academic honesty policy.

To facilitate analysis, the log file MUST the only file submitted whose name contains the string "log" and the estimate / total MUST be on the only line in that file that contains the string "ESTIMATE" / "TOTAL".

# Collaboration policy

## General Statement on Collaboration
TL;DR: Same as [CS323](https://zoo.cs.yale.edu/classes/cs323/current/syllabus.html)for the individual labs (which Labs 0-4 are).

Programming, like composition, is an individual creative process in which you must reach your own understanding of the problem and discover a path to its solution. During this time, discussions with others (including members of the teaching staff) are encouraged. But see the Gilligan's Island Rule below.

However, when the time comes to design the program and write the code, such discussions are no longer appropriate---your solution must be your own personal inspiration (although you may ask members of the teaching staff for help in understanding, designing, writing, and debugging).

Since code reuse is an important part of programming, you may study and/or incorporate published code (e.g., from text books or the Net) in your programs, provided that you give proper attribution in your source code and in your log file and that the bulk of the code submitted is your own. Note: Removing/rewriting comments, renaming functions/variables, or reformatting statements does not convey ownership.

But when you incorporate more than 25 lines of code from a single source, this code (prefaced by a comment identifying the source) must be isolated in a separate file that the rest of your code #include-s or links with. The initial submission of this file should contain only the identifying comment and the original code; revisions may only change types or function/variable names, turn blocks of code into functions, or add comments.

DO NOT UNDER ANY CIRCUMSTANCES COPY SOMEONE ELSE'S CODE OR GIVE A COPY OF YOUR CODE TO SOMEONE ELSE OR OTHERWISE MAKE IT PUBLICLY AVAILABLE---to do so is a clear violation of ethical/academic standards that, when discovered, will be referred to the Executive Committee of Yale College for disciplinary action. Modifying code to conceal copying only compounds the offense.

## The Gilligan's Island Rule

When discussing an assignment with anyone other than a member of the teaching staff, you may write on a board or a piece of paper, but you may not keep any written or electronic record of the discussion. Moreover, you must engage in some mind-numbing activity (e.g., watching an episode of Gilligan's Island) before you work on the assignment again. This will ensure that you can reconstruct what you learned, by yourself, using your own brain. The same rule applies to reading books or studying on-line sources.

## Tips on asking good questions / help us help you
- First, try to find the answer(s) by Googling. See above about rules re: code reuse and attribution.
- For technical questions, prefer posting to Canvas such that your classmates may answer your question(s) and benefit from the answer(s).
- Help your classmates by answering their questions! You are not in competition with one another, so help each other learn!
- Identify the [**minimum reproducible example**](https://myweb.uiowa.edu/pbreheny/reproducible.html) of your problem. E.g., for Go language related questions, attempt to construct a small example on [Go playground](https://go.dev/play/) (which has sharing functionality).
