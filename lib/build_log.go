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

import "time"

type BuildLog struct {
	loaded bool
}

type LogItem struct {
	MTime time.Time
	Hash  string
}

func NewBuildLog() *BuildLog {
	return &BuildLog{
		loaded: false,
	}
}

func (b *BuildLog) Load(path string) {

	b.loaded = true
}

func (b *BuildLog) IsLoaded() bool {
	return b.loaded
}

func (b *BuildLog) QueryOutput(path string) *LogItem {
	return nil
}

func (b *BuildLog) WriteItem() {

}
