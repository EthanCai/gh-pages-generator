---
title: 学习Elasticsearch之1：如何在Ubuntu 14.04上安装Oracle JDK
tags:
  - jdk
  - elasticsearch
categories:
  - 软件开发
date: 2016-08-06 00:29:11
---

{% asset_img oracle-jdk8-on-Ubuntu.png %}

# 如何选择JDK版本

[Elasticsearch 2.3官方文档](https://www.elastic.co/guide/en/elasticsearch/reference/current/_installation.html)中推荐的JDK版本是“Oracle JDK version 1.8.0_73”，最低Java 7.

# Oracle的JDK的安装方式

Oracle JDK的安装方式有两种方式

- 手动安装，参考[How to Install Oracle Java JDK on Ubuntu Linux](http://www.wikihow.com/Install-Oracle-Java-JDK-on-Ubuntu-Linux)
- 通过`apt`命令安装，参考[Installing the oracle JDK](https://www.elastic.co/guide/en/elasticsearch/reference/current/setup-service.html#_installing_the_oracle_jdk)

# 手动安装步骤

- 从[这里](http://www.oracle.com/technetwork/java/javase/downloads/jdk8-downloads-2133151.html)下载JDK 8u91
- 拷贝下载安装包到`~/Download/`目录下：'scp ~/Downloads/jdk-8u91-linux-x64.tar.gz ethancai@172.16.210.149:~/Downloads/'
- SSH远程连接Ubuntu Server，`ssh ethancai@172.16.210.149`
- 删除OpenJDK
    - `sudo apt-get purge openjdk-\*`
- 拷贝安装文件到安装目录
    - `sudo mkdir -p /usr/local/java`
    - `sudo cp -r ~/Downloads/jdk-8u91-linux-x64.tar.gz /usr/local/java/`
- 解压
    - `cd /usr/local/java`
    - `sudo tar -xvzf jdk-8u91-linux-x64.tar.gz`
- 修改登录式Shell的全局启动配置
    - `sudo vim /etc/profile`，然后将如下内容拷贝到文件中
    ```
    JAVA_HOME=/usr/local/java/jdk1.8.0_91
    JRE_HOME=$JAVA_HOME/jre
    PATH=$PATH:$JAVA_HOME/bin:$JRE_HOME/bin
    export JAVA_HOME
    export JRE_HOME
    export PATH
    ```
- Notify the system that JRE/JDK/Java Web Start is available for use
    - `sudo update-alternatives --install "/usr/bin/java" "java" "/usr/local/java/jdk1.8.0_91/bin/java" 1`
    - `sudo update-alternatives --install "/usr/bin/javac" "javac" "/usr/local/java/jdk1.8.0_91/bin/javac" 1`
    - `sudo update-alternatives --install "/usr/bin/javaws" "javaws" "/usr/local/java/jdk1.8.0_91/bin/javaws" 1`
- Reload your system wide PATH
    - `source /etc/profile`

# 通过`apt`命令安装步骤

- 此命令目前能安装的jdk8最新版本是`jdk-8u77-linux-x64`
- `sudo apt-get purge openjdk-\*`
- `sudo add-apt-repository ppa:webupd8team/java`
- `sudo apt-get update`
- `sudo apt-get install oracle-java8-installer`

# 验证安装是否成功

- `java -version`
- `javac -version`
- 重启`sudo shutdown -r 0`
