package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"smarterp/internal/httpapi"
)

func main() {
	addr := env("ADDR", ":8080")

	srv := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewServer(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("ERP沙盘智能决策系统已启动，监听地址 %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
