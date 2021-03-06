package raft

import (
	"sync"
	"time"

	"6.824/labrpc"
)

type Role string

const (
	Follower  Role = "follower"
	Candidate Role = "candidate"
	Leader    Role = "leader"
)

type Raft struct {
	mu          sync.Mutex
	peers       []*labrpc.ClientEnd
	persister   *Persister
	me          int
	dead        int32
	lastRecv    time.Time
	role        Role
	currentTerm int
	votedFor    int
	log         Log
	commitIndex int
	lastApplied int
	nextIndex   []int
	matchIndex  []int
	applyCh     chan ApplyMsg
	snapshot    []byte
}

func (rf *Raft) ticker() {
	for !rf.killed() {
		rf.mu.Lock()

		/* Rules for Followers */
		/* 2. If election timeout elapses without receiving AppendEntries RPC from current leader or granting vote to candidate: convert to candidate */
		/* Rules for Candidates */
		/* 4. If election timeout elapses: start new election */
		if rf.role != Leader && time.Since(rf.lastRecv) > rf.electTimeout() {
			rf.elect()
		}

		rf.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
}

func (rf *Raft) StateSize() int {
	return rf.persister.RaftStateSize()
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	rf.role = Follower
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.lastRecv = time.Now()

	rf.log.Entries = append(rf.log.Entries, LogEntry{Term: 0})
	rf.log.Base = 0
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))

	rf.applyCh = applyCh

	rf.readPersist(persister.ReadRaftState())
	rf.snapshot = persister.ReadSnapshot()

	go func() {
		rf.mu.Lock()
		rf.applyCh <- ApplyMsg{
			SnapshotValid: true,
			SnapshotIndex: rf.log.Base,
			SnapshotTerm: rf.log.Entries[0].Term,
			Snapshot: rf.snapshot,
		}
		rf.mu.Unlock()
	}()
	rf.commitIndex = 0 // 为了重启后发的快照不被cond拒绝，如果=Base，会在cond返回false
	rf.lastApplied = rf.log.Base

	go rf.ticker()
	go rf.apply()

	return rf
}
