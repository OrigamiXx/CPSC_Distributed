package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"

	"6.824/labgob"

	//	"bytes"
	"sync"
	"sync/atomic"
	"time"

	//	"6.824/labgob"
	"6.824/labrpc"
)

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
	CommandTerm  int
}

type NodeState int

const (
	FOLLOWER NodeState = iota
	CANDIDATE
	LEADER
)

type Entry struct {
	Command interface{}
	Term    int
	Index   int
}

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	applyCh chan ApplyMsg
	state   NodeState

	currentTerm int
	votedFor    int
	logs        []Entry

	commitIndex int
	lastApplied int
	nextIndex   []int
	matchIndex  []int

	electionTimer  *time.Timer
	heartbeatTimer *time.Timer

	applyCond *sync.Cond
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	term = rf.currentTerm
	isleader = rf.state == LEADER
	return term, isleader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(rf.logs)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
	rf.mu.Lock()
	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	var logs []Entry
	var currentTerm int
	var voteFor int

	if d.Decode(&logs) != nil ||
		d.Decode(&currentTerm) != nil ||
		d.Decode(&voteFor) != nil {
		fmt.Println("Decoder error")
	} else {
		rf.logs = logs
		rf.currentTerm = currentTerm
		rf.votedFor = voteFor
	}
	rf.mu.Unlock()
}

//AppendEntries的参数
type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Entry
	LeaderCommit int
}

//AppendEntries的返回值
type AppendEntriesReply struct {
	Term    int
	Success bool
	XTerm   int //term in the conflicting entry (if any)
	XIndex  int //index of first entry with that term (if any)
	XLen    int //if the follower has no log at corresponding, log length

	// 	 Case 1 (leader doesn't have XTerm):
	//     nextIndex = XIndex
	//   Case 2 (leader has XTerm):
	//     nextIndex = leader's last entry for XTerm
	//   Case 3 (follower's log is too short):
	//     nextIndex = XLen
}

//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(request *RequestVoteArgs, response *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()
	//fmt.Println(rf.printState() + ":getVoteFor:" + strconv.Itoa(request.CandidateId) + "    voteForTerm:" + strconv.Itoa(request.Term))

	response.Term = rf.currentTerm
	if request.Term > rf.currentTerm {
		rf.state = FOLLOWER
		rf.currentTerm, rf.votedFor = request.Term, -1
	}

	if request.Term < rf.currentTerm {
		response.VoteGranted = false
		return
	}

	if (rf.votedFor == -1 || rf.votedFor == request.CandidateId) && rf.checkLogs(request) {
		rf.votedFor = request.CandidateId
		rf.electionTimer.Reset(electionTimeout())
		response.VoteGranted = true
	}
}

func (rf *Raft) checkLogs(request *RequestVoteArgs) bool {
	if request.LastLogTerm > rf.getLastTerm() {
		return true
	}
	if request.LastLogTerm == rf.getLastTerm() && rf.getLastIndex() <= request.LastLogIndex {
		return true
	}
	return false
}

