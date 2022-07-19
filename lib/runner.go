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
package lib

import (
	"bufio"
	"bytes"
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
	Run         chan *Edge
	Status      map[*Edge]uint8
	RunQueue    []*Edge
	buildOption *BuildOption
	runEdges    int
	execCmd     int
	done        chan *cmdResult
	failure     bool
}

type cmdError struct {
	command string
	err     error
}

type cmdResult struct {
	E *Edge
	C *cmdError
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
func NewRunner(buildOption *BuildOption) *Runner {
	return &Runner{
		Run:         make(chan *Edge),
		Status:      map[*Edge]uint8{},
		RunQueue:    []*Edge{},
		buildOption: buildOption,
		runEdges:    0,
		execCmd:     0,
		done:        make(chan *cmdResult),
		failure:     false,
	}
}

func (ce *cmdError) Error() string {
	return fmt.Sprintf("\x1B[31mError\x1B[0m:\n%s\n\x1b[31m%s\x1B[0m\n", ce.command, ce.err.Error())
}

func (r *Runner) execCommand(command string) *cmdError {

	bytesBuf := bytes.NewBuffer([]byte{})
	outWriter := bufio.NewWriter(bytesBuf)

	cmd := exec.Command("sh", "-c", command)
	cmd.Stderr = outWriter
	cmd.Stdout = outWriter
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	outWriter.Flush()

	if err != nil {
		fmt.Fprintf(os.Stderr, "\x1B[31mError\x1B[0m: %s\n\x1b[31m%s\x1B[0m\n%s\n", command, err.Error(), bytesBuf.String())
		return &cmdError{command: command, err: err}
	}
	fmt.Fprintf(os.Stdout, "%s", bytesBuf.String())
	return &cmdError{}
}

func (r *Runner) workProcess(edge *Edge) {

	if edge.IsPhony() {
		r.done <- &cmdResult{E: edge, C: &cmdError{}}
	}

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

	err := r.execCommand(edge.EvalCommand())

	// delete rspfile if exist
	if rspfile != "" {
		os.Remove(rspfile)
	}

	r.done <- &cmdResult{E: edge, C: err}
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

	parallel := runtime.NumCPU()
	running := 0
	if len(r.RunQueue) == 0 {
		fmt.Printf("No work to do\n")
		return
	}

	fmt.Printf("run total %d commands.\n", r.runEdges)

Loop:
	for {
		if len(r.RunQueue) > 0 && !r.failure {
			running++
			edge := r.RunQueue[0]
			r.RunQueue = r.RunQueue[1:]
			if !edge.IsPhony() {
				r.execCmd++
			}

			if r.buildOption.Verbose {
				fmt.Printf("[%d/%d] %s\n", r.execCmd, r.runEdges, edge.EvalCommand())
			} else {
				fmt.Printf("\r\x1B[K[%d/%d] %s", r.execCmd, r.runEdges, edge.QueryVar("description"))
			}
			go r.workProcess(edge)
		}

		if running < parallel && len(r.RunQueue) > 0 && !r.failure {
			continue
		}

		select {
		case ce := <-r.done:
			running--
			if ce.C.err != nil {
				r.failure = true
			}

			if !r.failure {
				r.finished(ce.E)
			}

			// if one edge fail, wait all running edge finishing.
			if (running == 0 && len(r.RunQueue) <= 0) || (running == 0 && r.failure) {
				break Loop
			}
		}
	}

	if r.failure {
		fmt.Fprintf(os.Stderr, "\x1B[31mFailed\x1B[0m. Executed:%d, total: %d\n", r.execCmd, r.runEdges)
		return
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
