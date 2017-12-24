---
title: Go语言中的错误处理（Error Handling in Go）
tags:
  - coding
  - go
categories:
  - 软件开发
---

# 概述

在实际工程项目中，我们希望通过程序的错误信息快速定位问题，但是又不喜欢错误处理代码写的冗余而又啰嗦。

`Go`语言没有提供像`Java`、`C#`语言中的`try...catch`异常处理方式，而是通过函数返回值逐层往上抛。这种设计，鼓励工程师在代码中显式的检查错误，而非忽略错误，好处就是避免漏掉本应处理的错误。但是带来一个弊端，让代码啰嗦。


# Go标准包提供的错误处理功能

`error`是个`interface`:

```go
type error interface {
    Error() string
}
```

如何创建`error`:

```go
// example 1
func Sqrt(f float64) (float64, error) {
    if f < 0 {
        return 0, errors.New("math: square root of negative number")
    }
    // implementation
}

// example 2
if f < 0 {
    return 0, fmt.Errorf("math: square root of negative number %g", f)
}
```

如何自定义`error`:

```go
// errorString is a trivial implementation of error.
type errorString struct {
    s string
}

func (e *errorString) Error() string {
    return e.s
}
```

自定义`error`类型可以拥有一些附加方法。比如`net.Error`定义如下：

```go
package net

type Error interface {
    error
    Timeout() bool   // Is the error a timeout?
    Temporary() bool // Is the error temporary?
}
```

网络客户端程序代码可以使用类型断言判断网络错误是瞬时错误还是永久错误。比如，一个网络爬虫可以在碰到瞬时错误的时候，等待一段时间然后重试。

```go
if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
    time.Sleep(1e9)
    continue
}
if err != nil {
    log.Fatal(err)
}
```

# 如何处理错误

go标准包提供的错误处理方式虽然简单，但是在实际项目开发、运维过程中，会碰到一些问题：

- 函数该如何返回错误，是用值，还是用特殊的错误类型
- 如何检查被调用函数返回的错误，是判断错误值，还是用类型断言
- 程序中每层代码在碰到错误的时候，是每层都处理，还是只用在最上层处理，如何做到优雅
- 日志中的异常信息不够完整、缺少stack strace，不方便定位错误原因

我一直思考go语言中该如何处理错误，直到看到了 [Dave Cheney 写的一个演讲文档](https://dave.cheney.net/paste/gocon-spring-2016.pdf)。本章节以下部分内容，主要来自于这篇文档。


## Go语言中三种错误处理策略

go语言中一般有三种错误处理策略：

- 返回和检查错误值(sentinel errors)：通过特定值表示成功和不同的错误，上层代码检查错误的值，来判断被调用`func`的执行状态
- 自定义错误类型(custom error types)：通过自定义的错误类型来表示特定的错误，上层代码通过类型断言判断错误的类型
- 不透明的错误处理(opaque error handling)：假设上层代码不知道被调用函数返回的错误任何细节，直接再向上返回错误


### 返回和检查错误值(Sentinel Errors)

这种方式在其它语言中，也很常见。比如，[C Error Codes in Linux](http://www.virtsync.com/c-error-codes-include-errno)。

go标准库中提供一些例子：

- `io.EOF`: 参考[这里](https://github.com/golang/go/blob/master/src/io/io.go#L38)
- `syscall.ENOENT`: 参考[这里](https://github.com/golang/go/blob/master/src/syscall/zerrors_linux_amd64.go#L1280)
- `go/build.NoGoError`: 参考[这里](https://github.com/golang/go/blob/master/src/go/build/build.go#L446)
- `path/filepath.SkipDir`: 参考[这里](https://github.com/golang/go/blob/master/src/path/filepath/path.go#L331)

这种策略最不灵活的错误处理策略，上层代码会比较返回错误结果和特定值。如果想修改返回的错误值，则会破坏上层调用代码的逻辑。

```go
buf := make([]byte, 100)
n, err := r.Read(buf)   // 如果修改 r.Read，在读到文件结尾时，返回另外一个 error，比如 io.END，而不是 io.EOF，则所有调用 r.Read 的代码都必须修改
buf = buf[:n]
if err == io.EOF {
    log.Fatal("read failed:", err)
}
```

另外一种场景也属于这类情况，上层代码通过检查错误的`Error`方法的返回值是否包含特定字符串，来判定如何进行错误处理。

```go
func readfile(path string) error {
    err := openfile(path)
    if err != nil {
        return fmt.Errorf("cannot open file: %v", err)
    }
    //...
}

func main() {
    err := readfile(".bashrc")
    if strings.Contains(error.Error(), "not found") {
        // handle error
    }
}
```

> **`error` interface 的 `Error` 方法的输出，是给人看的，不是给机器看的。我们通常会把`Error`方法返回的字符串打印到日志中，或者显示在控制台上。永远不要通过判断`Error`方法返回的字符串是否包含特定字符串，来决定错误处理的方式。**

如果你是开发一个公共库，库的API返回了特定值的错误值。那么必须把这个特定值的错误定义为`public`，写在文档中。“高内聚、低耦合”是衡量公共库质量的一个重要方面，而返回特定错误值的方式，增加了公共库和调用代码的耦合性。让之间产生了依赖。


### 自定义错误类型

todo

### 屏蔽细节的错误

todo

# 参考

- The Go Blog
    - [Error handling in Go](https://blog.golang.org/error-handling-and-go)
    - [Defer, Panic, and Recover](https://blog.golang.org/defer-panic-and-recover)
- Dave Cheney
    - [presentation on my philosophy for error handling](https://dave.cheney.net/paste/gocon-spring-2016.pdf)
    - [Don’t just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)
    - [Stack traces and the errors package](https://dave.cheney.net/2016/06/12/stack-traces-and-the-errors-package)
- Go Packages
    - [errors](https://golang.org/pkg/errors/)
    - [runtime](https://golang.org/pkg/runtime/)
- pkg: Artisanal, hand crafted, barrel aged, Go packages
    - [github.com/pkg/errors](https://godoc.org/github.com/pkg/errors)
