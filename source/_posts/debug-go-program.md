---
title: Debugging Go Program
tags:
  - coding
categories:
  - 软件开发
date: 2019-01-24 17:24:38
---


# 概述

定位**Go**程序的错误，通常有两种方式：

- 打印日志
- 调试

**Go**是编译型语言，且**IDE**对调试的支持不太好，绝大多数**Go**的初学者调试**Go**程序都是通过`log.Printf`等打印日志方式定位问题。通常过程如下：

1. 程序`panic`或者报错
2. 修改**Go**程序，添加打印调试日志代码
3. 编译**Go**程序
4. 重复错误出现时的操作，查看日志
    1. 如果定位问题原因，修复程序错误，删除打印`debug`日志代码，返回第3步的操作
    2. 如果未定位问题原因，返回到第2步的操作

如果程序比较复杂，需要反复增加日志输出才能找到问题原因。熟练的使用调试器能够提高我们面对这样问题的灵活性。此文就是介绍如何使用**delve**等调试器调试**Go**程序。本文重点是全面介绍调试相关知识，具体调试工具的操作网上相关资料已经很全面（见最后一章参考），不作为重点。

# 使用GDB调试**Go**程序

## 简介

**GDB**不能很好的理解**Go**程序。**Go**程序和**GDB**的**stack management**、**threading**、**runtime**模型差异很大，并可能导致调试器输出不正确的结果。因此，虽然**GDB**在某些场景下有用，比如调试**Cgo**代码、调试**runtime**，但是对于**Go**程序来说，尤其是高并发程序，**GDB**不是一个可靠的调试器。而且，对于**Go**语言项目本身来说，解决这些的问题很困难，也不是一个高优先级的事情。

当你在***Linux***、***macOS***、***FreeBSD***、***NetBSD***上使用**gc**工具链编译和连接**Go**程序的时候，产生的二进制包含***DWARFv4***调试信息，最近版本的**GDB**调试器可以利用这些信息观察一个运行的进程或者**core dump**。

可以通过`-w`标记告诉连接器去掉这些调试信息，比如：

```
go build -ldflags "-w" .
```

