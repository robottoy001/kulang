package main

import "fmt"

type Runner struct {
	Run    chan *Edge
	Status map[*Edge]uint8
}

const (
	RUNNING uint8 = iota
	READY_TO_RUN
	STOP
)

func NewRunner() *Runner {
	return &Runner{
		Run:    make(chan *Edge),
		Status: map[*Edge]uint8{},
	}
}

func (self *Runner) AddTarget(node *Node, dep *Node) error {

	go func() {
		for {
			select {
			case edge := <-self.Run:
				fmt.Printf("RUN: %s <-- %s -- %s", edge.Outs[0].Path, edge.Rule.Name, edge.Ins[0].Path)
			}
		}
	}()

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
		return fmt.Errorf("all outputs are ready")
	}
	status, ok := self.Status[node.InEdge]
	// not exists
	if !ok {
		self.Status[node.InEdge] = READY_TO_RUN
		status = READY_TO_RUN
	}

	if node.Status.Dirty && status == READY_TO_RUN {
		self.Status[node.InEdge] = RUNNING
		if node.InEdge.AllInputReady() {
			self.Run <- node.InEdge
			fmt.Printf("Node Input ready: %s\n", node.Path)
		}
	}

	// already exists && proceeded
	if ok {
		return nil
	}

	for _, in := range node.InEdge.Ins {
		err := self.AddTarget(in, node)
		if err != nil {
			return err
		}
	}

	return nil
}
