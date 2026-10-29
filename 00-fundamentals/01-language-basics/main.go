package main

import "fmt"

func main() {
	//Переменные и нулевые значения

	var age int
	var name string
	var isDev bool

	fmt.Println("Zero values:")
	fmt.Println(age, name, isDev)

	age = 25
	name = "Oleg"
	isDev = true

	fmt.Println("Assigned values:")
	fmt.Println(age, name, isDev)

	//Краткое обьявление

	c := "City"
	yearsOfExpirience := 5

	fmt.Println(c, yearsOfExpirience)

	// Константы

	const appName = "Go Conurrency Lab"

	fmt.Println("App:", appName)

	//Базовые типы и приведение типов

	var count int = 10
	var ratio float64 = float64(count) / 4

	fmt.Println("Count, Ratio:", count, ratio)

	//Операторы

	a := 10
	b := 3

	fmt.Println("Addition:", a+b)
	fmt.Println("Subtraction:", a-b)
	fmt.Println("Multiplication:", a*b)
	fmt.Println("Division:", a/b)
	fmt.Println("Remainder:", a%b)

	//Условия
	if age >= 18 {
		fmt.Println("Adult")
	} else {
		fmt.Println("Underage")
	}

	// Циклы

	for i := 0; i < 10; i++ {
		fmt.Println("Iteration:", i)
	}
}
