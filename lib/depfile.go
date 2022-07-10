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
	"fmt"
	"os"

	"gitee.com/kulang/utils"
)

type DepsLoader struct {
	Fs utils.FileSystem
}

func NewDepsLoader() *DepsLoader {
	return &DepsLoader{
		Fs: utils.RealFileSystem{},
	}
}

func (d *DepsLoader) DepsLoad(e *Edge) error {
	depfile := e.QueryVar("depfile")
	if depfile != "" {
		return d.loadDepsFile(e, depfile)
	}
	return nil
}

func (d *DepsLoader) loadDepsFile(e *Edge, depfile string) error {
	fileInfo, err := d.Fs.Stat(depfile)
	if err != nil {
		switch fileInfo.Exist {
		case utils.ExistenceStatusMissing:
			fmt.Fprintf(os.Stderr, "depfile %s is missing\n", depfile)
			return err
		case utils.ExistenceStatusExist:
			break
		}
	}

	content, err := d.Fs.ReadFile(depfile)
	if err != nil {
		return err
	}

	if len(content) == 0 {
		return fmt.Errorf("depfile %s is empty", depfile)
	}

	// Not Implement: Load deps file
	// return nil by default

	return nil
}
