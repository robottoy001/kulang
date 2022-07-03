package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
)

type Runner struct {
	Run      chan *Edge
	Status   map[*Edge]uint8
	RunQueue []*Edge
	runEdges int
}

const (
	RUNNING uint8 = iota
	READY_TO_RUN
	FINISHED
	STOP
)

func NewRunner() *Runner {
	return &Runner{
		Run:      make(chan *Edge),
		Status:   map[*Edge]uint8{},
		RunQueue: []*Edge{},
		runEdges: 0,
	}
}

func (self *Runner) execCommand(command string) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func (self *Runner) workProcess(edge *Edge, done chan *Edge) {

	for _, o := range edge.Outs {
		os.MkdirAll(path.Dir(o.Path), os.ModePerm)
	}

	fmt.Printf("%s\n", edge.QueryVar("description"))
	self.execCommand(edge.EvalCommand())

	done <- edge
}

func (self *Runner) finished(edge *Edge) {
	edge.OutPutReady = true
	// delete in status map
	//delete(self.Status, edge)
	self.Status[edge] = FINISHED

	// delete rspfile if exist
	rspfile := edge.QueryVar("rspfile")
	if rspfile != "" {
		os.Remove(rspfile)
	}

	// find next
	for _, outNode := range edge.Outs {
		for _, outEdge := range outNode.OutEdges {
			if _, ok := self.Status[outEdge]; !ok {
				continue
			}
			if outEdge.AllInputReady() {
				self.scheduleEdge(outEdge)
			}
		}
	}

}

func (self *Runner) Start() {
	done := make(chan *Edge)

	parallel := runtime.NumCPU()
	running := 0
	if len(self.RunQueue) == 0 {
		fmt.Printf("No work to do\n")
		return
	}
	fmt.Printf("run %d commands\n", self.runEdges)

Loop:
	for {
		if len(self.RunQueue) > 0 {
			running += 1
			edge := self.RunQueue[0]
			self.RunQueue = self.RunQueue[1:]
			if edge.IsPhony() {
				continue
			}

			go self.workProcess(edge, done)
		}

		if running < parallel && len(self.RunQueue) > 0 {
			continue
		}

		select {
		case e := <-done:
			running -= 1
			self.finished(e)
			if running == 0 && len(self.RunQueue) <= 0 {
				break Loop
			}
		}
	}

	fmt.Printf("Done.\n")
}

func (self *Runner) scheduleEdge(edge *Edge) {
	if status, ok := self.Status[edge]; ok {
		if status == RUNNING || status == FINISHED || status == READY_TO_RUN {
			return
		}
	}

	if !edge.IsPhony() {
		// create rspfile if needed
		rspfile := edge.QueryVar("rspfile")
		if rspfile != "" {
			rsp_content := edge.Rule.QueryVar("rspfile_content")
			if rsp_content != nil && !rsp_content.Empty() {
				ioutil.WriteFile(rspfile, []byte(rsp_content.Eval(edge.Scope)), fs.ModePerm)
			}
		}
		self.RunQueue = append(self.RunQueue, edge)
	}
}

func (self *Runner) AddTarget(node *Node, dep *Node, depth int) error {
	//fmt.Printf("Addtarget, %s,output: %v\n", node.Path, node.InEdge.OutPutReady)

	if node.InEdge == nil {
		var err error
		if node.Status.Dirty {
			var depstr string
			if dep != nil {
				depstr = ", needed by " + dep.Path
			}
			err = fmt.Errorf("%s%s missing", node.Path, depstr)
		}
		return err
	}

	if node.InEdge.OutPutReady {
		return nil
	}

	status, ok := self.Status[node.InEdge]
	if !ok {
		self.Status[node.InEdge] = READY_TO_RUN
		status = READY_TO_RUN
	}

	if node.Status.Dirty && status == READY_TO_RUN {
		self.Status[node.InEdge] = RUNNING

		self.runEdges += 1
		if node.InEdge.AllInputReady() {
			self.scheduleEdge(node.InEdge)
		}
	}

	// already exists && proceeded
	if ok {
		return nil
	}

	for _, in := range node.InEdge.Ins {
		err := self.AddTarget(in, node, depth+1)
		if err != nil {
			fmt.Printf("AddTarget, input are: %v\n", in)
			return err
		}
	}

	return nil
}
