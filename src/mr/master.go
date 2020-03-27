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
import "strconv"

// for debug
import "fmt"

// store tasks [ map, reduce ]
// store worker machines

type WorkerState uint8

const (
	WorkerIdle WorkerState = iota
	WorkerBusy
	WorkerLost
)

type MasterState uint8

const (
	InMapping MasterState = iota
	InReducing
	MasterDone
)

type TaskState uint8

const (
	TaskTodo TaskState = iota
	TaskInProgress
	TaskDone
)

type Task struct {
	Id     int
	State  TaskState
	Detail string
	Extra  int
	Type   string
}

// type MapTask struct {
//     Id       int
//     State    TaskState
//     FileName string
// }

// //
// // each word needs a reduce task
// //
// type ReduceTask struct {
//     Id    int
//     State TaskState
//     Word  string
// }

// worker record struct
type WorkerRec struct {
	Id            int
	State         WorkerState
	LastHeartbeat int64
	TaskId        int
}

// Your definitions here.
// MapTask from slice to queue ?
// TODO change slice to map ?
// k MapTasks
// each map task will produce nReduce files for reduce task
// so there would be nReduce reduce tasks to handle mr-*-hash
type Master struct {
	Workers     []WorkerRec
	MapTasks    []Task
	ReduceTasks []Task
	State       MasterState
	NReduce     int
	mLock       sync.Mutex
}

// Your code here -- RPC handlers for the worker to call.

//
// register worker
// must lock
//
func (m *Master) RegisterWorker(_ *RegisterWorkerArgs, reply *RegisterWorkerReply) error {
	m.mLock.Lock()
	defer m.mLock.Unlock()

	// generate new worker id by incr 1
	// starts with 1
	newWorkerId := len(m.Workers) + 1
	newWorker := WorkerRec{
		Id:    newWorkerId,
		State: WorkerIdle,
	}
	m.Workers = append(m.Workers, newWorker)
	reply.Id = newWorkerId
	reply.State = WorkerIdle
	reply.NReduce = m.NReduce
	fmt.Printf("Registed new worker id: %v\n", reply.Id)
	return nil
}

//
// worker's heartbeat
// must lock
// TODO
// maybe master can give instructions to worker in HEARTBEAT
//
func (m *Master) ListenHeartbeat(args *WorkerHeartbeatArgs, reply *WorkerHeartbeatReply) error {
	m.mLock.Lock()
	defer m.mLock.Unlock()
	// fmt.Println("Heartbeat from worker ", args.Id)
	workerId := args.Id
	for idx, worker := range m.Workers {
		if worker.Id == workerId {
			worker.LastHeartbeat = time.Now().Unix()
			m.Workers[idx] = worker
			reply.Ack = true
			return nil
		}
	}
	return nil
}

//
// init Map tasks
//
func (m *Master) RegisterMapTasks(files []string) {
	for _, file := range files {
		mapTask := Task{
			Id:     len(m.MapTasks) + 1,
			Detail: file,
			State:  TaskTodo,
			Type:   "Map",
		}
		m.MapTasks = append(m.MapTasks, mapTask)
	}
	m.State = InMapping
}

//
// init reduce tasks
//
func (m *Master) RegisterReduceTasks(nReduce int) {
	i := 0
	for i < nReduce {
		reduceTask := Task{
			Id:     len(m.ReduceTasks) + 1,
			State:  TaskTodo,
			Type:   "Reduce",
			Detail: strconv.Itoa(i),
			Extra:  len(m.MapTasks),
		}
		m.ReduceTasks = append(m.ReduceTasks, reduceTask)
		i++
	}
}

//
// get worker by id
//
// func (m *Master) GetWorker(workerId int) WorkerRec {
//     for _, worker := range m.Workers {
//         if worker.Id == workerId {
//             return worker
//         }
//     }
//     log.Fatal("There is no worker's id %v\n", workerId)
// }

//
// assign task to worker
// must lock
//
func (m *Master) AssignTask(args *TaskDistributeArgs, reply *TaskDistributeReply) error {
	// find a todo task and return it to worker
	// if there aren't any map tasks to do
	// tell the worker to wait until there are some reduce tasks
	m.mLock.Lock()
	defer m.mLock.Unlock()
	fmt.Printf("Get assignment from worker: %v\n", args.WorkerId)
	for _, worker := range m.Workers {
		if worker.Id == args.WorkerId && (worker.State == WorkerBusy) {
			reply.Message = "Forbidden"
			return nil
		}
		if worker.Id == args.WorkerId && (worker.State == WorkerLost) {
			reply.Message = "Lost"
			return nil
		}
	}
	// fmt.Printf("Looking for an idle task for worker %v\n", args.WorkerId)
	if m.State == MasterDone {
		reply.Message = "MasterDone"
		return nil
	}
	tasks := []Task{}
	if m.State == InMapping {
		tasks = m.MapTasks
	} else {
		tasks = m.ReduceTasks
	}
	for idx, task := range tasks {
		if task.State == TaskTodo {
			// return this task to the worker
			reply.Task = task
			reply.Message = ""
			task.State = TaskInProgress
			for idxWorker, worker := range m.Workers {
				if worker.Id == args.WorkerId {
					fmt.Printf("Assigning task %v to worker %v\n", task.Id, worker.Id)
					worker.TaskId = task.Id
					worker.State = WorkerBusy
					m.Workers[idxWorker] = worker
					if m.State == InMapping {
						m.MapTasks[idx] = task
					} else {
						m.ReduceTasks[idx] = task
					}
					break
				}
			}
			return nil
		}
	}

	reply.Message = "Wait"

	return nil
}

