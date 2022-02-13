# [Initial version] CS426 Lab 1: Single-node video recommendation service

## Overview
In this lab, you will use gRPC and protobuf to implement a specified IDL (interface definition language) for a single node stateless server.

This server will mimic the backend functionality of a video service's landing page and return information that is needed to render a user’s homepage (ranked video recommendation). We will provide several other microservices for user metadata, video metadata, as well as a ranking library. Your service will make RPCs to compute the final results.

As in production services, the provided backend services (as well as the network or the client library) have a certain failure rate and latency profile. You will build smart error handling, retries, and batching logic to improve the reliability of your service. You will also implement monitoring metrics and APIs to improve the observability of your service.

## Logistics
**Policies**
- Lab 1 is meant to be an **individual** assignment. Please see the [Collaboration Policy](#collaboration-policy) for details.
- We will help you strategize how to debug but WE WILL NOT DEBUG YOUR CODE FOR YOU.
- Please keep and submit a time log of time spent and major challenges you've encountered. This may be familiar to you if you've taken CS323. See [Time logging](#time-logging) for details.

- Questions? post to Canvas or email the teaching staff at cs426ta@cs.yale.edu.
  - Scott Pruett (spruett345@gmail.com)
  - Xiao Shi (xiao.shi@aya.yale.edu)
  - Richard Yang (yry@cs.yale.edu)

**Submission deadline: 23:59 ET Wednesday Feb 23, 2022**

**Submission logistics** Submit a `.tar.gz` archive named after your NetID via
Canvas. The Canvas assignment will be up a day or two before the deadline.

Your submission for this lab should include the following files:
```
video_rec_service/server_lib/server_lib.go
video_rec_service/server/server.go // though you may not need to modify this file
video_rec_service/server/server_test.go // create this file and add your own unittests
discussions.txt
time.log
```

## gRPC and protobuf

gRPC is an open-source framework for making Remote Procedure Calls (RPCs) that is widely used in industry. gRPC uses protobuf as both its Interface Definition Language (the schema) and the transport format. For a quick overview, read the official introduction: https://grpc.io/docs/what-is-grpc/introduction/

gRPC provides libraries for many languages, including Go. You will be using these libraries to both implement a server for an existing API and to call other services as a client.

For basic examples of how it can be used in Go, check out the [official examples](https://github.com/grpc/grpc-go/tree/master/examples/helloworld).

### Setup
 1. Check out the repository like you did for lab 0, or pull to update: https://github.com/shixiao/cs426-spring2022-labs

 2. [Optional] The repo has already run the protobuf compiler to generate the binding code. But if you'd like to learn to do so, you can use the following command (or check out the command in the `Makefile`):
`protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative video_rec_service/proto/video_rec_service.proto`
    * Note: you will need to install `protoc`, the protobuf compiler, to get started. You can find instructions in the gPRC documentation: https://grpc.io/docs/protoc-installation/
    * Additionally, you will need to install the Go plugins, following these steps: https://grpc.io/docs/languages/go/quickstart/

## The Video Recommendation Service

Your first job will be to implement a single RPC method for VideoRecService: `GetTopVideos`. This service will return personalized video recommendations for a given user that can be displayed as they open the video application and start to browse (don't worry, you'll only be handling network requests and sending replies, no frontend required).

The overall strategy for the video recommending service will be to look at videos all the videos their subscribed-to-users have liked and rank them according to some properties (these will be wrapped as opaque coefficients, and the "ranking" algorithm is provided in the `ranker` library). To do this, the video recommending service will need to combine results from two different backends: the UserService (for fetching subscriptions and liked videos) and VideoService (for fetching video data). Both UserService and VideoService are also gRPC microservices, so VideoRecService will be using gRPC clients to communicate.

We've provided you with skeleton code in `video_rec_service/server/server.go` and the accompanying server library `video_rec_service/server_lib/server_lib.go`. To try it out, `go run video_rec_service/server/server.go`. This code will start up and run the gRPC server on a given port (default `8080`) for you.


### Part A. Implementing GetTopVideos

With the skeleton, you will start by writing code inside the `GetTopVideos` method. Take a look at the protobuf IDL for the request and response:

```
message GetTopVideosRequest {
    uint64 user_id = 1;
    // optional limit of the number of results to return
    int32 limit = 2;
}

message GetTopVideosResponse {
    repeated video_service.VideoInfo videos = 1;
    bool stale_response = 2; // For part C
}
```

In short, GetTopVideos receives a `user_id` parameter, and optionally a `limit`. It must return a list of `videos` to recommend in ranked order (i.e., descending score) and a boolean if the response is fresh (you will implement this boolean in **Part C**).

The request starts for a given user based on `user_id`. You will use the `UserService` to fetch the `UserInfo` for the starting user---importantly, the other users they subscribe to and the starting user's ranking coefficients.

#### A1. Using a gRPC client

To communicate with the `UserService` you'll need to create a gRPC client. We'll start with the basics, following the Go gRPC guide: https://grpc.io/docs/languages/go/basics/#client

`grpc.Dial()` creates a "channel" (in this case, a network stream channel, not a Go built-in `chan`) which allows us to communicate with a server. Use this to create a channel to `UserService`, which you will use in **A3**. For the server address, use the global `userServiceAddr` which can be set from a configurable flag to your service.

Note: for this lab, we are not using TLS, but `grpc.Dial()` does expect transport credentials:
```
import "google.golang.org/grpc/credentials/insecure"
...
var opts []grpc.DialOption
opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
conn, err := grpc.Dial(address, opts...)
```
(This is what `cmd/frontend/frontend.go` does, for reference.)

**Discussion**: What networking/socket API(s) do you think `Dial` corresponds to or is similar to? (For reference, you can find example socket APIs [here](https://en.wikipedia.org/wiki/Berkeley_sockets#Socket_API_functions) or in the slides for Lecture 3.) Include your responses under a heading **A1** in `discussions.txt`.

#### A2. Handling errors

Many functions in Go can fail, including `grpc.Dial()`. Go handles this by having potentially-failing functions return a tuple of `(result, err)` with `err` being non-`nil` if the call failed. If you are unfamiliar with this, check out the Go blog on error handling: https://go.dev/blog/error-handling-and-go

gRPC method handlers follow this standard as well. `GetTopVideos` may return an error as part of its return value, which follows the gRPC error conventions: https://www.grpc.io/docs/guides/error/

For all cases where functions you are calling (like `Dial()`) fail, you must appropriately handle these errors and not crash the server. Propagate these back to the return value of `GetTopVideos`, and use `status.Errorf()` to add appropriate error codes and error messages: https://pkg.go.dev/google.golang.org/grpc/status. *All* errors you return from your handler methods **must** have an appropriate error code set.
For handling errors, consider also logging them to your output with `log.Printf` or similar methods to help yourself debug these cases later.

Client objects also must be closed when you are done with them -- be sure to call `conn.Close()` appropriately. The [`defer` statement](https://go.dev/tour/flowcontrol/12) may be useful.

**Discussion**: In what cases do you think `Dial()` will fail? What status code do you think you should return for these cases? Include your responses and reasoning (why you picked the particular code) under a heading **A2** in `discussions.txt`.

#### A3. Fetching the user and the users they subscribe to

Once you have a generic client connection from `Dial`, you can get a typed object to make RPCs to UserService:
```
import  upb "cs426.yale.edu/lab1/user_service/proto"
// ...
userClient := upb.NewUserServiceClient(conn)
```

`userClient` effectively implements all the methods from the `UserService` protobuf file---you can call `GetUser(ctx, GetUserRequest {...})` and get back a slice of `UserInfo`s from the response (after you start the UserService in **A6**).

To find the set of videos to rank, you will need to first find the users that the original user (specified by `UserId` on the original request) subscribes to, which is indicated by the `SubscribedTo` slice on the response from `UserService`.

With subscribed-to user IDs, make another call to the `UserService` to find *their* `LikedVideos`.
Be sure to handle errors from `GetUser` correctly, just as you did in **A2** for `Dial`.

**Discussion**: What networking/system calls do you think will be used under the hood when you call `GetUser()`? What cases do you think this call / these calls can return an error? You may find the slides from Lecture 3 helpful. Could `GetUser` return errors in cases where the network calls succeed? Include your responses under a heading **A2** in `discussions.txt`.

**ExtraCredit1**: How are these errors detected? Include your responses under a heading `ExtraCredit1` in `discussions.txt`.

#### A4. Fetching video data

Repeat the process from **A1-A3** but using the appropriate `videoServiceAddr` and `NewVideoServiceClient` to create a client to `VideoService`. Use this to fetch
`VideoInfo`s for the union of all `LikedVideos` from the `SubscribedTo` set of users. Be sure to handle errors and close connections
appropriately.

**Discussion**: What would happen if you used the same `conn` for both `VideoService` and `UserService`? Include your responses under a heading `A4` in `discussions.txt`. This is not intended to be a trick question.

#### A5. Ranking the returned videos

After fetching all of the `VideoInfo`s, you should have a set of candidate videos to return
back to the user. The final step is ranking them to put them in order---the ranking algorithm
does some complex "machine learning" to try to predict what the user might want to watch.

We've provided a ranking library for you in `cs426.yale.edu/lab1/ranker`
(in the `ranker/` directory). Import it and use an instance of `BcryptRanker` to rank the videos. You may create an instance of `BcryptRanker` for every request, and it can be used with no special construction (the zero-value is valid).

To rank a single video, you'll need the `UserCoefficients` from the original user ID (not the subscribed-to-users), and the `VideoCoefficients` for that video.

Return the list of candidate videos in **descending** rank order (highest score is a better match), and truncate the list based on the limit in the `GetTopVideosRequest`. If no limit is set (a value of `0`), return all videos.

Warning: because the ranking algorithm is extremely "sophisticated", for different user video  coefficients pairs, some may take longer than others to compute the scores. No need to be alarmed. The intention is to simulate the differences in computation latency in "real world" scenarios.

#### A6. Testing your implementation as-is

Congratulations! If you've done the preceeding steps all correctly, you should have a basic
working implementation of `VideoRecService`. To test all of this out locally you'll need to
run an instance all 3 microservices together. In 3 separate shells (or using background jobs in one shell), run:
 1. `go run user_service/server/server.go`
 2. `go run video_service/server/server.go`
 3. `go run video_rec_service/server/server.go`

Once all of those start successfully, run our provided test client with **your own NetId**! For example, Xiao's NetId is `xs66`, therefore, the command to run is the following:
```
go run cmd/frontend/frontend.go --net-id=xs66
```

You should see some fake user info and recommended movies printed out. Include the name of the
user picked for you as well as the top recommended video in your `discussions.txt` under heading **A6**.

`frontend.go` also includes tests for two hardcoded UserIds and results, which you can use to verify your implementation.

#### A7. Unittesting with mocks
Unit testing is a good way to catch server logic bugs in short and contained
test runs. You should add your own unit tests in a separate
`video_rec_service/server/server_test.go` file to test the functionality you
built as you go (for functionality up till now as well as the rest of the lab).

Since the bulk of the application logic goes into the server_lib of VideoRecServiceServer, here's how you can implement the unit tests:
 - Complete the function `MakeVideoRecServiceServerWithMocks`, which takes in options as well as the provided mock clients for the User and Video services. You can read the code for the mock clients in `user|video_service/mock_client/mock_client.go`, which are quite straight forward;
 - In `server_test.go`, add unittest skeleton code following the Golang testing tutorials such as [this](https://go.dev/doc/tutorial/add-a-test) and [this](https://pkg.go.dev/testing);
 - In each unittest, spin up a `VideoRecServiceServer` by calling `MakeVideoRecServiceServerWithMocks` with the desired options.

By the end of this lab, you should add at least 5 unit tests. Your tests should have coverage (i.e., at least part of one unit test) on the basic functionality, batching, stats, error handing, retrying, fallback to trending videos. You of course are welcome to add more tests than just 5.

`go test -run='.*' -v` should pass on your implementation as well as our private reference implementation. (**Extra credit** will be given to tests that caught bugs in our reference implementation.)

Note: your actual server **must** be able to communicate with the user|video services via grpc and not solely rely on the mock clients. When we grade the server binary, we will use different undisclosed random seeds for user|video services that your video recommendation server is unaware of; an attempt to use the mock clients to avoid implementing more complex logic such as error handling will be considered cheating.

#### A8. Batching

Both `VideoService` and `UserService` support a batch API---you can send multiple user IDs or video IDs and get back a batched response for all the requested IDs. In practice, many services have an upper bound on how many sub-requests they will handle at once, to prevent large requests from occupying too many resources (the smaller batches may also be easier to spread among backend machines in a distributed environment).

Use the flag value `maxBatchSize` as the maximum amount of IDs to send in any one request to `VideoService` or `UserService`. If the set of IDs is too large, send multiple requests splitting
the set of IDs among them.

You can test that your functionality works by starting `UserService` and `VideoService` locally (like before) and setting `--batch-size=5` on all three services, then running the steps from **A6** again to see that it does not error.


**Discussion**: Should you send the batched requests concurrently? Why or why not? What are the advantages or disadvantages? Include your responses under a heading `A8` in `discussions.txt`.

**ExtraCredit2**: Assume that handling one batch request is cheap up to the batch size limit. How could you reduce the total number of requests to `UserService` and `VideoService` using batching, assuming you have many incoming requests to `VideoRecService`? Describe how you would implement this at a high-level (bullet points and pseudocode are fine) but you do not need to implement it in your service.
Include your responses under a heading `ExtraCredit2` in `discussions.txt`.

### Part B. Implementing GetStats

A key part of operating a real service is **observability**---being able to monitor and inspect the service, see if it is healthy, and find out what is wrong if it is not healthy. To this end, you will implement a primitive stats API as an RPC (though for a production serivce, you may likely have stats baked in to the framework or libraries you are using).

#### B1. Implement `GetStats`
To implement the `GetStats` method, you'll need to keep track of calls to your `GetTopVideos` implementation including:

1. The total number of requests to `GetTopVideos` (ignore calls to `GetStats`) over the uptime of your service, as a simple counter.
2. Total number of requests to `GetTopVideos` which returned an error over the uptime of your service, as a counter.
3. The current number of active requests or *in-progress* calls to `GetTopVideos`.
4. The total number of errors returned in attempts to call `UserService` or `VideoService` respectively (in calls to `Dial` or the RPC methods themselves).
5. The average latency (in milliseconds) of processing `GetTopVideos` requests over all requests over the uptime of your service.
6. For **ExtraCredit3**, the 99th percentile latency (in ms) of processing `GetTopVideos` requests.

These stats can be tracked as state in your`VideoRecServiceServer` type or elsewhere, but they need to be thread-safe. You'll need to update them as calls to `GetTopVideos` progress as part of your request handling logic. Using the [`defer` statement](https://go.dev/tour/flowcontrol/12) may be helpful for some of these cases.

Add a skeleton for the `GetStats` method to your `server_lib.go`

```
func (server *VideoRecServiceServer) GetStats(
	ctx context.Context,
	req *pb.GetStatsRequest,
) (*pb.GetStatsResponse, error) {
}
```
and use the recorded stats to fill out the appropriate response.


#### B2. Testing your stats implementation

Start your services in separate shells following instructions in **A6**, and then start the provided
stats client with `go run cmd/stats/stats.go` in another. Initially all your metrics should be zero.

Run the continuous load generation tool with `go run cmd/loadgen/loadgen.go --target-qps=10` and watch your metrics go up. Note neither the stats client nor the loadgen tool terminates on its own, feel free to Ctrl+C after you see non-zero values. Include a copy of your metrics including average latency under heading **B2** in your `discussions.txt`.

Notes:
* You may need to run `go get golang.org/x/time/rate`, which is a rate limiting library that loadgen uses.
* In the process of running loadgen (especially if you attempt a higher target QPS), you may run into errors like "socket: too many open files" (i.e., hitting the file descriptor limit). Check `ulimit -n` and set it to a higher number (e.g., `ulimit -n 65535`).

### Part C. Handling errors from dependencies

When running a service in production, you may have to deal with transient (or extended) periods
of unavailability---networks are unreliable, services have bugs, and machines crash. Continuing
to provide useful responses in degraded scenarios is an important part of a healthy and reliable
service.

Start up the set of services as in **A6**, but add the option `--failure-rate 10`
(1/5 requests fail randomly) to the arguments
of the video server (i.e., `go run video_service/server/server.go --failure-rate 10`) and start the stats client to monitor your service.

Run the loadgen client with 10 target QPS (see instructions in **B2**) and see how your service responds when 1 out of 10 requests to `VideoService` fail.

#### C1. Retrying failed upstream requests

One strategy for improving reliability in the face of short-lived transient failures is to simply *retry* a request to an upstream service. A request may fail due to a network blip or a service quickly restarting, and retrying can help cover these scenarios. In a distributed environment, retries critically may try a *different* backend host that may not be experiencing issues.

In case of an error from `UserService` or `VideoService`, add a *single* retry on failed RPCs and *single* retry on failed calls to `Dial`.

Run the test from above (with a failure rate of 1 in 10 on VideoService) and see how your service performs now.

**Discussion**: Why might retrying be a bad option? In what cases should you *not* retry a request? Add your response under heading **C1** in `discussions.txt`.

#### C2. Fallback recommendations

Retrying can't fix every issue, sometimes services may be down for an extended period of time.
In the case of extended downtime, we still want to provide *some* video recommendations to the
user. `VideoService` has implemented a trending videos RPC `GetTrendingVideos` which returns
globally popular videos. We can use these trending videos as a *fallback* recommendation to the
user when we cannot provide personalized recommendations.

One catch with using this fallback is that `VideoService` may be the service experiencing
issues --- we can't call the trending API if it is down! Your job will to build a *cache*
of the trending videos and keep it relatively up-to-date to use as a fallback strategy.
Overall general strategy:

1. Store state about trending videos in a thread-safe way (recall Lab 0 techniques) in your server.
2. Proactively fetch trending video IDs using gRPC (similar to **A4**), and fetch their matching video infos using another RPC (keeping batching in mind as well).
3. Periodically refetch and update the trending videos so they are not stale. Use the `expiration_time_s` field of the `GetTrendingVideosResponse` as a guide (it is a unixtime in seconds, see `time.Now().Unix()` in Go) for the time to refetch. If you cannot successfully fetch from `VideoService`, back-off for a bit --- wait at least 10 seconds.
4. If `VideoRecService` service experiences errors that cannot be solved with retries, return a response using the fallback trending videos instead of the personalized ranked videos.

You may want to refactor your code, splitting it into smaller helper methods to make this manageable. Recall concurrent programming (in particular spawning Goroutines and using `RWMutex`) from Lab 0 as tools that may help you here. Considering adding more tests (as in **A6**) to unit test bits of your functionality.

In cases where this fallback strategy is used we may want the users of our API to know so
they can show a message that the recommendations may be degraded (or just log it to their logs).
If fallback recommendations are used on a request, set the `stale_response` field on the `GetTopVideosResponse`. On a similar note, keep track of the number of requests that returned
a fallback response and use it to fill the `stale_responses` part of the `GetStats` RPC.


**Discussion**: What should you do if `VideoService` is still down and your responses are past expiration? Return expired responses or an error? What are the tradeoffs?

For this lab, prefer to return expired responses over an error if you can cache at least one
successful set of trending videos. Note down the tradeoffs of the strategies under heading **C2**
in `discussions.txt`.

You can verify your fallback strategy works by running the test from above and stopping user service entirely mid-run.
You should see `stale_responses` increase in the loadgen output.

#### C3. Other reliability strategies

Name at least one additional strategy you could use to improve the reliability
of `VideoRecService` (successful, and useful responses) in the event of failures of `UserService`
or `VideoService`. What tradeoffs would you make and why? Add your response under heading **C3** in `discussions.txt`.

#### C4. Connection management
In part **A** you likely created new connections via `grpc.Dial()` to `UserService` and `VideoService` on every
request when you needed to use them. What might be costly about connection establishment? (hint: for a high-throughput service you would want to avoid repeated connection
establishment.) How could you change your implementation to avoid per-request connection establishment? Does it have any tradeoffs (consider
topics such as load balancing from the course lectures)? Note your discussion under **C4** in `discussions.txt`, but
you do not have to change your implementation.


# End of Lab 1


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
