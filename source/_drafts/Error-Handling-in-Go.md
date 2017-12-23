---
title: Go语言中的错误处理（Error Handling in Go）
tags:
  - coding
  - go
categories:
  - 软件开发
---

# 概述

在实际工程项目中，我们希望通过程序的异常信息快速定位问题，但是又不喜欢异常处理代码写的冗余而又啰嗦。

`Go`语言没有提供像`Java`、`C#`语言中的`try...catch`异常处理方式，而是通过返回值逐层往上抛。这种设计，鼓励工程师在代码中显式的检查错误，而非忽略错误，好处就是避免漏掉本应处理的异常。但是带来一个弊端，让代码啰嗦。

# Go标准包提供的异常处理功能

`error`是个`interface`:

```go
type error interface {
    Error() string
}
```

为什么是个`interface`，而不是`struct`？我想原因主要是：

- go语言中并无继承，使用`struct`，无法让go开发者自定义自己的错误类型，`interface`无此担忧
- 在go语言的语法体系下，使用`interface`，在碰到 函数返回、类型断言 时候，更方便

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

# 生产级别的异常处理包含哪些内容

todo

# 如何实现生产级别的异常处理需求

todo

# 分布式系统中的异常处理

todo

# 参考

- The Go Blog
    - [Error handling in Go](https://blog.golang.org/error-handling-and-go)
- Dave Cheney
    - [presentation on my philosophy for error handling](https://dave.cheney.net/paste/gocon-spring-2016.pdf)
    - [Don’t just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)
    - [Stack traces and the errors package](https://dave.cheney.net/2016/06/12/stack-traces-and-the-errors-package)
- Go Packages
    - [errors](https://golang.org/pkg/errors/)
    - [runtime](https://golang.org/pkg/runtime/)
- pkg: Artisanal, hand crafted, barrel aged, Go packages
    - [github.com/pkg/errors](https://godoc.org/github.com/pkg/errors)
