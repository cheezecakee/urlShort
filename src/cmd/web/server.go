package web

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	qrcode "github.com/skip2/go-qrcode"
)

func Run() {
	cwd, _ := os.Getwd()
	log.Printf("Current working directory: %s", cwd)

	mux := routes()

	log.Print("Listening...")
	http.ListenAndServe(":8080", mux)
}

func routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./src/ui/static/"))))

	mux.Handle("GET /{$}", http.HandlerFunc(home))
	mux.Handle("GET /{url}", MapHandler(http.NotFoundHandler()))
	mux.Handle("POST /shortenUrl", http.HandlerFunc(shortenUrl))

	return mux
}

func home(w http.ResponseWriter, r *http.Request) {
	templatePath := "./src/ui/templates/"
	homePath := templatePath + "base.templ"
	navPath := templatePath + "nav.templ"

	tmpl, err := template.ParseFiles(
		homePath,
		navPath,
	)
	if err != nil {
		http.Error(w, "Unable to load templates", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", nil)
	if err != nil {
		http.Error(w, "Unable to execute template", http.StatusInternalServerError)
		log.Printf("Execution error: %v", err)
	}
}

func shortenUrl(w http.ResponseWriter, r *http.Request) {
	longUrl := r.FormValue("longurl")

	if longUrl == "" {
		http.Error(w, "Long URL is required", http.StatusBadRequest)
		return
	}

	pathUrl := "./src/internal/data.yaml"
	existingMappings := loadYAML(pathUrl)

	var (
		response string
		qrPath   string
		qrName   string
	)

	path := generateShortUrl(longUrl)

	format := r.FormValue("format")
	switch format {
	case "short":
		response = fmt.Sprintf(`<div id="result"><p>Your shortened URL: <a href="/%s">[Link]</a></p></div>`, path)
		fmt.Println("Short url option selected")
	case "qr":
		fmt.Println("QR Code option selected")

		qrName, qrPath = generateQRCode(longUrl)
		response = fmt.Sprintf(`
			<div class="result" id="result">
				<img src="static/QRCodes/%s" alt="QR Code"/>
			</div>
		`, qrName)
	case "custom":
	default:
		http.Error(w, "Invalid format selected", http.StatusBadRequest)
	}

	fmt.Printf("Shortened URL: %s -> %s\n", longUrl, path)
	// Save mappings to data.yaml
	existingMappings[path] = PathList{
		Path:   path,
		URL:    longUrl,
		QRCode: qrPath,
	}
	SaveYAML(pathUrl, existingMappings)

	// Display result
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(response))
}

func generateShortUrl(longUrl string) string {
	hash := sha256.New()
	hash.Write([]byte(longUrl))
	hashBytes := hash.Sum(nil)

	shortUrl := hex.EncodeToString(hashBytes)[:6]
	return shortUrl
}

func generateQRCode(longUrl string) (fileName string, filePath string) {
	err := os.MkdirAll("./src/ui/static/QRCodes", os.ModePerm)
	if err != nil {
		fmt.Printf("failed to create output directory: %v", err)
	}

	fileName = fmt.Sprintf("%x.png", sha256.Sum256([]byte(longUrl)))
	filePath = filepath.Join("./src/ui/static/QRCodes", fileName)

	err = qrcode.WriteFile(longUrl, qrcode.Medium, 256, filePath)
	if err != nil {
		fmt.Printf("Failed to encode qrcode: %v", err)
	}

	return fileName, filePath
}
