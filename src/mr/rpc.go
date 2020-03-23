package mr

//
// RPC definitions.
//

// worker tells master that it's still alive
type WorkerHeartbeatArgs struct {
	Id int
}

type WorkerHeartbeatReply struct {
	Ack bool
}

// register worker
type RegisterWorkerArgs struct {
}

type RegisterWorkerReply struct {
	Id    int
	State WorkerState
}

// tasks distribute
type TaskDistributeArgs struct {
}

type TaskDistributeReply struct {
}

// Add your RPC definitions here.
