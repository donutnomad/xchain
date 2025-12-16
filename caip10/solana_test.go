package caip10

import (
	"encoding/json"
	"testing"

	"github.com/donutnomad/solana-web3/web3"
	"github.com/fxamacker/cbor/v2"
)

func TestSolanaParse(t *testing.T) {
	tests := []struct {
		input     string
		reference string
		address   string
	}{
		{
			input:     "solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp:7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv",
			reference: SolanaMainnet.String(),
			address:   "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv",
		},
		{
			input:     "solana:EtWTRABZaYq6iMfeYKouRu166VU2xqa1:11111111111111111111111111111111",
			reference: SolanaDevnet.String(),
			address:   "11111111111111111111111111111111",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			a, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) failed: %v", tc.input, err)
			}

			// Should implement SolanaAccountID interface
			sol, ok := a.(SolanaAccountID)
			if !ok {
				t.Fatalf("expected SolanaAccountID, got %T", a)
			}

			if sol.Namespace() != NamespaceSolana {
				t.Errorf("Namespace: got %q, want %q", sol.Namespace(), NamespaceSolana)
			}
			if sol.Reference() != tc.reference {
				t.Errorf("Reference: got %q, want %q", sol.Reference(), tc.reference)
			}
			if sol.Address() != tc.address {
				t.Errorf("Address: got %q, want %q", sol.Address(), tc.address)
			}
		})
	}
}

func TestSolanaNewFromBase58(t *testing.T) {
	a, err := NewSolanaFromBase58(SolanaMainnet, "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")
	if err != nil {
		t.Fatalf("NewSolanaFromBase58 failed: %v", err)
	}

	if a.Namespace() != NamespaceSolana {
		t.Errorf("Namespace: got %q", a.Namespace())
	}
	if a.Reference() != SolanaMainnet.String() {
		t.Errorf("Reference: got %q", a.Reference())
	}
	if a.CAIP2() != "solana:"+SolanaMainnet.String() {
		t.Errorf("CAIP2: got %q", a.CAIP2())
	}
}

func TestSolanaNewFromPublicKey(t *testing.T) {
	pk, err := web3.NewPublicKey("7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")
	if err != nil {
		t.Fatalf("NewPublicKey failed: %v", err)
	}

	a := NewSolana(SolanaMainnet, pk)

	if a.Account() != pk {
		t.Errorf("Account mismatch")
	}
}

func TestSolanaMainnetDevnet(t *testing.T) {
	pk, _ := web3.NewPublicKey("7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")

	mainnet := NewSolanaMainnet(pk)
	if !mainnet.IsMainnet() {
		t.Error("should be mainnet")
	}
	if mainnet.IsDevnet() {
		t.Error("should not be devnet")
	}

	devnet := NewSolanaDevnet(pk)
	if !devnet.IsDevnet() {
		t.Error("should be devnet")
	}
	if devnet.IsMainnet() {
		t.Error("should not be mainnet")
	}
}

func TestSolanaJSON(t *testing.T) {
	a := MustNewSolanaFromBase58(SolanaMainnet, "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b GenericAccountID
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if !a.Equal(&b) {
		t.Errorf("JSON roundtrip failed: got %v, want %v", b, a)
	}
}

func TestSolanaBinary(t *testing.T) {
	a := MustNewSolanaFromBase58(SolanaMainnet, "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")

	data, err := a.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary failed: %v", err)
	}

	var b GenericAccountID
	if err := b.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary failed: %v", err)
	}

	if !a.Equal(&b) {
		t.Errorf("Binary roundtrip failed: got %v, want %v", b, a)
	}
}

func TestSolanaCBOR(t *testing.T) {
	a := MustNewSolanaFromBase58(SolanaMainnet, "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")

	data, err := cbor.Marshal(a)
	if err != nil {
		t.Fatalf("cbor.Marshal failed: %v", err)
	}

	var b GenericAccountID
	if err := cbor.Unmarshal(data, &b); err != nil {
		t.Fatalf("cbor.Unmarshal failed: %v", err)
	}

	if !a.Equal(&b) {
		t.Errorf("CBOR roundtrip failed: got %v, want %v", b, a)
	}
}

func TestSolanaDatabase(t *testing.T) {
	a := MustNewSolanaFromBase58(SolanaMainnet, "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")

	// Value
	v, err := a.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}
	if v == nil {
		t.Error("Value should not be nil")
	}

	// Scan into GenericAccountID
	var b GenericAccountID
	if err := b.Scan(v); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if !a.Equal(&b) {
		t.Errorf("Scan roundtrip failed: got %v, want %v", b, a)
	}
}

