# Bug 是什么

并发构造列表、仪表盘和 JSON 响应时共享可变缓冲，造成数据竞争、跨请求数据污染和偶发崩溃。

# 如何触发

执行题面中的 `-race` 测试，让多个 goroutine 同时生成不同用户的返回数据。

# 错误信息

Race detector 输出 `WARNING: DATA RACE`；高并发下还可能出现 `slice bounds out of range`。
