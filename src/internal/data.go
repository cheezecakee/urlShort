package internal

import "time"

type URLMapping struct {
	Path         string    `yaml:"path"`
	URL          string    `yaml:"url"`
	QRCode       string    `yaml:"qrcode"`
	ClickCount   int       `yaml:"click_count"`
	CreationDate time.Time `yaml:"creation_date"`
	Expires      time.Time `yaml:"expires"`
}
