package main

import (
	"fmt"
	"log"

	"github.com/eulixir/lnk/extensions/config"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println(config)
}
