### XGo의 기본 타입은 다음과 같습니다
<pre>
bool
int8    int16   int32   int    int64    int128
uint8   uint16  uint32  uint   uint64   uint128
uintptr (C 언어의 size_t와 유사)
byte (uint8의 별칭)
rune (int32의 별칭, 유니코드 코드 포인트를 나타냄)
string
float32 float64
complex64 complex128
bigint bigrat
unsafe.Pointer (C 언어의 void*와 유사)
any (Go의 interface{}의 별칭)
</pre>

예제는 여러 타입의 변수를 보여주며, 변수 선언도 import 문처럼 블록으로 "묶을" 수 있음을 보여줍니다.
---
int, uint, uintptr 타입은 32비트 시스템에서는 보통 32비트, 64비트 시스템에서는 64비트입니다. 정수 값이 필요할 때는 특정 크기나 부호 없는 정수 타입을 사용할 이유가 없다면 int를 사용해야 합니다.