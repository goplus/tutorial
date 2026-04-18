### XGo 的基本类型包括
<pre>
bool
int8    int16   int32   int    int64    int128
uint8   uint16  uint32  uint   uint64   uint128
uintptr (类似于 C 语言的 size_t)
byte (uint8 的别名)
rune (int32 的别名，表示一个 Unicode 码点)
string
float32 float64
complex64 complex128
bigint bigrat
unsafe.Pointer (类似于 C 语言的 void*)
any (Go 中 interface{} 的别名)
</pre>

示例展示了多种类型的变量，同时也演示了变量声明可以像 import 语句一样被"分组"到代码块中。
---
int、uint 和 uintptr 类型在 32 位系统上通常是 32 位宽，在 64 位系统上是 64 位宽。当你需要一个整数值时，应该使用 int，除非有特定原因需要使用固定大小或无符号的整数类型。
