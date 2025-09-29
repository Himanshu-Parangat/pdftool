package main

import (
	"crypto/rand"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"
)

//go:embed components/*
var templateFiles embed.FS

//go:embed static/*
var staticFiles embed.FS



func GetServerAddress() string {
	host := os.Getenv("SERVER_HOST_IP")
	port := os.Getenv("SERVER_PORT")

	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "8080"
	}

	return fmt.Sprintf("%s:%s", host, port)
}

func ensureDirs() {
	dirs := []string{
		"./artifacts",
		"./artifacts/uploads",
		"./artifacts/chuncks",
		"./artifacts/previews",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const alphabetSize = byte(len(alphabet)) 

func NanoID(n int) (string, error) {
	if n <= 0 {
		return "", nil
	}

	threshold := byte(256 - (256 % int(alphabetSize)))

	result := make([]byte, n)
	buf := make([]byte, 1)

	for i := 0; i < n; {
		_, err := io.ReadFull(rand.Reader, buf)
		if err != nil {
			return "", err
		}
		b := buf[0]
		if b >= threshold {
			continue
		}
		result[i] = alphabet[b%alphabetSize]
		i++
	}

	return string(result), nil
}

func getNanoID(w http.ResponseWriter, r *http.Request) {
	data,_ := NanoID(12)
	fmt.Fprintf(w, "Nano ID is %s\n", data)
}

func oldDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templateFiles, "components/dashboard.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}


func dashboard(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templateFiles, "components/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

var (
	storedValue string
	mu          sync.Mutex
)

func setValue(val string) {
	mu.Lock()
	defer mu.Unlock()
	storedValue = val
}

func getValue() string {
	mu.Lock()
	defer mu.Unlock()
	return storedValue
}

func counterSSE(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")


	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	lastSent := ""

	for {
		val := getValue()
		if val != "" && val != lastSent {
			fmt.Fprintf(w, "data: %s\n\n", val)
			flusher.Flush()
			lastSent = val
		}

		time.Sleep(1 * time.Second)

		if r.Context().Err() != nil {
			return
		}
	}

}

func sethtml()  {
	data := "<div class='w-[20%] '>conneting..</div>"
	setValue(data)
}


func detectOrientation(width, height int) string {
	if width > height {
		return "landscape"
	}
	return "portrait"
}


func pdfUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["pdfs"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}


	for _, fileHeader := range files {
		if filepath.Ext(fileHeader.Filename) != ".pdf" {
			http.Error(w, "Only pdf files are allowed", http.StatusBadRequest)
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		safeName := filepath.Base(fileHeader.Filename)

		var fileName, destinationPath string
		for {
			id, err := NanoID(15)
			if err != nil {
				http.Error(w, "Can't generate new id", http.StatusInternalServerError)
				return
			}

			fileName = id + "_" + safeName
			destinationPath = filepath.Join("./artifacts/uploads/", fileName)

			if _, err := os.Stat(destinationPath); os.IsNotExist(err) {
				break 
			}
		}

		out, err := os.Create(destinationPath)
		if err != nil {
			http.Error(w, "Unable to create file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Saved as %s\n", fileName)
	}
}




func main() {
	ensureDirs()
	sethtml()
	http.HandleFunc("/", dashboard)
	http.HandleFunc("/dashboard", oldDashboard)
	http.HandleFunc("/events", counterSSE)
	http.HandleFunc("/upload", pdfUpload)
	http.HandleFunc("/id", getNanoID)


	staticFS, _ := fs.Sub(staticFiles, "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))


	address := GetServerAddress()
	log.Println("\n\nServer is running on http://" + address)
	log.Fatal(http.ListenAndServe(address, nil))
}




