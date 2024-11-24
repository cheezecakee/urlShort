package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cheezecakee/urlShort/src/components"
	i "github.com/cheezecakee/urlShort/src/internal"
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
	mux.Handle("GET /stats", http.HandlerFunc(stats))
	mux.Handle("GET /expired", http.HandlerFunc(expired))
	mux.Handle("POST /format", http.HandlerFunc(format))
	mux.Handle("POST /shortenUrl", http.HandlerFunc(shortenUrl))

	return mux
}

func format(w http.ResponseWriter, r *http.Request) {
	var custom bool

	format := r.FormValue("format")
	switch format {
	case "custom":
		custom = true
	default:
		custom = false
	}
	// fmt.Printf("format: %v\ncustom: %v\n", format, custom)
	err := renderTempl(r.Context(), w, components.CustomContainer(custom))
	if err != nil {
		http.Error(w, "Failed to render custom container", http.StatusInternalServerError)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	err := renderTempl(r.Context(), w, components.Home(false))
	if err != nil {
		http.Error(w, "Failed to render Home page", http.StatusInternalServerError)
	}
}

func expired(w http.ResponseWriter, r *http.Request) {
	err := renderTempl(r.Context(), w, components.Expired())
	if err != nil {
		http.Error(w, "Failed to render Home page", http.StatusInternalServerError)
	}
}

func stats(w http.ResponseWriter, r *http.Request) {
	pathUrl := "./src/internal/data.yaml"
	pathMap := loadYAML(pathUrl)

	var urlMappings []i.URLMapping
	for _, mapping := range pathMap {
		urlMappings = append(urlMappings, mapping)
	}

	err := renderTempl(r.Context(), w, components.Stats(urlMappings))
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
	var expiresIn time.Time

	format := r.FormValue("format")
	switch format {
	case "short":
		response = fmt.Sprintf(`<div id="result"><p>Your shortened URL: <a href="/%s">[Link]</a></p></div>`, path)
		fmt.Println("Short url option selected")
		expiresIn = time.Time{}
	case "qr":
		fmt.Println("QR Code option selected")

		qrName, qrPath = generateQRCode(longUrl)
		response = fmt.Sprintf(`
			<div class="result" id="result">
				<img class="img-qr" src="static/QRCodes/%s" alt="QR Code"/>
			</div>
		`, qrName)
		expiresIn = time.Time{}
	case "custom":
		customSlug := r.FormValue("custom_slug")
		expirationHours := r.FormValue("expiration")

		// Check what the user selected
		returnLink := r.FormValue("check-link") != ""
		returnQR := r.FormValue("check-qr") != ""

		// Generate the response dynamically
		if returnQR {
			qrName, qrPath = generateQRCode(customSlug)
		}

		fmt.Printf("customSlug: %v\nexpirationHours: %v\n", customSlug, expirationHours)
		fmt.Printf("Link: %v\nQR: %v\n", returnLink, returnQR)

		if customSlug == "" {
			http.Error(w, "Custom slug is required", http.StatusBadRequest)
			return
		}

		hours := 0
		if expirationHours != "" {
			hours, _ = strconv.Atoi(expirationHours)
		}
		if hours == 0 {
			expiresIn = time.Time{} // No expiration
		} else {
			expiresIn = time.Now().Add(time.Duration(hours) * time.Hour)
		}

		path = customSlug

		switch {
		case returnLink && returnQR:
			response = fmt.Sprintf(`
				<div class="result" id="result">
					<p>Your custom URL: <a href="/%s">[%s]</a></p>
					<img class="img-qr" src="static/QRCodes/%s" alt="QR Code"/>
				</div>
			`, path, path, qrName)

		case returnLink:
			response = fmt.Sprintf(`<div id="result"><p>Your custom URL: <a href="/%s">[%s]</a></p></div>`, path, path)
		case returnQR:
			response = fmt.Sprintf(`
				<div class="result" id="result">
					<img class="img-qr" src="static/QRCodes/%s" alt="QR Code"/>
				</div>
			`, qrName)
		}
	default:
		http.Error(w, "Invalid format selected", http.StatusBadRequest)
	}

	fmt.Printf("Shortened URL: %s -> %s\n", longUrl, path)
	// Save mappings to data.yaml
	existingMappings[path] = i.URLMapping{
		Path:         path,
		URL:          longUrl,
		QRCode:       qrPath,
		ClickCount:   0,
		CreationDate: time.Now(),
		Expires:      expiresIn, // 0 means never
	}
	SaveYAML(pathUrl, existingMappings)

	// Display result
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(response))
}
