# Bug 是什么

服务层把依赖错误转成只有相同文字的新错误，调用方无法继续用 `errors.Is` 识别原错误。

# 如何触发

让用户、短链、跳转或访问统计依赖返回 sentinel 错误，再运行题面给出的目标测试。

# 错误信息

测试报告 `error lost identity`，错误文本仍可见，但 `errors.Is` 返回 false。
