package main

import (
	"fmt"
	"os"

	"vitek/internal/config"
	"vitek/internal/tokens"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %v\n", tokens.ProductName, tokens.BinaryWorker, err)
		os.Exit(1)
	}
	fmt.Printf("%s %s ready env=%s\n", tokens.ProductName, tokens.BinaryWorker, cfg.AppEnv)
}
