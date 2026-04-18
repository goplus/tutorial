_Maps_ 是 XGo 内置的[关联数据类型](http://en.wikipedia.org/wiki/Associative_array)（在其他语言中有时称为 _哈希_ 或 _字典_）。
---
### Map 基础
---
使用 `make(map[key-type]val-type)` 创建一个空的 map。
---
使用常见的 `name[key] = val` 语法设置键值对。
---
使用例如 `println` 打印 map 时，会显示其所有的键值对。
---
使用 `name[key]` 获取某个键对应的值。
---
内置函数 `len` 在用于 map 时，返回键值对的数量。
---
内置函数 `delete` 从 map 中删除键值对。
---
从 map 中获取值时，可选的第二个返回值指示该键是否存在于 map 中。这可以用来区分缺失的键和零值的键（如 `0` 或 `""`）。这里我们不需要值本身，所以用 _空白标识符_ `_` 忽略了它。
