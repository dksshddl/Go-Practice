package main

import (
	"testing"
)

func TestPath(t *testing.T) {
	s := NewPath("hello/world/123")
	if s.Path != "hello/world" {
		t.Errorf("path does not match, expected path is %v but got %v", "hello/world", s.Path)
	}
	if s.Id != "123" {
		t.Errorf("id does not match, expected id is %v but got %v", "123", s.Id)
	}

	if !s.HasID() {
		t.Errorf("has id function err")
	}

	s = NewPath("hello/world")
	if s.HasID() {
		t.Errorf("has id function err")
	}
}
