# [Initial version] Lab 2: Raft with static cluster membership

This lab is adapted from the Raft lab from MIT's distributed systems course [6.824](https://pdos.csail.mit.edu/6.824/labs/lab-raft.html). Used with permission.

## Logistics
**Policies**
 - Lab 2 is meant to be an **individual** assignment. Please see the [Collaboration Policy](#collaboration-policy) for details.
 - We will help you strategize how to debug but WE WILL NOT DEBUG YOUR CODE FOR YOU.
 - Please keep and submit a time log of time spent and major challenges you've encountered. This may be familiar to you if you've taken CS323. See [Time logging](#time-logging) for details.
 - Questions? post to [Edstem](https://edstem.org/us/courses/20288/discussion/) or email the teaching staff at cs426ta@cs.yale.edu.
   - Xiao Shi (xiao.shi@aya.yale.edu)
   - Scott Pruett (spruett345@gmail.com)
   - Richard Yang (yry@cs.yale.edu)

**Submission deadline: 23:59 ET Wednesday Mar 16, 2022**

**Submission logistics** Submit a `.tar.gz` archive named after your NetID via
Canvas. The Canvas assignment will be up a day or two before the deadline.

Your submission for this lab should include the following files:
```
raft/raft.go
raft/my_raft_test.go // create this file and add your own unittests
raft/my_util.go // [optional] only if you add additional util functions
discussions.md // NOTE: switching from txt to markdown
time.log
```

We suggest that you **read through** this lab before starting to write any code. There is a section of [miscellaneous hints](#miscellaneous-implementation-hints) that you will likely find helpful. You get **three weeks** in total and each part should be achievable in a week. However, it takes non-trivial amount of reading to get started; additionally, later parts might reveal bugs in the previous parts, even though all corresponding tests pass. You should budget ample time to debug and don't be afraid to rewrite / clean up parts of your implementation (use version control to your advantage).


## Introduction
Distributed consensus protocols such as Raft underly nearly all modern fault-tolerant storage systems. In this lab you'll implement Raft, a replicated state machine protocol. Subsequent 6.824 labs will instruct you to build [a key/value service](https://pdos.csail.mit.edu/6.824/labs/lab-kvraft.html) on top of Raft, and to [“shard” your service](https://pdos.csail.mit.edu/6.824/labs/lab-shard.html) over multiple replicated state machines for higher performance. We will not officially assign the subsequent labs, but you may choose to complete them as your final project, more details TBA.

A replicated service achieves fault tolerance by storing copies of its state (i.e., data) on multiple replica servers. Replication allows the service to continue operating even if some of its servers experience failures (crashes or a broken or flaky network). The challenge is that failures may cause the replicas to hold differing copies of the data.

Raft organizes client requests into a sequence, called the log, and ensures that all the replica servers see the same log. Each replica executes client requests in log order, applying them to its local copy of the service's state. Since all the live replicas see the same log contents, they all execute the same requests in the same order, and thus continue to have identical service state. If a server fails but later recovers, Raft takes care of bringing its log up to date. Raft will continue to operate as long as at least a majority of the servers are alive and can talk to each other. If there is no such majority, Raft will make no progress, but will pick up where it left off as soon as a majority can communicate again.

In this lab you'll implement Raft as a Go object type with associated methods, meant to be used as a module in a larger service. A set of Raft instances talk to each other with RPC to maintain replicated logs. Your Raft interface will support an indefinite sequence of numbered commands, also called log entries. The entries are numbered with _index numbers_. The log entry with a given index will eventually be committed. At that point, your Raft should send the log entry to the larger service for it to execute.

You should follow the design in the [extended Raft paper](https://raft.github.io/raft.pdf), with particular attention to Figure 2. You'll implement most of what's in the paper, including saving persistent state and reading it after a node fails and then restarts. You will _not_ implement cluster membership changes (Section 6) or log compaction (Section 7).

You may find this [guide](https://thesquareplanet.com/blog/students-guide-to-raft/) useful, as well as this advice about [locking](https://pdos.csail.mit.edu/6.824/labs/raft-locking.txt) and [structure](https://pdos.csail.mit.edu/6.824/labs/raft-structure.txt) for concurrency. The [visualization](https://raft.github.io/#raftscope) on the Raft website may help you build intuition. For a wider perspective, have a look at Paxos, Chubby, Paxos Made Live, Spanner, Zookeeper, Harp, Viewstamped Replication, and [Bolosky et al.](http://static.usenix.org/event/nsdi11/tech/full_papers/Bolosky.pdf) (Note: the student's guide was written several years ago, and includes parts of the labs we chose to leave out. Make sure you understand why a particular implementation strategy makes sense before blindly following it!)

Keep in mind that the most challenging part of this lab may not be implementing your solution, but debugging it. To help address this challenge, you may wish to spend time thinking about how to make your implementation more easily debuggable. You might refer to the [Guidance](https://pdos.csail.mit.edu/6.824/labs/guidance.html) page and to [this blog post about effective print statements](https://blog.josejg.com/debugging-pretty/). The functions in `raft/util.go` should help as well.

We also provide a [diagram of Raft interactions](https://pdos.csail.mit.edu/6.824/notes/raft_diagram.pdf) that can help clarify how your Raft code interacts with the layers on top of it.

## Getting Started

We supply you with skeleton code `raft/raft.go`. We also supply a set of tests, which you should use to drive your implementation efforts, and which we'll use to grade your submitted lab. The tests are in `raft/raft_test.go`.

To get up and running, execute the following commands. Don't forget the git pull to get the latest code.

```
$ cd raft
$ go test -race
Test (2A): initial election ...
--- FAIL: TestInitialElection2A (5.04s)
        config.go:326: expected one leader, got none
Test (2A): election after network failure ...
--- FAIL: TestReElection2A (5.03s)
        config.go:326: expected one leader, got none
...
```

You can turn on debug logging by passing `-debug` flag:
```
$ go test -race -debug
2022/02/13 22:30:41.208830 disconnect(0) (config.go:273)
2022/02/13 22:30:41.211028 disconnect(1) (config.go:273)
2022/02/13 22:30:41.211237 disconnect(2) (config.go:273)
2022/02/13 22:30:41.211409 connect(0) (config.go:250)
2022/02/13 22:30:41.211493 connect(1) (config.go:250)
2022/02/13 22:30:41.211511 connect(2) (config.go:250)
Test (2A): initial election ...
--- FAIL: TestInitialElection2A (4.92s)
    config.go:344: expected one leader, got none
2022/02/13 22:30:46.130352 disconnect(0) (config.go:273)
2022/02/13 22:30:46.130722 disconnect(1) (config.go:273)
...
```

## The code

Implement Raft by adding code to `raft/raft.go`. In that file you'll find skeleton code, plus examples of how to send and receive RPCs.

Your implementation must support the following interface, which the tester and (eventually) your key/value server will use. You'll find more details in comments in raft.go.

```
// create a new Raft server instance:
rf := Make(peers, me, persister, applyCh)

// start agreement on a new log entry:
rf.Start(command interface{}) (index, term, isleader)

// ask a Raft for its current term, and whether it thinks it is leader
rf.GetState() (term, isLeader)

// each time a new entry is committed to the log, each Raft peer
// should send an ApplyMsg to the service (or tester).
type ApplyMsg
```

A service calls `Make(peers, me, ...)` to create a Raft peer. The peers argument is an array of network identifiers of the Raft peers (including this one), for use with RPC. The `me` argument is the index of this peer in the peers array. `Start(command)` asks Raft to start the processing to append the command to the replicated log. `Start()` should **return immediately**, without waiting for the log appends to complete---note that this is one spot where the implementation differs from Figure 2 of the Raft paper. The service expects your implementation to send an `ApplyMsg` for each newly committed log entry **in order** to the `applyCh` channel argument to `Make()`.

`raft.go` contains example code that sends an RPC (`sendRequestVote()`) and that handles an incoming RPC (`RequestVote()`). Your Raft peers should exchange RPCs using the labrpc Go package (source in `labrpc`). The tester can tell labrpc to delay RPCs, re-order them, and discard them to simulate various network failures. While you can temporarily modify labrpc, make sure your Raft works with the original labrpc, since that's what we'll use to test and grade your lab. Your Raft instances must interact only with RPC; for example, they are not allowed to communicate using shared Go variables or files.

Subsequent parts of this lab depend on the preceding parts, so it is important to give yourself enough time to write solid code.

## Part 2A: leader election ([moderate](https://pdos.csail.mit.edu/6.824/labs/guidance.html))

### Task 2A-1
Implement Raft leader election and heartbeats (`AppendEntries` RPCs with no log entries). The goal for Part **2A** is for a single leader to be elected, for the leader to remain the leader if there are no failures, and for a new leader to take over if the old leader fails or if packets to/from the old leader are lost. Run `go test -run 2A -race` to test your **2A** code.

### Hints
* You can't easily run your Raft implementation directly; instead you should run it by way of the tester, i.e. `go test -run 2A -race`.
* Follow the paper's Figure 2. At this point you care about sending and receiving RequestVote RPCs, the Rules for Servers that relate to elections, and the State related to leader election,
* Add the Figure 2 state for leader election to the Raft struct in `raft.go`. You'll also need to define a struct to hold information about each log entry.
* Fill in the `RequestVoteArgs` and `RequestVoteReply` structs. Modify `Make()` to create a background goroutine that will kick off leader election periodically by sending out `RequestVote` RPCs when it hasn't heard from another peer for a while. This way a peer will learn who is the leader, if there is already a leader, or become the leader itself. Implement the `RequestVote()` RPC handler so that servers will vote for one another.
* To implement heartbeats, define an `AppendEntries` RPC struct (though you may not need all the arguments yet), and have the leader send them out periodically. Write an `AppendEntries` RPC handler method that resets the election timeout so that other servers don't step forward as leaders when one has already been elected.
* Make sure the election timeouts in different peers don't always fire at the same time, or else all peers will vote only for themselves and no one will become the leader. You may find Go's [`math/rand` package](https://golang.org/pkg/math/rand/) useful.
* The tester requires that the leader send heartbeat RPCs no more than ten times per second.
* The tester requires your Raft to elect a new leader within five seconds of the failure of the old leader (if a majority of peers can still communicate). Remember, however, that leader election may require multiple rounds in case of a split vote (which can happen if packets are lost or if candidates unluckily choose the same random backoff times). You must pick election timeouts (and thus heartbeat intervals) that are short enough that it's very likely that an election will complete in less than five seconds even if it requires multiple rounds.
* The paper's Section 5.2 mentions election timeouts in the range of 150 to 300 milliseconds. Such a range only makes sense if the leader sends heartbeats considerably more often than once per 150 milliseconds. Because the tester limits you to 10 heartbeats per second, you will have to use an election timeout larger than the paper's 150 to 300 milliseconds, but not too large, because then you may fail to elect a leader within five seconds.
* You'll need to write code that takes actions periodically or after delays in time. The easiest way to do this is to create a goroutine with a loop that calls [time.Sleep()](https://golang.org/pkg/time/#Sleep); (see the `ticker()` goroutine that `Make()` creates for this purpose). Don't use Go's `time.Timer` or `time.Ticker`, which are difficult to use correctly.
* The [Guidance page](https://pdos.csail.mit.edu/6.824/labs/guidance.html) has some tips on how to develop and debug your code.
* If your code has trouble passing the tests, read the paper's Figure 2 again; the full logic for leader election is spread over multiple parts of the figure.
* Don't forget to implement `GetState()`.
* The tester calls your Raft's `rf.Kill()` when it is permanently shutting down an instance. You can check whether `Kill()` has been called using `rf.killed()`. You may want to do this in all loops, to avoid having dead Raft instances print confusing messages.
* Go RPC sends only struct fields whose names start with capital letters. Sub-structures must also have capitalized field names (e.g. fields of log records in an array). The `labgob` package will warn you about this; don't ignore the warnings.

Be sure you pass the **2A** tests before proceeding, so that you see something like this:

```
$ go test -run 2A -race
Test (2A): initial election ...
  ... Passed --   4.0  3   48   13200    0
Test (2A): election after network failure ...
  ... Passed --   5.5  3   58   12635    0
Test (2A): multiple elections ...
  ... Passed --   8.5  7  408   71875    0
PASS
ok      6.824/raft      18.498s
```

Each "Passed" line contains five numbers; these are the time that the test took in seconds, the number of Raft peers (usually 3 or 5), the number of RPCs sent during the test, the total number of bytes in the RPC messages, and the number of log entries that Raft reports were committed. Your numbers will differ from those shown here. You can ignore the numbers if you like, but they may help you sanity-check the number of RPCs that your implementation sends. For all of labs 2, the grading script will fail your solution if it takes more than 400 seconds for all of the tests (`go test` without `-race`), or if any individual test takes more than 120 seconds.

Given the non-determinism of timers and RPCs, for all tests in this lab, it is a good idea to run each multiple times and check that each run passes.

```
$ for i in {0..10}; do go test; done
```

### Task 2A-2 Theory and practice of leader election liveness
**In theory**, a carefully-orchestrated (e.g., an omniscient adversary with control of `rand` and `time`) sequence of events could prevent a leader from ever being elected in Raft (or Paxos). This lack of liveness corroborates with the [FLP impossibility theorem](https://groups.csail.mit.edu/tds/papers/Lynch/jacm85.pdf), which states "in an asynchronous network where messages may be delayed but not lost, there is no consensus algorithm that is guaranteed to terminate in every execution for all starting conditions, if at least one node may experience failure."

(1) Construct and describe a scenario with a Raft cluster of 3 or 5 nodes where the leader election protocol fails to elect a leader. Hint: in your description, you may decide when timers time out or not time out, or arbitrate when RPCs get sent or processed.

(2) **In practice**, why is this not a major concern? i.e., how does Raft get around this theoretical possibility?

Include your answers under the heading **2A-2** in `discussions.md`. Reminder: you must cite any sources you consult or any discussions you may have had with your peers or the teaching staff.

**ExtraCredit1.** Another issue that affects Raft liveness in the real world (e.g., this [Cloudfare outage](https://blog.cloudflare.com/a-byzantine-failure-in-the-real-world/)---though this is **not** a Byzantine failure.) is related to "term-inflation". Dr. Diego Ongaro described the problem and his idea of addressing this in Section 9.6 of his [thesis](https://web.stanford.edu/~ouster/cgi-bin/papers/OngaroPhD.pdf); [here](https://www.openlife.cc/sites/default/files/4-modifications-for-Raft-consensus.pdf) is MongoDB's detailed account of the "Pre-Vote" modification they implemented; [this blog post](https://decentralizedthoughts.github.io/2020-12-12-raft-liveness-full-omission/) further describe the ramification and limitations of Pre-Vote and CheckQuorum. Does the scenario you constructed above resolve if the Raft instances implement Pre-Vote and CheckQuorum? If so, could you construct a scenario where Raft leader election can be theoretically stuck even with Pre-Vote and CheckQuorum? If not, explain why not. Include your response under the heading **ExtraCredit1** in `discussions.md`.

### Task 2A-3
Implement your own unit tests in a new file `my_raft_test.go`. By the end of this lab, you should write at least 5 tests. They could be tests on **2A** leader election, **2B** log replication, or **2C** persistence.

## Part 2B: log replication ([hard](https://pdos.csail.mit.edu/6.824/labs/guidance.html))

### Task 2B-1
Implement the leader and follower code to append new log entries, so that the `go test -run 2B -race` tests pass.

### Hints
* Your first goal should be to pass `TestBasicAgree2B()`. Start by implementing `Start()`, then write the code to send and receive new log entries via `AppendEntries` RPCs, following Figure 2.
* You will need to implement the election restriction (Section 5.4.1 in the paper).
* One way to fail to reach agreement in the early Lab **2B** tests is to hold repeated elections even though the leader is alive. Look for bugs in election timer management, or not sending out heartbeats immediately after winning an election.
* Your code may have loops that repeatedly check for certain events. Don't have these loops execute continuously without pausing, since that will slow your implementation enough that it fails tests. Use Go [channels and `select` statement](https://gobyexample.com/select) to trigger certain events, Go's [condition variables](https://golang.org/pkg/sync/#Cond), or insert a `time.Sleep(10 * time.Millisecond)` in each loop iteration.
* Do yourself a favor for readability and debugging and write (or re-write) code that's clean and clear. For ideas, re-visit our the [Guidance page](https://pdos.csail.mit.edu/6.824/labs/guidance.html) with tips on how to develop and debug your code.
* If you fail a test, look over the code for the test in `raft/config.go` and `raft_test.go` to get a better understanding of what the test is testing. `config.go` also illustrates how the tester uses the Raft API. You may insert `DPrintf` statements in `config.go` to help debugging as well.

The tests for upcoming labs may fail your code if it runs too slowly. You can check how much real time and CPU time your solution uses with the time command. Here's typical output:

```
$ time go test -run 2B
Test (2B): basic agreement ...
  ... Passed --   1.6  3   18    5158    3
Test (2B): RPC byte count ...
  ... Passed --   3.3  3   50  115122   11
Test (2B): agreement despite follower disconnection ...
  ... Passed --   6.3  3   64   17489    7
Test (2B): no agreement if too many followers disconnect ...
  ... Passed --   4.9  5  116   27838    3
Test (2B): concurrent Start()s ...
  ... Passed --   2.1  3   16    4648    6
Test (2B): rejoin of partitioned leader ...
  ... Passed --   8.1  3  111   26996    4
Test (2B): leader backs up quickly over incorrect follower logs ...
  ... Passed --  28.6  5 1342  953354  102
Test (2B): RPC counts aren't too high ...
  ... Passed --   3.4  3   30    9050   12
PASS
ok      raft    58.142s

real    0m58.475s
user    0m2.477s
sys     0m1.406s
```

The "ok raft 58.142s" means that Go measured the time taken for the 2B tests to be 58.142 seconds of real (wall-clock) time. The "user 0m2.477s" means that the code consumed 2.477 seconds of CPU time, or time spent actually executing instructions (rather than waiting or sleeping). If your solution uses much more than a minute of real time for the **2B** tests, or much more than 5 seconds of CPU time, you may run into trouble later on. Look for time spent sleeping or waiting for RPC timeouts, loops that run without sleeping or waiting for conditions or channel messages, or large numbers of RPCs sent.

## Part 2C: persistence ([hard](https://pdos.csail.mit.edu/6.824/labs/guidance.html))

If a Raft-based server reboots it should resume service where it left off. This requires that Raft keep persistent state that survives a reboot. The paper's Figure 2 mentions which state should be persistent.

A real implementation would write Raft's persistent state to disk each time it changed, and would read the state from disk when restarting after a reboot. Your implementation won't use the disk; instead, it will save and restore persistent state from a `Persister` object (see `raft/persister.go`). Whoever calls `Raft.Make()` supplies a `Persister` that initially holds Raft's most recently persisted state (if any). Raft should initialize its state from that `Persister`, and should use it to save its persistent state each time the state changes. Use the `Persister`'s `ReadRaftState()` and `SaveRaftState()` methods.

### Task 2C-1
Complete the functions `persist()` and `readPersist()` in `raft.go` by adding code to save and restore persistent state. You will need to encode (or "serialize") the state as an array of bytes in order to pass it to the `Persister`. Use the `labgob` encoder; see the comments in `persist()` and `readPersist()`. `labgob` is like Go's `gob` encoder but prints error messages if you try to encode structures with lower-case field names.

### Task 2C-2
Insert calls to `persist()` at the points where your implementation changes persistent state. Once you've done this, you should pass the remaining tests.

Note: in order to avoid running out of memory, Raft must periodically discard old log entries, but you **do not** have to worry about it for this lab.

### Task 2C-3 fast log backtracking
Implement the optimization that backs up `nextIndex` by more than one entry at a time. Look at the [extended Raft paper](https://raft.github.io/raft.pdf) starting at the bottom of page 7 and top of page 8 (marked by a gray line). The paper is vague about the details; you will need to fill in the gaps, perhaps with the help of the 6.824 Raft lectures ([here](https://www.youtube.com/watch?v=R2-9bsKmEbo) and [here](https://www.youtube.com/watch?v=h3JiQ_lnkE8)).

### Hints
* Many of the **2C** tests involve servers failing and the network losing RPC requests or replies. These events are non-deterministic, and you may get lucky and pass the tests, even though your code has bugs. Typically running the test several times will expose those bugs.
* While **2C** only requires you to implement persistence and fast log backtracking, **2C** test failures might be related to previous parts of your implementation. Even if you pass **2A** and **2B** tests consistently, you may still have leader election or log replication bugs that are exposed on **2C** tests.

Your code should pass all the **2C** tests (as shown below), as well as the **2A** and **2B** tests.

```
$ go test -run 2C -race
Test (2C): basic persistence ...
  ... Passed --   7.2  3  206   42208    6
Test (2C): more persistence ...
  ... Passed --  23.2  5 1194  198270   16
Test (2C): partitioned leader and one follower crash, leader restarts ...
  ... Passed --   3.2  3   46   10638    4
Test (2C): Figure 8 ...
  ... Passed --  35.1  5 9395 1939183   25
Test (2C): unreliable agreement ...
  ... Passed --   4.2  5  244   85259  246
Test (2C): Figure 8 (unreliable) ...
  ... Passed --  36.3  5 1948 4175577  216
Test (2C): churn ...
  ... Passed --  16.6  5 4402 2220926 1766
Test (2C): unreliable churn ...
  ... Passed --  16.5  5  781  539084  221
PASS
ok      raft    142.357s
```

## Task 2C-4
Kudos for making it all the way through the lab! Reflect on your implementation and debugging experience. This is a reminder to include in the `time.log` a brief discussion (100 words MINIMUM) of the major conceptual and coding difficulties that you encountered in developing and debugging the program (and there will always be some).

# Miscellaneous implementation hints
* A reasonable amount of time to consume for the full set of Lab 2 tests (2A+2B+2C) is 8 minutes of real time and one and a half minutes of CPU time.
* Both the extended Raft paper and the tester uses 1-indexed log entries, but Golang slices are 0-indexed.
* Think about how you would "wait for successful reply from a majority" and whether you would need to. You likely shouldn't wait for replies from all nodes since disconnected nodes or delayed RPCs could significally delay the progress.
* Below is one strategy to manage shared state for your Raft instance. Every member variable should be one of the following categories:
  * (1) once initialized (in `Make()`, read-only and never change), in which case you can access it how/whenever;
  * (2) protected under a mutex (e.g., `rf.mu`);
  * (3) is an [atomic](https://pkg.go.dev/sync/atomic) variable and every access is an atomic access;
  * (4) only ever accessed by a single (perhaps long-running) goroutine, in which case no synchronization is necessary.

  You might find it tempting to minimize contention upfront, however, we strongly recommend you to focus on correctness first, and do the obvious thing first (in this case, (1) and (2)) since "premature optimization is the root of all evil" (Tony Hoare, popularized by Donald Knuth).
* You might find it helpful to think of mutexes as a way to ensure logic in the critical section is carried out atomically (in addition to ensuring mutual exclusion).
* Safety vs. liveness. We encourage you to reason about distributed consensus protocols in both their safety guarantees (i.e., bad things do not ever happen) and liveness guarantees (i.e., good things eventually happen [in a reasonable amount of time]). This will be reflected in our grading rubric: we will not give any credit to tests that violate the safety property (e.g., two nodes in the cluster disagree on the value for a committed log entry with the same index, i.e., `apply error`), but we might give partial credit for test failures (especially in **2C** for tests under unreliable networks) due to lack of progress (e.g., `config.one(_): fail to reach agreement`).
* This [recitation](https://www.cs.princeton.edu/courses/archive/fall16/cos418/docs/P9-raft-assignments.pdf) from the Princeton distributed systems course might be helpful as well. Their Assignment 3 corresponds to Part **2A** and Assignment 4 corresponds to Parts **2B** and **2C**.
* The [TLA+](https://lamport.azurewebsites.net/tla/tla.html) spec for Raft can be found [here](https://github.com/ongardie/raft.tla/blob/master/raft.tla). It might help clarify details the Raft paper does not elaborate on.
* This [script](https://gist.github.com/jonhoo/f686cacb4b9fe716d5aa#file-go-test-many-sh) that runs many go tests in parallel might be helpful.

# End of Lab 2

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
TL;DR: Same as [CS323](https://zoo.cs.yale.edu/classes/cs323/current/syllabus.html) for the individual labs (which Labs 0-4 are).

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
