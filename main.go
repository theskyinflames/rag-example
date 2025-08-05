package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"

	_ "embed" // for embedding the PDF file

	"github.com/sashabaranov/go-openai"
	"rsc.io/pdf"
)

//go:embed the-egg.pdf
var pdfContent []byte

type document struct {
	text      string
	embedding []float32
}

type vectorStore struct {
	documents []document
}

func newVectorStore() *vectorStore {
	return &vectorStore{
		documents: make([]document, 0),
	}
}

func (vs *vectorStore) add(doc document) {
	vs.documents = append(vs.documents, doc)
}

func (vs *vectorStore) retrieveTopK(queryVec []float32, k int) []document {
	type scored struct {
		Doc   document
		Score float32
	}
	var scoredDocs []scored
	for _, doc := range vs.documents {
		s := cosineSim(queryVec, doc.embedding)
		scoredDocs = append(scoredDocs, scored{Doc: doc, Score: s})
	}
	sort.Slice(scoredDocs, func(i, j int) bool {
		return scoredDocs[i].Score > scoredDocs[j].Score
	})

	var top []document
	for i := 0; i < k && i < len(scoredDocs); i++ {
		top = append(top, scoredDocs[i].Doc)
	}
	return top
}

func main() {
	ctx := context.Background()
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Missing OPENAI_API_KEY env var")
	}

	// Configure client for OpenAI API
	client := openai.NewClient(apiKey)

	// Initialize vector store
	vectorStore := newVectorStore()

	// Step 1: Extract text from PDF
	text, err := extractTextFromPDF(pdfContent)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Extracted text from PDF: %s\n", text)

	// Step 2: Chunk text
	chunks := chunkText(text, 300)
	fmt.Printf("Created %d chunks\n", len(chunks))

	// Step 3: Embed and store
	i := 0
	for _, chunk := range chunks {
		i++
		fmt.Printf("Processing chunk %d/%d\n", i, len(chunks))
		vec, err := embedText(ctx, client, chunk)
		if err != nil {
			log.Fatal(err)
		}
		vectorStore.add(document{text: chunk, embedding: vec})
	}

	// Step 4: Get user query
	query := "What is the story about?"
	qVec, _ := embedText(ctx, client, query)

	// Step 5: Retrieve top-k chunks
	retrieved := vectorStore.retrieveTopK(qVec, 3)

	// Debug: Print retrieved chunks
	fmt.Println("=== RETRIEVED CHUNKS ===")
	for i, doc := range retrieved {
		fmt.Printf("Chunk %d: %s\n", i+1, doc.text)
		fmt.Println("---")
	}
	fmt.Println("=== END CHUNKS ===")
	fmt.Println()

	// Step 6: Send to LLM
	var contextBuilder strings.Builder
	for _, doc := range retrieved {
		contextBuilder.WriteString(doc.text)
		contextBuilder.WriteString("\n")
	}
	contextStr := contextBuilder.String()

	prompt := fmt.Sprintf("Use the context below to answer the question.\n\nContext:\n%s\n\nQuestion: %s", contextStr, query)
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4o, // OpenAI's GPT-4 model
		Messages: []openai.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Answer:", resp.Choices[0].Message.Content)
}

// --- UTILITIES ---
// extractTextFromPDF extracts text from a PDF file with better error handling and text cleaning.
func extractTextFromPDF(pdfContent []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(pdfContent), int64(len(pdfContent)))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		content := p.Content()
		for _, text := range content.Text {
			// Clean and filter text
			cleanText := strings.TrimSpace(text.S)
			if len(cleanText) > 0 {
				// Replace common problematic characters
				cleanText = strings.ReplaceAll(cleanText, "\x00", "")
				cleanText = strings.ReplaceAll(cleanText, "\ufffd", "") // replacement character

				// Decode Caesar cipher (shift back by 3)
				cleanText = decodeCaesarCipher(cleanText, 3)

				buf.WriteString(cleanText + " ")
			}
		}
		buf.WriteString("\n") // Add line break between pages
	}

	// Final cleaning
	result := buf.String()
	result = strings.ReplaceAll(result, "\r\n", "\n")
	result = strings.ReplaceAll(result, "\r", "\n")

	return result, nil
}

// decodeCaesarCipher decodes text that has been encoded with a Caesar cipher
func decodeCaesarCipher(text string, shift int) string {
	var result strings.Builder

	for _, char := range text {
		if char >= 'A' && char <= 'Z' {
			// Uppercase letters
			decoded := ((int(char-'A') - shift + 26) % 26) + int('A')
			result.WriteRune(rune(decoded))
		} else if char >= 'a' && char <= 'z' {
			// Lowercase letters
			decoded := ((int(char-'a') - shift + 26) % 26) + int('a')
			result.WriteRune(rune(decoded))
		} else {
			// Keep other characters unchanged (numbers, spaces, punctuation)
			result.WriteRune(char)
		}
	}

	return result.String()
}

// chunkText splits text into chunks of approximately maxLen characters.
func chunkText(text string, maxLen int) []string {
	if maxLen <= 0 {
		// For non-positive maxLen, return each word as separate chunk
		return strings.Fields(text)
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var chunks []string
	var buf []string

	for _, word := range words {
		// Try adding this word to current buffer
		testBuf := append(buf, word)
		testChunk := strings.Join(testBuf, " ")

		// If this would exceed maxLen and we have words in buffer, create chunk
		if len(testChunk) > maxLen && len(buf) > 0 {
			chunks = append(chunks, strings.Join(buf, " "))
			buf = []string{word} // Start new chunk with current word
		} else {
			buf = append(buf, word)
		}
	}

	// Add remaining words as final chunk
	if len(buf) > 0 {
		chunks = append(chunks, strings.Join(buf, " "))
	}

	return chunks
}

// embedText uses the OpenAI API to embed text and returns the embedding vector.
func embedText(ctx context.Context, client *openai.Client, input string) ([]float32, error) {
	resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.AdaEmbeddingV2, // OpenAI's text-embedding-ada-002 model
		Input: []string{input},
	})
	if err != nil {
		return nil, err
	}
	return resp.Data[0].Embedding, nil
}

// cosineSim calculates the cosine similarity between two vectors.
func cosineSim(a, b []float32) float32 {
	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	return dot / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}
