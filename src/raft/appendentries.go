package raft

// heartbeat
type AppendEntriesArgs struct {
	Term         int // leader's term
	LeaderID     int // leader's ID
	PrevLogIndex int // index of log entry immediately preceding new ones
	PrevLogTerm  int // term of PrevLogIndex entry
	LeaderCommit int // leader's commit index

	// TODO: use entry struct, log entries to store
	// (empty for heartbeat; may send more than one for efficiency)
	Entries []int
}

type AppendEntriesReply struct {
	Term    int  // current term for leader to update itself
	Success bool // true if follower contained entry matching prevLogIndex and prevLogTerm
}

// NYI:
func (rf *Raft) AppendEntriesHandler(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.receiveAppendEntriesCh <- true
}

// NYI:
func (rf *Raft) appendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	rf.rwmu.RLock()
	defer rf.rwmu.RUnlock()

	args.LeaderID = rf.me
	ok := rf.peers[server].Call("Raft.AppendEntriesHandler", args, reply)
	return ok
}
