package main

import (
	"errors"
	"fmt"
)

// 1. Функция с параметром и возвращаемым значением

func add(a int, b int) int {
	return a + b
}

// Go позволяет групировать параметры

func multiply(a, b int) int {
	return a * b
}

// 2. Результат и ошибка

var errDivisionByZero = errors.New("division by zero")

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errDivisionByZero
	}

	return a / b, nil
}

// 3. Возврат значения в стиле error

func findMax(numbers []int) (int, bool) {
	if len(numbers) == 0 {
		return 0, false
	}

	max := numbers[0]

	for _, number := range numbers[1:] {
		if number > max {
			max = number
		}
	}

	return max, true
}

// 4. Именованные возвращаемые значения

func rectangleArea(width, height float64) (area float64) {
	area = width * height
	return
}

// 5. Переменное число аргументов

func sum(numbers ...int) int {
	total := 0

	for _, number := range numbers {
		total += number
	}

	return total
}

// 6. Передача значений в функции

func changeValue(value int) {
	value = 100
}

func changeCopy(value int) {
	value = 99
}

// 7. Передача указателя

func changeThroughPointer(value *int) {
	*value = 100
}

func changeOriginal(value *int) {
	*value = 99
}

// 8. Функция как значение

func square(number int) int {
	return number * number
}

// 9. Функция как параметр

func apply(number int, operation func(int) int) int {
	return operation(number)
}

// 10. Анонимная функция

func runAnonymousFunction() {
	message := "Hello from anonymous function"

	func() {
		fmt.Println(message)
	}()
}

// 11. Замыкание

func newCounter() func() int {
	counter := 0

	return func() int {
		counter++
		return counter
	}
}

func makeCounter() func() int {
	return newCounter()
}

// 12. Рекурсия

func factorial(number int) int {
	if number <= 1 {
		return 1
	}

	return number * factorial(number-1)
}

func main() {
	// 1. Базовая функция

	fmt.Println("=== Basic functions ===")
	fmt.Println("add:", add(10, 20))
	fmt.Println("multiply:", multiply(4, 5))

	// 2. Результат и ошибка

	fmt.Println("\n=== Result and error ===")

	quotient, err := divide(17, 5)

	fmt.Println("Quotient:", quotient)
	fmt.Println("Error:", err)

	// 3. Работа с пустым слайсом

	fmt.Println("\n=== Find max ===")

	numbers := []int{12, 5, 37, 9, 21}

	if max, ok := findMax(numbers); ok {
		fmt.Println("Maximum:", max)
	}

	if _, ok := findMax(nil); !ok {
		fmt.Println("Cannot find max in an empty slice")
	}

	// 4. Именованное возвращаемое значение

	fmt.Println("\n=== Named return ===")
	fmt.Println("Area:", rectangleArea(5, 4))

	// 5. Функция с переменным числом аргументов

	fmt.Println("\n=== Variadic ===")
	fmt.Println("Sum:", sum(1, 2, 3, 4, 5))

	values := []int{10, 20, 30}
	fmt.Println("Sum of slice:", sum(values...))

	// 6. Пропуск по значению

	fmt.Println("\n=== Pass by value ===")

	value := 10
	changeValue(value)

	fmt.Println("Original value:", value)

	// 7. Указатель как параметр

	fmt.Println("\n=== Pointer ===")

	changeThroughPointer(&value)

	fmt.Println("Value after pointer call:", value)

	// 8. Функция как значение

	fmt.Println("\n=== Function as a value ===")

	operation := square

	fmt.Println("Square:", operation(6))

	// 9. Функция как параметр

	fmt.Println("\n=== Higher-order function ===")

	result := apply(6, square)

	fmt.Println("Result:", result)

	// 10. Анонимная функция

	fmt.Println("\n=== Anonymous function ===")

	runAnonymousFunction()

	// 11. Замыкание

	fmt.Println("\n=== Closure ===")

	counterA := makeCounter()
	counterB := makeCounter()

	fmt.Println(counterA()) // 1
	fmt.Println(counterA()) // 2
	fmt.Println(counterA()) // 3

	fmt.Println(counterB()) // 1
	fmt.Println(counterA()) // 4

	// 12. Рекурсия

	fmt.Println("\n=== Recursion ===")

	fmt.Println("Factorial of 5:", factorial(5))
}
