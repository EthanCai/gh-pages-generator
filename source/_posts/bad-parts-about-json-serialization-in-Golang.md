---
title: Golang中遇到的一些关于JSON处理的坑
categories:
  - 编程开发
tags:
  - golang
  - json
date: 2016-06-23 08:09:39
---


# 前言

一个人不会两次掉进同一个坑里，但是如果他（她）忘记了坑的位置，那就不一定了。

这篇文章记录了最近使用Golang处理JSON遇到的一些坑。

# 坑

## 1号坑：`omitempty`的行为

C#中最常用的JSON序列化类库`Newtonsoft.Json`中，把一个类的实例序列化成JSON，如果我们不想让某个属性输出到JSON中，可以通过`property annotation`或者`ShouldSerialize method`等方法，告知序列化程序。如下：

```c#
// 通过ShouldSerialize method指示不要序列化ObsoleteSetting属性
class Config
{
    public Fizz ObsoleteSetting { get; set; }

    public bool ShouldSerializeObsoleteSetting()
    {
        return false;
    }
}

// 通过JsonIgnore的annotation指示不需要序列化ObsoleteSetting属性
class Config
{
    [JsonIgnore]
    public Fizz ObsoleteSetting { get; set; }

    public Bang ReplacementSetting { get; set; }
}
```

关于`Newtonsoft.Json`的Conditional Property Serialization的更多内容参考：

