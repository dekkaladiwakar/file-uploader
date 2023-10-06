function uploadFile() {
  const file = document.getElementById("file-input").files[0];
  const numThreads = document.getElementById("num-threads").value;

  if (!file || !numThreads) {
    alert("Please select a file and specify the number of threads.");
    return;
  }

  const chunkSize = Math.ceil(file.size / numThreads);
  const uploadProgress = document.getElementById("upload-progress");
  const uploadStatus = document.getElementById("upload-status");

  const uploadPromises = Array.from({ length: numThreads }, (_, i) => {
    const start = i * chunkSize;
    const end = Math.min(start + chunkSize, file.size);
    const chunk = file.slice(start, end);
    const formData = new FormData();
    formData.append("file", chunk);
    formData.append("chunkNumber", i);
    formData.append("totalChunks", numThreads);
    formData.append("fileID", file.name);

    return fetch("http://localhost:8080/upload", {
      method: "POST",
      body: formData,
    }).then((response) => {
      if (!response.ok) {
        return response.text().then((text) => Promise.reject(text));
      }
      // Update the progress bar and status message
      const progress = ((i + 1) / numThreads) * 100;
      uploadProgress.value = progress;
      uploadStatus.textContent = `Uploaded chunk ${i + 1} of ${numThreads}`;
      return response.text();
    });
  });

  Promise.all(uploadPromises)
    .then((texts) => {
      console.log("All chunks uploaded successfully:", texts);
      uploadStatus.textContent = "Upload complete!";
    })
    .catch((error) => {
      console.error("Error uploading chunks:", error);
      uploadStatus.textContent = `Upload failed: ${error}`;
    });
}

function resetUpload() {
  // Clear the file input, number of threads input, progress bar, and status message
  document.getElementById("file-input").value = "";
  document.getElementById("num-threads").value = "";
  document.getElementById("upload-progress").value = 0;
  document.getElementById("upload-status").textContent = "";
}
