package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/rs/cors"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // limit your max memory size
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the file data, and thread count from the request
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	threadCountStr := r.FormValue("threads")
	threadCount, err := strconv.Atoi(threadCountStr)
	if err != nil || threadCount <= 0 {
		http.Error(w, "Invalid thread count", http.StatusBadRequest)
		return
	}

	fileID := r.FormValue("fileID")

	if fileID == "" {
		http.Error(w, "FileId is missing", http.StatusBadRequest)
		return
	}

	// Get the file size to determine chunk size
	fileSize := r.ContentLength
	chunkSize := fileSize / int64(threadCount)

	var wg sync.WaitGroup

	for i := 0; i < threadCount; i++ {
		wg.Add(1)

		go func(threadNum int) {
			defer wg.Done()

			// Determine the byte range for this chunk
			startByte := int64(threadNum) * chunkSize
			endByte := startByte + chunkSize - 1
			if threadNum == threadCount-1 {
				endByte = fileSize // ensure the last chunk goes to the end of the file
			}

			// Create a unique filename for this chunk
			chunkFilename := fmt.Sprintf("%s-chunk%d", fileID, threadNum)

			// Create a new file to store the chunk
			dst, err := os.Create(chunkFilename)
			if err != nil {
				log.Println(err) // Log the error and continue
				return
			}
			defer dst.Close()

			// Seek to the start byte and read the chunk data from the file
			file.Seek(startByte, 0)
			buf := make([]byte, endByte-startByte+1)
			file.Read(buf)

			// Write the chunk data to the file
			dst.Write(buf)

		}(i)
	}

	wg.Wait() // Wait for all goroutines to finish

	fmt.Fprintf(w, "File %s uploaded and processed successfully with %d threads", fileID, threadCount)
}

func main() {
	// Create a new router
	router := http.NewServeMux()

	// Register the chunk upload handler
	router.HandleFunc("/upload", UploadHandler)

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
