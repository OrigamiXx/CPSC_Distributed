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