func (rf *Raft) AppendEntries(request *AppendEntriesArgs, response *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()
	// defer func() {
	// 	log := rf.printState() + "Success="
	// 	if response.Success {
	// 		log += "true"
	// 	} else {
	// 		log += "false"
	// 	}
	// 	log += "    XTerm=" + strconv.Itoa(response.XTerm) + "    XIndex=" + strconv.Itoa(response.XIndex) + "    XLen=" + strconv.Itoa(response.XLen)
	// 	fmt.Println(log)
	// }()

	if request.Term < rf.currentTerm {
		response.Term, response.Success = rf.currentTerm, false
		return
	}
	if request.Term > rf.currentTerm {
		rf.currentTerm, rf.votedFor = request.Term, -1
	}

	rf.state = FOLLOWER
	rf.electionTimer.Reset(electionTimeout())

	if request.PrevLogIndex < rf.getFirstIndex()-1 {
		response.Term, response.Success = 0, false
		return
	}

	if request.PrevLogIndex > rf.getLastIndex() {
		response.Term, response.Success = rf.currentTerm, false
		response.XTerm = -1
		response.XLen = rf.getLastIndex() + 1
		//fmt.Println(rf.printState()+"returnForIndexBig")
		return
	}
	if rf.getTerm(request.PrevLogIndex) != request.PrevLogTerm {
		response.Term, response.Success = rf.currentTerm, false
		response.XTerm = rf.logs[request.PrevLogIndex-rf.getFirstIndex()].Term
		response.XIndex = rf.firstIndexByTerm(response.XTerm)
		//fmt.Println(rf.printState()+"returnForLogCheck")
		return
	}
	for index, entry := range request.Entries {
		if entry.Index-rf.getFirstIndex() >= len(rf.logs) || rf.logs[entry.Index-rf.getFirstIndex()].Term != entry.Term {
			logs := append([]Entry{}, append(rf.logs[:entry.Index-rf.getFirstIndex()], request.Entries[index:]...)...)
			rf.logs = logs // delete conflicting logs
			break
		}
	}
	if request.LeaderCommit > rf.commitIndex {
		if request.LeaderCommit < rf.getLastIndex() {
			rf.commitIndex = request.LeaderCommit
		} else {
			rf.commitIndex = rf.getLastIndex()
		}
		rf.applyCond.Signal()
	}
	//fmt.Println(rf.printState() + ":receiveEntries    from:" + strconv.Itoa(request.LeaderId) + "  fromTerm:" + strconv.Itoa(request.Term))
	response.Term, response.Success = rf.currentTerm, true
}

