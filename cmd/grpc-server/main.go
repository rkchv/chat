package main

import (
	"context"
	"log"

	"github.com/fatih/color"

	"github.com/rkchv/chat/internal/app"
)

func main() {
	ap := app.NewApp(context.Background())

	go func() {
		err := ap.StartPrometheusServer()
		if err != nil {
			log.Printf("failed to start prometheus: %v\n", err)
		}
	}()

	err := ap.Start()
	if err != nil {
		log.Fatal(color.RedString("failed to start app: %v", err))
	}
}
