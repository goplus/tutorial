XGo では、インターフェースを明示的に実装する必要はありません。つまり `implement` というキーワードはありません。代わりに、インターフェースは暗黙的に満たされます。
---
型がインターフェースに要求されるすべてのメソッドを持っていれば、そのインターフェースを満たします。例えば、*os.File は io.Reader、io.Writer、io.Closer、io.ReadWriter を満たします。*bytes.Buffer は io.Reader、io.Writer、io.ReadWriter を満たしますが、Close メソッドがないため io.Closer は満たしません。
---
インターフェースの代入可能性ルールは非常にシンプルです：式の型がインターフェースを満たす場合にのみ、その式をインターフェースに代入できます。右辺自体がインターフェースの場合でも、このルールは同様に適用されます。例えば：
---
io.ReadWriter と io.ReadWriteCloser は io.Writer のすべてのメソッドを含むため、io.ReadWriter または io.ReadWriteCloser を満たす型は必然的に io.Writer も満たします。
