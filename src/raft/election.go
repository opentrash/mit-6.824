package raft

// func (rf *Raft) electionTimeoutScan(electionTimeout time.Duration) {
//     for {
//         select {
//         case <-time.After(electionTimeout * time.Millisecond):
//             // @todo figure out if this[startNewElection] needs to be a goroutine
//             rf.startNewElection()
//             return
//         case <-rf.receiveLeaderHeartbeatCh:
//             continue
//         }
//     }
// }

// @todo start new election
func (rf *Raft) startNewElection() {

}
