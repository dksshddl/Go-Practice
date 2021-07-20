package main

import (
	"testing"
)

func TestPath(t *testing.T) {
	s := NewPath("hello/world/123")
	t.Log(s)
	t.Fail()
}

func TestSettings(t *testing.T) {
	t.Log(s)
}
