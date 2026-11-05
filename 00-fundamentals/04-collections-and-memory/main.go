package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	fmt.Println("=== 1. Arrays ===")
	arrays()

	fmt.Println("\n=== 2. Slices: len, cap, nil ===")
	slicesBasics()

	fmt.Println("\n=== 3. Shared backing array ===")
	sharedBackingArray()

	fmt.Println("\n=== 4. Append and reallocation ===")
	appendBehavior()

	fmt.Println("\n=== 5. Copy ===")
	copyExample()

	fmt.Println("\n=== 6. Full slice expression ===")
	fullSliceExpression()

	fmt.Println("\n=== 7. Range over slices ===")
	rangeSlices()

	fmt.Println("\n=== 8. Maps ===")
	mapsBasics()

	fmt.Println("\n=== 9. Strings, bytes and runes ===")
	stringsBasics()
}

func arrays() {
	var numbers [3]int
	fmt.Println("Zero-value array:", numbers)

	numbers[0] = 10
	numbers[1] = 20
	numbers[2] = 30

	fmt.Println("Updated array:", numbers)
	fmt.Println("Length:", len(numbers))

	// Массив - значение
	// Присваивание копирует все значения
	copied := numbers
	copied[0] = 999

	fmt.Println("Original:", numbers)
	fmt.Println("Copy:", copied)

	// Массив
	prices := [3]int{100, 200, 300}
	fmt.Println("Prices:", prices)
}

func slicesBasics() {
	// Слайс.
	numbers := []int{10, 20, 30}

	fmt.Println("Slice:", numbers)
	fmt.Println("Length:", len(numbers))
	fmt.Println("Capacity:", cap(numbers))

	// nil слайс
	var nilSlice []int

	// Пустой но не nil слайс
	emptySlice := []int{}

	fmt.Println("Nil slice:", nilSlice)
	fmt.Println("Nil slice is nil:", nilSlice == nil)
	fmt.Println("Empty slice is nil:", emptySlice == nil)

	// make создает slice определенной length и capacity.
	created := make([]int, 3, 5)

	fmt.Println("Created with make:", created)
	fmt.Println("Length:", len(created))
	fmt.Println("Capacity:", cap(created))

	// append добавляет элемент в слайс.
	created = append(created, 40, 50)

	fmt.Println("After append:", created)
	fmt.Println("Length:", len(created))
	fmt.Println("Capacity:", cap(created))
}

func sharedBackingArray() {
	original := []int{10, 20, 30, 40}

	// Оба слайса ссылаются на один массив.
	first := original[1:3]
	second := original[2:4]

	fmt.Println("Original:", original)
	fmt.Println("First:", first)
	fmt.Println("Second:", second)

	first[1] = 999

	// Изменения видны с обоих слайсах
	fmt.Println("After modification:")
	fmt.Println("Original:", original)
	fmt.Println("First:", first)
	fmt.Println("Second:", second)
}

func appendBehavior() {
	// Capacity должен быть всегда больше length.
	base := make([]int, 3, 5)
	base[0] = 10
	base[1] = 20
	base[2] = 30

	first := base
	second := base

	first = append(first, 40)

	// append использовал существующий внутренний массив, поскольку
	// в нём было достаточно ёмкости.
	fmt.Println("First:", first)
	fmt.Println("Second:", second)

	// second по-прежнему имеет длину 3, но его ёмкость составляет 5.
	// Повторное взятие среза позволяет увидеть значение, добавленное через first.
	fmt.Println("Second resliced:", second[:4])

	// Теперь выполним добавление за пределами доступной ёмкости.
	small := make([]int, 3, 3)
	small[0] = 1
	small[1] = 2
	small[2] = 3

	before := &small[0]

	small = append(small, 4)

	after := &small[0]

	fmt.Println("After append:", small)
	fmt.Println("Capacity:", cap(small))
	fmt.Println("Backing array changed:", before != after)

	// Не полагайтесь на какой-то конкретный коэффициент увеличения вместимости.
	// В реализации Go могут быть выбраны другие значения вместимости.
}

