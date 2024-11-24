package web

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/a-h/templ"
	qrcode "github.com/skip2/go-qrcode"
)

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

func renderTempl(ctx context.Context, w http.ResponseWriter, component templ.Component) error {
	// Set the content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Render the component
	return component.Render(ctx, w)
}

// func
