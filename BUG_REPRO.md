# Bug 是什么

HTTP 请求的取消状态没有传递到服务与仓储调用，底层工作在客户端离开后仍继续执行。

# 如何触发

给登录、列表、仪表盘或跳转请求传入已取消的 context，再观察仓储层收到的 context 状态。

# 错误信息

目标测试报告 `repository did not receive the canceled request context`。
