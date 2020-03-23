/**
 * The workers will talk to the master via RPC.
 * Each worker process will ask the master for a task,
 * read the task's input from one or more files,
 * execute the task, and write the task's output to one or more files.
 */

package mr

import (
	"fmt"
	"time"
)
import "log"
import "net/rpc"
import "hash/fnv"

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	// register the worker
	workerId := registerWorker().Id

	// heartbeat every one second
	go heartbeat(workerId)

	// 1. fetch task
	// 2. excute the task
	// 3. return the task result to master
	// 4. go back to 1
	// heartbeat at every second

	time.Sleep(30 * time.Minute)
}

//
// register worker to master
//
func registerWorker() RegisterWorkerReply {
	args := RegisterWorkerArgs{}

	reply := RegisterWorkerReply{}

	call("Master.RegisterWorker", &args, &reply)

	fmt.Printf("workerId: %v\n", reply.Id)

	return reply
}

//
// send heartbeat to master
// every one second
//
func heartbeat(workerId int) {
	for {
		args := WorkerHeartbeatArgs{Id: workerId}
		reply := WorkerHeartbeatReply{Ack: false}
		call("Master.ListenHeartbeat", &args, &reply)
		if !reply.Ack {
			fmt.Println("Something goes wrong in master.")
		}
		time.Sleep(time.Second)
	}
}

//
// fetch a task
//
func fetchTask() {

}

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	c, err := rpc.DialHTTP("unix", "mr-socket")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println("holy fuck here")
	fmt.Println(err)
	return false
}
