# kulang

#### 介绍
kulang是一款小型构建系统,兼容ninja文件语法和大部分ninja特性。

#### 软件架构



#### 安装教程

1.  依赖[go 1.7](https://golang.google.cn/dl/)及以上版本
2.  下载 `git clone https://gitee.com/robottoy/kulang.git`
3.  `cd kulang`
4.  `go build -o kulang`
5.  `kulang help`





#### 使用说明

```
kulang is yet another build system
usage:
  kulang [option] <command> [args...]
option:
  -C string
    	directory which include .ninja file (default ".")
  -f string
    	specified .ninja file (default "build.ninja")
  -perf string
    	enable cpu profile
  -v	show commandline when building
commands
  build           Build targets which specified
  clean           Clean built files
  commands        Show commands required to build the target
  help            Show help message
  targets         Show the targets need to build
  version         Show version of kulang
```

#### 参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request

#### kulang由来
[kulangsu](https://kulangsuisland.org)
