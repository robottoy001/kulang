/*
 * Copyright (c) 2022 Huawei Device Co., Ltd.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
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
		fmt.Fprintf(os.Stderr, "\x1B[31mError\x1B[0m:\n%s\n\x1b[31m%s\x1B[0m\n", command, err.Error())
		os.Exit(KulangError)
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

	fmt.Printf("run total %d commands.\n", r.runEdges)

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

	fmt.Printf("\n\x1B[1;32mSucceed\x1B[0m. Executed:%d, total: %d\n", r.execCmd, r.runEdges)
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
			return err
		}
	}

	return nil
}
