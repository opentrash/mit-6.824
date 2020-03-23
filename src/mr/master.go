/**
 * What does master do ?
 * 1. Manage workers and tasks
 * 2. Hand out tasks to workers
 * The master should notice if a worker hasn't completed
 * its task in a reasonable amount of time
 * (for this lab, use ten seconds),
 * and give the same task to a different worker.
 */

package mr

import "log"
import "net"
import "os"
import "net/rpc"
import "net/http"
import "sync"
import "time"

// for debug
import "fmt"

// store tasks [ map, reduce ]
// store worker machines

type WorkerState uint8

const (
	Worker_Idle WorkerState = iota
	Worker_Busy
	Worker_Lost
)

type WorkerRec struct {
	Id        int
	State     WorkerState
	LastAlive int64
}

var mLock sync.Mutex

type Master struct {
	// Your definitions here.
	Workers []WorkerRec
	Tasks   []string
}

// Your code here -- RPC handlers for the worker to call.

//
// register worker
// must lock
//
func (m *Master) RegisterWorker(_ *RegisterWorkerArgs, reply *RegisterWorkerReply) error {
	mLock.Lock()
	defer mLock.Unlock()

	// generate new worker id by incr 1
	// starts with 1
	newWorkerId := len(m.Workers) + 1
	newWorker := WorkerRec{
		Id:    newWorkerId,
		State: Worker_Idle,
	}
	m.Workers = append(m.Workers, newWorker)
	reply.Id = newWorkerId
	reply.State = Worker_Idle
	return nil
}

//
// worker's heartbeat
// must lock
//
func (m *Master) ListenHeartbeat(args *WorkerHeartbeatArgs, reply *WorkerHeartbeatReply) error {
	mLock.Lock()
	defer mLock.Unlock()
	fmt.Println("Heartbeat from worker ", args.Id)
	workerId := args.Id
	for idx, worker := range m.Workers {
		if worker.Id == workerId {
			worker.LastAlive = time.Now().Unix()
			m.Workers[idx] = worker
			reply.Ack = true
			return nil
		}
	}
	return nil
}

//
// assign task to worker
// must lock
//
func (m *Master) AssignTask(args *TaskDistributeArgs, reply *TaskDistributeReply) error {
	return nil
}

//
// check workers every 2 seconds
// mark the worker as lost if this worker didn't
// response in the last 10 seconds
//
func (m *Master) scanWorkers() {
	mLock.Lock()
	defer mLock.Unlock()
	for idx, worker := range m.Workers {
		// worker lost
		timeGap := time.Now().Unix() - worker.LastAlive
		if timeGap > 10 {
			// if this worker is dealing with some task
			// assign the task to another idle worker
			if worker.State == Worker_Busy {

			}

			// mark the worker as lost
			worker.State = Worker_Lost
			m.Workers[idx] = worker
		}
	}
	m.printWorkers()
}

//
// scan each worker every 2 (for example) seconds
//
func (m *Master) scanWorkersEvery(seconds time.Duration) {
	for {
		m.scanWorkers()
		time.Sleep(seconds * time.Second)
	}
}

func (m *Master) printWorkers() {
	for _, worker := range m.Workers {
		fmt.Printf("Worker Id: %v    ", worker.Id)
		state := "Idle"
		switch worker.State {
		case Worker_Idle:
			state = "Idle"
			break
		case Worker_Busy:
			state = "Busy"
			break
		case Worker_Lost:
			state = "Lost"
			break
		default:
			break
		}
		fmt.Printf("State: %v\n", state)
	}
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	os.Remove("mr-socket")
	l, e := net.Listen("unix", "mr-socket")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil) // concurrence running for RPC listener
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
// need to use lock ?
func (m *Master) Done() bool {
	ret := false

	// Your code here.

	// check if all tasks are done

	return ret
}

//
// create a Master.
// files: os.Args[1:] pg-*.txt
// nReduce: 10 for example
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{
		Workers: []WorkerRec{},
		Tasks:   []string{},
	}

	// Your code here.

	// get files from mrmaster.go

	for _, file := range files {
		fmt.Println("file:", file)
	}

	// register tasks

	// rpc
	// listen to workers
	m.server()
	go m.scanWorkersEvery(2)

	// Your code here.

	return &m
}
