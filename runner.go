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
	execCmd  int
}

const (
	STATUS_INIT uint8 = iota
	STATUS_RUNNING
	STATUS_FINISHED
)

func NewRunner() *Runner {
	return &Runner{
		Run:      make(chan *Edge),
		Status:   map[*Edge]uint8{},
		RunQueue: []*Edge{},
		runEdges: 0,
		execCmd:  0,
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

	if !edge.IsPhony() {

		for _, o := range edge.Outs {
			os.MkdirAll(path.Dir(o.Path), os.ModePerm)
		}
		// create rspfile if needed
		rspfile := edge.QueryVar("rspfile")
		if rspfile != "" {
			rsp_content := edge.Rule.QueryVar("rspfile_content")
			if rsp_content != nil && !rsp_content.Empty() {
				ioutil.WriteFile(rspfile, []byte(rsp_content.Eval(edge.Scope)), fs.ModePerm)
			}
		}

		fmt.Printf("%s\n", edge.QueryVar("description"))
		self.execCommand(edge.EvalCommand())

		// delete rspfile if exist
		if rspfile != "" {
			os.Remove(rspfile)
		} else {
			fmt.Printf("PHONY: %s\n", edge.String())
		}
	}

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
				//fmt.Printf("schedule, %s, output: %v\n", outEdge.String(), outEdge.OutPutReady)
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

	fmt.Printf("run %d commands, status: %d\n", self.runEdges, len(self.Status))

Loop:
	for {
		if len(self.RunQueue) > 0 {
			running += 1
			edge := self.RunQueue[0]
			self.RunQueue = self.RunQueue[1:]

			go self.workProcess(edge, done)
			self.execCmd += 1
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

	fmt.Printf("Done. Executed commands:%d \n", self.execCmd)
}

func (self *Runner) scheduleEdge(edge *Edge) {
	if status, ok := self.Status[edge]; ok {
		if status == STATUS_FINISHED {
			fmt.Printf("schedule finished Edge : %s\n", edge.String())
			return
		}
	}

	//if !edge.IsPhony() {
	// create rspfile if needed
	rspfile := edge.QueryVar("rspfile")
	if rspfile != "" {
		rsp_content := edge.Rule.QueryVar("rspfile_content")
		if rsp_content != nil && !rsp_content.Empty() {
			ioutil.WriteFile(rspfile, []byte(rsp_content.Eval(edge.Scope)), fs.ModePerm)
		}
	}

	self.Status[edge] = STATUS_FINISHED
	self.RunQueue = append(self.RunQueue, edge)
	//}
}

func (self *Runner) AddTarget(node *Node, dep *Node, depth int) error {

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
		self.Status[node.InEdge] = STATUS_INIT
		status = STATUS_INIT
	}

	if node.Status.Dirty && status == STATUS_INIT {
		self.Status[node.InEdge] = STATUS_RUNNING

		self.runEdges += 1
		if node.InEdge.AllInputReady() {
			fmt.Printf("Addtarget, rule: %s, %s, output: %v\n", node.InEdge.Rule.Name, node.Path, node.InEdge.OutPutReady)
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
			fmt.Printf("AddTarget failed, input are: %v\n", in)
			return err
		}
	}

	return nil
}
