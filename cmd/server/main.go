package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/cgcodeboy/sim-studio/api"
	"github.com/cgcodeboy/sim-studio/internal/app"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	uri := os.Getenv("SIM_MONGO")
	if uri == "" {
		uri = "mongodb://127.0.0.1:27017"
	}
	application, err := app.Open(ctx, uri)
	if err != nil {
		log.Printf("数据库没连上，工程保存不可用：%v", err)
		application = app.Wire()
	}
	addr := os.Getenv("SIM_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Fatal(api.Listen(application, addr))
}
