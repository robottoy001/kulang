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
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"gitee.com/kulang/utils"
)

type BuildLog struct {
	Items          map[string]*LogItem
	startBuildTime int64
	logFile        *os.File
	logFileName    string
}

type LogItem struct {
	Path             string
	CommandEndTime   int64
	CommandStartTime int64
	MTime            int64
	Hash             uint64
}

func NewBuildLog() *BuildLog {
	return &BuildLog{
		Items:          map[string]*LogItem{},
		startBuildTime: time.Now().UnixMilli(),
		logFile:        nil,
		logFileName:    "",
	}
}

func (b *BuildLog) Load(path string) {
	var logName string = ".ninja_log"
	if _, err := os.Stat(path + "/.ninja_log"); err != nil {
		logName = "/.kulang"
	}
	b.logFileName = path + logName

	logFile, err := os.OpenFile(path+logName, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return
	}
	defer logFile.Close()

	bufLogFile := bufio.NewReader(b.logFile)
	// ingnore ninja version
	bufLogFile.ReadLine()

	for {
		line, err := bufLogFile.ReadString('\n')
		if line == "" || (err != nil && err != io.EOF) {
			break
		}
		var item LogItem = LogItem{}
		fmt.Sscanf(line, "%d%d%d%s%x", &item.CommandStartTime, &item.CommandEndTime, &item.MTime, &item.Path, &item.Hash)
		b.Items[item.Path] = &item
	}
}

func (b *BuildLog) QueryOutput(path string) *LogItem {
	if v, exist := b.Items[path]; exist {
		return v
	}
	return nil
}

func (b *BuildLog) WriteEdge(e *Edge, startTime int64, endTime int64, logTime int64) {
	command := e.EvalCommand()
	hash := utils.Hash([]byte(command))

	var itemForWrite *LogItem
	for _, outNode := range e.Outs {
		if item, exists := b.Items[outNode.Path]; exists {
			itemForWrite = item
		} else {
			itemForWrite = new(LogItem)
			b.Items[outNode.Path] = itemForWrite
		}

		itemForWrite.Path = outNode.Path
		itemForWrite.Hash = hash
		itemForWrite.MTime = logTime
		itemForWrite.CommandStartTime = startTime
		itemForWrite.CommandEndTime = endTime

		if b.logFile == nil {
			file, err := os.OpenFile(b.logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
			if err != nil {
				log.Println(err)
				return
			}
			b.logFile = file
		}

		line := fmt.Sprintf("%d\t%d\t%d\t%s\t%x\n",
			itemForWrite.CommandStartTime, itemForWrite.CommandEndTime, itemForWrite.MTime, itemForWrite.Path, itemForWrite.Hash)
		b.logFile.WriteString(line)
	}
}

func (b *BuildLog) Close() {
	if b.logFile != nil {
		b.logFile.Close()
	}
	b.logFile = nil
}