func (rf *Raft) firstIndexByTerm(term int) int {
	left := 1
	right := len(rf.logs) - 1
	for left <= right {
		mid := left + (right-left)/2
		if rf.logs[mid].Term < term {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	if left >= len(rf.logs) || rf.logs[left].Term != term {
		return -1
	}

	return rf.logs[left].Index
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).
	rf.mu.Lock()
	if rf.state != LEADER {
		rf.mu.Unlock()
		return index, term, false
	}
	index = rf.getLastIndex() + 1
	term = rf.currentTerm
	entry := Entry{
		Command: command,
		Term:    term,
		Index:   index,
	}
	rf.logs = append(rf.logs, entry)
	rf.persist()
	rf.BroadcastHeartbeat()
	rf.mu.Unlock()

	return index, term, isLeader
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently.
func (rf *Raft) ticker() {
	// Your code here to check if a leader election should
	// be started and to randomize sleeping time using
	// time.Sleep().
	for !rf.killed() {
		select {
		case <-rf.electionTimer.C:
			rf.mu.Lock()
			if rf.state != LEADER {
				rf.state = CANDIDATE
				rf.currentTerm += 1
				rf.votedFor = rf.me
				rf.persist()
				request := rf.GetRequestVoteArgs()
				//fmt.Println(strconv.Itoa(rf.me) + ":startElection    term:" + strconv.Itoa(rf.currentTerm))

				rf.StartElection(request)
				rf.electionTimer.Reset(electionTimeout())
			}
			rf.mu.Unlock()
		case <-rf.heartbeatTimer.C:
			rf.mu.Lock()
			if rf.state == LEADER {
				rf.heartbeatTimer.Reset(heartbeatTimeout())
				rf.BroadcastHeartbeat()
			}
			rf.mu.Unlock()
		}
	}
}

func (rf *Raft) StartElection(request *RequestVoteArgs) {
	grantedVotes := 1
	lock := sync.Mutex{}
	for peer := range rf.peers {
		if peer == rf.me {
			continue
		}
		go rf.sendRequestVoteRoutine(peer, request, &grantedVotes, &lock)
	}

}

func (rf *Raft) GetRequestVoteArgs() *RequestVoteArgs {
	args := &RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateId:  rf.me,
		LastLogIndex: rf.getLastIndex(),
		LastLogTerm:  rf.getLastTerm(),
	}
	return args
}

func (rf *Raft) sendRequestVoteRoutine(peer int, request *RequestVoteArgs, grantedVotes *int, lock *sync.Mutex) {
	response := &RequestVoteReply{
		Term:        0,
		VoteGranted: false,
	}
	if rf.sendRequestVote(peer, request, response) {
		rf.mu.Lock()
		if rf.currentTerm == request.Term && rf.state == CANDIDATE {
			//fmt.Println(rf.printState() + ":sendRequestVoteTO:" + strconv.Itoa(peer) + "    term:" + strconv.Itoa(request.Term))
			if response.VoteGranted {
				lock.Lock()
				*grantedVotes += 1
				if *grantedVotes > len(rf.peers)/2 {
					//fmt.Println(rf.printState() + ":electionOK    term:" + strconv.Itoa(rf.currentTerm))
					for i := 0; i < len(rf.nextIndex); i++ {
						rf.nextIndex[i] = rf.commitIndex + 1
					}
					rf.BroadcastHeartbeat()
					rf.state = LEADER
					rf.heartbeatTimer.Reset(heartbeatTimeout())
				}
				lock.Unlock()
			} else if response.Term > rf.currentTerm {
				rf.state = FOLLOWER
				rf.currentTerm, rf.votedFor = response.Term, -1
				rf.persist()
				rf.electionTimer.Reset(electionTimeout())
			}
		}
		rf.mu.Unlock()
	}
}

func (rf *Raft) BroadcastHeartbeat() {
	for peer := range rf.peers {
		if peer == rf.me {
			continue
		}
		args := &AppendEntriesArgs{
			Term:         rf.currentTerm,
			LeaderId:     rf.me,
			PrevLogIndex: rf.nextIndex[peer] - 1,
			PrevLogTerm:  rf.getTerm(rf.nextIndex[peer] - 1),
			Entries:      rf.logs[rf.nextIndex[peer]-rf.getFirstIndex():],
			LeaderCommit: rf.commitIndex,
		}
		reply := &AppendEntriesReply{
			Term:    0,
			Success: false,
		}
		targetIndex := rf.nextIndex[peer] + len(args.Entries)
		go rf.sendAppendEntriesRoutine(peer, args, reply, targetIndex)
	}
}

func (rf *Raft) sendAppendEntriesRoutine(peer int, args *AppendEntriesArgs, reply *AppendEntriesReply, targetIndex int) {
	rf.mu.Lock()
	if rf.state != LEADER || args.Term != rf.currentTerm {
		rf.mu.Unlock()
		return
	}
	rf.mu.Unlock()
	if rf.sendAppendEntries(peer, args, reply) {
		rf.handleAppendEntriesResponse(peer, args, reply, targetIndex)
	}
}

func (rf *Raft) handleAppendEntriesResponse(peer int, args *AppendEntriesArgs, reply *AppendEntriesReply, targetIndex int) {
	rf.mu.Lock()
	if rf.state != LEADER || args.Term != rf.currentTerm {
		rf.mu.Unlock()
		return
	}
	//fmt.Println(rf.printState() + ":doSendAppend to:" + strconv.Itoa(peer) + "    argsEntriesLen=" + strconv.Itoa(len(args.Entries)))
	if reply.Term > rf.currentTerm {
		rf.state = FOLLOWER
		rf.currentTerm, rf.votedFor = reply.Term, -1
		rf.persist()
		rf.electionTimer.Reset(electionTimeout())
		rf.mu.Unlock()
		return
	}
	if reply.Success {
		if rf.nextIndex[peer]+len(args.Entries) == targetIndex {
			rf.nextIndex[peer] = targetIndex
			rf.matchIndex[peer] = rf.nextIndex[peer] - 1
		}
		commitCount := 1
		for i := 0; i < len(rf.peers); i++ {
			if i == rf.me {
				continue
			}
			// commit when quorum
			if rf.matchIndex[i] >= rf.matchIndex[peer] {
				commitCount++
			}
		}

		if commitCount >= len(rf.peers)/2+1 && rf.commitIndex < rf.matchIndex[peer] && rf.logs[rf.matchIndex[peer]-rf.getFirstIndex()].Term == rf.currentTerm {
			rf.commitIndex = rf.matchIndex[peer]
			// majority commited, signal CV
			rf.applyCond.Signal()
		}
	} else {
		if reply.XTerm == -1 {
			rf.nextIndex[peer] = reply.XLen
		} else {
			rf.nextIndex[peer] = reply.XIndex
		}
		argsT := &AppendEntriesArgs{
			Term:         rf.currentTerm,
			LeaderId:     rf.me,
			PrevLogIndex: rf.nextIndex[peer] - 1,
			PrevLogTerm:  rf.getTerm(rf.nextIndex[peer] - 1),
			Entries:      rf.logs[rf.nextIndex[peer]-rf.getFirstIndex():],
			LeaderCommit: rf.commitIndex,
		}
		replyT := &AppendEntriesReply{
			Term:    0,
			Success: false,
		}
		targetIndexT := rf.nextIndex[peer] + len(argsT.Entries)
		go rf.sendAppendEntriesRoutine(peer, argsT, replyT, targetIndexT)
	}
	rf.mu.Unlock()
}

func (rf *Raft) applyLogs() {
	for !rf.killed() {
		rf.mu.Lock()
		if rf.lastApplied >= rf.commitIndex {
			rf.applyCond.Wait()
		}
		if rf.lastApplied < rf.commitIndex {
			commitIndex, lastApplied := rf.commitIndex, rf.lastApplied
			currentTerm := rf.currentTerm
			entries := make([]Entry, commitIndex-lastApplied)
			copy(entries, rf.logs[lastApplied+1-rf.getFirstIndex():commitIndex+1-rf.getFirstIndex()])
			rf.mu.Unlock()
			for _, entry := range entries {
				rf.applyCh <- ApplyMsg{
					CommandValid: true,
					Command:      entry.Command,
					CommandIndex: entry.Index,
					CommandTerm:  currentTerm,
				}
			}
			rf.mu.Lock()
			//fmt.Println(rf.printState() + "lastApplied=" + strconv.Itoa(lastApplied))
			if rf.lastApplied < commitIndex {
				rf.lastApplied = commitIndex
			}
		}
		rf.mu.Unlock()
	}
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int, persister *Persister, applyCh chan ApplyMsg) *Raft {
	//fmt.Println("[Make a raft]... ...")
	rf := &Raft{
		peers:          peers,
		persister:      persister,
		me:             me,
		dead:           0,
		applyCh:        applyCh,
		state:          FOLLOWER,
		currentTerm:    0,
		votedFor:       -1,
		logs:           make([]Entry, 1),
		nextIndex:      make([]int, len(peers)),
		matchIndex:     make([]int, len(peers)),
		heartbeatTimer: time.NewTimer(heartbeatTimeout()),
		electionTimer:  time.NewTimer(electionTimeout()),
		commitIndex:    0,
		lastApplied:    0,
	}
	rf.logs[0] = Entry{
		Command: nil,
		Term:    0,
	}
	rf.applyCond = sync.NewCond(&rf.mu)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())
	// start ticker goroutine to start elections
	go rf.ticker()
	go rf.applyLogs()
	return rf
}

