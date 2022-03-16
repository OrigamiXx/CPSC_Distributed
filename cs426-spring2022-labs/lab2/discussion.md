# 2A-2
(1) Construct and describe a scenario with a Raft cluster of 3 or 5 nodes where the leader election protocol fails to elect a leader. 
Hint: in your description, you may decide when timers time out or not time out, or arbitrate when RPCs get sent or processed.

This can happen if the timers are not randomized. Consider the scenario where someone carelessly wrote the timers such that the election timer resets are not randomized.
Consider a 5 node Raft system, after a period of normal execution, the leader fails. An unfortuntely coincidence caused 3 nodes(or more) to timeout their heartbeat timers and start
sending request vote RPCs, and each of these 3 nodes turned into candidate states before they received the request vote RPCs from other candidate nodes. In the first round of voting,
it will fail because we have 1 failed node, 3 candidate nodes and 1 follower node who will vote for one of the three candidates(though its vote wouldn't matter), and no candidate node
will be able to get a majority vote since all candidates are voting themselves. Now of course they will attempt for the second round of voting, but if the election timer is not
randomized, obviously this cycle of everyone voting themselves would continue as the request vote RPCs are always sent at the same time, and no candidate would ever be able to get
majority vote.

(2) In practice, why is this not a major concern? i.e., how does Raft get around this theoretical possibility?
First of all, to let this happen, the first conincidence have to appear, that "unfortuntely coincidence caused 3 nodes(or more) to timeout their heartbeat timers roughly at the same time", at
least they need to time out in quick succession before the request vote RPCs come in. This is not actually that common.

But just to make sure the above is solved, we can tapple the problem by randomizing the election timer. By doing this eventually some candidate can send out their request vote RPC before other candidates' election timeout. This vote will succeed because this request RPC will have higher term number than other candidates have currently. Problem sovled.
