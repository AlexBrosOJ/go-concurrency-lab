package main

import (
	"fmt"

	"go-concurrency-lab/00-fundamentals/09-packages-and-modules/internal/textstats"
)

func main() {
	fmt.Println("=== Вызов функции из другого пакета ===")
	for _, input := range []string{"Go lab", "ёж и кот", ""} {
		stats := textstats.Analyze(input)
		fmt.Printf("%q: слов %d, рун %d\n", input, stats.Words, stats.Runes)
	}
}
