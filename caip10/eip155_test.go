package caip10

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/donutnomad/eths/ecommon"
	"github.com/fxamacker/cbor/v2"
)

func TestEIP155Parse(t *testing.T) {
	tests := []struct {
		input     string
		chainID   int64
		reference string
		address   string
	}{
		{
			input:     "eip155:1:0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb",
			chainID:   1,
			reference: "1",
			address:   "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb",
		},
		{
			input:     "eip155:137:0x1234567890123456789012345678901234567890",
			chainID:   137,
			reference: "137",
			address:   "0x1234567890123456789012345678901234567890",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			a, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) failed: %v", tc.input, err)
			}

			// Should implement EIP155AccountID interface
			eip, ok := a.(EIP155AccountID)
			if !ok {
				t.Fatalf("expected EIP155AccountID, got %T", a)
			}

			if eip.Namespace() != NamespaceEIP155 {
				t.Errorf("Namespace: got %q, want %q", eip.Namespace(), NamespaceEIP155)
			}
			if eip.Reference() != tc.reference {
				t.Errorf("Reference: got %q, want %q", eip.Reference(), tc.reference)
			}
			if eip.Account().String() != tc.address {
				t.Errorf("Address: got %q, want %q", eip.Address(), tc.address)
			}
			if eip.ChainID().Int64() != tc.chainID {
				t.Errorf("ChainID: got %d, want %d", eip.ChainID().Int64(), tc.chainID)
			}
		})
	}
}

func TestEIP155NewFromHex(t *testing.T) {
	// Test with int
	a := NewEIP155FromHex(1, "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")

	if a.Namespace() != NamespaceEIP155 {
		t.Errorf("Namespace: got %q", a.Namespace())
	}
	if a.Reference() != "1" {
		t.Errorf("Reference: got %q", a.Reference())
	}
	if a.CAIP2() != "eip155:1" {
		t.Errorf("CAIP2: got %q", a.CAIP2())
	}
	if a.ChainID().Int64() != 1 {
		t.Errorf("ChainID: got %s, want 1", a.ChainID())
	}
}

func TestEIP155NewFromAddress(t *testing.T) {
	addr := ecommon.HexToAddress("0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")

	// Test with int64
	a := NewEIP155(int64(1), addr)

	if a.Account() != addr {
		t.Errorf("Account mismatch")
	}
	if a.ChainID().Int64() != 1 {
		t.Errorf("ChainID mismatch")
	}
}

