package raft

// heartbeat
type AppendEntriesArgs struct {
	PeerID int
}

type AppendEntriesReply struct {
	Ack bool
}

func (rf *Raft) AppendEntriesHandler(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.receiveLeaderAppendEntriesCh <- true
}

func (rf *Raft) appendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	rf.rwmu.RLock()
	defer rf.rwmu.RUnlock()

	args.PeerID = rf.me
	ok := rf.peers[server].Call("Raft.AppendEntriesHandler", args, reply)
	return ok
}
