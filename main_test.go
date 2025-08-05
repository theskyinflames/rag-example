package main

import (
	"strings"
	"testing"
)

func TestVectorStore_RetrieveTopK(t *testing.T) {
	tests := []struct {
		name     string
		docs     []document
		queryVec []float32
		k        int
		expected int // expected number of documents returned
	}{
		{
			name: "retrieve top 2 from 3 documents",
			docs: []document{
				{text: "doc1", embedding: []float32{1.0, 0.0, 0.0}},
				{text: "doc2", embedding: []float32{0.0, 1.0, 0.0}},
				{text: "doc3", embedding: []float32{1.0, 1.0, 0.0}},
			},
			queryVec: []float32{1.0, 0.0, 0.0},
			k:        2,
			expected: 2,
		},
		{
			name: "k larger than available documents",
			docs: []document{
				{text: "doc1", embedding: []float32{1.0, 0.0}},
				{text: "doc2", embedding: []float32{0.0, 1.0}},
			},
			queryVec: []float32{1.0, 0.0},
			k:        5,
			expected: 2,
		},
		{
			name:     "empty vector store",
			docs:     []document{},
			queryVec: []float32{1.0, 0.0},
			k:        3,
			expected: 0,
		},
		{
			name: "k is zero",
			docs: []document{
				{text: "doc1", embedding: []float32{1.0, 0.0}},
			},
			queryVec: []float32{1.0, 0.0},
			k:        0,
			expected: 0,
		},
		{
			name: "single document",
			docs: []document{
				{text: "only doc", embedding: []float32{1.0, 0.0, 0.0}},
			},
			queryVec: []float32{1.0, 0.0, 0.0},
			k:        1,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vs := newVectorStore()
			for _, doc := range tt.docs {
				vs.add(doc)
			}

			result := vs.retrieveTopK(tt.queryVec, tt.k)

			if len(result) != tt.expected {
				t.Errorf("retrieveTopK() returned %d documents, expected %d", len(result), tt.expected)
			}

			// Verify that results are sorted by similarity (highest first)
			if len(result) > 1 {
				for i := 0; i < len(result)-1; i++ {
					score1 := cosineSim(tt.queryVec, result[i].embedding)
					score2 := cosineSim(tt.queryVec, result[i+1].embedding)
					if score1 < score2 {
						t.Errorf("Results not properly sorted: score[%d]=%.3f < score[%d]=%.3f", i, score1, i+1, score2)
					}
				}
			}
		})
	}
}

func TestVectorStore_RetrieveTopK_OrderCorrectness(t *testing.T) {
	vs := newVectorStore()

	// Add documents with known similarity scores to query vector [1.0, 0.0]
	vs.add(document{text: "perfect match", embedding: []float32{1.0, 0.0}}) // cosine sim = 1.0
	vs.add(document{text: "orthogonal", embedding: []float32{0.0, 1.0}})    // cosine sim = 0.0
	vs.add(document{text: "partial match", embedding: []float32{0.5, 0.5}}) // cosine sim ≈ 0.707

	queryVec := []float32{1.0, 0.0}
	result := vs.retrieveTopK(queryVec, 3)

	expectedOrder := []string{"perfect match", "partial match", "orthogonal"}
	for i, doc := range result {
		if doc.text != expectedOrder[i] {
			t.Errorf("Expected document %d to be '%s', got '%s'", i, expectedOrder[i], doc.text)
		}
	}
}

func TestExtractTextFromPDF(t *testing.T) {
	tests := []struct {
		name        string
		pdfContent  []byte
		expectError bool
		expectEmpty bool
	}{
		{
			name:        "invalid PDF content",
			pdfContent:  []byte("not a pdf"),
			expectError: true,
			expectEmpty: false,
		},
		{
			name:        "empty content",
			pdfContent:  []byte{},
			expectError: true,
			expectEmpty: false,
		},
		{
			name:        "nil content",
			pdfContent:  nil,
			expectError: true,
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractTextFromPDF(tt.pdfContent)

			if tt.expectError && err == nil {
				t.Errorf("extractTextFromPDF() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("extractTextFromPDF() unexpected error: %v", err)
			}
			if tt.expectEmpty && result != "" {
				t.Errorf("extractTextFromPDF() expected empty result but got: %s", result)
			}
		})
	}
}

func TestExtractTextFromPDF_TextCleaning(t *testing.T) {
	// Test the text cleaning functionality with embedded PDF
	if len(pdfContent) == 0 {
		t.Skip("No embedded PDF content available for testing")
	}

	result, err := extractTextFromPDF(pdfContent)
	if err != nil {
		t.Fatalf("extractTextFromPDF() failed with embedded PDF: %v", err)
	}

	// Verify text cleaning
	if strings.Contains(result, "\x00") {
		t.Errorf("Result contains null characters that should have been cleaned")
	}
	if strings.Contains(result, "\ufffd") {
		t.Errorf("Result contains replacement characters that should have been cleaned")
	}
	if strings.Contains(result, "\r\n") {
		t.Errorf("Result contains \\r\\n that should have been normalized to \\n")
	}
	if strings.Contains(result, "\r") {
		t.Errorf("Result contains \\r that should have been normalized to \\n")
	}

	// Basic sanity check - result should not be empty for a real PDF
	if strings.TrimSpace(result) == "" {
		t.Errorf("Expected non-empty result from PDF extraction")
	}
}

