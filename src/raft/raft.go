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
	"math/rand"

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
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

const (
	Leader    = 0
	Candidate = 1
	Follower  = 2
)

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
	currentTerm int
	votedFor    int
	log         []*logEntry
	role        int

	// commitIndex int
	// lastApplied int

	// Only in leader
	// nextIndex  []int
	// matchIndex []int
	lastPlayEntry time.Time

	// Only in follower
	lastReceiveEntry time.Time
}

type logEntry struct {
	Term int
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// Your code here (2A).
	return rf.currentTerm, rf.role == Leader
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
}

//
// A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
//
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

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

type AppendEntriesArgs struct {
	Term int
	// LeaderId     int
	// PrevLogIndex int
	// PrevLogTrem  int
	// Entries      []logEntry
	// LeaderCommit int
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

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) lastLogTerm() int {
	return 0
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	//dprintf("RaftNode[%d] Handle RequestVote, CandidatesId[%d] Term[%d] CurrentTerm[%d] LastLogIndex[%d] LastLogTerm[%d] votedFor[%d]",
	// rf.me, args.CandidateId, args.Term, rf.currentTerm, args.LastLogIndex, args.LastLogTerm, rf.votedFor)
	// defer func() {
	//dprintf("RaftNode[%d] Return RequestVote, CandidatesId[%d] VoteGranted[%v] ", rf.me, args.CandidateId, reply.VoteGranted)
	// }()

	reply.Term = rf.currentTerm
	reply.VoteGranted = false

	if args.Term < rf.currentTerm {
		return
	}
	if args.Term > rf.currentTerm {
		// DPrintf("RequestVote: Node %d args.Id %d args.Term %d currentTerm %d", rf.me, args.CandidateId, args.Term, rf.currentTerm)
		rf.currentTerm = args.Term
		rf.role = Follower
		rf.votedFor = -1
	}

	if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
		if args.LastLogTerm < rf.lastLogTerm() || args.LastLogIndex < len(rf.log) {
			return
		}
		rf.votedFor = args.CandidateId
		reply.VoteGranted = true
		rf.lastReceiveEntry = time.Now()
	}
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	//dprintf("RaftNode[%d] Handle AppendEntries, Term[%d] CurrentTerm[%d] role=[%d]",
	// rf.me, args.Term, rf.currentTerm, rf.role)
	// defer func() {
	//dprintf("RaftNode[%d] Return AppendEntries, Term[%d] CurrentTerm[%d] role=[%d]",
	// rf.me, args.Term, rf.currentTerm, rf.role)
	// }()

	reply.Term = rf.currentTerm
	reply.Success = false

	if args.Term < rf.currentTerm {
		return
	}
	if args.Term > rf.currentTerm {
		// DPrintf("AppendEntries: Node %d args.Term %d currentTerm %d", rf.me, args.Term, rf.currentTerm)
		rf.currentTerm = args.Term
		rf.role = Follower
		rf.votedFor = -1
	}
	rf.lastReceiveEntry = time.Now()
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
func (rf *Raft) sync() {
	for !rf.killed() {
		time.Sleep(1 * time.Millisecond)

		func() {
			rf.mu.Lock()
			defer rf.mu.Unlock()

			if rf.role != Leader || time.Since(rf.lastPlayEntry) < 100*time.Second {
				return
			}

			rf.lastPlayEntry = time.Now()

			type AppendResult struct {
				peerId int
				resp   *AppendEntriesReply
			}
			for peerId := 0; peerId < len(rf.peers); peerId++ {
				if peerId == rf.me {
					continue
				}

				args := AppendEntriesArgs{}
				args.Term = rf.currentTerm
				// args.LeaderId = rf.me
				// log相关字段在lab-2A不处理
				go func(id int, args1 *AppendEntriesArgs) {
					// DPrintf("RaftNode[%d] appendEntries starts, myTerm[%d] peerId[%d]", rf.me, args1.Term, id)
					reply := AppendEntriesReply{}
					if ok := rf.sendAppendEntries(id, args1, &reply); ok {
						rf.mu.Lock()
						defer rf.mu.Unlock()
						if reply.Term > rf.currentTerm { // 变成follower
							rf.role = Follower
							// rf.leaderId = -1
							rf.currentTerm = reply.Term
							rf.votedFor = -1
							// rf.persist()
						}
						// DPrintf("RaftNode[%d] appendEntries ends, peerTerm[%d] myCurrentTerm[%d] myRole[%s]", rf.me, reply.Term, rf.currentTerm, rf.role)
					}
				}(peerId, &args)
			}

			// for i := 0; i < len(rf.peers); i++ {
			// 	if i == rf.me {
			// 		continue
			// 	}
			// 	go func(peerId int, args *AppendEntriesArgs) {
			// 		// DPrintf("RaftNode[%d] appendEntries starts, myTerm[%d] peerId[%d]", rf.me, args.Term, peerId)
			// 		reply := AppendEntriesReply{}
			// 		if ok := rf.sendAppendEntries(peerId, args, &reply); ok {
			// 			rf.mu.Lock()
			// 			defer rf.mu.Unlock()
			// 			if reply.Term > rf.currentTerm {
			// 				rf.role = Follower
			// 				DPrintf("sync: Node %d reply.Id %d, reply.Term %d, currentTerm %d", rf.me, peerId, reply.Term, rf.currentTerm)
			// 				rf.currentTerm = reply.Term
			// 				rf.votedFor = -1
			// 			}
			// 		}
			// 		// DPrintf("RaftNode[%d] appendEntries ends, peerTerm[%d] myCurrentTerm[%d] myRole[%d]", rf.me, reply.Term, rf.currentTerm, rf.role)
			// 	}(i, &AppendEntriesArgs{
			// 		Term: rf.currentTerm,
			// 	})
			// }
		}()
	}
}

