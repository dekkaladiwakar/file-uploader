package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/rs/cors"
)

var mu sync.Mutex

func UploadChunkHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // limit your max memory size
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the chunk data, chunk number, total chunks, and file ID from the request
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	chunkNumber := r.FormValue("chunkNumber")
	totalChunks := r.FormValue("totalChunks")
	fileID := r.FormValue("fileID")

	// Create a unique filename for this chunk
	chunkFilename := fmt.Sprintf("%s-chunk%s", fileID, chunkNumber)

	// Protect concurrent writes with a mutex
	mu.Lock()
	defer mu.Unlock()

	// Create a new file to store the chunk
	dst, err := os.Create(chunkFilename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Write the chunk data to the file
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Chunk %s of %s for file %s uploaded successfully", chunkNumber, totalChunks, fileID)
}

func main() {
	// Create a new router
	router := http.NewServeMux()

	// Register the chunk upload handler
	router.HandleFunc("/upload", UploadChunkHandler)

	// Set up CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Allow all origins
		AllowedMethods: []string{"POST"},
	})

	log.Println("Server is starting on Port 8080.....")
	// Start the server with CORS middleware
	handler := c.Handler(router)
	http.ListenAndServe(":8080", handler)

}
