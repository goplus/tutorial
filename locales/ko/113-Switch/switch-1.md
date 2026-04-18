_Switch 문_ 은 여러 분기에 걸친 조건 판단을 표현하는 데 사용됩니다.
---
다음은 기본적인 `switch` 예제입니다.
---
동일한 `case` 문에서 쉼표로 여러 표현식을 구분할 수 있습니다. 이 예제에서는 선택적인 `default` 분기도 사용합니다.
---
표현식이 없는 `switch`는 if/else 로직을 표현하는 또 다른 방법입니다. 여기서는 `case` 표현식이 상수가 아닐 수도 있음을 보여줍니다.
---
XGo의 switch는 기본적으로 각 case 끝에서 자동으로 break됩니다. fallthrough를 사용하면 후속 case의 코드를 강제로 실행할 수 있습니다.
---
`fallthrough`가 있는 `switch`: