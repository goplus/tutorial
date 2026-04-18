在 XGo 中，接口不需要显式实现——即没有 implement 关键字。相反，接口是隐式满足的。
---
一个类型如果拥有接口所要求的所有方法，就满足了该接口。例如，*os.File 满足 io.Reader、io.Writer、io.Closer 和 io.ReadWriter。*bytes.Buffer 满足 io.Reader、io.Writer 和 io.ReadWriter，但不满足 io.Closer，因为它没有 Close 方法。
---
接口的可赋值规则非常简单：只有当表达式的类型满足接口时，才能将其赋值给接口。即使右侧本身就是接口，该规则同样适用。例如：
---
因为 io.ReadWriter 和 io.ReadWriteCloser 包含了 io.Writer 的所有方法，任何满足 io.ReadWriter 或 io.ReadWriteCloser 的类型也必然满足 io.Writer。
