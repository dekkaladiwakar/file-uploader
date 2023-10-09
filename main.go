package main

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/rs/cors"
)

const maxWorkers = 10
const maxQueueSize = 1000

type ChunkRequest struct {
	File      multipart.File
	FileID    string
	ChunkNum  int
	StartByte int64
	EndByte   int64
}

var chunkQueue = make(chan ChunkRequest, maxQueueSize)

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for chunk := range chunkQueue {
		chunkFilename := fmt.Sprintf("%s-chunk%d", chunk.FileID, chunk.ChunkNum)

		dst, err := os.Create(chunkFilename)
		if err != nil {
			log.Println(err)
			continue
		}

		chunk.File.Seek(chunk.StartByte, 0)
		buf := make([]byte, chunk.EndByte-chunk.StartByte+1)
		chunk.File.Read(buf)
		dst.Write(buf)
		dst.Close()
	}
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	threadCountStr := r.FormValue("threads")
	threadCount, err := strconv.Atoi(threadCountStr)
	if err != nil || threadCount <= 0 {
		http.Error(w, "Invalid thread count, must be between greater than 1", http.StatusBadRequest)
		return
	}

	fileID := r.FormValue("fileID")
	if fileID == "" {
		http.Error(w, "Missing fileID", http.StatusBadRequest)
		return
	}

	fileSize := r.ContentLength
	chunkSize := fileSize / int64(threadCount)

	for i := 0; i < threadCount; i++ {
		startByte := int64(i) * chunkSize
		endByte := int64(0)

		if i == threadCount-1 {
			endByte = fileSize
		} else {
			endByte = startByte + chunkSize - 1
		}
		chunk := ChunkRequest{
			File:      file,
			FileID:    fileID,
			ChunkNum:  i + 1,
			StartByte: startByte,
			EndByte:   endByte,
		}

		chunkQueue <- chunk
	}

	fmt.Fprintf(w, "File %s upload initiated with %d threads", fileID, threadCount)
}

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/upload", UploadHandler)

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST"},
	})

	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(i, &wg)
	}

	log.Println("Server is starting on Port 8080.....")
	handler := c.Handler(router)
	http.ListenAndServe(":8080", handler)

	close(chunkQueue)
	wg.Wait()
}
