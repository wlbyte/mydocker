package network

import (
	"testing"
)

func TestCharGet(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		err      bool
	}{
		{
			name:     "first character is 0",
			input:    "0abc",
			expected: 0,
			err:      false,
		},
		{
			name:     "middle character is 0",
			input:    "abc0def",
			expected: 3,
			err:      false,
		},
		{
			name:     "last character is 0",
			input:    "abcdef0",
			expected: 6,
			err:      false,
		},
		{
			name:     "multiple 0s, return first occurrence",
			input:    "a0b0c0",
			expected: 1,
			err:      false,
		},
		{
			name:     "no 0 in string",
			input:    "abcdef",
			expected: 0,
			err:      true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			err:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetChar(&tt.input, '0')
			if (err != nil) != tt.err {
				t.Errorf("CharGet() error = %v, wantErr %v", err, tt.err)
				return
			}
			if !tt.err && got != uint(tt.expected) {
				t.Errorf("CharGet() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCharSet(t *testing.T) {
	tests := []struct {
		name     string
		n        uint
		s        string
		expected string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "index equal to length",
			n:        3,
			s:        "000",
			expected: "",
			wantErr:  true,
			errMsg:   "bitStr.Set: index 3 out of range [0-2]",
		},
		{
			name:     "index greater than length",
			n:        5,
			s:        "000",
			expected: "",
			wantErr:  true,
			errMsg:   "bitStr.Set: index 5 out of range [0-2]",
		},
		{
			name:     "set first bit",
			n:        0,
			s:        "000",
			expected: "100",
			wantErr:  false,
		},
		{
			name:     "set middle bit",
			n:        1,
			s:        "000",
			expected: "010",
			wantErr:  false,
		},
		{
			name:     "set last bit",
			n:        2,
			s:        "000",
			expected: "001",
			wantErr:  false,
		},
		{
			name:     "bit already set",
			n:        1,
			s:        "010",
			expected: "010",
			wantErr:  false,
		},
		{
			name:     "empty string",
			n:        0,
			s:        "",
			expected: "",
			wantErr:  true,
			errMsg:   "bitStr.Set: index 0 out of range [0--1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.s
			err := SetChar(tt.n, &s, '1')

			if (err != nil) != tt.wantErr {
				t.Errorf("CharSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("CharSet() error message = %v, want %v", err.Error(), tt.errMsg)
			}

			if !tt.wantErr && s != tt.expected {
				t.Errorf("CharSet() = %v, want %v", s, tt.expected)
			}
		})
	}
}