func TestEIP155JSON(t *testing.T) {
	a := NewEIP155FromHex(1, "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal into GenericAccountID (since interface can't be unmarshaled directly)
	var b GenericAccountID
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if !a.Equal(&b) {
		t.Errorf("JSON roundtrip failed: got %v, want %v", b, a)
	}
}

func TestEIP155Binary(t *testing.T) {
	a := NewEIP155FromHex(1, "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")

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

func TestEIP155CBOR(t *testing.T) {
	a := NewEIP155FromHex(1, "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")

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

func TestEIP155Database(t *testing.T) {
	a := NewEIP155FromHex(1, "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")

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

func TestEIP155ToColumns(t *testing.T) {
	a := NewEIP155FromHex(1, "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")

	cols := a.ToColumns()
	if Namespace(cols.Namespace) != NamespaceEIP155 {
		t.Errorf("Namespace: got %q", cols.Namespace)
	}
	if cols.Reference != "1" {
		t.Errorf("Reference: got %q", cols.Reference)
	}

	// Convert back - should use EIP155 parser
	b, err := cols.ToAccountID()
	if err != nil {
		t.Fatalf("ToAccountID failed: %v", err)
	}
	if _, ok := b.(EIP155AccountID); !ok {
		t.Errorf("expected EIP155AccountID, got %T", b)
	}
}

func TestEIP155ZeroValue(t *testing.T) {
	var a *eip155AccountID
	if !a.IsZero() {
		t.Error("nil pointer should be IsZero")
	}
}

func TestEIP155Equal(t *testing.T) {
	a1 := NewEIP155FromHex(1, "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")
	a2 := NewEIP155FromHex(1, "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")
	a3 := NewEIP155FromHex(1, "0x1234567890123456789012345678901234567890")

	if !a1.Equal(a2) {
		t.Error("identical addresses should be equal")
	}
	if a1.Equal(a3) {
		t.Error("different addresses should not be equal")
	}
}

func TestEIP155ParserRegistered(t *testing.T) {
	p, ok := GetParser(NamespaceEIP155)
	if !ok {
		t.Fatal("EIP155 parser not registered")
	}
	if p.Namespace() != NamespaceEIP155 {
		t.Errorf("Namespace: got %q", p.Namespace())
	}
}

func TestEIP155ChainIDNil(t *testing.T) {
	// Test nil receiver returns nil
	var a *eip155AccountID
	if a.ChainID() != nil {
		t.Error("nil receiver should return nil chain ID")
	}
}

func TestEIP155LargeChainID(t *testing.T) {
	// Test large chain ID (e.g., some L2 chains have large IDs)
	a := NewEIP155FromHex(42161, "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb") // Arbitrum One

	if a.Reference() != "42161" {
		t.Errorf("Reference: got %q, want %q", a.Reference(), "42161")
	}
	if a.ChainID().Int64() != 42161 {
		t.Errorf("ChainID mismatch")
	}
}

func TestEIP155SetChainID(t *testing.T) {
	a := NewEIP155FromHex(1, "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")

	// Set new chain ID
	b := a.SetChainID(big.NewInt(137))
	if b == nil {
		t.Fatal("SetChainID returned nil")
	}

	// Original should be unchanged
	if a.ChainID().Int64() != 1 {
		t.Error("original chain ID should not change")
	}

	// New should have updated chain ID
	if b.ChainID().Int64() != 137 {
		t.Errorf("new chain ID: got %d, want 137", b.ChainID().Int64())
	}

	// Address should be the same
	if a.Account() != b.Account() {
		t.Error("address should be the same")
	}

	// Test nil receiver
	var nilAccount *eip155AccountID
	if nilAccount.SetChainID(big.NewInt(1)) != nil {
		t.Error("nil receiver should return nil")
	}
}

func TestEIP155SetAddress(t *testing.T) {
	a := NewEIP155FromHex(1, "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")
	newAddr := ecommon.HexToAddress("0x1234567890123456789012345678901234567890")

	// Set new address
	b := a.SetAddress(newAddr)
	if b == nil {
		t.Fatal("SetAddress returned nil")
	}

	// Original should be unchanged
	if a.Address() != "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb" {
		t.Error("original address should not change")
	}

	// New should have updated address
	if b.Account() != newAddr {
		t.Errorf("new address mismatch")
	}

	// Chain ID should be the same
	if a.ChainID().Cmp(b.ChainID()) != 0 {
		t.Error("chain ID should be the same")
	}

	// Test nil receiver
	var nilAccount *eip155AccountID
	if nilAccount.SetAddress(newAddr) != nil {
		t.Error("nil receiver should return nil")
	}
}

func TestEIP155GenericTypes(t *testing.T) {
	addr := ecommon.HexToAddress("0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")

	// Test various integer types
	tests := []struct {
		name    string
		create  func() EIP155AccountID
		chainID int64
	}{
		{"int", func() EIP155AccountID { return NewEIP155(1, addr) }, 1},
		{"int64", func() EIP155AccountID { return NewEIP155(int64(137), addr) }, 137},
		{"uint64", func() EIP155AccountID { return NewEIP155(uint64(42161), addr) }, 42161},
		{"*big.Int", func() EIP155AccountID { return NewEIP155(big.NewInt(10), addr) }, 10},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.create()
			if a.ChainID().Int64() != tc.chainID {
				t.Errorf("chainID: got %d, want %d", a.ChainID().Int64(), tc.chainID)
			}
		})
	}
}

func TestEIP155MaxChainID(t *testing.T) {
	addr := ecommon.HexToAddress("0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")

	// Create max chain ID (10^32 - 1)
	maxChainID := new(big.Int)
	maxChainID.Exp(big.NewInt(10), big.NewInt(32), nil)
	maxChainID.Sub(maxChainID, big.NewInt(1))

	// Test with exactly max value
	a := NewEIP155(maxChainID, addr)
	if a.ChainID().Cmp(maxChainID) != 0 {
		t.Errorf("chain ID should be max value")
	}
	// Reference should be 32 characters of '9'
	if len(a.Reference()) != 32 {
		t.Errorf("Reference length: got %d, want 32", len(a.Reference()))
	}

	// Test with value exceeding max (10^32)
	overMax := new(big.Int)
	overMax.Exp(big.NewInt(10), big.NewInt(32), nil) // 10^32

	b := NewEIP155(overMax, addr)
	// Should be capped to max
	if b.ChainID().Cmp(maxChainID) != 0 {
		t.Errorf("chain ID should be capped to max value, got %s", b.ChainID().String())
	}

	// Test with very large value (10^50)
	veryLarge := new(big.Int)
	veryLarge.Exp(big.NewInt(10), big.NewInt(50), nil)

	c := NewEIP155(veryLarge, addr)
	// Should be capped to max
	if c.ChainID().Cmp(maxChainID) != 0 {
		t.Errorf("chain ID should be capped to max value")
	}
}
