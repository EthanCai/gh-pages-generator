---
title: 基于Docker Swarm，构建你自己的开发容器集群环境
tags:
  - docker
categories:
  - 软件开发
---

# 简介

微服务和容器集群技术，是天生的好搭档。微服务让系统的模块划分更细，而容器集群技术降低了这种模块细分带来部署复杂度。

目前我司的服务端程序主要采用微服务架构，测试、生产环境的服务端程序都部署在基于Docker Swarm搭建的容器集群中。而本地开发微服务，相比以前开发单块应用，开发、运行、调试的复杂度都增加了不少。

本文介绍一种方法，通过开源容器技术搭建本地微服务运行环境，降低开发成本，提高开发效率。

# 总体结构


# 搭建步骤


# 如何运行应用程序


# 如何收集查看程序日志


# 参考

- Docker Engine
  - Install Docker
    - Mac上安装Docker: https://docs.docker.com/engine/installation/
    - CentOS上安装Docker Engine:
      - https://docs.docker.com/engine/installation/linux/centos/
      - https://docs.docker.com/v1.13/engine/installation/linux/centos/
    - Config Registry Mirror: https://yq.aliyun.com/articles/29941
  - User Guide
    - Build and Manage Image
      - Best Practices: https://docs.docker.com/engine/userguide/eng-image/dockerfile_best-practices/
    - Network configuration
      - Docker container networking: https://docs.docker.com/engine/userguide/networking/
  - Admin Guide
    - Limit a container's resources: https://docs.docker.com/engine/admin/resource_constraints/
    - Logging: https://docs.docker.com/engine/admin/logging/view_container_logs/
    - Using Ansible: https://docs.docker.com/engine/admin/ansible/
    - Runtime Metrics: https://docs.docker.com/engine/admin/runmetrics/
- Docker Registry
  - https://docs.docker.com/v1.13/registry/deploying/
  - https://hub.docker.com/r/h3nrik/registry-ldap-auth/
  - https://github.com/kwk/docker-registry-setup
  - 生成签名证书：
    - http://www.akadia.com/services/ssh_test_certificate.html
    - https://my.oschina.net/aiguozhe/blog/121764
- Docker Swarm
  - Guides - Manage a Swarm: https://docs.docker.com/engine/swarm/
  - Superseded products and tools: https://docs.docker.com/swarm/overview/
- Docker Compose
  - Overview of Docker Compose: https://docs.docker.com/compose/overview/
- Docker Machine
  - Overview: https://docs.docker.com/machine/overview/
- Docker Security
