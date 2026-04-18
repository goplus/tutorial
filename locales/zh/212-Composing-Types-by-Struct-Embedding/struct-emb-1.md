XGo 不提供典型的、基于类型驱动的子类化概念，但它可以通过在结构体或接口中嵌入类型来"借用"部分实现。只有接口可以嵌入到接口中。
---
示例中有两个结构体类型，Hello 结构体和 Goodbye 结构体，它们各自实现了 Talk 接口。HelloGoodbye 结构体也实现了 Talk 接口，它通过嵌入的方式将 Hello 结构体和 Goodbye 结构体组合到一个结构体中：在结构体中列出类型但不给它们字段名。
---
嵌入的元素是指向结构体的指针，当然在使用之前必须初始化为指向有效的结构体。
---
HelloGoodbye 结构体也有一个写作 forward *Forward 的 forward 成员，但如果要提升 forward 的方法并满足 Talk 接口，我们还需要提供转发方法，例如：hg.forward.Say()。通过直接嵌入结构体，我们避免了这种繁琐的操作。嵌入类型的方法会自动继承过来，这意味着 HelloGoodbye 结构体也拥有 Hello 结构体和 Goodbye 结构体的方法。
---
嵌入与子类化之间有一个重要区别。当我们嵌入一个类型时，该类型的方法成为外部类型的方法，但当它们被调用时，方法的接收者是内部类型，而不是外部类型。在我们的示例中，当 HelloGoodbye 的 Sleep 方法被调用时，它的效果与转发方法 helloGoodbye.Hello.Sleep() 完全相同；接收者是 helloGoodbye.Hello，而不是 helloGoodbye 本身。
