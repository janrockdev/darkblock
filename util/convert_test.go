package util

import (
	"testing"
)

func TestStringToHex(t *testing.T) {
	tests := []struct {
		input    string
		expected []byte
		hasError bool
	}{
		{"48656c6c6f", []byte("Hello"), false},
		{"4a6f686e", []byte("John"), false},
		{"invalidhex", nil, true},
	}

	for _, test := range tests {
		result, err := StringToHex(test.input)
		if (err != nil) != test.hasError {
			t.Errorf("StringToHex(%s) error = %v, wantErr %v", test.input, err, test.hasError)
			continue
		}
		if !test.hasError && string(result) != string(test.expected) {
			t.Errorf("StringToHex(%s) = %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestStringToHexInt(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
		hasError bool
	}{
		{"48656c6c6f", []int{72, 101, 108, 108, 111}, false},
		{"4a6f686e", []int{74, 111, 104, 110}, false},
		{"invalidhex", nil, true},
	}

	for _, test := range tests {
		result, err := StringToHexInt(test.input)
		if (err != nil) != test.hasError {
			t.Errorf("StringToHexInt(%s) error = %v, wantErr %v", test.input, err, test.hasError)
			continue
		}
		if !test.hasError {
			for i, v := range result {
				if v != test.expected[i] {
					t.Errorf("StringToHexInt(%s) = %v, want %v", test.input, result, test.expected)
					break
				}
			}
		}
	}
}
