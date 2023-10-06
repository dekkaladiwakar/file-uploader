# File Uploading System

This system allows users to upload large files in a chunked manner through parallel processing.

## Server Used

- Go (Golang)

## Running the Go Server

go run main.go

## Frontend

index.html

## Endpoint

### Upload File Chunks

- **URL**: `/upload`
- **Method**: `POST`
- **Form Data**:
  - `file`: The file chunk.
  - `chunkNumber`: The number of the current chunk.
  - `totalChunks`: The total number of chunks.
  - `fileID`: An identifier for the file being uploaded.

**Example Request**:

```bash
curl -X POST \
  http://localhost:8080/upload \
  -H 'content-type: multipart/form-data' \
  -F file=@path/to/chunk \
  -F chunkNumber=1 \
  -F totalChunks=5 \
  -F fileID='example-file'
```
