### XGo の基本型
<pre>
bool
int8    int16   int32   int    int64    int128
uint8   uint16  uint32  uint   uint64   uint128
uintptr（C 言語の size_t に類似）
byte（uint8 のエイリアス）
rune（int32 のエイリアス、Unicode コードポイントを表す）
string
float32 float64
complex64 complex128
bigint bigrat
unsafe.Pointer（C 言語の void* に類似）
any（Go の interface{} のエイリアス）
</pre>

この例では複数の型の変数を示すとともに、変数宣言が import 文と同様にブロックとして「グループ化」できることも示しています。
---
int、uint、uintptr 型は通常、32ビットシステムでは32ビット幅、64ビットシステムでは64ビット幅です。整数値が必要な場合は、固定サイズまたは符号なし整数型を使用する特定の理由がない限り、int を使用すべきです。
