package main

import "fmt"

func main() {
	// 1. Условия

	fmt.Println("=== Conditions ===")

	score := 82

	if score >= 90 {
		fmt.Println("Grade: A")
	} else if score >= 75 {
		fmt.Println("Grade: B")
	} else if score >= 60 {
		fmt.Println("Grade: C")
	} else {
		fmt.Println("Grade: F")
	}

	// 2. Быстрые обьявления в if

	if number := 15; number%2 == 0 {
		fmt.Println(number, "is even")
	} else {
		fmt.Println(number, "is odd")
	}

	// 3. Логические операции

	age := 25
	hasExperience := true

	if age >= 18 && hasExperience {
		fmt.Println("Candidate meets the basic requirements")
	}

	value := 0

	if value != 0 && 100/value > 1 {
		fmt.Println("Condition is true")
	} else {
		fmt.Println("Second operand was not evaluated")
	}

	// 4. Switch с значениями

	fmt.Println("\n=== Switch ===")

	day := "Saturday"

	switch day {
	case "Monday", "Tuesday", "Wednesday", "Thursday", "Friday":
		fmt.Println("Working day")
	case "Saturday", "Sunday":
		fmt.Println("Weekend")
	default:
		fmt.Println("Unknown day")
	}

	// 5. Switch с выражениями

	temperature := 28

	switch {
	case temperature < 0:
		fmt.Println("Freezing")
	case temperature < 20:
		fmt.Println("Cold")
	case temperature < 30:
		fmt.Println("Warm")
	default:
		fmt.Println("Hot")
	}

	// 6. For с 3 параметрами

	fmt.Println("\n=== Classic for ===")

	for i := 1; i <= 5; i++ {
		fmt.Println(i)
	}

	// 7. цикл for как цикл while

	fmt.Println("\n=== Condition-based for ===")

	counter := 1

	for counter <= 3 {
		fmt.Println("Counter:", counter)
		counter++
	}

	// 8. бесконечный цикл с break

	fmt.Println("\n=== Break ===")

	number := 1

	for {
		if number > 3 {
			break
		}

		fmt.Println("Number:", number)
		number++
	}

	// 9. Continue

	fmt.Println("\n=== Continue ===")

	for i := 1; i <= 10; i++ {
		if i%2 == 0 {
			continue
		}

		fmt.Println("Odd number:", i)
	}

	// 10. Range по слайсу

	fmt.Println("\n=== Range over slice ===")

	numbers := []int{10, 20, 30, 40}

	for index, value := range numbers {
		fmt.Printf("index=%d value=%d\n", index, value)
	}

	// 11. Range только по значениям

	for _, value := range numbers {
		fmt.Println("Value:", value)
	}

	// 12. Range по map

	fmt.Println("\n=== Range over map ===")

	ages := map[string]int{
		"Alice": 25,
		"Bob":   30,
		"Carol": 22,
	}

	for name, age := range ages {
		fmt.Printf("%s: %d\n", name, age)
	}

	// Порядок итераций в map случайный

	// 13. Range по строке

	fmt.Println("\n=== Range over string ===")

	word := "GoЖ"

	for index, r := range word {
		fmt.Printf("byte index=%d rune=%c codepoint=%U\n", index, r, r)
	}

	// 14. Вложенные циклы и break

	fmt.Println("\n=== Labeled break ===")

outer:
	for i := 1; i <= 3; i++ {
		for j := 1; j <= 3; j++ {
			fmt.Printf("i=%d j=%d\n", i, j)

			if i == 2 && j == 2 {
				break outer
			}
		}
	}

	// 15. Defer выполняются в обратном порядке

	fmt.Println("\n=== Defer ===")

	defer fmt.Println("First deferred call")
	defer fmt.Println("Second deferred call")

	fmt.Println("End of main body")
}
