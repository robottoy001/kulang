package build

import (
	"fmt"
	"log"
	"path"

	"gitee.com/kulang/parser"
)

type BuildOption struct {
	ConfigFile string
	BuildDir   string
}

type AppBuild struct {
	Option *BuildOption
}

func NewAppBuild(option *BuildOption) *AppBuild {
	return &AppBuild{
		Option: option,
	}
}

func (self *AppBuild) RunBuild() error {
	fmt.Println("start building...")

	// parser load .ninja file
	// default, Ninja parser & scanner
	p := parser.NewParser()
	err := p.Load(path.Join(self.Option.BuildDir, self.Option.ConfigFile))
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = p.Parse()
	if err != nil {
		return err
	}

	err = self._RunBuild()
	if err != nil {
		return err
	}
	return nil
}

func (self *AppBuild) _RunBuild() error {
	return nil
}
