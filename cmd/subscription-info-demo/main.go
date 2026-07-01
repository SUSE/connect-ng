package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"os"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/registration"
)

func bold(format string, args ...any) {
	fmt.Printf("\033[1m"+format+"\033[0m", args...)
}

func main() {
	var saveInfoPath string
	var saveProductsPath string

	flag.StringVar(&saveInfoPath, "save-info", "", "Save subscription info JSON to file")
	flag.StringVar(&saveProductsPath, "save-products", "", "Save subscription products JSON to file")
	flag.Parse()

	fmt.Println("subscription-info-demo: Demonstrates FetchSubscriptionInfo and FetchSubscriptionProducts")
	fmt.Println()

	regcode := os.Getenv("REGCODE")
	if regcode == "" {
		fmt.Fprintln(os.Stderr, "Error: REGCODE environment variable is required")
		fmt.Fprintln(os.Stderr, "Usage: REGCODE=your-reg-code ./subscription-info-demo [flags]")
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Setup connection
	opts := connection.DefaultOptions("subscription-info-demo", "1.0", "US")

	if url := os.Getenv("SCC_HOST"); url != "" {
		opts.URL = url
	}

	if disableTokens := os.Getenv("DISABLE_TOKEN_HANDLING"); disableTokens != "" {
		opts.DisableTokenHandling = true
	}

	if certificatePath := os.Getenv("API_CERT"); certificatePath != "" {
		crt, certReadErr := os.ReadFile(certificatePath)
		if certReadErr != nil {
			fmt.Fprintf(os.Stderr, "Error reading certificate: %v\n", certReadErr)
			os.Exit(1)
		}

		block, _ := pem.Decode(crt)
		if block == nil {
			fmt.Fprintln(os.Stderr, "Could not decode the server's certificate")
			os.Exit(1)
		}

		cert, parseErr := x509.ParseCertificate(block.Bytes)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Error parsing certificate: %v\n", parseErr)
			os.Exit(1)
		}

		opts.Certificate = cert
	}

	conn := connection.New(opts, connection.NoCredentials{})

	// Fetch subscription info
	bold("=== Fetching Subscription Info ===\n\n")
	info, infoErr := registration.FetchSubscriptionInfo(conn, regcode)
	if infoErr != nil {
		fmt.Fprintf(os.Stderr, "Error fetching subscription info: %v\n", infoErr)
		os.Exit(1)
	}

	// Pretty print subscription info
	infoJSON, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(infoJSON))

	// Save to file if requested
	if saveInfoPath != "" {
		if err := os.WriteFile(saveInfoPath, infoJSON, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving info to %s: %v\n", saveInfoPath, err)
			os.Exit(1)
		}
		fmt.Printf("\n✓ Saved subscription info to %s\n", saveInfoPath)
	}

	fmt.Println()
	bold("Subscription Details:\n")
	fmt.Printf("  Kind:          %s\n", info.Kind)
	fmt.Printf("  Name:          %s\n", info.Name)
	fmt.Printf("  Starts At:     %s\n", info.StartsAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Expires At:    %s\n", info.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Limit:         %d\n", info.Limit)
	fmt.Printf("  Notifications: %s\n", info.Notifications)
	fmt.Println()
	bold("Product Classes (%d):\n", len(info.ProductClasses))
	for i, pc := range info.ProductClasses {
		fmt.Printf("  [%d] %s\n", i+1, pc.Name)
		fmt.Printf("      %s\n", pc.Description)
	}

	fmt.Println()
	fmt.Println()

	// Fetch subscription products
	bold("=== Fetching Subscription Products ===\n\n")
	products, productsErr := registration.FetchSubscriptionProducts(conn, regcode)
	if productsErr != nil {
		fmt.Fprintf(os.Stderr, "Error fetching subscription products: %v\n", productsErr)
		os.Exit(1)
	}

	// Pretty print products
	productsJSON, _ := json.MarshalIndent(products, "", "  ")
	fmt.Println(string(productsJSON))

	// Save to file if requested
	if saveProductsPath != "" {
		if err := os.WriteFile(saveProductsPath, productsJSON, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving products to %s: %v\n", saveProductsPath, err)
			os.Exit(1)
		}
		fmt.Printf("\n✓ Saved subscription products to %s\n", saveProductsPath)
	}

	fmt.Println()
	bold("Products Summary (%d total):\n", len(products))
	for i, product := range products {
		fmt.Printf("  [%d] %s (%s)\n", i+1, product.Name, product.Identifier)
		fmt.Printf("      Version: %s, Arch: %s\n", product.Version, product.Arch)
		if len(product.Repositories) > 0 {
			fmt.Printf("      Repositories: %d\n", len(product.Repositories))
		}
		if len(product.Extensions) > 0 {
			fmt.Printf("      Extensions: %d\n", len(product.Extensions))
		}
		fmt.Println()
	}

	bold("✓ Demo complete\n")
}
