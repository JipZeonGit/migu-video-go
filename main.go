package main

import (
	"fmt"
	"os"

	"migu-video-go/config"
	"migu-video-go/internal/log"
	"migu-video-go/server"
)

func main() {
	cfg := config.Load()
	log.DebugMode = cfg.Debug

	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	srv := server.New(cfg, workDir)
	if err := srv.Start(); err != nil {
		log.Red(fmt.Sprintf("Server failed: %v", err))
		os.Exit(1)
	}
}
