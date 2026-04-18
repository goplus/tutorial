### Recover
---
要将错误作为返回值报告，你必须在调用 panic 函数的同一个 goroutine 中调用 recover 函数，从 recover 函数获取错误结构体，并将其传递给一个变量：
---
每个被推迟的函数都会在函数调用之后、return 语句之前执行。因此，你可以在 return 语句执行之前设置返回变量的值。
