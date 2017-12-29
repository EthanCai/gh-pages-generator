---
title: Go语言中的错误处理（Error Handling in Go）
tags:
  - coding
  - go
categories:
  - 软件开发
date: 2017-12-29 10:23:06
---



{% asset_img d2a2d12e1a00f819d7fd7af4b536efa2.png %}


# 概述

在实际工程项目中，我们希望通过程序的错误信息快速定位问题，但是又不喜欢错误处理代码写的冗余而又啰嗦。`Go`语言没有提供像`Java`、`C#`语言中的`try...catch`异常处理方式，而是通过函数返回值逐层往上抛。这种设计，鼓励工程师在代码中显式的检查错误，而非忽略错误，好处就是避免漏掉本应处理的错误。但是带来一个弊端，让代码啰嗦。


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

go标准包提供的错误处理方式虽然简单，但是在实际项目开发、运维过程中，会经常碰到如下问题：

- 函数该如何返回错误，是用值，还是用特殊的错误类型
- 如何检查被调用函数返回的错误，是判断错误值，还是用类型断言
- 程序中每层代码在碰到错误的时候，是每层都处理，还是只用在最上层处理，如何做到优雅
- 日志中的异常信息不够完整、缺少stack strace，不方便定位错误原因

我一直思考`Go`语言中该如何处理错误。下面的内容介绍了生产级`Go`语言代码中如何处理错误。

