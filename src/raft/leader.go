package raft

import (
	"time"
)

const (
	heartbeatInterval = 5000
)

// run only when this peer is leader
func (rf *Raft) sendHeartbeatToFollowers() {
	for {
		time.Sleep(heartbeatInterval * time.Millisecond)
		select {
		case <-rf.quitLeaderBroadcastHeartbeatCh:
			return
		default:
			for idx := range rf.peers {
				if idx != rf.me {
					args := SendHeartbeatArgs{}
					reply := SendHeartbeatReply{}
					ok := rf.sendHeartbeat(idx, &args, &reply)
					// no heartbeat back
					// @todo need to handle
					if !ok {

					}
				}
			}
		}
	}
}
