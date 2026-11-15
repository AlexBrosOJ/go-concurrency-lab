package main

import "fmt"

// Task объединяет название задачи и её состояние.
// Нулевое значение — задача без названия, которая ещё не выполнена.
type Task struct {
	Title string
	Done  bool
}

// Status читает состояние через получатель-значение.
func (t Task) Status() string {
	if t.Done {
		return "done"
	}
	return "pending"
}

// WithTitle возвращает изменённую копию, сохраняя исходную задачу.
func (t Task) WithTitle(title string) Task {
	t.Title = title
	return t
}

// Complete меняет исходную задачу через указатель. Получатель должен быть не nil.
func (t *Task) Complete() {
	t.Done = true
}

func main() {
	fmt.Println("=== 1. Zero value ===")
	var empty Task
	fmt.Printf("Title: %q, status: %s\n", empty.Title, empty.Status())

	fmt.Println("\n=== 2. Struct and value copy ===")
	original := Task{Title: "Learn structs"}
	copied := original
	copied.Title = "Learn methods"
	fmt.Println("Original:", original.Title, original.Status())
	fmt.Println("Copy:", copied.Title, copied.Status())

	fmt.Println("\n=== 3. Value receiver ===")
	renamed := original.WithTitle("Practice receivers")
	fmt.Println("Original:", original.Title)
	fmt.Println("Returned copy:", renamed.Title)

	fmt.Println("\n=== 4. Pointer receiver ===")
	pointer := &original
	pointer.Complete()
	fmt.Println("Original:", original.Status())
	fmt.Println("Copy:", copied.Status())

	// Для переменной Go позволяет вызвать метод с *Task без явного &.
	copied.Complete()
	fmt.Println("Copy after Complete:", copied.Status())
}
