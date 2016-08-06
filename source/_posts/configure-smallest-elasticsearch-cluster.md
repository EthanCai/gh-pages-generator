---
title: 学习Elasticsearch之4：配置一个3节点Elasticsearch集群(不区分主节点和数据节点)
tags:
  - elasticsearch
categories:
  - 软件开发
date: 2016-08-06 09:46:38
---


{% asset_img es-cluster.jpg %}

# 安装版本

Elasticsearch（通过`apt`方式安装）:

- Elasticsearch 2.3.2

Jar插件:

- [elasticsearch-analysis-ik](https://github.com/medcl/elasticsearch-analysis-ik): 1.9.2
- [elasticsearch-analysis-pinyin](https://github.com/medcl/elasticsearch-analysis-pinyin): 1.7.2
- [elasticsearch-analysis-stconvert](https://github.com/medcl/elasticsearch-analysis-stconvert): 1.8.2

Site插件:

- [elasticsearch-head](https://github.com/mobz/elasticsearch-head): 最新版

# 各节点服务器

- Node1
    - 集群名：`search-1`
    - 节点名：`node-1`
    - 内网IP: `192.168.31.171`
- Node2
    - 集群名：`search-1`
    - 节点名：`node-2`
    - 内网IP: `192.168.31.221`
- Node3
    - 集群名：`search-1`
    - 节点名：`node-3`
    - 内网IP: `192.168.31.154`

# 各节点安装的插件

| Node  | Installed Plugins           |
|:------|:----------------------------|
| Node1 | ik, pinyin, stconvert, head |
| Node1 | ik, pinyin, stconvert       |
| Node1 | ik, pinyin, stconvert       |

# 各节点配置

三节点的配置：

- Node 1

```yaml
cluster.name: search-1
node.name: node-1

node.master: true
node.data: true

index.number_of_shards: 3
index.number_of_replicas: 1

network.host: 0.0.0.0                           # 绑定本机所有端口

discovery.zen.ping.multicast.enabled: false     # 禁止多播
discovery.zen.minimum_master_nodes: 2           # 配置最少节点数量，防止脑裂
discovery.zen.ping.unicast.hosts: ['192.168.31.171', '192.168.31.221', '192.168.31.154']
```

- Node 2

```yaml
cluster.name: search-1
node.name: node-2

node.master: true
node.data: true

index.number_of_shards: 3
index.number_of_replicas: 1

network.host: 0.0.0.0                           # 绑定本机所有端口

discovery.zen.ping.multicast.enabled: false     # 禁止多播
discovery.zen.minimum_master_nodes: 2           # 配置最少节点数量，防止脑裂
discovery.zen.ping.unicast.hosts: ['192.168.31.171', '192.168.31.221', '192.168.31.154']
```

- Node 3

```yaml
cluster.name: search-1
node.name: node-3

node.master: true
node.data: true

index.number_of_shards: 3
index.number_of_replicas: 1

network.host: 0.0.0.0                           # 绑定本机所有端口

discovery.zen.ping.multicast.enabled: false     # 禁止多播
discovery.zen.minimum_master_nodes: 2           # 配置最少节点数量，防止脑裂
discovery.zen.ping.unicast.hosts: ['192.168.31.171', '192.168.31.221', '192.168.31.154']
```

# 常用操作

上传插件

```bash
$ scp -r -i ~/dev.pem elasticsearch-analysis-ik-1.9.2 ubuntu@192.168.31.171:~/
$ scp -r -i ~/dev.pem elasticsearch-analysis-pinyin-1.7.2 ubuntu@192.168.31.171:~/
$ scp -r -i ~/dev.pem elasticsearch-analysis-stconvert-1.8.2 ubuntu@192.168.31.171:~/

$ scp -r -i ~/dev.pem elasticsearch-analysis-ik-1.9.2 ubuntu@192.168.31.221:~/
$ scp -r -i ~/dev.pem elasticsearch-analysis-pinyin-1.7.2 ubuntu@192.168.31.221:~/
$ scp -r -i ~/dev.pem elasticsearch-analysis-stconvert-1.8.2 ubuntu@192.168.31.221:~/

$ scp -r -i ~/dev.pem elasticsearch-analysis-ik-1.9.2 ubuntu@192.168.31.154:~/
$ scp -r -i ~/dev.pem elasticsearch-analysis-pinyin-1.7.2 ubuntu@192.168.31.154:~/
$ scp -r -i ~/dev.pem elasticsearch-analysis-stconvert-1.8.2 ubuntu@192.168.31.154:~/

```

拷贝插件到`/usr/share/elasticsearch/plugins`目录

```bash
$ mv elasticsearch-analysis-ik-1.9.2 /usr/share/elasticsearch/plugins/
$ mv elasticsearch-analysis-pinyin-1.7.2 /usr/share/elasticsearch/plugins/
$ mv elasticsearch-analysis-stconvert-1.8.2 /usr/share/elasticsearch/plugins/

$ chown -R elasticsearch:elasticsearch elasticsearch-analysis-ik-1.9.2/             # 修改插件访问权限，允许elasticsearch服务访问插件
$ chown -R elasticsearch:elasticsearch elasticsearch-analysis-pinyin-1.7.2/         # 修改插件访问权限，允许elasticsearch服务访问插件
$ chown -R elasticsearch:elasticsearch elasticsearch-analysis-stconvert-1.8.2/      # 修改插件访问权限，允许elasticsearch服务访问插件
```
