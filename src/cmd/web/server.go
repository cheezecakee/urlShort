package web

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cheezecakee/urlShort/src/components"
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
	mux.Handle("GET /format", http.HandlerFunc(format))
	mux.Handle("GET /{url}", MapHandler(http.NotFoundHandler()))
	mux.Handle("POST /shortenUrl", http.HandlerFunc(shortenUrl))

	return mux
}

type Format struct {
	Custom bool
}

func format(w http.ResponseWriter, r *http.Request) {
	var f Format
	format := r.FormValue("format")
	fmt.Println("format: ", format)
	switch format {
	case "custom":
		f.Custom = true
	default:
		f.Custom = false
	}

	fmt.Println("custom: ", f.Custom)
}

func home(w http.ResponseWriter, r *http.Request) {
	err := renderTempl(r.Context(), w, components.Home())
	if err != nil {
		http.Error(w, "Failed to render Home page", http.StatusInternalServerError)
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
