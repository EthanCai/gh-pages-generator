---
title: 记一次结对开发Golang组件的过程
categories:
  - 编程开发
tags:
  - golang
date: 2016-04-20 02:09:39
---


# 目录

<!-- TOC depthFrom:1 depthTo:2 withLinks:0 updateOnSave:1 orderedList:0 -->

- 前言
- `dlog`的用途
- 对`dlog`的一些非功能性需求
- 碰到问题及解决方案
    - 何时使用`panic`，何时使用`return error`
    - 如何实现一个`logger`只能接收对应类型的`data log`
    - 如何实现批量发送`data log`
    - 如何实现对`Logger.Log`方法的调用超时机制
    - 如何在`logger`没有收到新`msg`情况下，保证`buf`中的数据依然会定期发送给AWS Kinesis
    - 如何向程序外部暴露运行指标
    - 如何在单元测试中实现`Setup`和`TearDown`
    - 如何实现`kinesisMock`
    - 如何模拟AWS Kinesis响应慢或者不可用
    - 提交到代码库中的测试代码是否可以保留`log.Print`
- 踩过的一些坑
- 未来可以优化的地方
- 参考

<!-- /TOC -->

# 前言

本文记录了前段时间我和[王益](https://segmentfault.com/a/1190000002416822)使用Go语言合作开发一个log组件[dlog](https://github.com/topicai/dlog)的过程中学到的一些知识。在整个合作开发的过程中，王益严谨认真的态度，对开发质量的严格要求，给我留下了极其深刻的印象。能够和王益这样的顶级工程师切磋技艺，对我学习Go语言帮助非常大。也谨以此文表达对王益的感谢。

> 注：本文假设读者已经对Go语法已经有基本了解。

# `dlog`的用途

首先引用项目**readme文档**的第一段文字介绍一下`dlog`的用途：

> dlog is a Go package for distributed structure logging using Amazon AWS Kinesis/Firehose.

更多介绍和设计请阅读[readme文档](https://github.com/topicai/dlog/blob/develop/README.md)

`dlog`主要是用来记录程序的`data log`的这样一个Golang package，那什么是`data log`？这里先简要解释一下。一般程序运行过程中主要产生两类日志：

- `status log`：主要用于帮助调试、定位程序Bug、或者找到性能瓶颈，比如方法调用日志、错误日志、方法执行时间日志等
- `data log`：主要用于记录用户行为，收集的`data log`用于后期的个性化搜索、智能推荐等，比如搜索行为、点击行为等


# 对`dlog`的一些非功能性需求

- 每一种类型的`data log`对应一种`logger`，一个`logger`只能记录对应类型的`data log`
- `dlog`内部发生的错误，不能影响调用的程序代码的执行
    - 应考虑到AWS Kinesis服务响应慢或者不可用的场景（暂未实现）
- 程序代码中通过调用`dlog`的方法记录`data log`，`dlog`的方法不能阻塞调用的程序代码的执行（这一点`dlog`暂时未满足要求，需要后期改进）
- AWS Kinesis提供两个API接收数据，一个是[PutRecord](http://docs.aws.amazon.com/kinesis/latest/APIReference/API_PutRecord.html), 另一个是[PutRecords](http://docs.aws.amazon.com/kinesis/latest/APIReference/API_PutRecords.html)，为了减少对Kinesis的调用次数，采用后者批量发送`data log`
    - `PutRecords`对一次调用的`record`数量限制是`500`，每个`record`大小必须小于等于1MB，整个`request`的大小必须小于等于5MB
    - 每一个Kinesis Stream能够承受的最大TPS和写数据量，与这个stream拥有的shard的数量有关。一个shard支持最大TPS是`1000 records per second`， 写数据量是`1MB per second`
- 通过单元测试保证功能正确性


# 碰到问题及解决方案

## 何时使用`panic`，何时使用`return error`

先看看`panic`和`return error`的执行机制。

### `panic`的执行机制

`panic`会中断当前`goroutine`的执行，如果不对`panic`的错误进行`recover`，那么整个进程都会崩溃。

```go
package main

import (
    "fmt"
    "log"
    "time"
)

func main() {
    go func() {
        log.Panic("some error before work2")
        fmt.Println("do some work2")
    }()

    time.Sleep(time.Second)
    fmt.Println("do some work1")
}
```

_执行上面代码请点击[这里](https://play.golang.org/p/off1y9tBax)_

可以通过`recover`捕捉当前`goroutine`中`panic`的错误并进行错误处理，整个进程的正常运行不受影响。

```go
package main

import (
    "fmt"
    "log"
    "time"
)

func main() {
    go func() {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("error: %v", err)
            }
        }()
        log.Panic("some error before work2")
        fmt.Println("do some work2")
    }()

    time.Sleep(time.Second)
    fmt.Println("do some work1")
}
```

_执行上面代码请点击[这里](https://play.golang.org/p/RLWyb813Uw)_

我们可以发现Go语言中的`panic`、`recover`机制，和Java、.NET中的`throw`、`try...catch`机制非常类似。

### `return error`的执行机制

`return error`是利用Go语言函数的多值返回的特性，通过函数的其中一个返回值（一般是第一个或者最后一个），向`caller`返回函数执行过程中产生的异常，其它值返回执行结果。

这种方式的问题，主要在于：如果函数调用层次比较多，每一层函数都通过`return error`方式返回错误，都需要处理被调用函数的`return error`，增加代码复杂度。对于无法恢复的错误也没有必要一层一层往上抛，直接`panic/recover`更加简洁。

```go
package main

import (
    "errors"
    "log"
)

type R struct {
}

func f1() (error, *R) {
    return errors.New("an error"), nil
}

func f2() (error, *R) {
    err, r := f1()
    if err != nil {
        return err, nil
    }

    return nil, r
}

func f3() (error, *R) {
    err, r := f2()
    if err != nil {
        return err, nil
    }

    return nil, r
}

func main() {
    err, _ := f3()
    if err != nil {
        log.Print(err)
    }
}
```

_执行上面代码请点击[这里](https://play.golang.org/p/GhE5JpzZvn)_

### `dlog`错误处理原则

使用`panic`还是`return error`的方式处理错误，要区分不同的场景。重要是不论使用`panic`还是`return error`，都需要符合架构上更高层面错误处理需求。

`dlog`是一个日志记录`package`，暴露给其它程序调用的方法如下：

- `func NewLogger(example interface{}, opts *Options) (*Logger, error)`
- `func (l *Logger) Log(msg interface{}) error`

这两个方法的使用场景并不一样，错误处理原则也不完全一致：

- `NewLogger`方法一般是在程序初始化的时候调用，用于创建记录程序运行过程中产生的data log的记录器。通过`NewLogger`创建一个`logger`的时候，如果传入参数不正确，使用`panic`方式，在上层调用程序不处理错误情况下会导致程序崩溃，所以使用`return error`方式向`caller`报告错误。大多数Golang package也是按此原则处理。
- 上层程序调用`logger.Log`时，如果`Log`方法内部发生的错误，不能影响调用的代码的执行，所以这里绝对不能用`panic`方式抛出错误。日志记录是辅助功能，如果日志记录行为失败，导致业务逻辑代码执行不下去，估计负责业务逻辑开发的工程师会和你拼命。
    - `logger.Log`可以使用`return error`方式返回`msg`校验类的错误
    - `logger.Log`发送日志采用的是异步批量方式向AWS Kinesis发送数据，向AWS Kinesis发送数据相关的错误无法通过`panic`或者`return error`方式直接报告给调用程序。最好的方式是允许调用程序向`logger`注册发送失败处理的`handler`，出现发送失败错误时执行`handler`逻辑。（暂未实现）

## 如何实现一个`logger`只能接收对应类型的`data log`

要实现一个`logger`只能接收对应类型的`data log`，主要思路如下：

- `Logger`的定义中通过属性`msgType reflect.Type`记住能够接受的消息类型
- 通过`NewLogger`方法创建`logger`的时候，指定`logger`可以接受的消息类型
- `Log`方法中首先校验`msg`的类型是否是创建`logger`时指定的类型

以下是相关代码：

```go
// msgType保存Logger能够接受的消息类型
type Logger struct {
    ...
    msgType    reflect.Type
    ...
}

// 获得msg的reflect.Type
func msgType(msg interface{}) (reflect.Type, error) {
    t := reflect.TypeOf(msg)

    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }

    if t.Kind() != reflect.Struct {
        return nil, fmt.Errorf("dlog message must be either *struct or struct")
    }

    return t, nil
}

func NewLogger(example interface{}, opts *Options) (*Logger, error) {
    t, e := msgType(example)
    if e != nil {
        return nil, e
    }

    ...

    l := &Logger{
        ...
        msgType:    t,
        ...
    }

    ...
    return l, nil
}

func (l *Logger) Log(msg interface{}) error {
    if t, e := msgType(msg); e != nil {
        return e
    } else if !t.AssignableTo(l.msgType) {
        return fmt.Errorf("parameter (%+v) not assignable to %v", msg, l.msgType)
    }

    ...
}
```

`Log`方法中为什么要用`AssignableTo`，而不是直接判断两个类型相等。其实都可以，在`msg`是`struct`情况下，`AssignableTo`返回`True`意味着两个类型相等。参考下面的例子：

```go
package main

import (
    "log"
    "reflect"
)

func main() {
    type Fn func(int) int
    id := func(x int) int {
        return x
    }
    var zeroFn Fn
    log.Println(reflect.TypeOf(id).AssignableTo(reflect.TypeOf(zeroFn)))

    type MyInt int
    mi := 1
    log.Println(reflect.TypeOf(2).AssignableTo(reflect.TypeOf(mi)))

    type S1 struct {
        name string
    }
    type S2 S1

    s1 := S1{
        name: "ethan",
    }
    s2 := S2{
        name: "ethan",
    }
    // s2 = s1	// if uncomment this line, will report "cannot use s1 (type S1) as type S2 in assignment" when compile
    log.Println(reflect.TypeOf(s1).AssignableTo(reflect.TypeOf(s2)))
}
```

_执行上面代码请点击[这里](https://play.golang.org/p/eDmzxW-ayk)_

## 如何实现批量发送`data log`

要实现批量发送，首先我们可以想到应该要有个`buffer`用来收集一定数量的的`message`，等待`buffer`中的数据积累到一定程度后，一次性发送给AWS Kinesis。设计`buffer`结构不难，难点在于如何解决多线程(goroutine)并发读写`buffer`的问题，主要的解决方案有两种：

- 基于锁机制实现对`buffer`访问控制
- 基于`channel`实现对`buffer`的访问控制

前者对于有Java、.NET等语言的并发编程经验的工程师来说，非常熟悉。而后者则体现了CSP(Communicating Sequential Processes)并发编程模型的优势。

{% asset_img channel.png CSP Model %}

`dlog`的`Log`方法把收到的`msg`写到名字叫`buffer`的`channel`中，另外一个单独的`goroutine`在`channel`的另一头收集编码后的日志信息，然后保存到`buf := make([][]byte, 0)`中。当`buf`中的数据量要达到一次向AWS Kinesis发送的最大量时，调用`flush`方法向AWS Kinesis发送数据。由于只有一个`goroutine`对`buf`进行访问，所以不需要通过锁机制控制对`buf`的读写。

<!--
digraph G {
    fontname="Microsoft YaHei";
    fontsize=10;
    rankdir = LR;

    "buffer channel" [shape=box];

    "Logger.Log goroutine 1" -> "buffer channel";
    "Logger.Log goroutine 2" -> "buffer channel";
    "Logger.Log goroutine 3" -> "buffer channel";
    "buffer channel" -> "sync goroutine";
    "sync goroutine" -> "AWS Kinesis Stream";
}
-->

<!-- ![Thread Model](http://g.gravizo.com/g?digraph%20G%20%7B%0A%20%20%20%20fontname%3D%22Microsoft%20YaHei%22%3B%0A%20%20%20%20fontsize%3D10%3B%0A%20%20%20%20rankdir%20%3D%20LR%3B%0A%0A%20%20%20%20%22buffer%20channel%22%20%5Bshape%3Dbox%5D%3B%0A%0A%20%20%20%20%22Logger.Log%20goroutine%201%22%20-%3E%20%22buffer%20channel%22%3B%0A%20%20%20%20%22Logger.Log%20goroutine%202%22%20-%3E%20%22buffer%20channel%22%3B%0A%20%20%20%20%22Logger.Log%20goroutine%203%22%20-%3E%20%22buffer%20channel%22%3B%0A%20%20%20%20%22buffer%20channel%22%20-%3E%20%22sync%20goroutine%22%3B%0A%20%20%20%20%22sync%20goroutine%22%20-%3E%20%22AWS%20Kinesis%20Stream%22%3B%0A%20%7D) -->

{% asset_img use_channel.png 使用Channel %}

具体代码实现：

```go
func NewLogger(example interface{}, opts *Options) (*Logger, error) {
    ...

    go l.sync()    // 启动sync goroutine
    return l, nil
}

func (l *Logger) Log(msg interface{}) error {
    ...

    en := encode(msg)       // 对msg进行编码
    ...
        select {
        case l.buffer <- en:    // 向buffer channel写入编码后的msg
        ...
        }
    ...
    return nil
}

func (l *Logger) sync() {
    ...

    buf := make([][]byte, 0) // 用于收集从buffer channel读取的日志数据
    bufSize := 0

    for {
        select {
        case msg := <-l.buffer:
            if bufSize+len(msg)+partitionKeySize >= maxBatchSize {  // 如果buf的大小接近一次批量发送的最大数据量
                l.flush(&buf, &bufSize)                             // 向AWS Kinesis批量发送数据
            }

            buf = append(buf, msg)                                  // 将从buffer channel读取日志数据保存到buf中
            bufSize += len(msg) + partitionKeySize

        ...
    }
}
```

## 如何实现对`Logger.Log`方法的调用超时机制

如果一个IO操作耗时较长，并且调用比较频繁的情况下，不仅会阻塞`caller`的执行，还会消耗大量系统资源。我们通常会使用超时机制，避免程序长时间等待或者对系统资源大量占用。

`Logger.Log`方法利用Go语言`channel`非常简洁的实现了超时机制：

```go
func (l *Logger) Log(msg interface{}) error {
    ...

    var timeout <-chan time.Time
    if l.WriteTimeout > 0 {
        timeout = time.After(l.WriteTimeout)    // 初始化时长为l.WriteTimeout的计时器
    }

    ...
        select {
        case l.buffer <- en:
        case <-timeout: // 如果上一行代码一直阻塞，timeout计时器时间到点后会触发执行当前case下的代码
            return fmt.Errorf("dlog writes %+v timeout after %v", msg, l.WriteTimeout)
        }
    ...
    return nil
}
```

对比Java、.NET语言中超时机制的实现方法，Go语言的实现简洁的令人发指：

- C#
    - [Implementing .Net method timeout](http://weblogs.asp.net/israelio/159985)
    - [How to implement Task Async for a timer in C#?](http://stackoverflow.com/questions/18646650/how-to-implement-task-async-for-a-timer-in-c)
    - [Implementing a timeout in c#](http://stackoverflow.com/questions/10143980/implementing-a-timeout-in-c-sharp)
- Java
    - [How to implement timeout using threads](http://www.coderanch.com/t/232213/threads/java/implement-timeout-threads)
    - [How to timeout a thread](http://stackoverflow.com/questions/2275443/how-to-timeout-a-thread)

## 如何在`logger`没有收到新`msg`情况下，保证`buf`中的数据依然会定期发送给AWS Kinesis

`dlog`在`Logger.sync()`方法中通过一个定时器，定期将`buf`中数据发送给AWS Kinesis。

```go
func (l *Logger) sync() {
    if l.SyncPeriod <= 0 {
        l.SyncPeriod = time.Second
    }
    ticker := time.NewTicker(l.SyncPeriod)  // l.SyncPeriod是定期发送的数据的时间间隔，ticker定时触发器

    buf := make([][]byte, 0)
    bufSize := 0

    for {  // 无限循环保证sync goroutine一直工作
        select {
        case msg := <-l.buffer:
            ...

        case <-ticker.C: // ticker.C的类型是<-chan Time，每隔l.SyncPeriod时间会触发执行当前case的代码
            if bufSize > 0 {
                l.flush(&buf, &bufSize)
            }
        }
    }
}
```

通过`ticker`，`dlog`保证了即使没有收到新的`msg`的时候，保存在`buf`中的数据最长`l.SyncPeriod`时间后也会发送给AWS Kinesis。

互联网产品的生产环境的上线，通常的做法是，将现有服务分组，然后交替切流量、升级。如果没有类似的机制，那么在服务程序断掉流量，没有收到新的访问时候，保存在内存中的数据就不会发送出去，升级时就可能导致数据丢失。

## 如何向程序外部暴露运行指标

Go语言的官方Package `expvar`提供一种标准化的接口，允许程序暴露公开访问的变量。`expvar`通过HTTP地址`/debug/vars`提供访问入口，并以JSON格式展示这些变量。下面是关于`expvar`常见用法的一个例子：

```go
package main

import (
    "encoding/json"
    "expvar"
    "fmt"
    "net/http"
    "sync"
    "time"
)

// Stats is used to collect runtime metrics
type Stats struct {
    sync.Mutex
    TotalHit  int
    ErrorNums int
}

func (s *Stats) IncreaseTotalHit(i int) {
    s.Lock()
    defer s.Unlock()

    s.TotalHit += i
}

func (s *Stats) IncreaseErrorNums(i int) {
    s.Lock()
    defer s.Unlock()

    s.ErrorNums += i
}

func (s *Stats) String() string {
    s.Lock()
    defer s.Unlock()

    b, err := json.Marshal(*s)
    if err != nil {
        return "{}"
    } else {
        return string(b)
    }
}

var (
    stats *Stats
    hits  *expvar.Map
)

func init() {

    expvar.Publish("now", expvar.Func(func() interface{} {
        return time.Now().Format("\"2006-01-02 15:04:05\"")
    }))

    stats = &Stats{}
    expvar.Publish("stats", stats)

    hits = expvar.NewMap("hits").Init()
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    p := r.URL.Path[1:]
    hits.Add(p, 1)
    stats.IncreaseTotalHit(1)
    fmt.Fprintf(w, "Hey! I love %s! hits: %v\n", p, hits.Get(p))
}

func errHandler(w http.ResponseWriter, r *http.Request) {
    stats.IncreaseErrorNums(1)
    fmt.Fprintf(w, "Error Nums: %v\n", stats.ErrorNums)
}

func main() {
    http.HandleFunc("/err", errHandler)
    http.HandleFunc("/", homeHandler)
    http.ListenAndServe(":8080", nil)
}
```

按照如下步骤测试运行效果：

- `go run expvarexample.go`运行例子代码
- 在浏览器中访问`http://localhost:8080/ethan`
- 在浏览器中访问`http://localhost:8080/err`
- 在浏览器中访问`http://localhost:8080/debug/vars`，得到如下结果：
```json
{
    "cmdline": ["/var/folders/jf/65ft181j33j_d75ktgv67bsc0000gn/T/go-build467453980/command-line-arguments/_obj/exe/expvarsample"],
    "hits": {
        "ethan": 1,
        "favicon.ico": 2
    },
    "memstats": { ... },
    "now": "\"2016-04-19 20:17:40\"",
    "stats": {
        "TotalHit":3,
        "ErrorNums":1
    }
}
```

[expvarmon](https://github.com/divan/expvarmon)是一个帮助查看`expvar`暴露运行指标的工具，用法如下：

- 安装：`go get github.com/divan/expvarmon`
- 运行：`expvarmon -ports="8080" -vars="hits.ethan,stats.TotalHit,stats.ErrorNums,now"`
- 效果如下：
{% asset_img expvarmon_screen.png expvarmon screen %}

`dlog`使用`expvar`向程序外部（比如监控程序）暴露运行指标，目前`dlog`中定义的运行指标包括：

- `writtenRecords`: 成功写到AWS Kinesis的`msg`数量
- `writtenBatches`: 成功调用AWS Kinesis批量写数据API的次数
- `failedRecords`: 写到AWS Kinesis失败的`msg`数量
- `tooBigMesssages`: 编码后体积过大(加上partitionKeySize大于1MB)的`msg`数量

未来还需要根据运维的需求对运行指标进行调整，当前的用法也有一些问题，后期需要重构。

## 如何在单元测试中实现`Setup`和`TearDown`

Go语言提供一种**轻量级**的单元测试框架（无需第三方工具或者程序包）。通过使用`go test`命令和`testing` package，可以非常快速的实现单元测试。先借用官方文档中的[例子](http://docs.studygolang.com/doc/code.html#Testing)回顾一下Go单元测试框架的用法：

```go
//$GOPATH/src/github.com/user/stringutil/reverse_test.go
package stringutil

import "testing"

func TestReverse(t *testing.T) {
    cases := []struct {
        in, want string
    }{
        {"Hello, world", "dlrow ,olleH"},
        {"Hello, 世界", "界世 ,olleH"},
        {"", ""},
    }
    for _, c := range cases {
        got := Reverse(c.in)
        if got != c.want {
            t.Errorf("Reverse(%q) == %q, want %q", c.in, got, c.want)
        }
    }
}
```

运行测试只需要简单的输入命令：

```go
$ go test github.com/user/stringutil
ok  	github.com/user/stringutil 0.165s
```

很多情况下，要执行单元测试，我们需要依赖一些外部资源，比如已完成初始化数据的数据库、公有云上的一些IaaS服务等。这些依赖资源，我们希望在单元测试执行前，能够自动的被初始化；单元测试完成后，能够自动的被清理。[testify/suite](https://github.com/stretchr/testify/suite) package就提供这样的支持。通过[testify/suite](https://github.com/stretchr/testify/suite)，你可以构建一个测试集`struct`，建立测试集的`setup`(初始化)/`teardown`(清理)方法，和最终实现测试用例逻辑的方法。而运行测试，仍然只需要一句简单的`go test`。

以下是使用[testify/suite](https://github.com/stretchr/testify/suite)实现测试集的常见模式：

```go
package suite

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

type SuiteTester struct {
    // Include our basic suite logic.
    Suite

    // Other properties
    propertyN string
}

// The SetupSuite method will be run by testify once, at the very
// start of the testing suite, before any tests are run.
func (suite *SuiteTester) SetupSuite() {
    // ...
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *SuiteTester) TearDownSuite() {
    // ...
}

// The SetupTest method will be run before every test in the suite.
func (suite *SuiteTester) SetupTest() {
    // ...
}

// The TearDownTest method will be run after every test in the suite.
func (suite *SuiteTester) TearDownTest() {
    // ...
}

// a test method
func (suite *SuiteTester) TestOne() {
    // ...
}

// another test method
func (suite *SuiteTester) TestTwo() {
    // ...
}

// TestRunSuite will be run by the 'go test' command, so within it, we
// can run our suite using the Run(*testing.T, TestingSuite) function.
func TestRunSuite(t *testing.T) {
    suiteTester := new(SuiteTester)
    Run(t, suiteTester)
}
```

`dlog`中为了测试`Logger.Log`方法能否正常工作，按照上面的模式编写了相应的测试代码：

```go
package dlog

//...

type WriteLogSuiteTester struct {
    suite.Suite

    options     *Options
    seachLogger *Logger
    clickLogger *Logger
    streamNames []string // save the created AWS Kinesis Streams, which will be removed in TearDownSuite()
}

// The SetupSuite method will be run by testify once, at the very
// start of the testing suite, before any tests are run.
func (s *WriteLogSuiteTester) SetupSuite() {

    //...

    // create stream 1
    err = s.seachLogger.kinesis.CreateStream(s.seachLogger.streamName, testingShardCount)
    s.Nil(err)

    // create stream 2
    err = s.clickLogger.kinesis.CreateStream(s.clickLogger.streamName, testingShardCount)
    s.Nil(err)

    s.streamNames = []string{s.seachLogger.streamName, s.clickLogger.streamName}

    for { // waiting created stream's status to be active
        time.Sleep(1 * time.Second)
        resp1, err1 := s.seachLogger.kinesis.DescribeStream(s.seachLogger.streamName)
        s.Nil(err1)

        resp2, err2 := s.seachLogger.kinesis.DescribeStream(s.clickLogger.streamName)
        s.Nil(err2)

        status1 := strings.ToLower(string(resp1.StreamStatus))
        status2 := strings.ToLower(string(resp2.StreamStatus))
        if status1 == "active" && status2 == "active" {
            break
        }
    }
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (s *WriteLogSuiteTester) TearDownSuite() {
    if s.streamNames == nil || len(s.streamNames) == 0 {
        return
    }

    for _, streamName := range s.streamNames {
        err := s.seachLogger.kinesis.DeleteStream(streamName)
        s.Nil(err)
    }
}

func (s *WriteLogSuiteTester) TestWriteLog() {
    defer func() { // Recover if panicking to make sure TearDownSuite will be executed
        if r := recover(); r != nil {
            s.Fail(fmt.Sprint(r))
        }
    }()

    //...
}

func TestRunWriteLogSuite(t *testing.T) {
    suiteTester := new(WriteLogSuiteTester)
    suite.Run(t, suiteTester)
}
```

注：

- 很多场景下，测试程序自动创建依赖的资源需要运维部门的授权，所以实现前有必要先和运维部门沟通。
- 云环境下，出于安全上的考虑，需要对创建、删除测试资源的账户管理严格管理
    - 账户信息不能写在可以公开访问的测试代码、配置文件中
    - 只给账户分配必要资源的最小权限
    - 为账户能够创建的资源设定配额

## 如何实现`kinesisMock`

上一节我们提到在测试执行前初始化依赖资源，现实场景中，并不是任何情况下都能够获得依赖的测试资源，或者测试资源也会出现不可用的情况。通过Mock技术，可以减少测试代码对其它资源（或模块）的依赖。

`dlog`的测试代码中，首先定义了一个`KinesisInterface`:

```go
type KinesisInterface interface {
    PutRecords(streamName string, records []kinesis.PutRecordsRequestEntry) (resp *kinesis.PutRecordsResponse, err error)
    CreateStream(name string, shardCount int) error
    DescribeStream(name string) (resp *kinesis.StreamDescription, err error)
    DeleteStream(name string) error
}
```

`KinesisInterface`包含了`dlog`用到的[github.com/AdRoll/goamz/kinesis/kinesis](https://github.com/AdRoll/goamz/blob/master/kinesis/kinesis.go)的所有方法。因为Go语言`interface`实现**非侵入式**的特点，[github.com/AdRoll/goamz/kinesis/kinesis](https://github.com/AdRoll/goamz/blob/master/kinesis/kinesis.go)自动实现了`KinesisInterface`，我们再定义一个`kinesisMock`实现`KinesisInterface`：

```go
type kinesisMock struct {
    // Mapping from steam name to batches of batches
    storage map[string][][]kinesis.PutRecordsRequestEntry

    // simulate lantency that sync to Kinesis
    putRecordLatency time.Duration

    // created streams' names
    streamNames []string

    // lock to solve concurrent call
    lock sync.RWMutex
}

func newKinesisMock(putRecordsLatency time.Duration) *kinesisMock {
    return &kinesisMock{
        storage:          make(map[string][][]kinesis.PutRecordsRequestEntry),
        putRecordLatency: putRecordsLatency,
        streamNames:      make([]string, 0),
    }
}

func (mock *kinesisMock) PutRecords(streamName string, records []kinesis.PutRecordsRequestEntry) (resp *kinesis.PutRecordsResponse, err error) {
    // ...
}

func (mock *kinesisMock) CreateStream(name string, shardCount int) error {
    // ...
}

func (mock *kinesisMock) DescribeStream(name string) (resp *kinesis.StreamDescription, err error) {
    // ...
}

func (mock *kinesisMock) DeleteStream(name string) error {
    // ...
}
```

然后，把业务代码中所有类型`kinesis`的变量，替换成`KinesisInterface`类型。

```go
type Logger struct {
    //...
    kinesis    KinesisInterface
    //...
}
```

测试代码中，在构造`Logger`时传入`kinesisMock`，而不是真实的`kinesis`，这样就做到了“狸猫换太子”。

```go
func TestLoggingToMockKinesis(t *testing.T) {
    assert := assert.New(t)

    l, e := NewLogger(&impression{}, &Options{
        // ...
        UseMockKinesis: true,
        MockKinesis:    newKinesisMock(0),
    })

    // ...
}
```

## 如何模拟AWS Kinesis响应慢或者不可用

`kinesisMock`完全是我们“虚构”出来的一个`kinesis`，在它的基础上，我们完全可以模拟响应慢或者不可用的情况。

上一节中，不知道大家注意到没有，`kinesisMock`有个属性叫`putRecordLatency`，用来模拟调用`PutRecords`方法的延迟时间。

```go
type kinesisMock struct {
    // ...

    // simulate lantency that sync to Kinesis
    putRecordLatency time.Duration

    // ...
}

func (mock *kinesisMock) PutRecords(streamName string, records []kinesis.PutRecordsRequestEntry) (resp *kinesis.PutRecordsResponse, err error) {
    //...

    time.Sleep(mock.putRecordLatency) // 模拟延迟

    //...
}
```

模拟不可用的`kinesis`则重新定义了一个`brokenKinesisMock`：

```go
type brokenKinesisMock struct {
    *kinesisMock
}

func newBrokenKinesisMock() *brokenKinesisMock {
    return &brokenKinesisMock{
        kinesisMock: newKinesisMock(0),
    }
}

func (mock *brokenKinesisMock) PutRecords(streamName string, records []kinesis.PutRecordsRequestEntry) (resp *kinesis.PutRecordsResponse, err error) {
    return nil, fmt.Errorf("Kinesis is broken")
}
```

`kinesisMock`是`brokenKinesisMock`的嵌入`struct`，`brokenKinesisMock`会自动拥有`kinesisMock`的所有公开方法，这样也就实现了`KinesisInterface`。

## 提交到代码库中的测试代码是否可以保留`log.Print`

结论是“不可以”，原因总结如下：

- 测试代码中的`log.Print`，一般用于调试代码，或者在`stdout`打印出一些信息帮助判断测试失败原因。不论哪种目的，这样的代码目的都仅仅是为了辅助开发，而不应该出现在最终交付的产品代码中。
- `go test`命令会在控制台输出失败的测试方法，如果加上`-v`标志会打印出所有测试方法的执行结果，`log.Print`会影响执行结果的展示效果。团队合作开发，如果每个人都在测试代码中加上自己的`log.Print`，那么控制台打印出来的测试结果就没法看了。


# 踩过的一些坑

- [AWS Kinesis API - CreateStream](http://docs.aws.amazon.com/kinesis/latest/APIReference/API_CreateStream.html)是异步创建Stream，而且耗时10+秒，才能完成一个Stream的创建。开始以为是同步创建，结果执行测试逻辑的时候总是出错。
- [github.com/AdRoll/goamz/aws/regions.go](https://github.com/AdRoll/goamz/blob/master/aws/regions.go)中缺少中国区AWS Kinesis的URL地址，调用中国区AWS Kinesis会出错。
- Travis CI会Kill掉执行时间超过1分钟的CI过程，而不是如它文档中介绍的“10分钟”


# 未来可以优化的地方

- 发送失败的错误事件机制
- 实现Kinesis服务不可用或者响应慢的场景下`dlog`的容错处理


# 参考

- [Effective Go](http://docs.studygolang.com/doc/effective_go.html)
- [Amazon Kinesis Documentation](https://aws.amazon.com/cn/documentation/kinesis/)
- [Advanced Go Concurrency Patterns](http://blog.golang.org/advanced-go-concurrency-patterns)
- [hystrix-go](https://github.com/afex/hystrix-go)


# 招聘消息

我所在的[奥阁门科技有限公司](http://www.augmn.com)正在招聘后端、运维工程师，想加入的朋友、或者有朋友可以推荐的都可以联系我(ethancai@qq.com)。

{% asset_img 2016-04-21_07-12-24.png 办公环境1 %}

{% asset_img 2016-04-21_07-12-40.png 办公环境2 %}

**后端工程师 / Backend Engineer**

职责

- 研讨和设计产品功能特性；
- 设计研发系统后端的一个或多个独立服务（micro-service）模块；
- 设计研发业务运营管理系统；
- Code Review。

要求

- 有良好的编程习惯和代码风格；
- 精通至少一种后台开发语言，包括但不限于Go、Node.js、C++、Python；
- 对RESTful、RPC等架构有深刻理解和运用经验；
- 有丰富的web service、web app开发经验；使用过著名的开源应用框架，并完整阅读过源代码；
- 对Mysql、Redis、MongoDB或同类数据存储技术有丰富的使用经验；
- 有提交代码到著名开源库或创建过开源项目者优先；
- 能熟练查阅英文技术文档；
- 有开放、坦诚的沟通心态，乐于分享；
- 5年以上工作经验，3年以上后台系统开发经验。


**高级系统运维工程师 / Senior Ops Engineer**

职责

- 负责日常业务系统基础实施（AWS）、网络及各子系统的管理维护。
- 负责设计并部署相关应用平台，并提出平台的实施、运行报告。
- 负责配合开发搭建测试平台，协助开发设计、推行、实施和持续改进。
- 负责相关故障、疑难问题排查处理，编制汇总故障、问题，定期提交汇总报告。
- 负责网络监控和应急反应，以确保网络系统有7*24小时的持续运作能力。
- 负责日常系统维护，及监控，提供IT方面的服务和支持，保证系统的稳定。

要求

- 深入理解Linux/Unix操作系统并能熟练使用，了解Linux系统内核，有相关操作系统调优经验优先；
- 熟悉计算机网络基础知识，了解TCP/IP、HTTP等网络协议；
- 熟悉系统服务的管理和维护，例如：Nginx、DNS服务器、NTP服务等；
- 熟悉一种或者多种脚本语言，例如：Shell、Python、Perl 、Ruby等；
- 熟练掌握Linux管理相关命令行工具，例如：grep、awk、sed、tmux、vim等；
- 对数据库系统（MySQL）运维管理有一定的了解；
- 熟悉常见分布式系统系统架构部署管理，熟悉基础设施管理、并具有较强的故障排查和解决问题的能力；
- 具有 2 年以上中大型互联网系统或亚马逊AWS管理经验者优先；
- 有DevOps经验者优先；
- 学习能力和沟通能力较强，具有良好的团队协作精神；
- 工作中需要胆大心细，具备探索创新精神；
- 具有良好的文档编写能力；
- 具有一定的英文技术文档阅读能力。
