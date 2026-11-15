package main

import "testing"

func TestTaskStatus(t *testing.T) {
	tests := []struct {
		name string
		task Task
		want string
	}{
		{name: "zero value", want: "pending"},
		{name: "named pending task", task: Task{Title: "Read"}, want: "pending"},
		{name: "completed task", task: Task{Title: "Read", Done: true}, want: "done"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := tt.task
			if got := tt.task.Status(); got != tt.want {
				t.Fatalf("Status() = %q, want %q", got, tt.want)
			}
			if tt.task != before {
				t.Fatal("Status changed the task")
			}
		})
	}
}

func TestTaskWithTitle(t *testing.T) {
	for _, title := range []string{"New title", "", "Изучить Go"} {
		t.Run(title, func(t *testing.T) {
			original := Task{Title: "Original", Done: true}
			before := original
			got := original.WithTitle(title)
			if want := (Task{Title: title, Done: true}); got != want {
				t.Fatalf("WithTitle() = %+v, want %+v", got, want)
			}
			if original != before {
				t.Fatal("WithTitle changed the original")
			}
		})
	}
}

func TestTaskComplete(t *testing.T) {
	for _, initial := range []Task{{}, {Title: "Read"}, {Title: "Read", Done: true}} {
		t.Run(initial.Title+"/"+initial.Status(), func(t *testing.T) {
			task := initial
			copied := task
			pointer := &task
			pointer.Complete()
			pointer.Complete() // Повторное завершение сохраняет результат.
			if want := (Task{Title: initial.Title, Done: true}); task != want {
				t.Fatalf("after Complete: %+v, want %+v", task, want)
			}
			if copied != initial {
				t.Fatal("completing the original changed its value copy")
			}
		})
	}
}
