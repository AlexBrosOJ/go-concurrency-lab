package main

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestRenameUsesOriginalAndPreservesValueCopy(t *testing.T) {
	project := Project{name: "До"}
	var copied Namer = project
	var shared Namer = &project
	if got := renameProject(&project, "После"); got != "После" || shared.Name() != "После" || copied.Name() != "До" {
		t.Fatalf("result=%q pointer=%q copy=%q", got, shared.Name(), copied.Name())
	}
}

func TestTypedNilInsideInterface(t *testing.T) {
	var pointer *Project
	var boxed Namer = pointer
	got, ok := boxed.(*Project)
	if boxed == nil || !ok || got != nil {
		t.Fatalf("interface=%v assertion=%t pointer=%v", boxed, ok, got)
	}
}

type failingReader struct {
	err error
}

func (reader failingReader) Read([]byte) (int, error) {
	return 0, reader.err
}

func TestReadTextAcceptsDifferentReaders(t *testing.T) {
	cause := errors.New("read failed")
	reader := io.MultiReader(strings.NewReader("готово"), failingReader{err: cause})
	if got, err := readText(reader); got != "готово" || !errors.Is(err, cause) {
		t.Fatalf("got (%q, %v), want partial text and original error", got, err)
	}
	if got, err := readText(strings.NewReader("")); got != "" || err != nil {
		t.Fatalf("empty reader returned (%q, %v)", got, err)
	}
}
