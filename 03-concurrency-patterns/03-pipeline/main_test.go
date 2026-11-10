package main

import (
	"reflect"
	"testing"
)

func TestGenerate(t *testing.T) {
	var got []int
	for value := range generate(4) {
		got = append(got, value)
	}

	want := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestSquare(t *testing.T) {
	in := make(chan int, 3)
	in <- 2
	in <- 3
	in <- 4
	close(in)

	var got []int
	for value := range square(in) {
		got = append(got, value)
	}

	want := []int{4, 9, 16}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPipeline(t *testing.T) {
	var got []string
	for result := range format(square(generate(3))) {
		got = append(got, result)
	}

	want := []string{"квадрат: 1", "квадрат: 4", "квадрат: 9"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
