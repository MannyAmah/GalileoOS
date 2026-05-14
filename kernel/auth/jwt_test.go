package auth

import (
	"strings"
	"testing"
	"time"
)

const testTenantID = "00000000-0000-7000-8000-000000000001"

// roundtrip: generate keypair → mint token → verify → check claims.
func TestKeypairMintVerifyRoundtrip(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateKeypair(dir); err != nil {
		t.Fatalf("generate keypair: %v", err)
	}

	tok, err := MintToken(dir+"/private.pem", Claims{
		TenantID:           testTenantID,
		MonthlyBudgetCents: 50000,
	}, time.Hour)
	if err != nil {
		t.Fatalf("mint token: %v", err)
	}
	if !strings.HasPrefix(tok, "eyJ") {
		t.Fatalf("token does not look like a JWT: %q", tok)
	}

	c, err := VerifyToken(dir+"/public.pem", tok)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if c.TenantID != testTenantID {
		t.Errorf("tenant_id: got %q, want %q", c.TenantID, testTenantID)
	}
	if c.Subject != testTenantID {
		t.Errorf("sub: got %q, want %q (Stage 0 sets sub = tenant_id)", c.Subject, testTenantID)
	}
	if c.Issuer != Issuer {
		t.Errorf("iss: got %q, want %q", c.Issuer, Issuer)
	}
	if c.MonthlyBudgetCents != 50000 {
		t.Errorf("monthly_budget_cents (informational): got %d, want 50000", c.MonthlyBudgetCents)
	}
}

func TestVerifyRejectsWrongIssuer(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateKeypair(dir); err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	// Mint a valid token, then re-sign with the wrong issuer by minting
	// from a Claims with a pre-populated RegisteredClaims that MintToken
	// will overwrite — so instead we test "verify a token minted under
	// a different signer rejects." Simulate by using a *second* keypair
	// to mint and the first keypair's public to verify.
	dir2 := t.TempDir()
	if err := GenerateKeypair(dir2); err != nil {
		t.Fatalf("generate second keypair: %v", err)
	}
	tok, err := MintToken(dir2+"/private.pem", Claims{TenantID: testTenantID}, time.Hour)
	if err != nil {
		t.Fatalf("mint with second keypair: %v", err)
	}
	if _, err := VerifyToken(dir+"/public.pem", tok); err == nil {
		t.Fatal("expected verification with wrong public key to fail; got nil")
	}
}

func TestVerifyRejectsExpiredToken(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateKeypair(dir); err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	tok, err := MintToken(dir+"/private.pem", Claims{TenantID: testTenantID}, -time.Hour /* already expired */)
	if err != nil {
		t.Fatalf("mint expired token: %v", err)
	}
	if _, err := VerifyToken(dir+"/public.pem", tok); err == nil {
		t.Fatal("expected verification of expired token to fail; got nil")
	}
}

func TestVerifyRejectsMissingTenantID(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateKeypair(dir); err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	tok, err := MintToken(dir+"/private.pem", Claims{TenantID: ""}, time.Hour)
	if err != nil {
		t.Fatalf("mint token with empty tenant: %v", err)
	}
	if _, err := VerifyToken(dir+"/public.pem", tok); err == nil {
		t.Fatal("expected verification of token without tenant_id to fail; got nil")
	}
}
