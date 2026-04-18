### 切片字面量
切片字面量类似于没有长度的数组字面量。
<pre>
这是一个数组字面量：
[3]bool{true, true, false}
而下面这种写法会创建与上面相同的数组，然后构建一个引用该数组的切片：
[]bool{true, true, false}
</pre>
