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


# 实现
1. DE登录后10s左右后，在idle状态下分析Page Cache中file system的cache情况，并记录为cache list
2. 进入用户界面前，读取cache list并预加载对应文件．

- cache list的生成, 通过mmap(2) + mincore(2) 扫描/{,usr}/lib, /{,usr}/bin等目录下的文件．
- cache list的读取, 通过mmap(2) + madvise(2) 进行精准预读．

# 需要调研的事情
1. [ ] 实验环境下，分析应该对哪些目录进行扫描，以便生成cache list.
2. [ ] 实验分析在哪个阶段进行预热更合适．
3. [ ] 根据实际可用内存大小预热价值更高的文件．需要采用计数等方式统计出价值更高的文件．
4. [ ] 根据实际分析，清理进入桌面后明显不会被用到的Cache(比如plymouth等), 辅助kernel进行调度.

# 样本数据

[样本数据](./sample.list)
是在deepin 15.4.1 刚进入桌面后启动deepin-terminal后收集
`./warm-sched /bin /etc /boot /lib /lib32 /lib64 /libx32 /opt /sbin /srv /usr /var | sort -hr`
后生成
第一列为对应文件实际使用的RAM, 第二列为占用RAM与文件大小的比例，第三列为文件路径．

从样本数据可以发现一些　[问题](https://github.com/snyh/warm-sched/issues)

注意: 这些只是Page Cache的使用情况, 在内存压力较大时，只要最近没有访问，且是干净的(没修改过)，
那么在换页时的代价是非常小的．
