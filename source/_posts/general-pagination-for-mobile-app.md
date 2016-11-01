---
title: 支持游标和偏移量的通用翻页机制
tags:
  - API设计
categories:
  - 软件开发
date: 2016-11-01 15:54:00
---


# 前言

对于大多数mobile App，当 App 发出请求时，通常不会在单个响应中收到该请求的全部结果，而是以分片的方式获取部分结果。

随着业务需求的变化，某些情况下 App 的翻页机制可能会调整。一般我们通过 App 重新发版，服务端和客户端同步调整分页机制来完成调整。而本文提供了一种通用协议，支持仅通过服务端发版来调整 App 的分页机制。

# 常见分页机制

## 基于游标的分页

游标是指标记数据列表中特定项目的一个随机字符串。该项目未被删除时，游标将始终指向列表的相同部分，项目被删除时游标即失效。因此，客户端应用不应存储任何旧的游标，也不能假定它们仍然有效。

Request的结构一般如下：

```
https://api.sample.com/v3/users/?limit=30&before=NDMyNzQyODI3OTQw
https://api.sample.com/v3/users/?limit=30&after=MTAxNTExOTQ1MjAwNzI5NDE=
```

参数说明：

- **limit**：每个页面返回的单独对象的数量。请注意这是上限，如果数据列表中没有足够的剩余对象，那么返回的数量将小于这个数。为防止客户端传入过大的值，某些列表的 `limit` 值将设置上限。
- **before**：向后查询的起始位置。
- **after**：向前查询的起始位置。

Response结构一般如下：

```json
{
  "rows": [
     ... Endpoint data is here
  ],
  "paging": {
    "cursors": {
      "top": "MTAxNTExOTQ1MjAwNzI5NDE=",
      "last": "NDMyNzQyODI3OTQw"
    },
    "previous": "NDMyNzQyODI3OTQw",
    "next": "MTAxNTExOTQ1MjAwNzI5NDE="
  }
}
```

参数说明：

- **rows**：如果当前页没有数据，或者根据过滤规则（比如隐私）当前页所有数据都被过滤掉，返回空的数组。客户端程序不能根据`rows`是否为空数组来判断，是否已经滚动到列表的末尾，而应根据下面的`next`字段是否有值来决定是否滚动到了列表尾部。
- **top**：已返回的数据页面开头的游标。
- **last**：已返回的数据页面末尾的游标。
- **previous**：上一页数据的 API 端点。如果是`null`或者没有此字段，则表示返回的是第一页数据。
- **next**：下一页数据的 API 端点。如果是`null`或者没有此字段，则表示返回的是最后一页数据。

## 基于偏移量的分页

Request的结构一般如下：

```
https://api.sample.com/v3/users/?limit=30&offset=30
```

参数说明：

- **limit**：每个页面返回的单独对象的数量。
- **offset**：偏移量，查询的起始位置。

> 一般情况下，还会包含其他查询条件，比如根据关键字查找姓名和关键字匹配的用户

Response结构一般如下：

```json
{
  "rows": [
     ... Endpoint data is here
  ],
  "count": 10765
}
```

参数说明：

- **count**：符合查询条件的总记录数。客户端根据`offset`和`count`判断是否已经滚动到列表尾部。

> 注意，如果正分页的项目列表添加了新的对象，后续基于偏移量的查询的内容都将发生更改。

# 支持游标和偏移量的通用分页机制

每个API视场景需要实现部分规范（比如仅实现向后翻页，不实现向前翻页），没有实现的行为统一返回一个特定错误码 "not supported"

## HTTP Request

- `limit`: 必填项；期望返回的记录数量；整数类型；必须大于等于0
- `after`: 可选项；字符串类型；表示查询从`after`指向的记录之后（不包括`after`指向的当前记录）的`limit`条记录
- `before`: 可选项；字符串类型；表示查询从`before`指向的记录之前（不包括`before`指向的当前记录）的`limit`条记录
- 备注：
    - Request中`before`、`after`不能并存
    - 如果Request中没有`before`和`after`，视为从结果集起始位置向后查询
    - **MySQL和MongoDB均不支持基于游标位置的向前查询**，如需支持需要在程序逻辑中实现

## HTTP Response

```json
{
    "code": 0,
    "result": {
        "rows": [
            //"... Endpoint data is here"
        ],
        "paging": {
            "cursors": {
                "top": "MTAxNTExOTQ1MjAwNzI5NDE=" or "19",
                "last": "MTAxNTExOTQ1MjAwNzI5NDE=" or "19"
            },
            "previous": "MTAxNTExOTQ1MjAwNzI5NDE=" or "19",
            "next": "MTAxNTExOTQ1MjAwNzI5NDE=" or "19",
            "count": 1087
        }
    }
}
```

- `paging.cursors.top`: 必填项；字符串或者null；指向已返回的数据页面开头的游标
- `paging.cursors.last`: 必填项；字符串或者null；指向已返回的数据页面末尾的游标
- `paging.previous`: 必填项；字符串或者null；查询前一页数据的末尾位置，Request中将此值赋给`before`，为null时，表示没有前一页
- `paging.next`: 必填项；字符串或者null；查询后一页数据的起始位置，Request中将此值赋给`after`，为null时，表示没有后一页
- `paging.count`: 可选项；整型；结果集总数
- 备注：
    - `rows`为空结果集时：`paging.cursors.top`、`paging.cursors.last`为`null`，`paging.previous`、`paging.next`不一定为`null`
    - 仅一条结果集是，`paging.cursors.top`、`paging.cursors.last`相同

# 参考

- [twitter API - GET statuses/user_timeline](https://dev.twitter.com/rest/reference/get/statuses/user_timeline)
