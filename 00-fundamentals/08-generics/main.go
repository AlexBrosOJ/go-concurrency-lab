package main

import (
	"cmp"
	"fmt"
)

// first сохраняет тип элемента и отдельно сообщает, был ли он найден.
func first[T any](values []T) (T, bool) {
	if len(values) == 0 {
		var zero T
		return zero, false
	}
	return values[0], true
}

func contains[T comparable](values []T, wanted T) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

// minValue использует тот же порядок для NaN, что и cmp.Compare.
func minValue[T cmp.Ordered](values []T) (T, bool) {
	minimum, ok := first(values)
	if !ok {
		return minimum, false
	}
	for _, value := range values[1:] {
		if cmp.Less(value, minimum) {
			minimum = value
		}
	}
	return minimum, true
}

// reversed сохраняет именованный тип слайса и создаёт отдельный массив.
func reversed[S ~[]E, E any](source S) S {
	if source == nil {
		return nil
	}
	result := make(S, len(source))
	for index, value := range source {
		result[len(source)-1-index] = value
	}
	return result
}

type Scores []int

func (scores Scores) Count() int {
	return len(scores)
}

func main() {
	fmt.Println("=== 1. any и вывод типа ===")
	number, ok := first([]int{0, 20})
	fmt.Printf("Первый элемент: %d, найден: %t\n", number, ok)
	word, ok := first[string](nil)
	fmt.Printf("Пустой список: %q, найдено: %t\n", word, ok)

	fmt.Println("\n=== 2. comparable ===")
	fmt.Println("Есть Go:", contains([]string{"Go", "SQL"}, "Go"))
	fmt.Println("Есть 5:", contains([]int{1, 2, 3}, 5))

	fmt.Println("\n=== 3. cmp.Ordered ===")
	minimum, _ := minValue([]int{3, -2, 1})
	fmt.Println("Минимальное число:", minimum)
	name, _ := minValue([]string{"beta", "alpha"})
	fmt.Println("Первая строка по порядку:", name)

	fmt.Println("\n=== 4. Сохранение именованного типа ===")
	original := Scores{10, 20, 30}
	result := reversed(original)
	fmt.Println("Исходный:", original)
	fmt.Printf("Обратный: %v, тип: %T, элементов: %d\n", result, result, result.Count())
}
