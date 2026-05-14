// jwt-tool is the Stage 0 operator helper for the local Ed25519 dev
// keypair and for minting tokens against it. It is not deployed; it
// exists so `make stage0-jwt-setup` and `make stage0-jwt` are one-shot
// commands rather than ad-hoc scripts.
//
// Subcommands:
//
//	genkey -dir <path>
//	    Generate an Ed25519 keypair (private.pem + public.pem) at dir.
//	    Use kernel/auth/dev-keys/ as the canonical location — that path
//	    is gitignored and is the default the gateway reads from.
//
//	mint -priv <path> -tenant <uuid> [-ttl 1h] [-budget 50000]
//	    Mint a JWT for the given tenant. The token's monthly_budget_cents
//	    claim is informational only — the gateway re-reads the value
//	    from Postgres on every request (Drift-1).
//
// Stage 1's Supabase migration retires this tool; tokens come from
// Supabase Auth and the gateway swaps to Supabase JWKS for verification.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/MannyAmah/GalileoOS/kernel/auth"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "genkey":
		if err := genkey(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "mint":
		if err := mint(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: jwt-tool {genkey|mint} [flags]")
	fmt.Fprintln(os.Stderr, "  genkey -dir <path>")
	fmt.Fprintln(os.Stderr, "  mint   -priv <path> -tenant <uuid> [-ttl 1h] [-budget 50000]")
}

func genkey(args []string) error {
	fs := flag.NewFlagSet("genkey", flag.ExitOnError)
	dir := fs.String("dir", "kernel/auth/dev-keys", "directory to write private.pem and public.pem into")
	_ = fs.Parse(args)
	if err := auth.GenerateKeypair(*dir); err != nil {
		return fmt.Errorf("generate keypair: %w", err)
	}
	fmt.Printf("wrote %s/private.pem and %s/public.pem\n", *dir, *dir)
	return nil
}

func mint(args []string) error {
	fs := flag.NewFlagSet("mint", flag.ExitOnError)
	priv := fs.String("priv", "kernel/auth/dev-keys/private.pem", "path to Ed25519 private key (PEM/PKCS#8)")
	tenant := fs.String("tenant", "", "tenant_id (UUID v7) for the sub and tenant_id claims")
	ttl := fs.Duration("ttl", time.Hour, "token lifetime")
	budget := fs.Int64("budget", 0, "monthly_budget_cents (informational only — gateway re-reads from Postgres)")
	_ = fs.Parse(args)
	if *tenant == "" {
		return fmt.Errorf("mint: -tenant is required")
	}
	tok, err := auth.MintToken(*priv, auth.Claims{
		TenantID:           *tenant,
		MonthlyBudgetCents: *budget,
	}, *ttl)
	if err != nil {
		return fmt.Errorf("mint token: %w", err)
	}
	fmt.Println(tok)
	return nil
}
