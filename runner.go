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

// Runner exectuate commands
type Runner struct {
	Run      chan *Edge
	Status   map[*Edge]uint8
	RunQueue []*Edge
	runEdges int
	execCmd  int
}

const (
	// StatusInit initialize edge status
	StatusInit uint8 = iota
	// StatusRunning status of edge which is drity & ready to run
	StatusRunning
	// StatusFinished status of edge which is scheduled
	StatusFinished
)

// NewRunner create new Runner instance
func NewRunner() *Runner {
	return &Runner{
		Run:      make(chan *Edge),
		Status:   map[*Edge]uint8{},
		RunQueue: []*Edge{},
		runEdges: 0,
		execCmd:  0,
	}
}

func (r *Runner) execCommand(command string) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\x1B[31mError\x1B[0m: %s\n", command)
		log.Fatal(err)
	}
}

func (r *Runner) workProcess(edge *Edge, done chan *Edge) {

	if !edge.IsPhony() {

		for _, o := range edge.Outs {
			os.MkdirAll(path.Dir(o.Path), os.ModePerm)
		}
		// create rspfile if needed
		rspfile := edge.QueryVar("rspfile")
		if rspfile != "" {
			rspContent := edge.Rule.QueryVar("rspfile_content")
			if rspContent != nil && !rspContent.Empty() {
				ioutil.WriteFile(rspfile, []byte(rspContent.Eval(edge.Scope)), fs.ModePerm)
			}
		}

		r.execCommand(edge.EvalCommand())

		// delete rspfile if exist
		if rspfile != "" {
			os.Remove(rspfile)
		}
	}

	done <- edge
}

func (r *Runner) finished(edge *Edge) {
	edge.OutPutReady = true
	// delete in status map
	delete(r.Status, edge)

	// find next
	for _, outNode := range edge.Outs {
		for _, outEdge := range outNode.OutEdges {
			if _, ok := r.Status[outEdge]; !ok {
				continue
			}
			if outEdge.AllInputReady() {
				if r.Status[outEdge] != StatusInit {
					r.scheduleEdge(outEdge)
				} else {
					r.finished(outEdge)
				}
			}
		}
	}

}

// Start start run edges comand
func (r *Runner) Start() {
	done := make(chan *Edge)

	parallel := runtime.NumCPU()
	running := 0
	if len(r.RunQueue) == 0 {
		fmt.Printf("No work to do\n")
		return
	}

	fmt.Printf("run total %d commands, queue: %d\n", r.runEdges, len(r.Status))

Loop:
	for {
		if len(r.RunQueue) > 0 {
			running++
			edge := r.RunQueue[0]
			r.RunQueue = r.RunQueue[1:]
			if !edge.IsPhony() {
				r.execCmd++
			}
			fmt.Printf("\r\x1B[K[%d/%d] %s", r.execCmd, r.runEdges, edge.QueryVar("description"))
			go r.workProcess(edge, done)
		}

		if running < parallel && len(r.RunQueue) > 0 {
			continue
		}

		select {
		case e := <-done:
			running--
			r.finished(e)
			if running == 0 && len(r.RunQueue) <= 0 {
				break Loop
			}
		}
	}

	fmt.Printf("\nDone. Executed:%d, total: %d, status: %d\n", r.execCmd, r.runEdges, len(r.Status))
}

func (r *Runner) scheduleEdge(edge *Edge) {
	if status, ok := r.Status[edge]; ok {
		if status == StatusFinished {
			return
		}
	}

	r.Status[edge] = StatusFinished
	r.RunQueue = append(r.RunQueue, edge)
}

// AddTarget collect dirty nodes & edges
func (r *Runner) AddTarget(node *Node, dep *Node, depth int) error {

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

	status, ok := r.Status[node.InEdge]
	if !ok {
		r.Status[node.InEdge] = StatusInit
		status = StatusInit
	}

	if node.Status.Dirty && status == StatusInit {
		r.Status[node.InEdge] = StatusRunning

		if !node.InEdge.IsPhony() {
			r.runEdges++
		}
		if node.InEdge.AllInputReady() {
			r.scheduleEdge(node.InEdge)
		}
	}

	// already exists && proceeded
	if ok {
		return nil
	}

	for _, in := range node.InEdge.Ins {
		err := r.AddTarget(in, node, depth+1)
		if err != nil {
			fmt.Printf("AddTarget failed, input are: %v\n", in)
			return err
		}
	}

	return nil
}
