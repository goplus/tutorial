_Switch 语句_ 用于表达跨多个分支的条件判断。
---
这是一个基本的 `switch` 示例。
---
你可以在同一个 `case` 语句中使用逗号分隔多个表达式。在这个示例中我们还使用了可选的 `default` 分支。
---
不带表达式的 `switch` 是表达 if/else 逻辑的另一种方式。这里我们还展示了 `case` 表达式可以是非常量的。
---
XGo 中的 switch 默认在每个 case 末尾自动 break。使用 fallthrough 可以强制执行后续 case 的代码。
---
带有 `fallthrough` 的 `switch`：
