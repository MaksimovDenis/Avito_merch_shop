package main

import (
	"context"
	"log"

	"github.com/MaksimovDenis/Avito_merch_shop/internal/app"
)

func main() {
	ctx := context.Background()

	merchShop, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("failed to init app: %s", err.Error())
	}

	err = merchShop.Run()
	if err != nil {
		log.Fatalf("failed to run app: %s", err.Error())
	}
}
