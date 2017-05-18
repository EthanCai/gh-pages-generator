---
title: 1 - Install PHP | Study PHP and Symfony on Mac OS X
tags:
  - php
  - symfony
categories:
  - 软件开发
date: 2017-05-17 20:00:00
---


# Introduction

My mac's OS version is macOS Sierra Version 10.12.4

# Environment Requirements

- install Xcode: `xcode-select --install`
- install homebrew

# Install PHP

```sh
# 快速安装php，参考 https://github.com/Homebrew/homebrew-php
> brew install brew-php-switcher
> brew install php56

> which php
/usr/local/bin/php

> which php-fpm
lrwxr-xr-x  1 leeco  admin    32B May 17 15:08 /usr/local/bin/php -> ../Cellar/php56/5.6.30_6/bin/php

# 配置 /private/etc/php.ini
> sudo vim /usr/local/etc/php/5.6/php.ini
# date.timezone = Asia/Shanghai
> php -i | grep timezone
Default timezone => Asia/Shanghai
date.timezone => Asia/Shanghai => Asia/Shanghai
```

如果要安装多个版本的php：

```sh
# 用于在不同版本的php之间切换，参考 https://github.com/philcook/brew-php-switcher
> brew unlink php56
> brew install php71

> brew-php-switcher 56 -s # 切回 php 5.6
```

安装 PEAR 和 PECL:

```sh
# 参考 https://jason.pureconcepts.net/2012/10/install-pear-pecl-mac-os-x/
> curl -O http://pear.php.net/go-pear.phar
> sudo php -d detect_unicode=0 go-pear.phar
# Press return

> brew-php-switcher 56 -s # 在 /usr/local/sbin 目录下创建link，指向 /usr/local/Cellar/php56/5.6.30_6/bin 下的命令（包含新安装的pear/pecl)
```

# Install PHP Extension

## Install `intl`

```sh
# 安装 intl 扩展，参考 http://note.rpsh.net/posts/2015/10/07/installing-php-intl-extension-os-x-el-capitan/
> sudo pear channel-update pear.php.net
> sudo pecl channel-update pecl.php.net
> sudo pear upgrade-all

> brew install autoconf
> brew install icu4c
> sudo pecl install intl
# icu4c path: /usr/local/opt/icu4c/

> vim /usr/local/etc/php/5.6/php.ini
# 在最后一行添加 extension=intl.so

> php -m | grep intl # 检查是否安装成功，正常会返回 intl
```

## Install `OPcache`

```sh
> brew install php56-opcache

> brew info
...
To finish installing opcache for PHP 5.6:
  * /usr/local/etc/php/5.6/conf.d/ext-opcache.ini was created,
    do not forget to remove it upon extension removal.
  * Validate installation via one of the following methods:
...

> php -m | grep OPcache  # 检查 OPcache 是否已生效
```

# Install Composer

```sh
> brew install composer
```

# Install Xdebug

```sh
> brew install php56-xdebug
...
To finish installing xdebug for PHP 5.6:
  * /usr/local/etc/php/5.6/conf.d/ext-xdebug.ini was created,
    do not forget to remove it upon extension removal.
  * Validate installation via one of the following methods:
...

> brew install xdebug-osx
...
Signature:
  xdebug-toggle <on | off> [--no-server-restart]

Usage:
  xdebug-toggle         # outputs the current status
  xdebug-toggle on      # enables xdebug
  xdebug-toggle off     # disables xdebug

Options:
  --no-server-restart   # toggles xdebug without restarting apache or php-fpm
...
```

# 参考

- [macOS 10.12 Sierra Apache Setup: MySQL, APC & More...](https://getgrav.org/blog/macos-sierra-apache-mysql-vhost-apc)
- [Cannot find libz when install php56](https://github.com/Homebrew/homebrew-php/issues/1946)
