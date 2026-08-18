package main

import "testing"

func TestEncodeDropsUnmappable(t *testing.T) {
	s := &session{cp: codepages["cp850"]}
	if got := string(s.encode("a")); got != "a" {
		t.Fatalf("ascii: got %q", got)
	}
	if got := s.encode("ç"); len(got) != 1 {
		t.Fatalf("cedilla should map to one cp850 byte, got %v", got)
	}
	if got := s.encode("a→b"); string(got) != "ab" {
		t.Fatalf("unmappable rune should be dropped, got %q", got)
	}
}

func TestIsEscKey(t *testing.T) {
	bare := make(chan rune)
	close(bare)
	if !isEscKey(bare) {
		t.Fatal("bare ESC should count as an Esc press")
	}

	seq := make(chan rune, 2)
	seq <- '['
	seq <- 'A'
	close(seq)
	if isEscKey(seq) {
		t.Fatal("arrow-key sequence should not count as an Esc press")
	}
}
