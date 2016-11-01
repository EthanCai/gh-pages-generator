---
title: 应对业务增长，对MongoDB集群进行扩容的一种路径
tags:
  - mongodb
categories:
  - 软件开发
date: 2016-11-01 13:21:18
---


> 本文以用户`feed`作为案例

# MongoDB集群结构

## 数据量较小

采用MongoDB三节点副本集的方式构造集群

![replica-set-primary-with-secondary-and-arbiter](https://cloud.githubusercontent.com/assets/286882/15322892/e004e45c-1c71-11e6-8255-c938c1c8e12a.png)

## 数据量较大

使用`sharding`方式扩展单个集群的容量

![shardreplica](https://cloud.githubusercontent.com/assets/286882/15322822/80d4167e-1c71-11e6-996d-725800dd5531.jpg)

## 数据量非常大

不同时期的Feed数据写入到不同的MongoDB Cluster中，避免单个MongoDB集群规模过大带来各种运维上的问题

![multiple_cluster](https://cloud.githubusercontent.com/assets/286882/15310929/c702c694-1c27-11e6-9d4e-f4fbfa4a5abb.jpg)

- 每个MongoDB Cluster保存的数据包括：
    - 元数据
        - 时间范围：指定当前cluster保存那一段时间的feed信息
    - Feed数据
        - 使用一个collection保存所有用户的feed
        - 这个collection的根据用户的user_id进行分片，适应写、读扩容场景
- 客户端程序根据MongoDB Cluster的元数据将收到的Feed消息写入到对应的MongoDB Cluster
- 客户端程序启动时从所有的MongoDB Cluster中加载元数据

# Feed DB的结构

`metadata`集合

```json
{
    "_id": "cluster_1",
    "name": "cluster_1",
    "start_date": new Date("2016-05-01"),
    "end_date": new Date("2016-08-01"),
    "creator_name": "ethan",
    "created_at": new Date("2016-05-01 00:00:00")
}
```

`feed`集合

```json
{
    "_id": ObjectId(""),
    "data_key": "e6755cfae343b6719cc2121e888b0a41",
    "receiver_id": 1000386,
    "sender_id": 1000765,
    "event_time": new Date("2016-05-01 10:00:00"),
    "type": 1,
    "data": {
        "fabula_id": 1000983
    }
}
```

`feed.data_key`用于根据业务对象查找对应`feed`记录的标识，主要用于删除场景，生成算法如下：

- `feed.data_key = MMH3Hash("fabula_" + $fabulaId)`

`feed._id`的生成算法：

- 同`ObjectID`的生成算法，包含`time`, `machine identifier`, `process id`, `counter`四部分，使用`feed.event_time`作为第一部分
- `ObjectID`生成算法参考: [https://github.com/go-mgo/mgo/blob/v2/bson/bson.go](https://github.com/go-mgo/mgo/blob/v2/bson/bson.go)

# 参考

- [陌陌：日请求量过亿，谈陌陌的Feed服务优化之路](http://mp.weixin.qq.com/s?__biz=MzA5Nzc4OTA1Mw==&mid=2659597071&idx=1&sn=cd8df9f8c52dfbfb54e65adbe19fae27&scene=0#wechat_redirect)
- [几个大型网站的Feeds(Timeline)设计简单对比](http://www.tuicool.com/articles/BJRJja)
- [新浪微博：大数据时代的feed流架构](http://www.infoq.com/cn/presentations/feed-stream-architecture-in-big-data-era)
- [新浪微博：Feed架构-我们做错了什么](http://itindex.net/detail/52175-feed-%E6%9E%B6%E6%9E%84)
- [新浪微博：Feed消息队列架构分析](http://timyang.net/data/feed-message-queue/)
- [Pinterest：Pinterest的Feed架构与算法](http://ju.outofmemory.cn/entry/223817)
- [Pinterest：Building a smarter home feed](https://engineering.pinterest.com/blog/building-smarter-home-feed)
- [Pinterest：Building a scalable and available home feed](https://engineering.pinterest.com/blog/building-scalable-and-available-home-feed)
- [Pinterest：Pinterest 的 Smart Feed 架构与算法](https://mp.weixin.qq.com/s?__biz=MzA4OTk5OTQzMg==&mid=2449231037&idx=1&sn=c2fc8a7d2832ea109e2abe4b773ff1f5&scene=1&srcid=0509fzQ02Jubcqnw7WPzp6IO)
- [Pinterest：Pinnability: Machine learning in the home feed](https://engineering.pinterest.com/blog/pinnability-machine-learning-home-feed)
