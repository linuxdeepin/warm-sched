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
总体来说由于warm-sched自带一个高效的PageCache分析机制/proc/mincores, 因此可以更好的挖掘预加载潜能．

## systemd-readahead
systemd[自带组件](http://sourceforge.net/projects/preload)，目前由于无人维护，且systemd的维护者使用的都是固态盘因此此组件被废弃．

systemd-readahead只记录了systemd启动过程中的资源情况，且因为没有高效的PageCache分析机制．在固态盘下会影响整体效率．

## preload
[preload](http://sourceforge.net/projects/preload) is an adaptive readahead daemon. It monitors applications that users run, and by analyzing this data, predicts what applications users might run, and fetches those binaries and their dependencies into memory for faster startup times.

preload只收集了进程的elf文件，但UI程序启动还依赖大量的字体，图片等文件．因此preload对DE的效果十分不明显．

# 实现
1. DE登录后20s左右后，在idle状态下分析Page Cache中file system的cache情况，并记录为cache list
2. 进入用户界面前，读取cache list并预加载对应文件．

- cache list的生成, 通过mmap(2) + mincore(2) 扫描/{,usr}/lib, /{,usr}/bin等目录下的文件．
- cache list的读取, 通过mmap(2) + madvise(2) 进行精准预读．


# 样本数据

[样本数据](./sample.list)
是在deepin 15.4.1 刚进入桌面后启动deepin-terminal后收集

`./warm-sched -c | sort -hr`

后生成

第一列为对应文件实际使用的RAM, 第二列为占用RAM与文件大小的比例，第三列为文件路径．

从样本数据可以发现一些　[问题](https://github.com/snyh/warm-sched/issues)

注意: 这些只是Page Cache的使用情况, 在内存压力较大时，只要最近没有访问，且是干净的(没修改过)，
那么在换页时的代价是非常小的．


# 编译
1. vagrant up && vagrant ssh
2. dpkg-buildpackage && cp ../*.deb . && dh clean && exit

# 测试方式
1. 安装mincores-dkms.deb以及warm-sched.deb
2. 重启以后，登录dde, 等待20s
3. 下次重启即可对比实际效果．

# TODO
- [X] 分系统级与用户级阶段加载．分别在sysinit.target和greeter输入密码阶段
- [X] 使用kernel module 或 ebpf等机制 查询某个dentry是否在page cache中
- [X] 文件黑名单
- [ ] 使用startdde提供的launch app接口UI App进行snapshot
- [ ] 纳入应用启动时间参数到UI App snapshot
- [X] 实验环境下，分析应该对哪些目录进行扫描，以便生成cache list.
- [X] 实验分析在哪个阶段进行预热更合适．
- [ ] 根据实际可用内存大小预热价值更高的文件．
- [ ] 记录历史信息统计文件使用频率
- [ ] 根据实际分析，清理进入桌面后明显不会被用到的Cache(比如plymouth等), 辅助kernel进行调度.
- [ ] 提供snapshot inspection
