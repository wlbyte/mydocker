package utils

import (
	"testing"
)

func TestHashStr(t *testing.T) {
	// Test normal case
	str, err := HashStr("test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expectedHash := "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2"
	if str != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, str)
	}
	// test struct case
	c := struct {
		Name   string
		Volume string
	}{
		Name:   "test",
		Volume: "/root",
	}
	str, err = HashStr(c)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expectedHash = "62b4ebe29ee56fb9e603f774dd813b6127f4bbe4674baf69a5e04972d8da7ef5"
	if str != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, str)
	}
	// Test nil case
	str, err = HashStr(nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expectedHash = "cc925a337413bc197145439db5efb5f1ca846581cd25a184101c94bef41f0db2"
	if str != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, str)
	}
}