package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"vending-qris-client/internal/api"
	"vending-qris-client/internal/config"
	"vending-qris-client/internal/mcu"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return fmt.Errorf("missing command")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch args[0] {
	case "serve":
		return cmdServe(client)
	case "health":
		return cmdHealth(ctx, client)
	case "generate-qris":
		return cmdGenerateQRIS(ctx, client, args[1:])
	case "list-gateways":
		return cmdListGateways(ctx, client, cfg)
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

// cmdServe is the microcontroller mode: read JSON lines from stdin, write JSON lines to stdout.
func cmdServe(client *api.Client) error {
	return mcu.Serve(context.Background(), os.Stdin, os.Stdout, client)
}

func cmdHealth(ctx context.Context, client *api.Client) error {
	resp, err := client.Health(ctx)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func cmdGenerateQRIS(ctx context.Context, client *api.Client, args []string) error {
	fs := flag.NewFlagSet("generate-qris", flag.ExitOnError)
	description := fs.String("description", "client purchase", "payment description")
	invoice := fs.String("invoice", "", "optional invoice number")
	productName := fs.String("product", "Vending Item", "product name")
	quantity := fs.Int("quantity", 1, "product quantity")
	price := fs.Float64("price", 10000, "item price")
	if err := fs.Parse(args); err != nil {
		return err
	}

	resp, err := client.GenerateDynamicQRIS(ctx, api.GenerateDynamicQRISRequest{
		Description:   *description,
		InvoiceNumber: *invoice,
		Products: []api.ProductDetail{
			{
				Name:      *productName,
				Quantity:  *quantity,
				ItemPrice: *price,
			},
		},
	})
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func cmdListGateways(ctx context.Context, client *api.Client, cfg config.Config) error {
	if err := cfg.RequireAuthKey(); err != nil {
		return err
	}
	resp, err := client.ListPaymentGateways(ctx)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printUsage() {
	fmt.Println(`vending-qris-client — HTTP client for vending-qris-service

Environment:
  PAYMENT_SERVICE_BASE_URL  Service base URL (default: http://localhost:8080)
  PAYMENT_SERVICE_AUTH_KEY  Admin API key (required for admin commands; inject via CI/CD)

Commands:
  serve                     Microcontroller mode (JSON lines on stdin/stdout)
  health
  generate-qris [--description TEXT] [--invoice TEXT] [--product NAME] [--quantity N] [--price AMOUNT]
  list-gateways

Microcontroller protocol (serve):
  Input (one JSON object per line on stdin):
    {"product":"Water","quantity":1,"price":5000,"description":"optional","invoice":"optional"}
  Output (one JSON object per line on stdout):
    {"ok":true,"qr_string":"...","reference_id":"...","qr":{"size":29,"modules":[[1,0,...]],"ascii":"..."}}
    {"ok":false,"error":"reason"}`)
}