func copyExample() {
	source := []int{10, 20, 30}
	destination := make([]int, len(source))

	copied := copy(destination, source)

	fmt.Println("Source:", source)
	fmt.Println("Destination:", destination)
	fmt.Println("Elements copied:", copied)

	destination[0] = 999

	// copy переносит элементы; при этом целевой объект
	// не начинает использовать тот же внутренний массив, что и исходный.
	fmt.Println("Source after destination modification:", source)
	fmt.Println("Destination:", destination)

	// copy копирует лишь столько элементов,
	// сколько помещается в целевой срез.
	shortDestination := make([]int, 2)
	count := copy(shortDestination, source)

	fmt.Println("Short destination:", shortDestination)
	fmt.Println("Copied:", count)
}

func fullSliceExpression() {
	numbers := make([]int, 3, 5)
	numbers[0] = 10
	numbers[1] = 20
	numbers[2] = 30

	// Третий индекс ограничивает вместимость результирующего среза.
	limited := numbers[0:2:2]

	fmt.Println("Original:", numbers)
	fmt.Println("Limited:", limited)
	fmt.Println("Limited length:", len(limited))
	fmt.Println("Limited capacity:", cap(limited))

	// Поскольку у limited нет свободного,
	// append создает новый массив.
	limited = append(limited, 99)

	fmt.Println("After append:", limited)
	fmt.Println("Original remains:", numbers)
}

func rangeSlices() {
	numbers := []int{10, 20, 30}

	// Переменная value получает копию каждого элемента.
	for _, value := range numbers {
		value *= 2
	}

	fmt.Println("After changing range value:", numbers)

	//Используйте индекс для изменения исходного слайса.
	for i := range numbers {
		numbers[i] *= 2
	}

	fmt.Println("After changing by index:", numbers)
}

func mapsBasics() {
	// map.
	ages := map[string]int{
		"Alice": 25,
		"Bob":   30,
	}

	fmt.Println("Map:", ages)

	// Получить значение.
	fmt.Println("Alice:", ages["Alice"])

	// Идиома , ok позволяет выявить отсутствие ключа
	age, exists := ages["Charlie"]

	fmt.Println("Charlie age:", age)
	fmt.Println("Charlie exists:", exists)

	// Добавить или обновить
	ages["Charlie"] = 35
	ages["Alice"] = 26

	fmt.Println("Updated map:", ages)

	// Удаление ключа.
	delete(ages, "Bob")

	fmt.Println("After delete:", ages)

	// Из нил мапы можно читать, но запись вызовет панику.
	var nilMap map[string]int

	fmt.Println("Nil map:", nilMap)
	fmt.Println("Missing key:", nilMap["unknown"])

	// При создании через make можно писать.
	initialized := make(map[string]int)
	initialized["one"] = 1

	fmt.Println("Initialized map:", initialized)

	// Порядок итерации по мапе случайный.
	for key, value := range ages {
		fmt.Println(key, value)
	}
}

func stringsBasics() {
	text := "Привет, Go!"

	fmt.Println("String:", text)

	// len считает байты, не символы юникода.
	fmt.Println("Bytes:", len(text))

	// Колличество рун.
	fmt.Println("Runes:", utf8.RuneCountInString(text))

	// Строка может быть конвертирована в слайс байт.
	bytes := []byte(text)

	fmt.Println("Bytes:", bytes)

	// rune представляет Unicode.
	runes := []rune(text)

	fmt.Println("Runes:", runes)

	// Range по строке:
	// byte index и rune.
	for index, char := range text {
		fmt.Printf("Byte index: %d, rune: %c\n", index, char)
	}

	// Строки иммутбельны
	// Конвертируем в []byte что бы изменить байты.
	mutable := []byte("hello")
	mutable[0] = 'H'

	fmt.Println("Modified bytes:", string(mutable))
}
