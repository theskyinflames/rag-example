# RAG Example with Go

A Retrieval-Augmented Generation (RAG) application that processes PDF documents and answers questions using OpenAI's API.

## Prerequisites

- Docker installed
- OpenAI API key

## Quick Start

### Using Make (Recommended)

For easier management, use the provided Makefile:

1. **Set your OpenAI API key:**

   ```bash
   export OPENAI_API_KEY="your-openai-api-key-here"
   ```

2. **Build and run:**

   ```bash
   make quick-start
   ```

3. **Or build separately:**

   ```bash
   make build
   make run
   ```

4. **Run interactively:**

   ```bash
   make run-interactive
   ```

### Local Development (without Docker)

If you want to run locally without Docker:

```bash
export OPENAI_API_KEY="your-openai-api-key-here"
go run main.go
```

### Rebuilding

After making code changes:

```bash
make build
make run
```

## Configuration

The application uses the following environment variables:

- `OPENAI_API_KEY`: Your OpenAI API key (required)

## Features

- PDF text extraction using embedded PDF file
- Text chunking for optimal processing
- Vector embeddings using OpenAI's text-embedding-ada-002
- Semantic search with cosine similarity
- Question answering using GPT-4

## Troubleshooting

### Common Issues

1. **Missing API Key Error:**

   ```text
   Missing OPENAI_API_KEY env var
   ```

   Solution: Ensure your OpenAI API key is set as an environment variable.

2. **PDF Text Corruption:**
   If you see "heavily corrupted text" messages, the PDF extraction may have issues. Check the debug output for text quality.

3. **Memory Issues:**
   For large PDFs, you may need to increase Docker memory limits or optimize chunk sizes.

### Debug Mode

The application includes debug output that shows:

- Extracted text samples
- Number of chunks created
- Retrieved chunks for each query

### Available Make Commands

Run `make help` to see all available commands:

```text
build                Build the Docker image
run                  Run the container with Docker
run-interactive      Run the container interactively
run-detached         Run the container in detached mode
dev                  Run the application locally (requires Go)
test                 Run tests
clean                Remove the Docker image
quick-start          Build and run the application quickly
```

## Architecture

```text
PDF → Text Extraction → Chunking → Embeddings → Vector Store → Similarity Search → LLM → Answer
```

## License

MIT License
