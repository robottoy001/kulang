package main

import (
	"fmt"
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
}

const (
	RUNNING uint8 = iota
	READY_TO_RUN
	STOP
)

func NewRunner() *Runner {
	return &Runner{
		Run:      make(chan *Edge),
		Status:   map[*Edge]uint8{},
		RunQueue: []*Edge{},
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

	self.execCommand(edge.EvalCommand())
	fmt.Printf("%s\n", edge.EvalCommand())

	done <- edge
}

func (self *Runner) finished(edge *Edge) {
	edge.OutPutReady = true
	// delete in status map
	delete(self.Status, edge)

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
Loop:
	for {
		if len(self.RunQueue) > 0 {
			running += 1
			edge := self.RunQueue[0]
			self.RunQueue = self.RunQueue[1:]

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
				fmt.Printf("no edge should runing\n")
				break Loop
			}
		}
	}

	fmt.Printf("DONE : %d\n", running)
}

func (self *Runner) scheduleEdge(edge *Edge) {
	//fmt.Printf("scheduleEdge: %v\n", edge.Outs[0].Path)
	self.RunQueue = append(self.RunQueue, edge)
}

func (self *Runner) AddTarget(node *Node, dep *Node) error {
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
		fmt.Printf("AddTarget, inEdge is nil: %v\n", err)
		return err
	}

	if node.InEdge.OutPutReady {
		return nil
	}

	status, ok := self.Status[node.InEdge]
	// not exists
	if !ok {
		self.Status[node.InEdge] = READY_TO_RUN
		status = READY_TO_RUN
		fmt.Printf("AddTarget, %s,output: %v\n", node.Path, node.InEdge.OutPutReady)
	}

	if node.Status.Dirty && status == READY_TO_RUN {
		self.Status[node.InEdge] = RUNNING
		//fmt.Printf("AddTarget, check if we can schedule: %s\n", node.Path)
		if node.InEdge.AllInputReady() {
			fmt.Printf("Node Input ready: %s\n", node.Path)
			self.scheduleEdge(node.InEdge)
		}
	}

	// already exists && proceeded
	if ok {
		fmt.Printf("AddTarget, already proceeded %s\n", node.Path)
		return nil
	}

	for _, in := range node.InEdge.Ins {
		err := self.AddTarget(in, node)
		if err != nil {
			fmt.Printf("AddTarget, err: input are: %v\n", in)
			return err
		}
	}

	return nil
}
