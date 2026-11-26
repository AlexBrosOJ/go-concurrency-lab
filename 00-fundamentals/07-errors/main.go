package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var errEmptyPort = errors.New("порт не указан")

type PortRangeError struct {
	Port int
}

func (err *PortRangeError) Error() string {
	return fmt.Sprintf("порт %d вне диапазона 1–65535", err.Port)
}

func parsePort(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errEmptyPort
	}
	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("разобрать число %q: %w", raw, err)
	}
	if port < 1 || port > 65535 {
		return 0, &PortRangeError{Port: port}
	}
	return port, nil
}

func loadPort(service, raw string) (int, error) {
	port, err := parsePort(raw)
	if err != nil {
		return 0, fmt.Errorf("настройки сервиса %q: %w", service, err)
	}
	return port, nil
}

type configPanic struct {
	err error
}

// Внутренний пример panic. В обычном разборе настроек достаточно вернуть ошибку.
func mustLoadDefaultPort(raw string) int {
	port, err := loadPort("default", raw)
	if err != nil {
		panic(configPanic{err: err})
	}
	return port
}

func tryDefaultPort(raw string) (port int, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if known, ok := recovered.(configPanic); ok {
				err = known.err
				return
			}
			// Неизвестная паника должна уйти дальше с исходным значением.
			panic(recovered)
		}
	}()
	return mustLoadDefaultPort(raw), nil
}

func main() {
	fmt.Println("=== 1. Обычная проверка error ===")
	port, err := loadPort("api", " 8080 ")
	fmt.Printf("Порт: %d, ошибка: %v\n", port, err)

	fmt.Println("\n=== 2. Поиск известной ошибки через errors.Is ===")
	_, err = loadPort("api", "")
	fmt.Println("Причина — пустой порт:", errors.Is(err, errEmptyPort))

	fmt.Println("\n=== 3. Данные ошибки через errors.As ===")
	_, err = loadPort("api", "70000")
	var rangeErr *PortRangeError
	if errors.As(err, &rangeErr) {
		fmt.Println("Недопустимый порт:", rangeErr.Port)
	}
	_, err = loadPort("api", "abc")
	var numberErr *strconv.NumError
	if errors.As(err, &numberErr) {
		fmt.Println("Не удалось разобрать:", numberErr.Num)
	}

	fmt.Println("\n=== 4. Локальная граница panic/recover ===")
	_, err = tryDefaultPort("")
	fmt.Println("Известная паника преобразована в ошибку:", errors.Is(err, errEmptyPort))
}
