### 方法表达式
---
方法表达式可以将方法转换为一个以接收者作为第一个参数的函数。例如：
---
因此，如果你定义了一个 Point 结构体和一个方法 func (p Point) Add(another Point)，你可以写 Point.Add 来获得一个 func(p Point, another Point)。
