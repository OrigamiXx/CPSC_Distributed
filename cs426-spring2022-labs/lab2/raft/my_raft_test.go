package raft

//
// Raft tests.
//
// we will use the original test_test.go to test your code for grading.
// so, while you can modify this code to help you debug, please
// test with the original before submitting.
//

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestContinuousFail2D(t *testing.T) {
	servers := 7
	cfg := make_config(t, servers, false, false)
	defer cfg.cleanup()

	cfg.checkOneLeader()

	iters := 3
	for ii := 1; ii < iters; ii++ {
		leader := cfg.checkOneLeader()
		fmt.Println("Checkpoint 1")
		cfg.disconnect(leader)

		time.Sleep(1 * RaftElectionTimeout)
		leader = cfg.checkOneLeader()
		fmt.Println("Checkpoint 2")
		cfg.disconnect(leader)

		time.Sleep(1 * RaftElectionTimeout)
		leader = cfg.checkOneLeader()
		fmt.Println("Checkpoint 3")
		cfg.disconnect(leader)

		time.Sleep(1 * RaftElectionTimeout)
		leader = cfg.checkOneLeader()
		fmt.Println("Checkpoint 4")
		cfg.disconnect(leader)

		cfg.checkNoLeader()

		cfg.connect(leader)
		time.Sleep(1 * RaftElectionTimeout)
		cfg.checkOneLeader()
	}

	cfg.end()
}

func TestElection2D(t *testing.T) {
	servers := 7
	cfg := make_config(t, servers, false, false)
	defer cfg.cleanup()

	cfg.checkOneLeader()

	iters := 3
	for ii := 1; ii < iters; ii++ {
		leader := cfg.checkOneLeader()
		fmt.Println("Checkpoint 1")
		cfg.disconnect(leader)
		cfg.disconnect((leader + 1) % servers)
		cfg.disconnect((leader + 2) % servers)
		cfg.disconnect((leader + 3) % servers)

		fmt.Println("Checkpoint 2")
		time.Sleep(1 * RaftElectionTimeout)
		cfg.checkNoLeader()

		fmt.Println("Checkpoint 3")
		cfg.connect(leader)
		cfg.connect((leader + 1) % servers)
		cfg.connect((leader + 2) % servers)

		time.Sleep(1 * RaftElectionTimeout)
		cfg.checkOneLeader()
	}

	cfg.checkOneLeader()
	cfg.end()
}

func TestAgreeLeaderFailAgree2D(t *testing.T) {
	servers := 7
	cfg := make_config(t, servers, false, false)
	defer cfg.cleanup()

	cfg.one(77, servers, false)

	leader := cfg.checkOneLeader()
	fmt.Println("Checkpoint 1")

	cfg.disconnect((leader + 1) % servers)
	cfg.disconnect((leader + 2) % servers)

	time.Sleep(1 * RaftElectionTimeout)
	cfg.one(102, servers-2, false)
	cfg.one(103, servers-2, false)

	fmt.Println("Checkpoint 2")

	cfg.disconnect(leader)
	time.Sleep(2 * RaftElectionTimeout)

	cfg.one(102, servers-3, false)
	cfg.one(103, servers-3, false)

	cfg.connect((leader + 1) % servers)
	cfg.connect((leader + 2) % servers)

	fmt.Println("Checkpoint 3")

	time.Sleep(1 * RaftElectionTimeout)
	cfg.one(103, servers-1, false)
	cfg.one(102, servers-1, false)

	cfg.end()
}

func TestAgreeLeaderFailNoAgree2D(t *testing.T) {
	servers := 7
	cfg := make_config(t, servers, false, false)
	defer cfg.cleanup()

	cfg.one(77, servers, false)

	leader := cfg.checkOneLeader()
	fmt.Println("Checkpoint 1")

	cfg.disconnect((leader + 1) % servers)
	cfg.disconnect((leader + 2) % servers)

	time.Sleep(1 * RaftElectionTimeout)
	cfg.one(102, servers-2, false)
	cfg.one(103, servers-2, false)

	fmt.Println("Checkpoint 2")

	cfg.disconnect(leader)
	time.Sleep(2 * RaftElectionTimeout)

	cfg.one(102, servers-3, false)
	cfg.one(103, servers-3, false)

	leader2 := cfg.checkOneLeader()
	cfg.disconnect(leader2)

	index, _, _ := cfg.rafts[leader2].Start(177)
	if index != 6 {
		t.Fatalf("expected index 6, got %v", index)
	}

	fmt.Println("Checkpoint 3")

	time.Sleep(2 * RaftElectionTimeout)

	n, _ := cfg.nCommitted(index)
	if n > 0 {
		t.Fatalf("%v should not commit no quorom", n)
	}

	cfg.connect(leader2)

	fmt.Println("Checkpoint 4")

	time.Sleep(1 * RaftElectionTimeout)
	cfg.one(103, servers-3, false)
	cfg.one(102, servers-3, false)

	n, _ = cfg.nCommitted(index)
	if n == 0 {
		t.Fatalf("%v should commit have quorom", n)
	}

	cfg.end()
}

