以可变数量参数调用的 joinstr 函数被称为可变参数函数。在 joinstr 函数的声明中，最后一个参数的类型前面带有省略号，即 (…)。这表示该函数可以接受任意数量的 string 参数。调用 joinstr(elements...) 时，如果不加 ...，将无法编译，因为类型不匹配；elements 不是 string 类型。
