package util

import (
	"testing"
)

func TestRandomHash(t *testing.T) {
	hash := RandomHash()
	if len(hash) != 32 {
		t.Errorf("Expected hash length of 32, but got %d", len(hash))
	}
}

func TestRandomBlock(t *testing.T) {
	block := RandomBlock()
	if block == nil {
		t.Fatal("Expected non-nil block")
	}
	if block.Header == nil {
		t.Fatal("Expected non-nil block header")
	}
	if block.Header.Version != 1 {
		t.Errorf("Expected version 1, but got %d", block.Header.Version)
	}
	if block.Header.Height < 0 || block.Header.Height >= 1000 {
		t.Errorf("Expected height between 0 and 999, but got %d", block.Header.Height)
	}
	if len(block.Header.PrevHash) != 32 {
		t.Errorf("Expected PrevHash length of 32, but got %d", len(block.Header.PrevHash))
	}
	if len(block.Header.RootHash) != 32 {
		t.Errorf("Expected RootHash length of 32, but got %d", len(block.Header.RootHash))
	}
	if block.Header.Timestamp <= 0 {
		t.Errorf("Expected positive timestamp, but got %d", block.Header.Timestamp)
	}
}
