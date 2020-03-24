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
	Id      int
	State   WorkerState
	NReduce int
}

// tasks distribute
type TaskDistributeArgs struct {
	WorkerId int
}

type TaskDistributeReply struct {
	Task    Task
	Message string
}

type SubmitTaskResultArgs struct {
	WorkerId int
	TaskId   int
	Result   string
}

type SubmitTaskResultReply struct {
	Ack bool
}

// Add your RPC definitions here.
