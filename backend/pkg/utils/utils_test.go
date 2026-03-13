package utils

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal_file.txt", "normal_file.txt"},
		{"file/with\\slash.txt", "filewithslash.txt"},
		{"file:with*special?chars.txt", "filewithspecialchars.txt"},
		{"", "unknown"},
		{"   ", "unknown"},
	}

	for _, test := range tests {
		result := SanitizeFilename(test.input)
		if result != test.expected {
			t.Errorf("SanitizeFilename(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestChunkSlice(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	chunks := ChunkSlice(slice, 3)

	if len(chunks) != 4 {
		t.Errorf("Expected 4 chunks, got %d", len(chunks))
	}

	if len(chunks[0]) != 3 {
		t.Errorf("Expected first chunk size 3, got %d", len(chunks[0]))
	}

	if len(chunks[3]) != 1 {
		t.Errorf("Expected last chunk size 1, got %d", len(chunks[3]))
	}
}

func TestUnique(t *testing.T) {
	slice := []int{1, 2, 3, 3, 3, 4, 5, 5}
	result := Unique(slice)

	expected := []int{1, 2, 3, 4, 5}
	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Expected %d at index %d, got %d", v, i, result[i])
		}
	}
}
