XGo에서 인터페이스는 명시적으로 구현할 필요가 없습니다. 즉, `implement` 키워드는 없습니다. 대신 인터페이스는 암시적으로 만족됩니다.
---
타입이 인터페이스가 요구하는 모든 메서드를 보유하면 해당 인터페이스를 만족합니다. 예를 들어, *os.File은 io.Reader, io.Writer, io.Closer, io.ReadWriter를 만족합니다. *bytes.Buffer는 io.Reader, io.Writer, io.ReadWriter를 만족하지만, Close 메서드가 없으므로 io.Closer는 만족하지 않습니다.
---
인터페이스의 할당 규칙은 매우 간단합니다: 표현식의 타입이 인터페이스를 만족하는 경우에만 해당 인터페이스에 할당할 수 있습니다. 이 규칙은 우변이 인터페이스인 경우에도 동일하게 적용됩니다. 예를 들어:
---
io.ReadWriter와 io.ReadWriteCloser는 io.Writer의 모든 메서드를 포함하므로, io.ReadWriter나 io.ReadWriteCloser를 만족하는 모든 타입은 반드시 io.Writer도 만족합니다.