func TestChunkText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxLen   int
		expected []string
	}{
		{
			name:     "empty text",
			text:     "",
			maxLen:   10,
			expected: []string{},
		},
		{
			name:     "single word within limit",
			text:     "hello",
			maxLen:   10,
			expected: []string{"hello"},
		},
		{
			name:     "single word exceeding limit",
			text:     "verylongword",
			maxLen:   5,
			expected: []string{"verylongword"},
		},
		{
			name:     "multiple words within single chunk",
			text:     "hello world",
			maxLen:   20,
			expected: []string{"hello world"},
		},
		{
			name:     "multiple words requiring multiple chunks",
			text:     "this is a test of chunking functionality",
			maxLen:   10,
			expected: []string{"this is a", "test of", "chunking", "functionality"},
		},
		{
			name:     "exact boundary case",
			text:     "word1 word2",
			maxLen:   11, // exactly "word1 word2" length
			expected: []string{"word1 word2"},
		},
		{
			name:     "words with extra whitespace",
			text:     "  hello   world   test  ",
			maxLen:   10,
			expected: []string{"hello", "world test"}, // "hello world" is 11 chars > 10
		},
		{
			name:     "single character words",
			text:     "a b c d e f g",
			maxLen:   5,
			expected: []string{"a b c", "d e f", "g"},
		},
		{
			name:     "maxLen of 1",
			text:     "a b c",
			maxLen:   1,
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "maxLen of 0",
			text:     "hello world",
			maxLen:   0,
			expected: []string{"hello", "world"},
		},
		{
			name:     "negative maxLen",
			text:     "hello world",
			maxLen:   -5,
			expected: []string{"hello", "world"},
		},
		{
			name:     "text with newlines and tabs",
			text:     "hello\nworld\ttest",
			maxLen:   15,
			expected: []string{"hello world", "test"}, // "hello world test" is 16 chars > 15
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := chunkText(tt.text, tt.maxLen)

			if len(result) != len(tt.expected) {
				t.Errorf("chunkText() returned %d chunks, expected %d", len(result), len(tt.expected))
				t.Errorf("Got: %v", result)
				t.Errorf("Expected: %v", tt.expected)
				return
			}

			for i, chunk := range result {
				if chunk != tt.expected[i] {
					t.Errorf("chunkText() chunk %d = %q, expected %q", i, chunk, tt.expected[i])
				}
			}

			// Verify that no chunk exceeds maxLen (except for single words longer than maxLen)
			if tt.maxLen > 0 {
				for i, chunk := range result {
					words := strings.Fields(chunk)
					if len(words) > 1 && len(chunk) > tt.maxLen {
						t.Errorf("chunk %d exceeds maxLen: %d > %d, chunk: %q", i, len(chunk), tt.maxLen, chunk)
					}
				}
			}
		})
	}
}

func TestChunkText_LengthAccounting(t *testing.T) {
	// Test that the function properly counts character lengths
	text := "abc def ghi"
	maxLen := 7 // Should fit "abc def" (7 chars)

	result := chunkText(text, maxLen)
	expected := []string{"abc def", "ghi"}

	if len(result) != len(expected) {
		t.Errorf("Expected %d chunks, got %d", len(expected), len(result))
	}

	for i, chunk := range result {
		if i < len(expected) && chunk != expected[i] {
			t.Errorf("Chunk %d: expected %q, got %q", i, expected[i], chunk)
		}
	}
}

func TestChunkText_PreservesWordBoundaries(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog"
	chunks := chunkText(text, 15)

	// Verify that when we join all chunks back, we get the original text
	rejoined := strings.Join(chunks, " ")
	if rejoined != text {
		t.Errorf("Chunking and rejoining changed the text:\nOriginal: %q\nRejoined: %q", text, rejoined)
	}

	// Verify that no chunk contains partial words (words are never split)
	originalWords := strings.Fields(text)
	var chunkWords []string
	for _, chunk := range chunks {
		chunkWords = append(chunkWords, strings.Fields(chunk)...)
	}

	if len(originalWords) != len(chunkWords) {
		t.Errorf("Word count mismatch: original %d, chunked %d", len(originalWords), len(chunkWords))
	}

	for i, word := range originalWords {
		if i < len(chunkWords) && word != chunkWords[i] {
			t.Errorf("Word %d: expected %q, got %q", i, word, chunkWords[i])
		}
	}
}
