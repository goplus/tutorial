### 메서드 표현식
---
메서드 표현식은 메서드를 리시버를 첫 번째 인자로 받는 함수로 변환합니다. 예를 들어:
---
따라서 Point 구조체와 메서드 func (p Point) Add(another Point)를 정의했다면, Point.Add를 작성하여 func(p Point, another Point)를 얻을 수 있습니다.
