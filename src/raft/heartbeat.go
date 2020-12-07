package raft

// heartbeat
type SendHeartbeatArgs struct {
	PeerID int
}

type SendHeartbeatReply struct {
	Ack bool
}

func (rf *Raft) HeartbeatHandler(args *SendHeartbeatArgs, reply *SendHeartbeatReply) {

}

func (rf *Raft) sendHeartbeat(server int, args *SendHeartbeatArgs, reply *SendHeartbeatReply) bool {
	args.PeerID = rf.me
	ok := rf.peers[server].Call("Raft.HeartbeatHandler", args, reply)
	return ok
}