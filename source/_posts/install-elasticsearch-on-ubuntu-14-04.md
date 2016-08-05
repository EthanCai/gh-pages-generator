---
title: 学习Elasticsearch之2：在Ubuntu 14.04上安装Elasticsearch
tags:
  - jdk
  - elasticsearch
categories:
  - 软件开发
date: 2016-08-06 00:40:48
---

{% asset_img logo-elastic.png %}


#  Elasticsearch的安装有两种方式

- 通过`tar.gz`压缩包方式安装，参考[ElasticSearch Reference 2.3 - Installation](https://www.elastic.co/guide/en/elasticsearch/reference/current/_installation.html)
- 通过`apt`命令安装，参考[ElasticSearch Reference 2.3 - Repositories](https://www.elastic.co/guide/en/elasticsearch/reference/current/setup-repositories.html)

## 通过`tar.gz`压缩包安装步骤

- SSH远程连接Ubuntu Server，`ssh ethancai@###.###.###.###`
- 下载Elasticsearch
    - `mkdir ~/Downloads && cd ~/Downloads && curl -L -O https://download.elastic.co/elasticsearch/release/org/elasticsearch/distribution/tar/elasticsearch/2.3.1/elasticsearch-2.3.1.tar.gz`
- 解压
    - `tar -xzvf elasticsearch-2.3.1.tar.gz && sudo mv elasticsearch-2.3.1 /var/`
- 运行
    - `cd /var/elasticsearch-2.3.1/bin`
    - `./elasticsearch`，或者以指定`cluster`和`node`名称的方式运行`./elasticsearch --cluster.name my_cluster_name --node.name my_node_name`

## 通过`apt`命令安装

- Download and install the Public Signing Key
    - `wget -qO - https://packages.elastic.co/GPG-KEY-elasticsearch | sudo apt-key add -`
- Save the repository definition
    - `echo "deb http://packages.elastic.co/elasticsearch/2.x/debian stable main" | sudo tee -a /etc/apt/sources.list.d/elasticsearch-2.x.list`
- Install elasticsearch, 注意这里需要指定安装版本
    - `sudo apt-get update && sudo apt-get install elasticsearch=2.3.2`
- Configure Elasticsearch to automatically start during bootup
    - `sudo update-rc.d elasticsearch defaults 95 10`

## 验证安装是否成功

By default, Elasticsearch uses port 9200 to provide access to its REST API. 你可以通过下面命令验证是否启动成功：

- `curl 'localhost:9200/_cat/health?v'`
- `curl 'localhost:9200/_cat/nodes?v'`
- 验证后记得重启，`sudo shutdown -r 0`

# 默认目录

不同安装方式，默认的目录位置不一样，参考[ElasticSearch Reference 2.3 - Directory Layout](https://www.elastic.co/guide/en/elasticsearch/reference/current/setup-dir-layout.html)

通过`tar.gz`方式安装的默认目录如下:

| Type    | Description                                                                 | Location Debian/Ubuntu          |
|:--------|:----------------------------------------------------------------------------|:--------------------------------|
| home    | Home of elasticsearch installation.                                         | `{extract.path}`                |
| bin     | Binary scripts including elasticsearch to start a node.                     | `{extract.path}/bin`            |
| conf    | Configuration files elasticsearch.yml and logging.yml.                      | `{extract.path}/config`         |
| data    | The location of the data files of each index / shard allocated on the node. | `{extract.path}/data`           |
| logs    | Log files location                                                          | `{extract.path}/logs`           |
| plugins | Plugin files location. Each plugin will be contained in a subdirectory.     | `{extract.path}/plugins`        |
| repo    | Shared file system repository locations.                                    | Not configured                  |
| script  | Location of script files.                                                   | `{extract.path}/config/scripts` |

通过`apt`方式安装的默认目录如下：

| Type    | Description                                                                                                          | Location Debian/Ubuntu             |
|:--------|:---------------------------------------------------------------------------------------------------------------------|:-----------------------------------|
| home    | Home of elasticsearch installation.                                                                                  | `/usr/share/elasticsearch`         |
| bin     | Binary scripts including elasticsearch to start a node.                                                              | `/usr/share/elasticsearch/bin`     |
| conf    | Configuration files elasticsearch.yml and logging.yml.                                                               | `/etc/elasticsearch`               |
| conf    | Environment variables including heap size, file descriptors. **There isn't this folder when install using `tar.gz`** | `/etc/default/elasticsearch`       |
| data    | The location of the data files of each index / shard allocated on the node.                                          | `/var/lib/elasticsearch`           |
| logs    | Log files location                                                                                                   | `/var/log/elasticsearch`           |
| plugins | Plugin files location. Each plugin will be contained in a subdirectory.                                              | `/usr/share/elasticsearch/plugins` |
| repo    | Shared file system repository locations.                                                                             | Not configured                     |
| script  | Location of script files.                                                                                            | `/etc/elasticsearch/scripts`       |

# 安装Elasticsearch插件

Plugins are a way to enhance the core Elasticsearch functionality in a custom manner. They range from adding custom mapping types, custom analyzers, native scripts, custom discovery and more.

There are three types of plugins:

- `Java Plugins`: These plugins contain only JAR files, and must be installed on every node in the cluster. **After installation, each node must be restarted before the plugin becomes visible.**
- `Site Plugins`: These plugins contain static web content like Javascript, HTML, and CSS files, that can be served directly from Elasticsearch. Site plugins may only need to be installed on one node, and do not require a restart to become visible.
    - The content of site plugins is accessible via a URL like: `http://yournode:9200/_plugin/[plugin name]`
- `Mixed Plugins`: Mixed plugins contain both JAR files and web content.

## 安装`Site Plugins`

```bash
$ sudo su -
$ cd /

# 查看已安装插件
$ /usr/share/elasticsearch/bin/plugin list

# 安装elasticsearch-head，访问http://localhost:9200/_plugin/head
$ /usr/share/elasticsearch/bin/plugin install mobz/elasticsearch-head

# 安装elasticsearch-kopf，访问http://localhost:9200/_plugin/kopf
$ /usr/share/elasticsearch/bin/plugin install lmenezes/elasticsearch-kopf

# 安装elastichq，访问http://localhost:9200/_plugin/hq
$ /usr/share/elasticsearch/bin/plugin install royrusso/elasticsearch-HQ

# 安装elasticsearch-inquisitor，访问http://localhost:9200/_plugin/elasticsearch-inquisitor
$ /usr/share/elasticsearch/bin/plugin install polyfractal/elasticsearch-inquisitor
```

## 安装`Jar Plugins`

### 安装`elasticsearch-analysis-ik`

安装前请注意和Elasticsearch的适配版本，安装[elasticsearch-analysis-ik](https://github.com/medcl/elasticsearch-analysis-ik)有两种方式：

- 通过源代码手动编译安装
    - 请先[安装apache maven](http://maven.apache.org/install.html)
    - 下载`Apache Maven`的源代码
        - `git clone https://github.com/medcl/elasticsearch-analysis-ik.git`
    - 编译打包
        - `cd [elasticsearch-analysis-ik所在目录]`
        - `git checkout v1.9.2`
        - `mvn package`
    - 解压编译生成的打包文件`target/releases/elasticsearch-analysis-ik-{version}.zip`，并拷贝解压后的文件到Elasticsearch的plugins目录
        - `unzip elasticsearch-analysis-ik-{version}.zip -d elasticsearch-analysis-ik-{version}`
        - `scp -r elasticsearch-analysis-ik-{version} ubuntu@xxx.xxx.xxx.xxx:/var/tmp/`
        - `cp -R /var/tmp/elasticsearch-analysis-ik-{version} /usr/share/elasticsearch/plugins`
        - `chown -R elasticsearch:elasticsearch /usr/share/elasticsearch/plugins/elasticsearch-analysis-ik-{version}`
    - 重启elasticsearch服务
        - `service elasticsearch restart`
- 直接[下载安装程序包](https://github.com/medcl/elasticsearch-analysis-ik/releases)
    - 下载指定版本的打包文件
    - 后续步骤同上

注意：

- v2.2.1版本以后安装`elasticsearch-analysis-ik`并不需要修改`elasticsearch.yml`配置文件

### 安装`elasticsearch-analysis-pinyin`

安装前请注意和Elasticsearch的适配版本, 安装[elasticsearch-analysis-pinyin](https://github.com/medcl/elasticsearch-analysis-pinyin)有两种方式：

- 通过源代码手动编译安装
    - 具体操作参考`elasticsearch-analysis-ik`的安装步骤
- 直接[下载安装程序包](https://github.com/medcl/elasticsearch-analysis-pinyin/releases)

### 安装`elasticsearch-analysis-stconvert`

安装前请注意和Elasticsearch的适配版本, 安装[elasticsearch-analysis-stconvert](https://github.com/medcl/elasticsearch-analysis-stconvert)有两种方式：

- 通过源代码手动编译安装
    - 具体操作参考`elasticsearch-analysis-ik`的安装步骤
- 直接[下载安装程序包](https://github.com/medcl/elasticsearch-analysis-stconvert/releases)

# 总结

建议Oracle JDK和Elasticsearch均使用`apt`安装方式安装，原因如下：

- 安装方便
- 默认目录的配置更规范
- 和Elasticsearch官方的Docker镜像的目录结构一致，未来切换到Docker方式部署障碍少，参考[Dockerfile](https://github.com/docker-library/elasticsearch/blob/8267f6c29f06373373b4379473291d3082728cc0/2.3/Dockerfile)
