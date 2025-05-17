package config

import (
	"fmt"
	"os"
)

type Config struct {
	TargetURL         string
	MaxPages          int
	MyList            []string
	GeminiAPIKey      string
	LineChannelToken  string
	LineChannelSecret string
}

const (
	TARGET_URL        = "https://buyee.jp/item/search/category/2084261642?sort=end&order=d&aucmin_bidorbuy_price=6000&aucmax_bidorbuy_price=30000"
	DEFAULT_MAX_PAGES = 10
)

func LoadConfig() (*Config, error) {
	cfg := &Config{
		TargetURL: TARGET_URL,
	}

	maxPagesStr := os.Getenv("MAX_PAGES")
	if maxPagesStr == "" {
		cfg.MaxPages = DEFAULT_MAX_PAGES
	} else {
		_, err := fmt.Sscan(maxPagesStr, &cfg.MaxPages)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_PAGES: %w", err)
		}
	}

	var myList = []string{
		"Canon IXY 10",
		"Canon IXY 20",
		"Canon IXY 50",
		"Canon IXY 60",
		"Canon IXY 120",
		"Canon IXY 130",
		"Canon IXY 140",
		"Canon IXY 160",
		"Canon IXY 910",
		"Canon IXY 10s",
		"Canon IXY 30s",
		"Canon IXY 31s",
		"Canon IXY 32s",
		"Canon IXY 50s",
		"Canon IXY 20 IS",
		"Canon IXY 25 IS",
		"Canon IXY 95 IS",
		"Canon IXY 110 IS",
		"Canon IXY 510 IS",
		"Canon IXY 800 IS",
		"Canon IXY 900 IS",
		"Canon IXY 910 IS",
		"Canon IXY PC1249",
		"Canon IXY 920 IS",
		"Canon IXY 930 IS",
		"Canon IXY 100f",
		"Canon IXY 200f",
		"Canon IXY 210f",
		"Canon IXY 420f",
		"Canon IXY 600",
		"Canon PowerShot E1",
		"Canon PowerShot A800",
		"Canon PowerShot A1000",
		"Canon PowerShot A3100",
		"Casio Exilim EX-ZR20",
		"Casio Exilim EX-ZR100",
		"Casio Exilim EX-Z1080",
		"Casio Exilim EX-ZR1500",
		"Casio Exilim EX-ZR3600",
		"Fuji Finepix F10",
		"Fuji Finepix F11",
		"Fuji Finepix F440",
		"Nikon Coolpix S520",
		"Nikon Coolpix A10",
		"Nikon Coolpix L5",
		"Nikon Coolpix L21",
		"Nikon Coolpix L23",
		"Panasonic Lumix DMC-FX01",
		"Panasonic Lumix DMC-FX35",
		"Panasonic Lumix DMC-FX60",
		"Sony DSC-N1",
		"Sony DSC-N2",
		"Sony DSC-W5",
	}

	cfg.MyList = myList

	cfg.GeminiAPIKey = os.Getenv("GEMINI_API_KEY")
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	cfg.LineChannelToken = os.Getenv("LINE_CHANNEL_TOKEN")
	if cfg.LineChannelToken == "" {
		return nil, fmt.Errorf("LINE_CHANNEL_TOKEN environment variable not set")
	}

	cfg.LineChannelSecret = os.Getenv("LINE_CHANNEL_SECRET")
	if cfg.LineChannelSecret == "" {
		return nil, fmt.Errorf("LINE_CHANNEL_SECRET environment variable not set")
	}

	return cfg, nil
}