- [Conditional Property Serialization](http://www.newtonsoft.com/json/help/html/ConditionalProperties.htm)
- [Making a property deserialize but not serialize with json.net](http://stackoverflow.com/questions/11564091/making-a-property-deserialize-but-not-serialize-with-json-net)

开始使用Golang的时候，以为`omitempty`的行为和C#中一样用来控制是否序列化字段，结果使用的时候碰了一头钉子。回头阅读[encoding/json package的官方文档](http://docs.studygolang.com/pkg/encoding/json/#Marshal)，找到对`omitempty`行为的描述：

> Struct values encode as JSON objects. Each exported struct field becomes a member of the object unless
>
> - the field's tag is "-", or
> - the field is empty and its tag specifies the "omitempty" option.
>
> The empty values are false, 0, any nil pointer or interface value, and any array, slice, map, or string of length zero. The object's default key string is the struct field name but can be specified in the struct field's tag value. The "json" key in the struct field's tag value is the key name, followed by an optional comma and options. Examples:
>
> ```go
> // Field is ignored by this package.
> Field int `json:"-"`
>
> // Field appears in JSON as key "myName".
> Field int `json:"myName"`
>
> // Field appears in JSON as key "myName" and
> // the field is omitted from the object if its value is empty,
> // as defined above.
> Field int `json:"myName,omitempty"`
>
> // Field appears in JSON as key "Field" (the default), but
> // the field is skipped if empty.
> // Note the leading comma.
> Field int `json:",omitempty"`
> ```

Golang中，如果指定一个`field`序列化成JSON的变量名字为`-`，则序列化的时候自动忽略这个`field`。这种用法，才是和上面`JsonIgnore`的用法的作用是一样的。

而`omitempty`的作用是当一个`field`的值是`empty`的时候，序列化JSON时候忽略这个`field`（`Newtonsoft.Json`的类似用法参考[这里](http://stackoverflow.com/questions/6507889/how-to-ignore-a-property-in-class-if-null-using-json-net)和[例子](https://dotnetfiddle.net/VXqRnm)）。这里需要注意的是关于`emtpty`的定义：

> The empty values are false, 0, any nil pointer or interface value, and any array, slice, map, or string of length zero.

通过下面的例子，来加深对`empty values`的了解：

```go
package main

import (
    "bytes"
    "encoding/json"
    "log"
    "os"
)

type S1 struct {
    I1 int
    I2 int `json:",omitempty"`

    F1 float64
    F2 float64 `json:",omitempty"`

    S1 string
    S2 string `json:",omitempty"`

    B1 bool
    B2 bool `json:",omitempty"`

    Slice1 []int
    Slice2 []int `json:",omitempty"`
    Slice3 []int `json:",omitempty"`

    Map1 map[string]string
    Map2 map[string]string `json:",omitempty"`
    Map3 map[string]string `json:",omitempty"`

    O1 interface{}
    O2 interface{} `json:",omitempty"`
    O3 interface{} `json:",omitempty"`
    O4 interface{} `json:",omitempty"`
    O5 interface{} `json:",omitempty"`
    O6 interface{} `json:",omitempty"`
    O7 interface{} `json:",omitempty"`
    O8 interface{} `json:",omitempty"`

    P1 *int
    P2 *int               `json:",omitempty"`
    P3 *int               `json:",omitempty"`
    P4 *float64           `json:",omitempty"`
    P5 *string            `json:",omitempty"`
    P6 *bool              `json:",omitempty"`
    P7 *[]int             `json:",omitempty"`
    P8 *map[string]string `json:",omitempty"`
}

func main() {

    p3 := 0
    p4 := float64(0)
    p5 := ""
    p6 := false
    p7 := []int{}
    p8 := map[string]string{}

    s1 := S1{
        I1: 0,
        I2: 0,

        F1: 0,
        F2: 0,

        S1: "",
        S2: "",

        B1: false,
        B2: false,

        Slice1: []int{},
        Slice2: nil,
        Slice3: []int{},

        Map1: map[string]string{},
        Map2: nil,
        Map3: map[string]string{},

        O1: nil,
        O2: nil,
        O3: int(0),
        O4: float64(0),
        O5: "",
        O6: false,
        O7: []int{},
        O8: map[string]string{},

        P1: nil,
        P2: nil,
        P3: &p3,
        P4: &p4,
        P5: &p5,
        P6: &p6,
        P7: &p7,
        P8: &p8,
    }

    b, err := json.Marshal(s1)
    if err != nil {
        log.Printf("marshal error: %v", err)
        return
    }

    var out bytes.Buffer
    json.Indent(&out, b, "", "\t")
    out.WriteTo(os.Stdout)
    //Output:
    //{
    //	"I1": 0,
    //	"F1": 0,
    //	"S1": "",
    //	"B1": false,
    //	"Slice1": [],
    //	"Map1": {},
    //	"O1": null,
    //	"O3": 0,
    //	"O4": 0,
    //	"O5": "",
    //	"O6": false,
    //	"O7": [],
    //	"O8": {},
    //	"P1": null,
    //	"P2": 0
    //}%
}
```

*点击[这里](https://play.golang.org/p/6y_m27r8EO)执行上面的程序*

关于`empty value`的定义，这里面隐藏了一些坑。下面通过一个例子来说明。

假设我们有一个社交类App，通过Restful API形式从服务端获取当前登录用户基本信息及粉丝数量。如果服务端对Response中`User`对象的定义如下：

```go
type User struct {
    ID        int `json:"id"`                  // 用户id
    // 其它field
    FansCount int `json:"fansCount,omitempty"` // 粉丝数
}
```

如果正在使用App时一个还没有粉丝的用户，访问Restful API的得到Response如下：

```json
{
    "id": 1000386,
    ...
}
```

这时候你会发现Response的User对象中没有`fansCount`，因为`fansCount`是个`int`类型且值为0，序列化的时候会被忽略。语义上，`User`对象中没有`fansCount`应该理解为**粉丝数量未知**，而不是**没有粉丝**。

如果我们希望做到能够区分**粉丝数未知**和**没有粉丝**两种情况，需要修改`User`的定义：

```go
type User struct {
    ID        int  `json:"id"`                  // 用户id
    // 其它field
    FansCount *int `json:"fansCount,omitempty"` // 粉丝数
}
```

将`FansCount`修改为指针类型，如果为`nil`，表示粉丝数未知；如果为整数（包括0），表示粉丝数。

这么修改语义上没有漏洞了，但是代码中要给`FansCount`赋值的时候却要多一句废话。必须先将从数据源查询出粉丝数赋给一个变量，然后再将变量的指针传给`FansCount`。代码读起来实在是啰嗦：

```go
// FansCount是int类型时候
user := dataAccess.GetUserInfo(userId)
user.FansCount = dataAccess.GetFansCount(userId)

// FansCount是*int类型的时候
user := dataAccess.GetUserInfo(userId)
fansCount := dataAccess.GetFansCount(userId)
user.FansCount = &fansCount
```

## 2号坑：JSON反序列化成interface{}对Number的处理

[JSON的规范](http://json.org/)中，对于数字类型，并不区分是整型还是浮点型。

![](https://www.json.org/img/value.png)

对于如下JSON文本:

```json
{
    "name": "ethancai",
    "fansCount": 9223372036854775807
}
```

如果反序列化的时候指定明确的结构体和变量类型

```go
package main

import (
    "encoding/json"
    "fmt"
)

type User struct {
    Name      string
    FansCount int64
}

func main() {
    const jsonStream = `
        {"name":"ethancai", "fansCount": 9223372036854775807}
    `
    var user User  // 类型为User
    err := json.Unmarshal([]byte(jsonStream), &user)
    if err != nil {
        fmt.Println("error:", err)
    }

    fmt.Printf("%+v \n", user)
}
// Output:
//  {Name:ethancai FansCount:9223372036854775807}
```

*点击[这里](https://play.golang.org/p/203egccrea)执行上面的程序*

如果反序列化不指定结构体类型或者变量类型，则JSON中的数字类型，默认被反序列化成`float64`类型：

```go
package main

import (
    "encoding/json"
    "fmt"
    "reflect"
)

func main() {
    const jsonStream = `
        {"name":"ethancai", "fansCount": 9223372036854775807}
    `
    var user interface{}  // 不指定反序列化的类型
    err := json.Unmarshal([]byte(jsonStream), &user)
    if err != nil {
        fmt.Println("error:", err)
    }
    m := user.(map[string]interface{})

    fansCount := m["fansCount"]

    fmt.Printf("%+v \n", reflect.TypeOf(fansCount).Name())
    fmt.Printf("%+v \n", fansCount.(float64))
}

// Output:
// 	float64
//  	9.223372036854776e+18
```

*点击[这里](https://play.golang.org/p/l4GzgA4WDA)执行上面的程序*

```go
package main

import (
    "encoding/json"
    "fmt"
)

type User struct {
    Name      string
    FansCount interface{}  // 不指定FansCount变量的类型
}

func main() {
    const jsonStream = `
        {"name":"ethancai", "fansCount": 9223372036854775807}
    `
    var user User
    err := json.Unmarshal([]byte(jsonStream), &user)
    if err != nil {
        fmt.Println("error:", err)
    }

    fmt.Printf("%+v \n", user)
}

// Output:
// 	{Name:ethancai FansCount:9.223372036854776e+18}
```

*点击[这里](https://play.golang.org/p/SoD6SOGuCM)执行上面的程序*

从上面的程序可以发现，如果`fansCount`精度比较高，反序列化成`float64`类型的数值时存在丢失精度的问题。

如何解决这个问题，先看下面程序：

```go
package main

import (
    "encoding/json"
    "fmt"
    "reflect"
    "strings"
)

func main() {
    const jsonStream = `
        {"name":"ethancai", "fansCount": 9223372036854775807}
    `

    decoder := json.NewDecoder(strings.NewReader(jsonStream))
    decoder.UseNumber()    // UseNumber causes the Decoder to unmarshal a number into an interface{} as a Number instead of as a float64.

    var user interface{}
    if err := decoder.Decode(&user); err != nil {
        fmt.Println("error:", err)
            return
        }

    m := user.(map[string]interface{})
    fansCount := m["fansCount"]
    fmt.Printf("%+v \n", reflect.TypeOf(fansCount).PkgPath() + "." + reflect.TypeOf(fansCount).Name())

     v, err := fansCount.(json.Number).Int64()
    if err != nil {
        fmt.Println("error:", err)
            return
    }
    fmt.Printf("%+v \n", v)
}

// Output:
// 	encoding/json.Number
// 	9223372036854775807
```
*点击[这里](https://play.golang.org/p/KYrFshVMFD)执行上面的程序*

上面的程序，使用了`func (*Decoder) UseNumber`方法告诉反序列化JSON的数字类型的时候，不要直接转换成`float64`，而是转换成`json.Number`类型。`json.Number`内部实现机制是什么，我们来看看源码：

```go
// A Number represents a JSON number literal.
type Number string

// String returns the literal text of the number.
func (n Number) String() string { return string(n) }

// Float64 returns the number as a float64.
func (n Number) Float64() (float64, error) {
    return strconv.ParseFloat(string(n), 64)
}

// Int64 returns the number as an int64.
func (n Number) Int64() (int64, error) {
    return strconv.ParseInt(string(n), 10, 64)
}
```

`json.Number`本质是字符串，反序列化的时候将JSON的数值先转成`json.Number`，其实是一种延迟处理的手段，待后续逻辑需要时候，再把`json.Number`转成`float64`或者`int64`。

对比其它语言，Golang对JSON反序列化处理真是易用性太差（“蛋疼”）。

JavaScript中所有的数值都是双精度浮点数（参考[这里](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Numbers_and_dates)），反序列化JSON的时候不用考虑数值类型匹配问题。这里多说两句，JSON的全名JavaScript Object Notation（从名字上就能看出和JavaScript的关系非常紧密），发明人是Douglas Crockford，如果你自称熟悉JavaScript而不知道[Douglas Crockford](http://www.infoq.com/cn/news/2010/02/qconbeijing2010-douglas)是谁，就像是自称是苹果粉丝却不知道乔布斯是谁。

C#语言的第三方JSON处理library [Json.NET](http://www.newtonsoft.com/json)反序列化JSON对数值的处理也比Golang要优雅的多：

```c#
using System;
using Newtonsoft.Json;

public class Program
{
    public static void Main()
    {
        string json = @"{
  'Name': 'Ethan',
  'FansCount': 121211,
  'Price': 99.99
}";

        Product m = JsonConvert.DeserializeObject<Product>(json);

        Console.WriteLine(m.FansCount);
        Console.WriteLine(m.FansCount.GetType().FullName);

        Console.WriteLine(m.Price);
        Console.WriteLine(m.Price.GetType().FullName);

    }
}

public class Product
{
    public string Name
    {
        get;
        set;
    }

    public object FansCount
    {
        get;
        set;
    }

    public object Price
    {
        get;
        set;
    }
}

// Output:
//      121211
//      System.Int64
//      99.99
//      System.Double
```

*点击[这里](https://dotnetfiddle.net/IrlMae)执行上面的程序*

`Json.NET`在反序列化的时候自动识别数值是浮点型还是整型，这一点对开发者非常友好。

## 3号坑：选择什么格式表示日期

JSON的规范中并没有日期类型，不同语言的library对日期序列化的处理也不完全一致：

Go语言：

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"
)

func main() {
    type Product struct {
        Name      string
        CreatedAt time.Time
    }
    pdt := Product{
        Name:      "Reds",
        CreatedAt: time.Now(),
    }
    b, err := json.Marshal(pdt)
    if err != nil {
        fmt.Println("error:", err)
    }
    os.Stdout.Write(b)
}
// Output
//      {"Name":"Reds","CreatedAt":"2016-06-27T07:40:54.69292134+08:00"}
```

JavaScript语言：

```sh
➜  ~ node
> var jo = { name: "ethan", createdAt: Date.now() };
undefined
> JSON.stringify(jo)
'{"name":"ethan","createdAt":1466984665633}'
```

C#语言：

```c#
using System;
using Newtonsoft.Json;

public class Program
{
    public static void Main()
    {
        Product product = new Product();
        product.Name = "Apple";
        product.CreatedAt = DateTime.Now;

        string json = JsonConvert.SerializeObject(product,
                            Newtonsoft.Json.Formatting.Indented,
                            new JsonSerializerSettings {
                                NullValueHandling = NullValueHandling.Ignore
                            });
        Console.WriteLine(json);
    }
}

public class Product
{
    public string Name
    {
        get;
        set;
    }

    public DateTime CreatedAt
    {
        get;
        set;
    }
}
// Output:
//      {
//        "Name": "Apple",
//        "CreatedAt": "2016-06-26T23:46:57.3244307+00:00"
//      }
```

Go的`encoding/json` package、C#的Json.NET默认把日期类型序列化成[ISO 8601标准](http://www.w3.org/TR/NOTE-datetime)的格式，JavaScript默认把`Date`序列化从1970年1月1日0点0分0秒的毫秒数。但JavaScript的[`dateObj.toISOString()`](https://developer.mozilla.org/zh-CN/docs/Web/JavaScript/Reference/Global_Objects/Date/toISOString)能够将日期类型转成ISO格式的字符串，[`Date.parse(dateString)`](https://developer.mozilla.org/zh-CN/docs/Web/JavaScript/Reference/Global_Objects/Date/parse)方法能够将ISO格式的日期字符串转成日期。

个人认为ISO格式的日期字符串可读性更好，但序列化和反序列化时的性能应该比整数更低。这一点从Go语言中`time.Time`的定义看出来。

```go
type Time struct {
    // sec gives the number of seconds elapsed since
    // January 1, year 1 00:00:00 UTC.
    sec int64

    // nsec specifies a non-negative nanosecond
    // offset within the second named by Seconds.
    // It must be in the range [0, 999999999].
    nsec int32

    // loc specifies the Location that should be used to
    // determine the minute, hour, month, day, and year
    // that correspond to this Time.
    // Only the zero Time has a nil Location.
    // In that case it is interpreted to mean UTC.
    loc *Location
}
```

具体选择哪种形式在JSON中表示日期，有如下几点需要注意：

- 选择标准格式。曾记得.NET Framework官方序列化JSON的方法中，会把日期转成如`"\/Date(1343660352227+0530)\/"`的专有格式，这样的专有格式对跨语言的访问特别不友好。
- 如果你倾向性能，可以使用整数。如果你倾向可读性，可以使用ISO字符串。
- 如果使用整数表示日期，而你的应用又是需要支持跨时区的，注意一定要是从`1970-1-1 00:00:00 UTC`开始计算的毫秒数，而不是当前时区的`1970-1-1 00:00:00`。


# 参考

文章：

- [package encoding/json in Go](http://docs.studygolang.com/pkg/encoding/json/)
- [http://docs.studygolang.com/src/encoding/json/example_test.go](http://docs.studygolang.com/src/encoding/json/example_test.go)
- [The Go Blog: JSON and Go](https://blog.golang.org/json-and-go)
- [Go by example: JSON](https://gobyexample.com/json)
- [JSON decoding in Go](http://attilaolah.eu/2013/11/29/json-decoding-in-go/)
- [go and json](https://eager.io/blog/go-and-json/)
- [Decode JSON Documents In Go](https://www.goinggo.net/2014/01/decode-json-documents-in-go.html)
- [ffjson: faster JSON serialization for Golang](https://journal.paul.querna.org/articles/2014/03/31/ffjson-faster-json-in-go/)
- [Serialization in Go](http://www.slideshare.net/albertstrasheim/serialization-in-go)

第三方类库：

- [ffjson](https://github.com/pquerna/ffjson): faster JSON serialization for Go
- [go-simplejson](https://github.com/bitly/go-simplejson): a Go package to interact with arbitrary JSON
- [Jason](https://github.com/antonholmquist/jason): Easy-to-use JSON Library for Go
- [easyjson](https://github.com/mailru/easyjson)
- [gabs](https://github.com/Jeffail/gabs)
- [jsonparser](https://github.com/buger/jsonparser)

工具：

- [JSON-to-Go](https://mholt.github.io/json-to-go/): instantly converts JSON into a Go type definition
