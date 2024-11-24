package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	i "github.com/cheezecakee/urlShort/src/internal"
)

func MapHandler(fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathUrl := "./src/internal/data.yaml"
		ensureYAML(pathUrl)

		pathMap := loadYAML(pathUrl)

		url := r.PathValue("url")

		fmt.Printf("url: %v\n", url)

		if entry, ok := pathMap[url]; ok {
			if !entry.Expires.IsZero() && entry.Expires.Before(time.Now()) {
				deleteYAML(pathUrl, url)

				http.Redirect(w, r, "/expired", http.StatusFound)
				return
			}
			entry.ClickCount++

			// Update the map with the new entry
			pathMap[url] = entry

			SaveYAML(pathUrl, pathMap)

			fmt.Printf("clicks: %v\n", pathMap[url].ClickCount)
			if entry.QRCode != "" {
				http.ServeFile(w, r, filepath.Join("./src/ui/static/QRCodes", entry.QRCode))
				return
			}
			if entry.URL != "" {
				fmt.Printf("entry: %v\n", entry.URL)

				http.Redirect(w, r, entry.URL, http.StatusFound)

				return
			}
		}
		fallback.ServeHTTP(w, r)
	}
}

func deleteYAML(filePath, urlPath string) {
	pathMap := loadYAML(filePath)

	delete(pathMap, urlPath)

	SaveYAML(filePath, pathMap)

	log.Printf("Deleted URL path '%s' from YAML", urlPath)
}

func SaveYAML(filePath string, mappings map[string]i.URLMapping) {
	var pathList []i.URLMapping
	for _, entry := range mappings {
		pathList = append(pathList, entry)
	}

	data, err := yaml.Marshal(pathList)
	if err != nil {
		log.Fatalf("Failed to marshal YAML data: %v", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		log.Fatalf("Failed to write YAML File: %v", err)
	}

	log.Printf("Data saved to YAML file: %s", filePath) // Confirm save operation
}

func loadYAML(filePath string) map[string]i.URLMapping {
	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	var pathList []i.URLMapping
	err = yaml.Unmarshal(file, &pathList)
	if err != nil {
		log.Fatalf("Failed to unmarshal YAML data: %v", err)
	}

	pathMap := make(map[string]i.URLMapping)
	for _, entry := range pathList {
		pathMap[entry.Path] = entry
	}

	return pathMap
}

// Ensure data.yaml exists
func ensureYAML(filePath string) {
	var dataFile []i.URLMapping
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		data, err := yaml.Marshal(&dataFile)
		if err != nil {
			log.Fatalf("Failed to marshal default YAML: %v", err)
		}
		err = os.WriteFile(filePath, data, 0644)
		if err != nil {
			log.Fatalf("Failed to create YAML file: %v", err)
		}
		log.Printf("Created default YAML file at %s", filePath)
	}
}
