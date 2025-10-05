package main

import (
	"crypto/rand"
	"embed"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/gen2brain/go-fitz"
	"github.com/pdfcpu/pdfcpu/pkg/api"
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

func ensureDirs(session_id string) (string, string, string, error) {
	sessionDir := filepath.Join("artifacts", session_id)
	sessionUploadsDir := filepath.Join(sessionDir, "uploads")
	sessionPreviewsDir := filepath.Join(sessionDir, "previews")

	dirs := []string{sessionDir, sessionUploadsDir, sessionPreviewsDir}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", "", "", fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return sessionDir, sessionUploadsDir, sessionPreviewsDir, nil
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




type PageMeta struct {
	ID              string `json:"id"`
	PageNumber      int    `json:"pagenumber"`
	PageOrientation string `json:"pageorientation"`
	Flip            int    `json:"flip"`
	Rotate          int    `json:"rotate"`
	Status          string `json:"status"`
	PreviewPath     string `json:"preview_path"`
}

type PDFMeta struct {
	Filename         string     `json:"filename"`
	Version          string     `json:"version"`
	PageCount        int        `json:"page_count"`
	PageSize         string     `json:"page_size"`
	Title            string     `json:"title"`
	Author           string     `json:"author"`
	Subject          string     `json:"subject"`
	Producer         string     `json:"producer"`
	Creator          string     `json:"creator"`
	CreationDate     string     `json:"creation_date"`
	ModificationDate string     `json:"modification_date"`
	Pages            []PageMeta `json:"pages"`
}

func extractPDFMeta(pdfPath, filename string) (PDFMeta, error) {
	ctx, err := api.ReadContextFile(pdfPath)
	if err != nil {
		return PDFMeta{}, err
	}

	meta := PDFMeta{
		Filename:  filename,
		Version:   ctx.HeaderVersion.String(), 
		PageCount: ctx.PageCount,
	}

	if dims, err := ctx.PageDims(); err == nil && len(dims) > 0 {
		d := dims[0]
		meta.PageSize = fmt.Sprintf("%.2f x %.2f points", d.Width, d.Height)
	}

	if ctx.Title != "" {
		meta.Title = ctx.Title
	}
	if ctx.Author != "" {
		meta.Author = ctx.Author
	}
	if ctx.Subject != "" {
		meta.Subject = ctx.Subject
	}
	if ctx.Producer != "" {
		meta.Producer = ctx.Producer
	}
	if ctx.Creator != "" {
		meta.Creator = ctx.Creator
	}

	fmt.Printf("Context type: %T\n", ctx)
	if ctx.Info != nil {
		fmt.Printf("Info type: %T, value: %+v\n", ctx.Info, ctx.Info)
	}

	if ctx.ModDate != "" {
		meta.ModificationDate = ctx.ModDate
	}

	return meta, nil
}
func detectOrientation(width, height int) string {
	if width > height {
		return "landscape"
	}
	return "portrait"
}

func generatePreviews(pdfPath, filename, sessionPreviewsDir string) (PDFMeta, error) {
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return PDFMeta{}, fmt.Errorf("fitz open: %w", err)
	}
	defer doc.Close()

	pdfMeta, err := extractPDFMeta(pdfPath, filename)
	if err != nil {
		return PDFMeta{}, fmt.Errorf("pdfcpu metadata: %w", err)
	}

	for i := 0; i < doc.NumPage(); i++ {
		img, err := doc.Image(i)
		if err != nil {
			return PDFMeta{}, fmt.Errorf("render page %d: %w", i+1, err)
		}

		pageID, err := NanoID(15)
		if err != nil {
			return PDFMeta{}, fmt.Errorf("generate id: %w", err)
		}

		previewPath := filepath.Join(sessionPreviewsDir, pageID+".png")

		f, err := os.Create(previewPath)
		if err != nil {
			return PDFMeta{}, fmt.Errorf("create preview file: %w", err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			return PDFMeta{}, fmt.Errorf("encode png: %w", err)
		}
		if err := f.Close(); err != nil {
			return PDFMeta{}, fmt.Errorf("close preview file: %w", err)
		}

		orientation := detectOrientation(img.Bounds().Dx(), img.Bounds().Dy())

		p := PageMeta{
			ID:              pageID,
			PageNumber:      i + 1,
			PageOrientation: orientation,
			Flip:            0,
			Rotate:          0,
			Status:          "show",
			PreviewPath:     filepath.ToSlash(previewPath),
		}

		pdfMeta.Pages = append(pdfMeta.Pages, p)
	}

	return pdfMeta, nil
}

func streamPagesJSON(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")

    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }

    sessionID := generateCookie(w, r)
    sessionDir := filepath.Join("artifacts", sessionID)

    lastSent := ""

		for {
				jsonFile := filepath.Join(sessionDir, sessionID+"_store.json")

				data, err := os.ReadFile(jsonFile)
				if err == nil {
						// recompact JSON to single line
						var tmp map[string]any
						if json.Unmarshal(data, &tmp) == nil {
								compactData, _ := json.Marshal(tmp)
								jsonStr := string(compactData)

								if jsonStr != "" && jsonStr != lastSent {
										fmt.Fprintf(w, "data: %s\n\n", jsonStr)
										flusher.Flush()
										lastSent = jsonStr
								}
						}
				}

				if r.Context().Err() != nil {
						return
				}
				time.Sleep(1 * time.Second)
		}
}



func updatePagesJSON(meta PDFMeta, sessionID string,sessionDir string) error {
	jsonFile := filepath.Join(sessionDir, sessionID+"_store.json")
	allPDFs := make(map[string]PDFMeta)

	if data, err := os.ReadFile(jsonFile); err == nil {
		_ = json.Unmarshal(data, &allPDFs)
	}

	allPDFs[meta.Filename] = meta

	data, err := json.MarshalIndent(allPDFs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(jsonFile, data, 0644)
}

func generateCookie(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		return cookie.Value
	}

	id, _ := NanoID(15)
	cookie = &http.Cookie{
		Name:     "session_id",
		Value:    id,
		Path:     "/",
		HttpOnly: false,
		Expires:  time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, cookie)
	return id
}


func pdfUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := generateCookie(w,r)

	sessionDir , sessionUploadsDir, sessionPreviewsDir, err := ensureDirs(sessionID)
	if err != nil {
		http.Error(w, "Unable to create session directories", http.StatusInternalServerError)
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
			destinationPath = filepath.Join(sessionUploadsDir, fileName)

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

		meta, err := generatePreviews(destinationPath, fileName, sessionPreviewsDir)
		if err != nil {
			log.Printf("generatePreviews error: %v", err)
			http.Error(w, fmt.Sprintf("Error generating previews: %v", err), http.StatusInternalServerError)
			return
		}

		if err := updatePagesJSON(meta, sessionID, sessionDir); err != nil {
			log.Printf("updatePagesJSON error: %v", err)
			http.Error(w, fmt.Sprintf("Error updating pages.json: %v", err), http.StatusInternalServerError)
			return
		}

		var fileid string = fileName[:15]
		var fileDisplayName string = ""

		if len(fileName) > 16 {
				fileDisplayName = fileName[16:]
		} else if len(fileName) > 15 {
			fileDisplayName = "File_" + fileName[6:]
		}

		
		data := fmt.Sprintf(`<li id="%s" class="flex items-center rounded-l bg-secondary hover:border-2 hover:border-primary pl-3" style="display: list-item;">
				<div class="inline-flex flex-row justify-between p-2 rounded-xl w-[calc(100%%-2rem)]">
					<div>%s</div>
			<div class="rounded-full pl-1 pr-1 ml-2 hover:text-red-500" onclick="this.closest('li').remove()">&times;</div>
				</div>
			</li>`, fileid, fileDisplayName)

		 fmt.Fprintf(w,data)
	}
}


func main() {
	http.HandleFunc("/", dashboard)
	http.HandleFunc("/dashboard", oldDashboard)
	http.HandleFunc("/events", streamPagesJSON)
	http.HandleFunc("/upload", pdfUpload)
	http.HandleFunc("/id", getNanoID)


	staticFS, _ := fs.Sub(staticFiles, "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	http.Handle("/artifacts/", http.StripPrefix("/artifacts/", http.FileServer(http.Dir("./artifacts"))))

	address := GetServerAddress()
	log.Println("\n\nServer is running on http://" + address)
	log.Fatal(http.ListenAndServe(address, nil))
}
