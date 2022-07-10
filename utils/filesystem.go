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
package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type ExistenceStatus uint8

const (
	ExistenceStatusUnknown ExistenceStatus = iota
	ExistenceStatusMissing
	ExistenceStatusExist
)

type FileInfo struct {
	MTime time.Time
	Exist ExistenceStatus
}

type FileSystem interface {
	Stat(string) (*FileInfo, error)
	CreateFile(string, string)
	ReadFile(string) ([]byte, error)
}

type RealFileSystem struct {
}

func (rf RealFileSystem) Stat(path string) (*FileInfo, error) {
	info := &FileInfo{}
	finfo, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			info.Exist = ExistenceStatusExist
		} else if os.IsNotExist(err) {
			info.Exist = ExistenceStatusMissing
		} else {
			fmt.Printf("%s: %v", path, err)
			return info, err
		}
		return info, nil
	}

	info.MTime = finfo.ModTime()
	info.Exist = ExistenceStatusExist

	return info, nil
}

func (rf RealFileSystem) CreateFile(p string, c string) {

}

func (rf RealFileSystem) ReadFile(f string) ([]byte, error) {
	c, err := ioutil.ReadFile(f)
	if err != nil {
		return []byte{}, err
	}
	return c, nil
}

type File struct {
	Content string
	Info    FileInfo
}

type VirtualFileSystem struct {
	Files map[string]*File
}

func (vf VirtualFileSystem) CreateFile(path string, content string) {
	file := &File{
		Content: content,
		Info: FileInfo{
			Exist: ExistenceStatusExist,
			MTime: time.Now(),
		},
	}

	vf.Files[path] = file
}

func (vf VirtualFileSystem) Stat(path string) (*FileInfo, error) {
	if v, ok := vf.Files[path]; ok {
		return &v.Info, nil
	}

	info := FileInfo{
		MTime: time.Now(),
		Exist: ExistenceStatusMissing,
	}
	return &info, nil
}
func (rf VirtualFileSystem) ReadFile(f string) ([]byte, error) {
	return []byte{}, nil
}
