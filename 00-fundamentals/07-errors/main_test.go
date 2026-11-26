package main

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestLoadPortAcceptsBoundaries(t *testing.T) {
	for _, test := range []struct {
		raw  string
		want int
	}{{" 1 ", 1}, {"65535", 65535}} {
		if got, err := loadPort("api", test.raw); got != test.want || err != nil {
			t.Fatalf("loadPort(%q) = (%d, %v)", test.raw, got, err)
		}
	}
}

func TestLoadPortPreservesErrorChain(t *testing.T) {
	_, err := loadPort("api", " ")
	if !errors.Is(err, errEmptyPort) || !strings.Contains(err.Error(), "api") {
		t.Fatalf("missing cause or service context: %v", err)
	}
	for _, bad := range []int{-1, 0, 65536} {
		_, err := loadPort("api", strconv.Itoa(bad))
		var target *PortRangeError
		if !errors.As(err, &target) || target.Port != bad {
			t.Fatalf("port %d returned %v", bad, err)
		}
	}
	for _, bad := range []string{"abc", "999999999999999999999"} {
		_, err := loadPort("api", bad)
		var target *strconv.NumError
		if !errors.As(err, &target) || target.Num != bad {
			t.Fatalf("input %q returned %v", bad, err)
		}
	}
}

func TestTryDefaultPortRecoversKnownPanic(t *testing.T) {
	if port, err := tryDefaultPort(""); port != 0 || !errors.Is(err, errEmptyPort) {
		t.Fatalf("got (%d, %v), want a recovered configuration error", port, err)
	}
	if port, err := tryDefaultPort("80"); port != 80 || err != nil {
		t.Fatalf("got (%d, %v), want (80, nil)", port, err)
	}
}
