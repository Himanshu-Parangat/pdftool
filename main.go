package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
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

func pdfUpload(w http.ResponseWriter, r *http.Request)  {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	file, header, err := r.FormFile("pdf")
	if err != nil {
		http.Error(w, "No File upload", http.StatusBadRequest)
	}
	defer file.Close()

	if filepath.Ext(header.Filename) != ".pdf" {
		http.Error(w, "only Pdf files are allowed", http.StatusBadRequest)
	}

	out, err := os.Create("./uploads" + header.Filename)
	if err != nil {
		http.Error(w, "unable to create file", http.StatusInternalServerError)
	}
	defer out.Close()

	_ , err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Uploaded %s successfully\n", header.Filename)
	
}




func main() {
	http.HandleFunc("/", dashboard)
	http.HandleFunc("/upload", pdfUpload)


	staticFS, _ := fs.Sub(staticFiles, "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))


	address := GetServerAddress()
	log.Println("\n\nServer is running on http://" + address)
	log.Fatal(http.ListenAndServe(address, nil))
}




