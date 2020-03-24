/**
 * The workers will talk to the master via RPC.
 * Each worker process will ask the master for a task,
 * read the task's input from one or more files,
 * execute the task, and write the task's output to one or more files.
 */

package mr

import (
	"encoding/json"
	"fmt"
	"time"
)
import "log"
import "net/rpc"
import "hash/fnv"
import "os"
import "io/ioutil"
import "sort"

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

//
// save intermediate data to
// mr-X-Y
// X is map task number
// Y is reduce task number
//
func mapWork(mapf func(string, string) []KeyValue, filename string, taskId int, workerId int, nReduce int) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open %v", filename)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", filename)
	}
	file.Close()
	intermediate := mapf(filename, string(content))

	// array of { word, "1" }
	// put the same word to same file
	sort.Sort(ByKey(intermediate))

	// init kvs to store nReduce KeyValues
	kvs := [][]KeyValue{}
	for i := 0; i < nReduce; i++ {
		subKvs := []KeyValue{}
		kvs = append(kvs, subKvs)
	}

	i := 0
	for i < len(intermediate) {

		// word is intermediate[i].Key
		// TODO change this to temp file
		word := intermediate[i].Key
		hash := ihash(word) % nReduce
		kvs[hash] = append(kvs[hash], intermediate[i])
		i++
	}

	writeIntermediateToFiles(kvs, nReduce, taskId)

	// submit map work result to master
	submitArgs := SubmitTaskResultArgs{
		WorkerId: workerId,
		TaskId:   taskId,
		Result:   "",
	}
	submitReply := SubmitTaskResultReply{Ack: false}
	call("Master.SubmitTask", &submitArgs, &submitReply)
	fmt.Printf("Submit the task %v\n", taskId)
	if !submitReply.Ack {
		// fmt.Println("Something goes wrong in master.")
	}
}

//
// write intermediate to files
//
func writeIntermediateToFiles(kvs [][]KeyValue, nReduce int, taskId int) {
	i := 0
	for i < nReduce {
		subKvs := kvs[i]
		oname := fmt.Sprintf("%s-%d-%d", "mr", taskId, i)
		ofile, _ := os.Create(oname)
		enc := json.NewEncoder(ofile)

		for _, kv := range subKvs {
			err := enc.Encode(kv)
			if err != nil {
				log.Fatal(err)
			}
		}
		ofile.Close()
		i++
	}
}

//
// reduce work
//
func reduceWork(reducef func(string, []string) string, nReduceId string, taskId int, workerId int, extra int) {
	kvs := []KeyValue{}
	// read mr-*-nReduceId from disk
	i := 1
	// fmt.Printf("extra:%v\n", extra)
	for i <= extra {
		oname := fmt.Sprintf("%s-%d-%v", "mr", i, nReduceId)
		ofile, err := os.Open(oname)
		if err != nil {
			// fmt.Printf("open err: %v\n", err)
		}
		dec := json.NewDecoder(ofile)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			kvs = append(kvs, kv)
		}
		ofile.Close()
		i++
	}

	sort.Sort(ByKey(kvs))

	outputOname := fmt.Sprintf("mr-out-%v", nReduceId)
	ofile, _ := os.Create(outputOname)

	i = 0
	for i < len(kvs) {
		j := i + 1
		for j < len(kvs) && kvs[j].Key == kvs[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, kvs[k].Value)
		}

		// word, array of { word, "1" }
		// each word will reduce once
		output := reducef(kvs[i].Key, values)

		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(ofile, "%v %v\n", kvs[i].Key, output)

		i = j
	}

	ofile.Close()

	// submit map work result to master
	submitArgs := SubmitTaskResultArgs{
		WorkerId: workerId,
		TaskId:   taskId,
		Result:   "",
	}
	submitReply := SubmitTaskResultReply{Ack: false}
	call("Master.SubmitTask", &submitArgs, &submitReply)
	if !submitReply.Ack {
		// fmt.Println("Something goes wrong in master.")
	}
}

func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	// register the worker

	registerWorkerReply := registerWorker()
	workerId := registerWorkerReply.Id
	fmt.Printf("here id:%v\n", workerId)
	nReduce := registerWorkerReply.NReduce

	go working(workerId, mapf, reducef, nReduce)
	// heartbeat every one second
	go heartbeat(workerId)

	time.Sleep(time.Hour)

}

//
// 1. fetch task
// 2. excute the task
// 3. return the task result to master
// 4. go back to 1
// handle waiting
//
func working(workerId int, mapf func(string, string) []KeyValue, reducef func(string, []string) string, nReduce int) {
	for {
		task := fetchTask(workerId)
		// fmt.Printf("Excuting task %v.\n", task.Id)
		if task.Type == "Map" {
			// map task
			filename := task.Detail
			mapWork(mapf, filename, task.Id, workerId, nReduce)
		} else {
			// reduce task
			// fmt.Printf("Reduce task here %v\n", task.Id)
			reduceWork(reducef, task.Detail, task.Id, workerId, task.Extra)
		}
	}
}

//
// fetch a task
//
func fetchTask(workerId int) Task {
	args := TaskDistributeArgs{
		WorkerId: workerId,
	}
	for {
		reply := TaskDistributeReply{}
		task := Task{}
		fmt.Printf("asking for a task\n")
		call("Master.AssignTask", &args, &reply)
		if reply.Message == "" {
			fmt.Printf("Fetch task %v from master\n", reply.Task.Id)
			task = reply.Task
			return task
		} else if reply.Message == "Wait" {
			fmt.Printf("Keep waiting until there are some idle tasks.\n")
			time.Sleep(time.Second)
		} else if reply.Message == "MasterDone" {
			fmt.Printf("Shutting down the worker because of the MapReduce job finishes.\n")
			os.Exit(0)
		} else if reply.Message == "Forbidden" {
			fmt.Println("Finish temp task before asking for another one.\n")
			time.Sleep(time.Second)
			// TODO handle something here
			// os.Exit(0)
		} else if reply.Message == "Lost" {
			fmt.Println("Marked Lost from master, exiting.\n")
			// os.Exit(0)
		}
	}
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
			// fmt.Println("Something goes wrong in master.")
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

//
// mock that maybe this worker would be broken
//
func maybeBreak() {

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

	fmt.Println(err)
	return false
}
