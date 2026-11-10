package main

import (
	"fmt"
	"strings"
)

// generate создаёт канал, в который отправляются числа от 1 до n,
// после чего канал закрывается. Это первый этап конвейера — источник данных.
func generate(n int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)

		for i := 1; i <= n; i++ {
			out <- i
		}
	}()

	return out
}

// square принимает канал чисел и возвращает канал их квадратов.
// Каждый этап конвейера завершается, когда закрывается его входной канал.
func square(in <-chan int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)

		for value := range in {
			out <- value * value
		}
	}()

	return out
}

// format превращает числа в строки — последний этап конвейера.
func format(in <-chan int) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		for value := range in {
			out <- fmt.Sprintf("квадрат: %d", value)
		}
	}()

	return out
}

// repeatWords демонстрирует, что этапы конвейера можно комбинировать
// с функциями обработки данных любого типа.
func repeatWords(in <-chan string, times int) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		for word := range in {
			out <- strings.Repeat(word+" ", times)
		}
	}()

	return out
}

func main() {
	// Конвейер: generate -> square -> format.
	// Данные проходят через все этапы по одному значению,
	// горотины выполняются параллельно.
	for result := range format(square(generate(5))) {
		fmt.Println(result)
	}

	// Этапы можно соединять и со строковыми каналами.
	words := make(chan string)
	go func() {
		defer close(words)

		words <- "привет"
		words <- "мир"
	}()

	for result := range repeatWords(words, 2) {
		fmt.Println(strings.TrimSpace(result))
	}
}
