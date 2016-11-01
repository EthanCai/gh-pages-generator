---
title: 在Ubuntu Server 14.04上配置一个最小的MongoDB副本集
tags:
  - mongodb
categories:
  - 软件开发
date: 2016-11-01 12:59:40
---


# 在Ubuntu Server 14.04上安装MongoDB 3.2.6

- Import the public key used by the package management system
    - `sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv EA312927`
- Create a list file for MongoDB
    - `echo "deb http://repo.mongodb.org/apt/ubuntu trusty/mongodb-org/3.2 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-3.2.list`
- Reload local package database
    - `sudo apt-get update`
- Install the MongoDB packages
    - `sudo apt-get install -y mongodb-org=3.2.6 mongodb-org-server=3.2.6 mongodb-org-shell=3.2.6 mongodb-org-mongos=3.2.6 mongodb-org-tools=3.2.6`
- Pin a specific version of MongoDB
```sh
echo "mongodb-org hold" | sudo dpkg --set-selections
echo "mongodb-org-server hold" | sudo dpkg --set-selections
echo "mongodb-org-shell hold" | sudo dpkg --set-selections
echo "mongodb-org-mongos hold" | sudo dpkg --set-selections
echo "mongodb-org-tools hold" | sudo dpkg --set-selections
```
- 修改MongoDB的配置文件`/etc/mongod.conf`
    - 修改`net.bindIp`为`0.0.0.0`
    - 增加配置
    ```yaml
    storage:
      directoryPerDB: true
    ```
- 验证MongoDB是否成功安装
    - `sudo service mongod restart`
    - `mongo`
- Disable Transparent Huge Pages，参考[这里](https://docs.mongodb.com/manual/tutorial/transparent-huge-pages/)
    - Create the init.d script
    ```sh
    #!/bin/sh
    ### BEGIN INIT INFO
    # Provides:          disable-transparent-hugepages
    # Required-Start:    $local_fs
    # Required-Stop:
    # X-Start-Before:    mongod mongodb-mms-automation-agent
    # Default-Start:     2 3 4 5
    # Default-Stop:      0 1 6
    # Short-Description: Disable Linux transparent huge pages
    # Description:       Disable Linux transparent huge pages, to improve
    #                    database performance.
    ### END INIT INFO

    case $1 in
      start)
        if [ -d /sys/kernel/mm/transparent_hugepage ]; then
          thp_path=/sys/kernel/mm/transparent_hugepage
        elif [ -d /sys/kernel/mm/redhat_transparent_hugepage ]; then
          thp_path=/sys/kernel/mm/redhat_transparent_hugepage
        else
          return 0
        fi

        echo 'never' > ${thp_path}/enabled
        echo 'never' > ${thp_path}/defrag

        unset thp_path
        ;;
    esac
    ```
    - Make it executable
    ```sh
    sudo chmod 755 /etc/init.d/disable-transparent-hugepages
    ```
    - Configure your operating system to run it on boot
    ```sh
    sudo update-rc.d disable-transparent-hugepages defaults
    ```
    - 重启OS

# 配置MongoDB ReplicaSet副本集

副本集结构：

![](https://docs.mongodb.com/manual/_images/replica-set-primary-with-secondary-and-arbiter.png)

- 2数据节点，1个仲裁节点

## 配置步骤

- 准备2台高配ec2（假设为A、B）和1台低配ec2（假设为C）
- 在A、B、C上参考上一节的步骤安装MongoDB
- 在A的Shell中执行`mongo`命令，然后创建超级管理员
```javascript
$ mongo
> admin = db.getSiblingDB("admin");
> admin.createUser({ user: "ethan", pwd: "{ethan的密码}", roles: [{ role: "root", db: "admin" }] });
```
- 准备keyfile
    - 生成keyfile
        - `openssl rand -base64 755 > rs0.key`
    - 上传`rs0.key`到A、B、C的`/etc`目录
    - 修改`rs0.key`的权限和所有者
        - `chmod 400 rs0.key`
        - `chown mongodb:mongodb /etc/rs0.key`
- 修改A、B的配置文件`/etc/mongod.conf`
    - 增加配置
    ```yaml
    security:
      keyFile: "/etc/rs0.key"
      authorization: enabled

    replication:
      replSetName: rs0
    ```
- 修改C的配置文件`/etc/mongod.conf`
    - 增加配置
    ```yaml
    security:
      keyFile: "/etc/rs0.key"
      authorization: enabled

    replication:
      replSetName: rs0
    ```
    - 修改配置
    ```yaml
    storage:
      journal:
        enabled: false
    ```
- 重启A、B、C上的`mongod`实例
- 配置集群
    - 连接A上的`mongod`实例
    ```sh
    $ mongo -u ethan -p {ethan的密码} --authenticationDatabase admin
    ```
    - 通过下面的命令配置ReplicaSet：
    ```javascript
    > rs.initiate()
    > cfg = rs.conf()
    > cfg.members[0].host = "{A的IP}:27017"
    > rs.reconfig(cfg)
    > rs.add("{B的IP}")
    > rs.addArb("{C的IP}")

    // 等待几秒...
    > rs.status()  // 检查副本集状态
    ```
- 为副本集客户端创建访问账户
```javascript
> use {dbname};
> db.createUser({ user: "{账户名}", pwd: "{账户密码}", roles: [{ role: "readWrite", db: "{dbname}" }] });
```

# 参考

- [Install MongoDB on Ubuntu](https://docs.mongodb.com/manual/tutorial/install-mongodb-on-ubuntu/)， 注意：
    - 安装过程需要指定MongoDB的安装版本
    - 锁定MongoDB的安装版本，避免执行`apt-get`升级命令的时候连带升级MongoDB
- [Enforce Keyfile Access Control in a Replica Set](https://docs.mongodb.com/manual/tutorial/enforce-keyfile-access-control-in-existing-replica-set/)
    - Security between members of the replica set using Internal Authentication
    - Security between connecting clients and the replica set using User Access Controls
- [MongoDB configuration file options](https://docs.mongodb.com/manual/reference/configuration-options/#core-options)
