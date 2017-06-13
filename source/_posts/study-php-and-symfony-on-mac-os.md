---
title: 1 - Install PHP | Study PHP
tags:
  - php
categories:
  - 软件开发
date: 2017-05-17 20:00:00
---

# Environment

- macOS Sierra Version 10.12.4
- Xcode: `xcode-select --install`
- `Homebrew`

# Install PHP

安装PHP：

```sh
# 快速安装php，参考 https://github.com/Homebrew/homebrew-php
> brew install brew-php-switcher
> brew install php56

> which php
/usr/local/bin/php

> which php-fpm
lrwxr-xr-x  1 leeco  admin    32B May 17 15:08 /usr/local/bin/php -> ../Cellar/php56/5.6.30_6/bin/php
```

配置PHP：

```sh
# 配置 /private/etc/php.ini
> sudo vim /usr/local/etc/php/5.6/php.ini
# date.timezone = Asia/Shanghai

> php -i | grep timezone
Default timezone => Asia/Shanghai
date.timezone => Asia/Shanghai => Asia/Shanghai
```

安装 PEAR 和 PECL:

```sh
# 参考 https://jason.pureconcepts.net/2012/10/install-pear-pecl-mac-os-x/
> curl -O http://pear.php.net/go-pear.phar
> sudo php -d detect_unicode=0 go-pear.phar
# Press return

# 安装 intl 扩展，参考 http://note.rpsh.net/posts/2015/10/07/installing-php-intl-extension-os-x-el-capitan/
> sudo pear channel-update pear.php.net
> sudo pecl channel-update pecl.php.net
> sudo pear upgrade-all
```

## 如果要安装多个版本的PHP

```sh
# 用于在不同版本的php之间切换，参考 https://github.com/philcook/brew-php-switcher
> brew unlink php56
> brew install php71

> brew-php-switcher 56 -s # 切回 php 5.6
```

# Install Composer

```sh
> brew install composer

> composer
```

{% asset_img study-php-and-symfony-on-mac-os-a5831.png %}

# Install PHP Extension

## Install `intl`

安装

```sh
> brew install autoconf
> brew install icu4c
> brew install php56-intl   # /usr/local/etc/php/5.6/conf.d/ext-intl.ini was created
```

检查是否安装成功：

```
> php -m | grep intl # 正常会包含 intl
```

## Install `OPcache`

```sh
> brew install php56-opcache

> brew info php56-opcache
...
To finish installing opcache for PHP 5.6:
  * /usr/local/etc/php/5.6/conf.d/ext-opcache.ini was created,
    do not forget to remove it upon extension removal.
  * Validate installation via one of the following methods:
...

> php -m | grep OPcache  # 检查 OPcache 是否已生效
```

# Install `Xdebug`

安装：

```sh
> brew install php56-xdebug
...
To finish installing xdebug for PHP 5.6:
  * /usr/local/etc/php/5.6/conf.d/ext-xdebug.ini was created,
    do not forget to remove it upon extension removal.
  * Validate installation via one of the following methods:
...
```

The Xdebug extension will be enabled per default after the installation, additional configuration of the extension should be done by adding a custom ini-file to `/usr/local/etc/php/<php-version>/conf.d/`.

配置：

```
> sudo echo 'xdebug.remote_enable=1
xdebug.remote_host=127.0.0.1
xdebug.remote_port=9000
xdebug.profiler_enable=1
xdebug.profiler_output_dir="/tmp/xdebug-profiler-output"' >> /usr/local/etc/php/5.6/conf.d/ext-xdebug.ini
```

安装`xdebug-osx`（`xdebug`开关工具）：

```sh
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

# Use PHPStorm as IDE

## Config Language & Frameworks

打开`Preferences > Languages & Frameworks > PHP`

{% asset_img study-php-and-symfony-on-mac-os-0c6a2.png %}

添加CLI Intepreter

{% asset_img study-php-and-symfony-on-mac-os-2aa39.png %}


## Config `Xdebug`

打开`Preferences > Languages & Frameworks > PHP > Debug`

{% asset_img study-php-and-symfony-on-mac-os-33b82.png %}

如何配置参考[Configuring Xdebug in PhpStorm](https://www.jetbrains.com/help/phpstorm/configuring-xdebug.html)

## 验证

创建一个PHP项目，新建一个`php`文件，创建执行配置：

{% asset_img study-php-and-symfony-on-mac-os-6c7bd.png %}

打上断点，以`Debug`方式运行：

{% asset_img study-php-and-symfony-on-mac-os-aec25.png %}

# 参考

- [macOS 10.12 Sierra Apache Setup: MySQL, APC & More...](https://getgrav.org/blog/macos-sierra-apache-mysql-vhost-apc)
- [Cannot find libz when install php56](https://github.com/Homebrew/homebrew-php/issues/1946)
- [How to install Xdebug - Xdebug Documents](https://xdebug.org/docs/install)
- [Xdebug Installation Guide](https://confluence.jetbrains.com/display/PhpStorm/Xdebug+Installation+Guide)
- [PHPStorm Help - Configuring Xdebug](https://www.jetbrains.com/help/phpstorm/configuring-xdebug.html)
