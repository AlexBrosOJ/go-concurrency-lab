package main

import (
	"errors"
	"fmt"
	"strings"
)

// ordinary — обычная функция: параметры, именованный возвращаемый результат.
// Именованные результаты видны в return, но их главное назначение —
// читаемость сигнатуры и работа с defer.
func divide(a, b float64) (result float64, err error) {
	if b == 0 {
		return 0, errors.New("деление на ноль")
	}

	return a / b, nil
}

// variadic — вариативная функция: принимает произвольное число аргументов.
// Внутри args ведёт себя как срез []int.
func sum(args ...int) int {
	total := 0

	for _, value := range args {
		total += value
	}

	return total
}

// callback — функция высшего порядка: принимает другую функцию как параметр.
func applyEach(values []int, transform func(int) int) []int {
	result := make([]int, len(values))

	for i, value := range values {
		result[i] = transform(value)
	}

	return result
}

// closure — функция возвращает замыкание. Замыкание "запоминает"
// переменную count из окружающего контекста и изменяет её при каждом вызове.
func counter() func() int {
	count := 0

	return func() int {
		count++
		return count
	}
}

// multiReturn — множественное возвращаемое значение: типичный идиоматичный
// способ вернуть результат вместе с ошибкой или дополнительными данными.
func splitName(full string) (first, last string, ok bool) {
	parts := strings.SplitN(full, " ", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

// method — метод со значением-получателем. Методы с указателем (*Stack)
// могут изменять получателя; здесь мы только читаем.
type Stack struct {
	items []string
}

func (s Stack) peek() (string, bool) {
	if len(s.items) == 0 {
		return "", false
	}

	return s.items[len(s.items)-1], true
}

func main() {
	result, err := divide(10, 3)
	fmt.Printf("divide(10, 3) = %.2f, err = %v\n", result, err)

	_, err = divide(10, 0)
	fmt.Printf("divide(10, 0): err = %v\n", err)

	fmt.Println("sum(1, 2, 3) =", sum(1, 2, 3))
	fmt.Println("sum(slice...) =", sum([]int{10, 20, 30}...))

	doubled := applyEach([]int{1, 2, 3}, func(value int) int {
		return value * 2
	})
	fmt.Println("applyEach doubled:", doubled)

	next := counter()
	fmt.Println("counter:", next(), next(), next())

	// Каждое замыкание имеет свою собственную копию состояния.
	second := counter()
	fmt.Println("second counter:", second())

	first, last, ok := splitName("Алан Тьюринг")
	fmt.Printf("splitName: first=%q last=%q ok=%v\n", first, last, ok)

	stack := Stack{items: []string{"go", "channels", "goroutines"}}
	top, ok := stack.peek()
	fmt.Printf("peek: %q ok=%v\n", top, ok)
}
