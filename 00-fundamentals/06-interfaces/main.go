package main

import (
	"fmt"
	"io"
	"strings"
)

type Namer interface {
	Name() string
}

type Renamer interface {
	Namer
	Rename(string)
}

type Project struct {
	name string
}

func (p Project) Name() string {
	return p.name
}

func (p *Project) Rename(name string) {
	p.name = name
}

// Компилятор проверяет соответствие типов интерфейсам.
var (
	_ Namer   = Project{}
	_ Renamer = (*Project)(nil)
)

func renameProject(project Renamer, name string) string {
	project.Rename(name)
	return project.Name()
}

func describe(value any) string {
	switch value := value.(type) {
	case nil:
		return "нет значения"
	case string:
		return "строка: " + value
	case Namer:
		return "имя: " + value.Name()
	default:
		return fmt.Sprintf("тип: %T", value)
	}
}

func readText(reader io.Reader) (string, error) {
	data, err := io.ReadAll(reader)
	return string(data), err
}

func main() {
	fmt.Println("=== 1. Значение и указатель ===")
	project := Project{name: "Go Lab"}
	var named Namer = project
	fmt.Println("После переименования:", renameProject(&project, "Go Workshop"))
	fmt.Println("Копия в интерфейсе:", named.Name())

	fmt.Println("\n=== 2. Утверждение типа ===")
	original, ok := named.(Project)
	fmt.Printf("Project: %s, найден: %t\n", original.Name(), ok)
	_, ok = named.(*Project)
	fmt.Println("В интерфейсе лежит *Project:", ok)

	fmt.Println("\n=== 3. Выбор по типу ===")
	for _, value := range []any{nil, "Go", project, 42} {
		fmt.Println(describe(value))
	}

	fmt.Println("\n=== 4. nil внутри интерфейса ===")
	var pointer *Project
	var empty Namer
	var boxed Namer = pointer
	fmt.Println("Пустой интерфейс == nil:", empty == nil)
	fmt.Println("Интерфейс с nil-указателем == nil:", boxed == nil)
	extracted, ok := boxed.(*Project)
	fmt.Printf("Извлечён *Project: %t, указатель == nil: %t\n", ok, extracted == nil)

	fmt.Println("\n=== 5. Интерфейс стандартной библиотеки ===")
	text, err := readText(strings.NewReader("читать через io.Reader"))
	fmt.Printf("Текст: %s, ошибка: %v\n", text, err)
}
