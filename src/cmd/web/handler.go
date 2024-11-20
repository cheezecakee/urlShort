package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type PathList struct {
	Path   string `yaml:"path"`
	URL    string `yaml:"url"`
	QRCode string `yaml:"qrcode"`
}

func MapHandler(fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathUrl := "./src/internal/data.yaml"
		ensureYAML(pathUrl)

		pathMap := loadYAML(pathUrl)

		url := r.PathValue("url")

		fmt.Printf("url: %v\n", url)

		if entry, ok := pathMap[url]; ok {
			if entry.QRCode != "" {
				http.ServeFile(w, r, filepath.Join("./src/internal/", entry.QRCode))
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

func SaveYAML(filePath string, mappings map[string]PathList) {
	var pathList []PathList
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
}

func loadYAML(filePath string) map[string]PathList {
	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	var pathList []PathList
	err = yaml.Unmarshal(file, &pathList)
	if err != nil {
		log.Fatalf("Failed to unmarshal YAML data: %v", err)
	}

	pathMap := make(map[string]PathList)
	for _, entry := range pathList {
		pathMap[entry.Path] = entry
	}

	return pathMap
}

// Ensure data.yaml exists
func ensureYAML(filePath string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		defaultData := []PathList{
			{Path: "github", URL: "https://github.com", QRCode: ""},
			{Path: "google", URL: "https://google.com", QRCode: ""},
		}
		data, err := yaml.Marshal(&defaultData)
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
