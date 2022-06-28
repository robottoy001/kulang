package main

import (
	"fmt"
	"os"
	"time"
)

type FileInfo struct {
	MTime time.Time
	Exist ExistenceStatus
}

type FileSystem interface {
	Stat(string) (*FileInfo, error)
	CreateFile(string, string)
}

type RealFileSystem struct {
}

func (self RealFileSystem) Stat(path string) (*FileInfo, error) {
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

func (self RealFileSystem) CreateFile(p string, c string) {

}

type File struct {
	Content string
	Info    FileInfo
}

type VirtualFileSystem struct {
	Files map[string]*File
}

func (self VirtualFileSystem) CreateFile(path string, content string) {
	file := &File{
		Content: content,
		Info: FileInfo{
			Exist: ExistenceStatusExist,
			MTime: time.Now(),
		},
	}

	self.Files[path] = file
}

func (self VirtualFileSystem) Stat(path string) (*FileInfo, error) {
	if v, ok := self.Files[path]; ok {
		return &v.Info, nil
	}

	info := FileInfo{
		MTime: time.Now(),
		Exist: ExistenceStatusMissing,
	}
	return &info, nil
}