func TestNoBadLog2D(t *testing.T) {
	servers := 7
	cfg := make_config(t, servers, false, false)
	defer cfg.cleanup()

	// cfg.one(101, servers, true)
	// cfg.one(102, servers, true)

	leader1 := cfg.checkOneLeader()
	_, _, isLeader := cfg.rafts[(leader1+1)%servers].Start(rand.Int())
	if isLeader != false {
		t.Fatal("follower cannot start")
	}

	cfg.disconnect(leader1)
	time.Sleep(1 * RaftElectionTimeout)

	fmt.Println("Checkpoint 1")

	for i := 0; i < 30; i++ {
		cfg.rafts[leader1].Start(rand.Int())
	}

	fmt.Println("Checkpoint 2")

	for i := 0; i < 10; i++ {
		cfg.one(175, servers-2, true)
	}

	leader2 := cfg.checkOneLeader()

	for i := 0; i < 20; i++ {
		index, _, _ := cfg.rafts[leader2].Start(177)
		if index != i+11 {
			t.Fatalf("wrong index, got %v", index)
		}
	}

	fmt.Println("Checkpoint 3")

	time.Sleep(5 * RaftElectionTimeout) //wait for the RPCs to catch up thus updating followers
	cfg.disconnect(leader2)
	time.Sleep(1 * RaftElectionTimeout)

	for i := 0; i < 20; i++ { // should be irrelevant to working nodes
		index, _, _ := cfg.rafts[leader2].Start(rand.Int())
		if index != i+31 {
			t.Fatalf("leader 2 wrong index, got %v", index)
		}
	}

	fmt.Println("Checkpoint 4")

	leader3 := cfg.checkOneLeader()
	// time.Sleep(5 * RaftElectionTimeout)

	for i := 0; i < 20; i++ {
		index, _, _ := cfg.rafts[leader3].Start(rand.Int())
		if index != i+31 {
			t.Fatalf("leader 3 wrong index, got %v", index)
		}
	}

	cfg.connect(leader1)
	cfg.connect(leader2)
	cfg.one(rand.Int(), servers, true)
	cfg.one(rand.Int(), servers, true)
	cfg.one(rand.Int(), servers, true)

	cfg.end()
}

func TestPersistHotCrashNotBreak2D(t *testing.T) {
	servers := 7
	cfg := make_config(t, servers, false, false)
	defer cfg.cleanup()

	cfg.one(11, servers, true)

	for i := 0; i < servers; i++ {
		cfg.start1(i, cfg.applier)
	}
	for i := 0; i < servers; i++ {
		cfg.disconnect(i)
		cfg.connect(i)
	}

	fmt.Println("Checkpoint 1")

	cfg.one(rand.Intn(100), servers, true)

	leader1 := cfg.checkOneLeader()
	cfg.crash1((leader1 + 0) % servers)
	cfg.crash1((leader1 + 1) % servers)

	fmt.Println("Checkpoint 2")

	cfg.one(rand.Intn(100), servers-2, true)
	cfg.one(rand.Intn(100), servers-2, true)

	leader2 := cfg.checkOneLeader()
	cfg.disconnect(leader2)
	cfg.start1(leader2, cfg.applier)
	cfg.connect(leader2)

	for i := 0; i < 10; i++ {
		cfg.one(175, servers-2, true)
	}

	fmt.Println("Checkpoint 3")

	cfg.start1((leader1+0)%servers, cfg.applier)
	for i := 0; i < 20; i++ {
		cfg.rafts[leader1].Start(rand.Int()) //these should not commit
	}
	cfg.connect((leader1 + 0) % servers)

	fmt.Println("Checkpoint 4")

	cfg.start1((leader1+1)%servers, cfg.applier)
	cfg.connect((leader1 + 1) % servers)

	for i := 0; i < 10; i++ {
		cfg.one(175, servers, true)
	}

	fmt.Println("Checkpoint 5")

	cfg.end()
}
