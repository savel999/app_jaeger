package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/99designs/gqlgen/api"
	"github.com/Yamashou/gqlgenc/clientgenv2"
	"github.com/Yamashou/gqlgenc/config"
	"github.com/Yamashou/gqlgenc/generator"
)

// nolint:gomnd
func main() {
	ctx := context.Background()

	conf := flag.String("config", ".gqlgenc.yml", "--config=some-path.yml")
	flag.Parse()

	cfg, err := config.LoadConfig(*conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err.Error())
		os.Exit(2)
	}

	clientGen := api.AddPlugin(clientgenv2.New(cfg.Query, cfg.Client, cfg.Generate))
	if err := generator.Generate(ctx, cfg, clientGen); err != nil {
		fmt.Fprintf(os.Stderr, "Nooo: %+v", err.Error())
		os.Exit(4)
	}
}