func (rf *Raft) election() {
	for !rf.killed() {
		time.Sleep(1 * time.Millisecond)
		func() {
			rf.mu.Lock()
			defer rf.mu.Unlock()

			randomTimeout := time.Duration(200+rand.Intn(200)) * time.Millisecond
			interval := time.Since(rf.lastReceiveEntry)

			if rf.role != Follower || interval < randomTimeout {
				return
			}

			//dprintf("RaftNode[%d] Follower -> Candidate", rf.me)
			rf.role = Candidate
			rf.currentTerm++
			rf.votedFor = rf.me
			rf.lastReceiveEntry = time.Now()

			args := RequestVoteArgs{
				Term:         rf.currentTerm,
				CandidateId:  rf.me,
				LastLogIndex: len(rf.log),
			}
			rf.mu.Unlock()
			type VoteResult struct {
				peerId int
				resp   *RequestVoteReply
			}
			voteCount := 1
			finishCount := 1
			voteResultChan := make(chan *VoteResult, len(rf.peers))
			for peerId := 0; peerId < len(rf.peers); peerId++ {
				go func(id int) {
					if id == rf.me {
						return
					}
					resp := RequestVoteReply{}
					if ok := rf.sendRequestVote(id, &args, &resp); ok {
						voteResultChan <- &VoteResult{peerId: id, resp: &resp}
					} else {
						voteResultChan <- &VoteResult{peerId: id, resp: nil}
					}
				}(peerId)
			}
			maxTerm := 0
			for {
				select {
				case voteResult := <-voteResultChan:
					finishCount += 1
					if voteResult.resp != nil {
						if voteResult.resp.VoteGranted {
							voteCount += 1
						}
						if voteResult.resp.Term > maxTerm {
							maxTerm = voteResult.resp.Term
						}
					}
					// 得到大多数vote后，立即离开
					if finishCount == len(rf.peers) || voteCount > len(rf.peers)/2 {
						goto VOTE_END
					}
				}
			}
		VOTE_END:
			rf.mu.Lock()

			if rf.role != Candidate {
				return
			}
			if maxTerm > rf.currentTerm {
				rf.role = Follower
				rf.currentTerm = maxTerm
				rf.votedFor = -1
				return
			}
			if voteCount > len(rf.peers)/2 {
				rf.role = Leader
				rf.lastPlayEntry = time.Unix(0, 0)
				return
			}
		}()
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
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.lastReceiveEntry = time.Now()
	rf.role = Follower
	// rf.lastPlayEntry = time.Now()
	rf.votedFor = -1
	rf.currentTerm = 0
	rand.Seed(time.Now().UnixNano())

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.election()
	go rf.sync()

	return rf
}
