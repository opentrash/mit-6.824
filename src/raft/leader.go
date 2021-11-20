package raft

// run only when this peer is leader
func (rf *Raft) appendEntriesToFollowers() {
	for peerID := range rf.peers {
		if peerID != rf.me {
			args := AppendEntriesArgs{}
			reply := AppendEntriesReply{}
			ok := rf.appendEntries(peerID, &args, &reply)
			// handle append entries response
			if !ok {

			}
		}
	}
}
