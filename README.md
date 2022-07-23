# kulang

#### 介绍
kulang是一款小型构建系统，兼容ninja文件语法和特性。

#### 软件架构

![](/docs/figures/kulang_arch.png)

- **CommandFunction**

    命令行接口模块，提供构建命令入口。

- **AppBuild**

    构建流程管理类，提供所有命令的实现，控制构建流程和基本数据结构建立。

- **Basic Structure**

    基础数据结构，提供词法(scanner)语法(parser)解析、图、命名空间、并法池、构建规则等基本数据结构。

- **CommandRunner**

    命令执行器，执行构建命令，控制并发数量和执行顺序，输出构建结果。

- **BuildLog**
    
    构建日志，记录构建目标使用的时间、 生成时间、目标名称、构建命令的Hash值。

#### 编译构建

1.  依赖[go 1.7](https://golang.google.cn/dl/)及以上版本
2.  下载 `git clone https://gitee.com/robottoy/kulang.git`
3.  `cd kulang`
4.  `go build -o kulang`

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
