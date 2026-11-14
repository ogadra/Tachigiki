package main

import (
	"log"
	"time"

	"github.com/ogadra/Tachigiki/internal/config"
)

func main() {
	day := time.Now().In(time.FixedZone("JST", 9*60*60)).Format("2006-01-02")
	sources, err := config.Load(".")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s: %d sources", day, len(sources))
}