func TestSolanaToColumns(t *testing.T) {
	a := MustNewSolanaFromBase58(SolanaMainnet, "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")

	cols := a.ToColumns()
	if Namespace(cols.Namespace) != NamespaceSolana {
		t.Errorf("Namespace: got %q", cols.Namespace)
	}
	if cols.Reference != SolanaMainnet.String() {
		t.Errorf("Reference: got %q", cols.Reference)
	}

	// Convert back - should use Solana parser
	b, err := cols.ToAccountID()
	if err != nil {
		t.Fatalf("ToAccountID failed: %v", err)
	}
	if _, ok := b.(SolanaAccountID); !ok {
		t.Errorf("expected SolanaAccountID, got %T", b)
	}
}

func TestSolanaZeroValue(t *testing.T) {
	var a *solanaAccountID
	if !a.IsZero() {
		t.Error("nil pointer should be IsZero")
	}
}

func TestSolanaEqual(t *testing.T) {
	a1 := MustNewSolanaFromBase58(SolanaMainnet, "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")
	a2 := MustNewSolanaFromBase58(SolanaMainnet, "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")
	a3 := MustNewSolanaFromBase58(SolanaMainnet, "11111111111111111111111111111111")

	if !a1.Equal(a2) {
		t.Error("identical addresses should be equal")
	}
	if a1.Equal(a3) {
		t.Error("different addresses should not be equal")
	}
}

func TestSolanaParserRegistered(t *testing.T) {
	p, ok := GetParser(NamespaceSolana)
	if !ok {
		t.Fatal("Solana parser not registered")
	}
	if p.Namespace() != NamespaceSolana {
		t.Errorf("Namespace: got %q", p.Namespace())
	}
}

func TestValidateSolanaAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "valid on-curve address",
			address: "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv",
			wantErr: false,
		},
		{
			name:    "system program (on-curve)",
			address: "11111111111111111111111111111111",
			wantErr: false,
		},
		{
			name:    "too short",
			address: "abc",
			wantErr: true,
		},
		{
			name:    "invalid base58 character (0)",
			address: "0S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv",
			wantErr: true,
		},
		{
			name:    "invalid base58 character (O)",
			address: "OS3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv",
			wantErr: true,
		},
		{
			name:    "invalid base58 character (I)",
			address: "IS3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv",
			wantErr: true,
		},
		{
			name:    "invalid base58 character (l)",
			address: "lS3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv",
			wantErr: true,
		},
		{
			name:    "empty string",
			address: "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSolanaAddress(tc.address)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateSolanaAddress(%q) error = %v, wantErr %v", tc.address, err, tc.wantErr)
			}
		})
	}
}

func TestValidateSolanaAddressLoose(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "valid on-curve address",
			address: "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv",
			wantErr: false,
		},
		{
			name:    "system program",
			address: "11111111111111111111111111111111",
			wantErr: false,
		},
		{
			name:    "token program",
			address: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
			wantErr: false,
		},
		{
			name:    "too short",
			address: "abc",
			wantErr: true,
		},
		{
			name:    "empty string",
			address: "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSolanaAddressLoose(tc.address)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateSolanaAddressLoose(%q) error = %v, wantErr %v", tc.address, err, tc.wantErr)
			}
		})
	}
}

func TestIsOnCurve(t *testing.T) {
	tests := []struct {
		name    string
		address string
		onCurve bool
	}{
		{
			name:    "normal wallet address (on-curve)",
			address: "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv",
			onCurve: true,
		},
		{
			name:    "system program (on-curve)",
			address: "11111111111111111111111111111111",
			onCurve: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pk, err := web3.NewPublicKey(tc.address)
			if err != nil {
				t.Fatalf("NewPublicKey failed: %v", err)
			}
			got := IsOnCurve(pk)
			if got != tc.onCurve {
				t.Errorf("IsOnCurve(%q) = %v, want %v", tc.address, got, tc.onCurve)
			}
		})
	}
}

func TestSolanaAccountID_IsOnCurve(t *testing.T) {
	// Test on-curve address
	a := MustNewSolanaFromBase58(SolanaMainnet, "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")
	if !a.IsOnCurve() {
		t.Error("normal wallet address should be on curve")
	}

	// Test nil receiver
	var nilAccount *solanaAccountID
	if nilAccount.IsOnCurve() {
		t.Error("nil account should return false for IsOnCurve")
	}
}
