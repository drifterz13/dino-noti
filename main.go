package main

import (
	"fmt"
	"net/http"

	"github.com/drifterz13/dino-noti/config"
	"github.com/drifterz13/dino-noti/service"

	"os"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	srv := service.NewService(cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	http.HandleFunc("/callback", srv.HandleLineMessageReq)

	fmt.Printf("Listening on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		os.Exit(1)
	}
}
