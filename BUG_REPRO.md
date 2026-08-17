# Bug 是什么

没有过期时间的短链在多个读取路径中触发 nil 指针解引用。

# 如何触发

创建 `ExpiresAt` 为空的短链，执行列表转换、过期判断、跳转解析或列表接口测试。

# 错误信息

`panic: runtime error: invalid memory address or nil pointer dereference`。
