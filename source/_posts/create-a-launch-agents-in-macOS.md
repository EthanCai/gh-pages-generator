---
title: 在macOS上创建自启动服务解决软件启动时的注册验证问题
tags:
  - macOS
categories:
  - 善用佳软
date: 2017-11-18 19:54:23
---


{% asset_img 2a73dea7a35b4a2057eec46b15c17133.png %}

经常使用的一个开发软件需要注册。网上找了一个注册证书服务程序，这个程序启动后会提供一个注册服务。软件启动的时候，提供这个服务监听的本地**http**端口，就能自动通过注册验证。但是这个注册证书服务程序是个命令行程序，每次手动启动这个程序实在太麻烦，我就想把这个程序做成一个后台服务，这样就节省了每次手动操作的时间。在macOS可以通过`launchd`启动后台服务，关于`launchd`，wiki2上介绍如下：

> In computing, launchd, a unified service-management framework, starts, stops and manages daemons, applications, processes, and scripts.
>
> There are two main programs in the launchd system: launchd and launchctl.
>
> launchd manages the daemons at both a system and user level. Similar to xinetd, launchd can start daemons on demand. Similar to watchdogd, launchd can monitor daemons to make sure that they keep running. launchd also has replaced init as PID 1 on macOS and as a result it is responsible for starting the system at boot time.
>
> Configuration files define the parameters of services run by launchd. Stored in the LaunchAgents and LaunchDaemons subdirectories of the Library folders, the property list-based files have approximately thirty different keys that can be set. launchd itself has no knowledge of these configuration files or any ability to read them - that is the responsibility of "launchctl".
>
> launchctl is a command line application which talks to launchd using IPC and knows how to parse the property list files used to describe launchd jobs, serializing them using a specialized dictionary protocol that launchd understands. launchctl can be used to load and unload daemons, start and stop launchd controlled jobs, get system utilization statistics for launchd and its child processes, and set environment settings.


**具体操作步骤如下：**

在`/Users/{your_username}/Library/LaunchAgents`下创建`com.myutils.licsrv.plist`文件

```
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>com.myutils.licsrv.activator</string>
    <key>ProgramArguments</key>
    <array>
       <string>/Users/{your_username}/Applications/myutils/licsrv.darwin.amd64</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>/dev/null</string>
    <key>StandardOutPath</key>
    <string>/dev/null</string>
  </dict>
</plist>
```

将注册证书服务程序拷贝到`/Users/{your_username}/Applications/myutils`目录下。

```
> launchctl load com.myutils.licsrv.plist
> launchctl start com.myutils.licsrv.plist
> ps aux | grep licsrv
```

然后重启再测试一下，这样一个后台服务就创建好了。


# 参考

- [Daemons and Services Programming Guide](https://developer.apple.com/library/content/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/Introduction.html)
  - [Creating Launch Daemons and Agents](https://developer.apple.com/library/content/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html)
- [launchd info](http://www.launchd.info/)
- [wiki2 - launchd](https://wiki2.org/en/Launchd)
