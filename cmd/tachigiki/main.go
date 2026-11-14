package main

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"github.com/ogadra/Tachigiki/internal/config"
	"github.com/ogadra/Tachigiki/internal/fetch"
	_ "github.com/ogadra/Tachigiki/internal/fetch/adapters"
)

func main() {
	ctx := context.Background()
	now := time.Now().In(time.FixedZone("JST", 9*60*60))
	day := now.Format(time.DateOnly)
	sources, err := config.Load(".")
	if err != nil {
		log.Fatal(err)
	}
	if err := fetch.Run(ctx, now, sources, filepath.Join("data", "raw", day)); err != nil {
		log.Fatal(err)
	}
}