**gc**编译器生成的程序包含[函数内联](https://wiki2.org/en/Inline_function)和变量注册。这些优化可能会让**gdb**调试更加困难。如果你需要禁用这些优化，使用下面的参数构建程序：

```
go build -gcflags "all=-N -l" .
```

如果你想要使用**gdb**检查一个**core dump**，你可以在程序崩溃的时候触发一个dump。在支持dump的**OS**上，使用`GOTRACEBACK=crash`环境变量（参考[runtime package documentation](https://golang.org/pkg/runtime/#hdr-Environment_Variables)）。

**Go 1.11**版本中，由于编译器会产生更多更准确的调试信息，为了减少二进制的大小，**DWARF**调试信息编译时候会默认被压缩。这对于大多数**ELF**工具来说这是透明的，也得到**Delve**支持。但是macOS和Windows上一些工具不支持。如果要禁用**DWARF**压缩，可以在编译的时候传入参数`-ldflags "-compressdwarf=false"`。

**Go 1.11**添加了一个实验性的功能，允许在调试器中调用函数。目前这个特性仅得到**Delve**(version 1.1.0及以上)的支持。

## 常用命令和教程

可以参考下面几篇文章，这里不做赘述：

- [Debugging Go Code with GDB](https://golang.org/doc/gdb)
  - [Go 1.11 Release Notes | Debugging](https://golang.org/doc/go1.11#debugging)
  - [Using the gdb debugger with Go](https://blog.codeship.com/using-gdb-debugger-with-go/)
  - [GDB Tutorial | Debugging with GDB](https://sourceware.org/gdb/current/onlinedocs/gdb/)

# 使用LLDB调试Go程序

## 简介

Mac下如果你安装XCode，应该会自动安装了LLDB，LLDB是XCode的默认调试器。LLDB的安装方法可以参考[这里](https://github.com/vadimcn/vscode-lldb/wiki/Installing-LLDB)。

GDB的命令格式非常自由，和GDB的命令不同，LLDB命令格式非常结构化（“严格”的婉转说法）。LLDB的命令格式如下：

```
<command> [<subcommand> [<subcommand>...]] <action> [-options [option-value]] [argument [argument...]]
```

解释一下：

- `<command>`(命令)和`<subcommand>`(子命令)：LLDB调试命令的名称。命令和子命令按层级结构来排列：一个命令对象为跟随其的子命令对象创建一个上下文，子命令又为其子命令创建一个上下文，依此类推。
- `<action>`：执行命令的操作
- `<options>`：命令选项。需要注意的是，如果aguments的第一个字母是"-"，`<options>`和`<arguments>`中间必须以"--"分隔开。所以如果你想启动一个程序，并给这个程序传入`-program_arg value`参数，可以输入`(lldb) process launch --stop-at-entry -- -program_arg value`
- `<arguement>`：命令的参数
- `[]`：表示命令是可选的，可以有也可以没有

LLDB也减少了gdb中一些命令的特殊写法，让用户更加容易理解命令的意图。可以阅读[LLDB文档](http://lldb.llvm.org/tutorial.html)中下面一段文字了解细节：

> We also tried to reduce the number of special purpose argument parsers, which sometimes forces the user to be a little more explicit about stating their intentions.
>
> ......


LLDB的命令同样给很多命令提供了缩写形式，可以通过`(lldb) help`查看所有的缩写命令。

[gdb和LLDB的命令之间的差别](http://lldb.llvm.org/lldb-gdb.html)可以访问这里查看。

## 常用命令

使用LLDB需要熟悉的常用命令如下：

### 帮助

> (lldb) help help
> Show a list of all debugger commands, or give details about a specific command.
>
> Syntax: help [<cmd-name>]
>

### 使用LLDB加载一个程序

```
$lldb /binary-path
Current executable set to '/binary-path'(x86_64).

$lldb
(lldb) file /binary-path
Current executable set to '/binary-path'(x86_64).
```

### 设置断点（breakpoints）

常见的设置断点的命令如下：

```
(lldb) breakpoint set --file source-file.go --line 11
Breakpoint 1: where = sample1`github.com/ethancai/go-debug-practice/sample1/model.(*MyStruct).Print + 19 at my_struct.go:11, address = 0x00000000010b2713
```

`breakpoint`命令会创建一个**逻辑的**断点，一个逻辑的断点可以对应一个或者多个位置`location`。比如，通过`selector`设置的断点对应所有实现了`selector`的方法。

`breakpoint`命令：

```
(lldb) help breakpoint
Commands for operating on breakpoints (see 'help b' for shorthand.)

Syntax: breakpoint <subcommand> [<command-options>]

...
```



### 设置观察点（Watchpoints）

`watchpoint`命令：

```
(lldb) help watchpoint
  Commands for operating on watchpoints.

Syntax: watchpoint <subcommand> [<command-options>]

...
```



### 运行程序或者附着程序

`process`命令：

```
(lldb) help process
  Commands for interacting with processes on the current platform.

Syntax: process <subcommand> [<subcommand-options>]

...
```



### 控制程序执行或者检查Thread状态

`thread`命令

```
(lldb) help thread
  Commands for operating on one or more threads in the current process.

Syntax: thread <subcommand> [<subcommand-options>]

...
```



### 检查堆栈结构（Stack Frame）状态

`frame`命令

```
(lldb) help frame
  Commands for selecting and examing the current thread's stack frames.

Syntax: frame <subcommand> [<subcommand-options>]

...
```

`expression`命令

```
(lldb) help expression
  Evaluate an expression on the current thread.  Displays any returned value with LLDB's
  default formatting.  Expects 'raw' input (see 'help raw-input'.)

Syntax: expression <cmd-options> -- <expr>

...
```



## 操作教程

可以参考下面几篇文章：

- Debugging Go Code with LLDB](http://ribrdb.github.io/lldb/): ([中文翻译](https://colobu.com/2018/03/12/Debugging-Go-Code-with-LLDB/))
- [熟练使用 LLDB，让你调试事半功倍](http://ios.jobbole.com/83393/)
- [LLDB Tutorial](http://lldb.llvm.org/tutorial.html)

# 使用Delve调试**Go**程序

可以参考下面几篇文章：

- [Debugging Go programs with Delve](https://blog.gopheracademy.com/advent-2015/debugging-with-delve/)
  - [Golang调试工具Delve](https://juejin.im/entry/5aa1f98d6fb9a028c522c84b)

# 不要使用调试器

对于调试器，一众计算机大牛都给出了明确而且强烈的建议：不要使用调试器。

- [Linus Torvalds](https://en.wikipedia.org/wiki/Linus_Torvalds), the creator of Linux, [does not use a debugger](https://lwn.net/2000/0914/a/lt-debugger.php3).
- [Robert C. Martin](https://en.wikipedia.org/wiki/Robert_Cecil_Martin), one of the inventors of agile programming, thinks that debuggers are [a wasteful timesink](http://www.artima.com/weblogs/viewpost.jsp?thread=23476).
- [John Graham-Cumming](https://en.wikipedia.org/wiki/John_Graham-Cumming) [hates debuggers](http://blog.jgc.org/2007/01/tao-of-debugging.html).
- [Brian W. Kernighan](https://en.wikipedia.org/wiki/Brian_Kernighan) and [Rob Pike](https://en.wikipedia.org/wiki/Rob_Pike) wrote that *stepping through a program less productive than thinking harder and adding output statements and self-checking code at critical places*. Kernighan once wrote that *the most effective debugging tool is still careful thought, coupled with judiciously placed print statements*.
- The author of Python, [Guido van Rossum](https://en.wikipedia.org/wiki/Guido_van_Rossum) has been quoted as saying that uses print statements for 90% of his debugging.

调试技术是一众纯手工的技术，诞生于计算机程序的规模还不是很大的时期。在当今软件规模不断扩展的情况下，调试无法解决软件质量问题。深入的思考、合理的架构、优美的代码、充分的单元测试才是提高软件质量的正确方向。调试应该仅作为调查问题最后一种办法。

# 参考

- Debugging  Go Program
  - [Debugging Go Code with GDB](https://golang.org/doc/gdb)
    - [Go 1.11 Release Notes | Debugging](https://golang.org/doc/go1.11#debugging)
    - [Using the gdb debugger with Go](https://blog.codeship.com/using-gdb-debugger-with-go/)
    - [GDB Tutorial | Debugging with GDB](https://sourceware.org/gdb/current/onlinedocs/gdb/)
  - [Debugging Go Code with LLDB](http://ribrdb.github.io/lldb/): ([中文翻译](https://colobu.com/2018/03/12/Debugging-Go-Code-with-LLDB/))
    - [熟练使用 LLDB，让你调试事半功倍](http://ios.jobbole.com/83393/)
    - [LLDB Tutorial](http://lldb.llvm.org/tutorial.html)
  - [Debugging Go programs with Delve](https://blog.gopheracademy.com/advent-2015/debugging-with-delve/)
    - [Golang调试工具Delve](https://juejin.im/entry/5aa1f98d6fb9a028c522c84b)
  -  Post-mortem debugging
    - [Go Post-mortem](https://fntlnz.wtf/post/gopostmortem/)
    - [Debugging Go core dumps](https://rakyll.org/coredumps/)
  - Debugging Concurrent Programs
    - [Debugging Concurrent Programs](https://users.soe.ucsc.edu/~dph/mypubs/debugConcProg89.pdf)
    - [Data Race Detector](https://golang.org/doc/articles/race_detector.html)
- Other
  - [Diagnostics](https://golang.org/doc/diagnostics.html): Profiling, Tracing, Debugging, Rutime statistics and events
  - Build Go Program
    - [Compile packages and dependencies](https://golang.org/cmd/go/#hdr-Compile_packages_and_dependencies)
  - ELF format and Tools
    - [The 101 of ELF files on Linux: Understanding and Analysis](https://linux-audit.com/elf-binaries-on-linux-understanding-and-analysis/)
    - [Executable and Linkable Format](https://wiki2.org/en/Executable_and_Linkable_Format)
  - Methodology
    - [I don't use a debugger](https://lemire.me/blog/2016/06/21/i-do-not-use-a-debugger/)
    - [Debugging golang programs](https://ttboj.wordpress.com/2016/02/15/debugging-golang-programs/)
  - Debugging in IDE
    - VSCode
      - [Debugging Go code using VS Code](https://github.com/Microsoft/vscode-go/wiki/Debugging-Go-code-using-VS-Code)
      - [DEBUGGING GO WITH VS CODE AND DELVE](https://flaviocopes.com/go-debugging-vscode-delve/)
    - Goland
      - [Goland Help | Debugging code](https://www.jetbrains.com/help/go/debugging-code.html)
- Tools
  - [GDB](https://www.gnu.org/software/gdb/)
  - [LLDB](https://lldb.llvm.org/)
  - [Delve](https://github.com/go-delve/delve): Delve is a debugger for the Go programming language
  - [Spew](https://github.com/davecgh/go-spew): Implements a deep pretty printer for Go data structures to aid in debugging
  - [panicparse](https://github.com/maruel/panicparse): Crash your app in style (Golang)