//
// mark task with TaskID as Done
//
func (m *Master) FinishTask(taskId int) {
	tasks := []Task{}
	if m.State == InMapping {
		tasks = m.MapTasks
	} else {
		tasks = m.ReduceTasks
	}
	for idx, task := range tasks {
		if task.Id == taskId {
			task.State = TaskDone
			if m.State == InMapping {
				m.MapTasks[idx] = task
			} else {
				m.ReduceTasks[idx] = task
			}
			break
		}
	}
	// TODO what if there is no task with taskId
}

//
// check if all tasks are done
// to change master state
//
func (m *Master) ScanTasks() {
	allMapDone := true
	for _, task := range m.MapTasks {
		if task.State == TaskTodo || task.State == TaskInProgress {
			allMapDone = false
		}
	}
	if allMapDone {
		m.State = InReducing
	} else {
		return
	}
	allReduceDone := true
	for _, task := range m.ReduceTasks {
		if task.State == TaskTodo || task.State == TaskInProgress {
			allReduceDone = false
		}
	}
	if allReduceDone {
		m.State = MasterDone
	}
}

//
// mark worker state
//
func (m *Master) MarkWorkerState(workerId int, state WorkerState) {
	// fmt.Printf("worker id:%v\n", workerId)
	for idx, worker := range m.Workers {
		if worker.Id == workerId {
			worker.State = state
			m.Workers[idx] = worker
			break
		}
	}
	// TODO what if this worker does no exist
}

//
// submit a task result by worker
//
func (m *Master) SubmitTask(args *SubmitTaskResultArgs, reply *SubmitTaskResultReply) error {
	m.mLock.Lock()
	defer m.mLock.Unlock()
	tasks := []Task{}
	if m.State == InMapping {
		tasks = m.MapTasks
	} else {
		tasks = m.ReduceTasks
	}
	for idx, task := range tasks {
		if task.Id == args.TaskId {
			task.State = TaskDone
			if m.State == InMapping {
				m.MapTasks[idx] = task
			} else {
				m.ReduceTasks[idx] = task
			}
			break
		}
	}
	for idx, worker := range m.Workers {
		if worker.Id == args.WorkerId {
			worker.State = WorkerIdle
			m.Workers[idx] = worker
			break
		}
	}
	reply.Ack = true

	// check if all map tasks are be done
	m.ScanTasks()

	return nil
}

//
// find a todo task
//
func (m *Master) FetchTodoTask() {
}

//
// check workers every 2 seconds
// mark the worker as lost if this worker didn't
// response in the last 10 seconds
//
func (m *Master) ScanWorkers() {
	m.mLock.Lock()
	defer m.mLock.Unlock()
	for idx, worker := range m.Workers {
		timeGap := time.Now().Unix() - worker.LastHeartbeat
		fmt.Printf("TimeGap: %v from worker %v\n", timeGap, worker.Id)
		// worker lost
		if timeGap > 10 {
			// if this worker is dealing with some task
			// assign this task to an idle worker
			// by marking the task to be todo
			if worker.State == WorkerBusy {
				taskId := worker.TaskId
				tasks := []Task{}
				if m.State == InMapping {
					tasks = m.MapTasks
				} else {
					tasks = m.ReduceTasks
				}
				for taskIdx, task := range tasks {
					if task.Id == taskId {
						task.State = TaskTodo
						if m.State == InMapping {
							m.MapTasks[taskIdx] = task
						} else {
							m.ReduceTasks[taskIdx] = task
						}
						break
					}
				}
			}

			// mark the worker as lost
			worker.State = WorkerLost
			m.Workers[idx] = worker
		}
	}
	m.printWorkers()
}

//
// scan each worker every 2 (for example) seconds
//
func (m *Master) ScanWorkersEvery(seconds time.Duration) {
	for {
		m.ScanWorkers()
		time.Sleep(seconds * time.Second)
	}
}

func (m *Master) printWorkers() {
	fmt.Printf("Master state: %v \n", m.State)
	remainTask := 0
	if m.State == InMapping {
		for _, task := range m.MapTasks {
			if task.State != TaskDone {
				remainTask++
			}
		}
	} else {
		for _, task := range m.ReduceTasks {
			if task.State != TaskDone {
				remainTask++
			}
		}
	}
	fmt.Printf("remainint tasks: %v\n", remainTask)
	for _, worker := range m.Workers {
		fmt.Printf("Worker Id: %v    ", worker.Id)
		state := "Idle"
		switch worker.State {
		case WorkerIdle:
			state = "Idle"
			break
		case WorkerBusy:
			state = "Busy"
			break
		case WorkerLost:
			state = "Lost"
			break
		default:
			break
		}
		fmt.Printf("State: %v\n", state)
		if state == "Busy" {
			fmt.Printf("Busy on : %v\n", worker.TaskId)
		}
	}
	fmt.Println()
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
	go http.Serve(l, nil)
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
	if m.State == MasterDone {
		ret = true
	}

	return ret
}

//
// create a Master.
// files: os.Args[1:] pg-*.txt
// nReduce: 10 for example
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{
		Workers:     []WorkerRec{},
		MapTasks:    []Task{},
		ReduceTasks: []Task{},
		NReduce:     nReduce,
	}

	// Your code here.

	// init map tasks
	m.RegisterMapTasks(files)
	m.RegisterReduceTasks(nReduce)

	// rpc
	// listen to workers
	m.server()
	go m.ScanWorkersEvery(2)

	return &m
}