func electionTimeout() time.Duration {
	rand.Seed(time.Now().UnixNano())
	return time.Millisecond * time.Duration(1000+rand.Intn(1000))
}

func heartbeatTimeout() time.Duration {
	return time.Millisecond * 100
}

func (rf *Raft) printState() string {
	state := strconv.Itoa(rf.me) + ",    logsLen=" + strconv.Itoa(len(rf.logs)) + "    term=" + strconv.Itoa(rf.currentTerm) + "    state=" + strconv.Itoa(int(rf.state))
	state += "    lastLogTerm=" + strconv.Itoa(rf.getLastTerm())
	state += "    lastLogIndex=" + strconv.Itoa(rf.getLastIndex())
	for peer := range rf.peers {
		if peer == rf.me {
			continue
		}
		state = state + "   " + strconv.Itoa(peer) + ":" + strconv.Itoa(rf.nextIndex[peer])
	}
	return state
}

func (rf *Raft) getLastTerm() int {
	return rf.logs[len(rf.logs)-1].Term
}

func (rf *Raft) getLastIndex() int {
	return rf.logs[len(rf.logs)-1].Index
}

func (rf *Raft) getFirstIndex() int {
	return rf.logs[0].Index
}

func (rf *Raft) getTerm(index int) int {
	return rf.logs[index-rf.getFirstIndex()].Term
}
