# Warm-Sched

一个快速进入热启动状态的调度工具．

# 想法
冷启动和热启动在机械硬盘上的响应速度非常大．而内存中一般只有
1. file system' cache/buffer memory page
2. program's anonymous memory page  (heap, stack等)
3. kernel's slab memory page.

除了file system的cache/buffer需要从低速的磁盘处理，其他两类page都是进程正常执行过程
中逐步创建的，不受冷热启动的影响．

因此只要能有办法一次性准确的重构file system's cache memory page就能快速进入类似热启动的状态下．
而这个过程不论是服务器还是桌面都是比较固定的．因此完全是可以牺牲少量启动时间换来极大的用户体验．

# kernel mincores
warm-sched早期版本(参见mincore_syscall.go)采用传统的
mincore(2)配合mmap方式来探测具体的某一个文件在PageCache中的情况．

但这样就必须"先知道具体的文件列表才能逐一进行询问"．
如果进行全盘扫描则dentry的读取会引起大量磁盘IO，且速度缓慢．类似执行了一次updatedb(8)的操作.

这是类似软件都会遇到的一个问题，一般采用以下策略进行规避．
1. 只扫描被系统中正在运行进程所打开的文件．
2. 扫描少量指定目录,比如/usr/lib, /lib.

warm-shed通过kernel module的方式直接导出VFS中inode的情况来大大增加探测能力．
安装mincores-dkms.deb后，系统会生成一个/proc/mincores文件．此文件的输出会根据当前进程的PWD进行输出.

由于直接通过VFS系统，所以资源消耗只取决与在PageCache中的文件数量与磁盘大小无关．一般都在500ms内完成．
因此warm-sched在用户态通过配合systemd, startdde可以多次进行全盘扫描对比差异.

# 其他软件对比
1. warm-sched实现了一个高效的PageCache分析机制/proc/mincores, 因此可以更好的挖掘预加载潜能.
2. 基于事件源提供预热配置框架, 达到在不同阶段进行精准的记录.

目前支持多种事件来源, 并且提供接口实现新的事件源.
- 基于文件是否存在
- 基于X11 应用进程是否出现
- 基于systemd unit是否出现
- 基于其他snapshot是否已经加载
- 基于某个进程名是否出现


## systemd-readahead
systemd[自带组件](http://sourceforge.net/projects/preload)，目前由于无人维护，且systemd的维护者使用的都是固态盘因此此组件被废弃．

systemd-readahead只记录了systemd启动过程中的资源情况，且因为没有高效的PageCache分析机制．在固态盘下会影响整体效率．

## preload
[preload](http://sourceforge.net/projects/preload) is an adaptive readahead daemon. It monitors applications that users run, and by analyzing this data, predicts what applications users might run, and fetches those binaries and their dependencies into memory for faster startup times.

preload只收集了进程的elf文件，但UI程序启动还依赖大量的字体，图片等文件．因此preload对DE的效果十分不明显．

# 编译
1. vagrant up && vagrant ssh
2. dpkg-buildpackage && cp ../*.deb . && dh clean && exit

# 测试方式
1. 安装mincores-dkms.deb以及warm-sched.deb
2. 重启以后，登录dde, 在30分钟打开chrome,firefox等默认支持的应用程序.
3. 下次重启即可对比实际效果．

# TODO
- [X] 分系统级与用户级阶段加载．分别在sysinit.target和greeter输入密码阶段
- [X] 使用kernel module 或 ebpf等机制 查询某个dentry是否在page cache中
- [X] 文件黑名单
- [X] 实验环境下，分析应该对哪些目录进行扫描，以便生成cache list.
- [X] 实验分析在哪个阶段进行预热更合适．
- [X] 根据实际可用内存大小预热价值更高的文件.
- [X] 记录历史信息统计文件使用频率.
- [X] 提供snapshot inspection.
- [ ] 根据实际分析，清理进入桌面后明显不会被用到的Cache(比如plymouth等), 辅助kernel进行调度.
- [ ] 完善黑名单机制.
- [ ] 在配置文件变动时, 标记对应snapshot为dirty, 以便下次执行实际的capture操作.
- [ ] 在TryFile变动时, 标记对应snapshot为dirty, 以便下次执行实际的capture操作.
- [ ] 剥离事件源到独立的组件中,以便支持外部事件源.
