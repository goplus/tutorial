### 使用 make 创建切片
切片可以使用内置的 make 函数创建；这是创建动态大小数组的方式。

make 函数会分配一个零值数组，并返回一个引用该数组的切片：

<pre>
a := make([]int, 5)			// len(a)=5
要指定容量，可以向 make 传递第三个参数：
b := make([]int, 0, 5) 		// len(b)=0, cap(b)=5
b = b[:cap(b)] 				// len(b)=5, cap(b)=5
b = b[1:]      				// len(b)=4, cap(b)=4
</pre>