> 以下内容，主要来自于[Dave Cheney 写的一个演讲文档](https://dave.cheney.net/paste/gocon-spring-2016.pdf)。


## Go语言中三种错误处理策略

`go`语言中一般有三种错误处理策略：

- 返回和检查错误值：通过特定值表示成功和不同的错误，上层代码检查错误的值，来判断被调用`func`的执行状态
- 自定义错误类型：通过自定义的错误类型来表示特定的错误，上层代码通过类型断言判断错误的类型
- 隐藏内部细节的错误处理：假设上层代码不知道被调用函数返回的错误任何细节，直接再向上返回错误


### 返回和检查错误值

这种方式在其它语言中，也很常见。比如，[C Error Codes in Linux](http://www.virtsync.com/c-error-codes-include-errno)。

go标准库中提供一些例子：

- `io.EOF`: 参考[这里](https://github.com/golang/go/blob/master/src/io/io.go#L38)
- `syscall.ENOENT`: 参考[这里](https://github.com/golang/go/blob/master/src/syscall/zerrors_linux_amd64.go#L1280)
- `go/build.NoGoError`: 参考[这里](https://github.com/golang/go/blob/master/src/go/build/build.go#L446)
- `path/filepath.SkipDir`: 参考[这里](https://github.com/golang/go/blob/master/src/path/filepath/path.go#L331)

这种策略是最不灵活的错误处理策略，上层代码需要判断返回错误值是否等于特定值。如果想修改返回的错误值，则会破坏上层调用代码的逻辑。

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

如果你是开发一个公共库，库的API返回了特定值的错误值。那么必须把这个特定值的错误定义为`public`，写在文档中。

“高内聚、低耦合”是衡量公共库质量的一个重要方面，而返回特定错误值的方式，增加了公共库和调用代码的耦合性。让模块之间产生了依赖。


### 自定义错误类型

这种方式的典型用法如下：

```go
// 定义错误类型
type MyError struct {
    Msg string
    File string
    Line int
}

func (e *MyError) Error() string {
    return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Msg)
}


// 被调用函数中
func doSomething() {
    // do something
    return &MyError{"Something happened", "server.go", 42}
}

// 调用代码中
err := doSomething()
if err, ok := err.(SomeType); ok {    // 使用 类型断言 获得错误详细信息
    //...
}
```

这种方式相比于“返回和检查错误值”，很大一个优点在于可以将 底层错误 包起来一起返回给上层，这样可以提供更多的上下文信息。比如`os.PathError`：

```go
// PathError records an error and the operation
// and file path that caused it.
type PathError struct {
    Op string
    Path string
    Err error
}

func (e *PathError) Error() string
```

然而，这种方式依然会增加模块之间的依赖。


### 隐藏内部细节的错误处理

这种策略之所以叫“隐藏内部细节的错误处理”，是因为当上层代码碰到错误发生的时候，不知道错误的内部细节。

**作为上层代码，你需要知道的就是被调用函数是否正常工作。** 如果你接受这个原则，将极大降低模块之间的耦合性。

```go
import “github.com/quux/bar”

func fn() error {
    x, err := bar.Foo()
    if err != nil {
        return err
    }
    // use x
}
```

上面的例子中，`Foo`这个方法不承诺返回的错误的具体内容。这样，`Foo`函数的开发者可以不断调整返回错误的内容来提供更多的错误信息，而不会破坏`Foo`提供的协议。这就是“隐藏内部细节”的内涵。


## 最合适的错误处理策略

上面我们提到了三种错误处理策略，其中第三种策略耦合性最低。然而，第三种方式也存在一些问题：

- 如何获得更详细错误信息，比如`stack trace`，帮助定位错误原因
- 如何优雅的处理错误
    - 有些场景需要了解错误细节，比如网络调用，需要知道是否是瞬时的中断
    - 是否每层捕捉到错误的时候都需要处理


### 如何输出更详细的错误信息

```go
func AuthenticateRequest(r *Request) error {
    err := authenticate(r.User)
    if err != nil {
        return err  // No such file or directory
    }
    return nil
}
```

上面这段代码，在我看来，在顶层打印错误的时候，只看到一个类似于"No such file or directory"的文字，从这段文字中，无法了解到错误是哪行代码产生的，也无法知道当时出错的调用堆栈。

我们调整一下代码，如下：

```go
func AuthenticateRequest(r *Request) error {
    err := authenticate(r.User)
    if err != nil {
        return fmt.Errorf("authenticate failed: %v", err)    // authenticate failed: No such file or directory
    }
    return nil
}
```

通过`fmt.Errorf`创建一个新的错误，添加更多的上下文信息到新的错误中，但这样仍不能解决上面提出的问题。

这里，我们通过一个很小的包`github.com/pkg/errors`来试图解决上面的问题。这个包提供这样几个主要的API：

```go
// Wrap annotates cause with a message.
func Wrap(cause error, message string) error

// Cause unwraps an annotated error.
func Cause(err error) error
```

[goErrorHandlingSample](https://github.com/EthanCai/goErrorHandlingSample)这个repo中的例子演示了，不同错误处理方式，输出的错误信息的区别。

```sh
> go run sample1/main.go

open /Users/caichengyang/.settings.xml: no such file or directory
exit status 1
```

```sh
> go run sample2/main.go

could not read config: open failed: open /Users/caichengyang/.settings.xml: no such file or directory
exit status 1
```

```sh
> go run sample3/main.go

could not read config: open failed: open /Users/caichengyang/.settings.xml: no such file or directory
exit status 1
```

```sh
> go run sample4/main.go

open /Users/caichengyang/.settings.xml: no such file or directory
open failed
main.ReadFile
        /Users/caichengyang/go/src/github.com/ethancai/goErrorHandlingSample/sample4/main.go:15
main.ReadConfig
        /Users/caichengyang/go/src/github.com/ethancai/goErrorHandlingSample/sample4/main.go:27
main.main
        /Users/caichengyang/go/src/github.com/ethancai/goErrorHandlingSample/sample4/main.go:32
runtime.main
        /usr/local/Cellar/go/1.9.2/libexec/src/runtime/proc.go:195
runtime.goexit
        /usr/local/Cellar/go/1.9.2/libexec/src/runtime/asm_amd64.s:2337
could not read config
main.ReadConfig
        /Users/caichengyang/go/src/github.com/ethancai/goErrorHandlingSample/sample4/main.go:28
main.main
        /Users/caichengyang/go/src/github.com/ethancai/goErrorHandlingSample/sample4/main.go:32
runtime.main
        /usr/local/Cellar/go/1.9.2/libexec/src/runtime/proc.go:195
runtime.goexit
        /usr/local/Cellar/go/1.9.2/libexec/src/runtime/asm_amd64.s:2337
exit status 1

```

`sample4/main.go`中将出错的代码行数也打印了出来，这样的日志，可以帮助我们更方便的定位问题原因。


### 为了行为断言错误，而非为了类型

在有些场景下，仅仅知道是否出错是不够的。比如，和进程外其它服务通信，需要了解错误的属性，以决定是否需要重试操作。

这种情况下，不要判断错误值或者错误的类型，我们可以判断错误是否实现某个行为。

```go
type temporary interface {
    Temporary() bool    // IsTemporary returns true if err is temporary.
}

func IsTemporary(err error) bool {
    te, ok := err.(temporary)
    return ok && te.Temporary()
}
```

这种实现方式的好处在于，不需要知道具体的错误类型，也就不需要引用定义了错误类型的三方`package`。如果你是底层代码的开发者，哪天你想更换一个实现更好的`error`，也不用担心影响上层代码逻辑。如果你是上层代码的开发者，你只需要关注`error`是否实现了特定行为，不用担心引用的三方`package`升级后，程序逻辑失败。

习大大说：“听其言、观其行”。小时候父母总是教导：“不要以貌取人”。原来生活中的道理，在程序开发中也是适用的。


### 不要忽略错误，也不要重复处理错误

遇到错误，而不去处理，导致信息缺失，会增加后期的运维成本

```go
func Write(w io.Writer, buf []byte) {
    w.Write(buf)    // Write(p []byte) (n int, err error)，Write方法的定义见 https://golang.org/pkg/io/#Writer
}
```

重复处理，添加了不必要的处理逻辑，导致信息冗余，也会增加后期的运维成本

```go
func Write(w io.Writer, buf []byte) error {
    _, err := w.Write(buf)
    if err != nil {
        log.Println("unable to write:", err)    // 第1次错误处理

        return err
    }
    return nil
}

func main() {
    // create writer and read data into buf

    err := Write(w, buf)
    if err != nil {
        log.Println("Write error:", err)        // 第2次错误处理
        os.Exit(1)
    }

    os.Exit(0)
}
```

# 总结

- 使用“隐藏内部细节的错误处理”
    - 使用`errors.Wrap`封装原始`error`
    - 使用`errors.Cause`找出原始`error`
- 为了行为而断言，而不是类型
- 尽量减少错误值的使用


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
