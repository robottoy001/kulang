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
	"os"
	"time"
)

type BuildLog struct {
	Items          map[string]*LogItem
	startBuildTime int64
	loaded         bool
	logFile        *os.File
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
		loaded:         false,
		logFile:        nil,
	}
}

func (b *BuildLog) Load(path string) {
	logFile, err := os.OpenFile(".ninja_log", os.O_RDONLY, os.ModePerm)
	if err != nil {
		return
	}
	defer logFile.Close()
	b.logFile = logFile

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

	b.loaded = true
}

func (b *BuildLog) IsLoaded() bool {
	return b.loaded
}

func (b *BuildLog) QueryOutput(path string) *LogItem {
	if v, exist := b.Items[path]; exist {
		return v
	}
	return nil
}

func (b *BuildLog) WriteItem() {

}

func (b *BuildLog) Close() {

}
