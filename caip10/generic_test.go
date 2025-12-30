package caip10

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestGenericParse(t *testing.T) {
	tests := []struct {
		input     string
		namespace string
		reference string
		address   string
		chainID   string
	}{
		{
			input:     "cosmos:cosmoshub-3:cosmos1t2uflqwqe0fsj0shcfkrvpukewcw40yjj6hdc0",
			namespace: "cosmos",
			reference: "cosmoshub-3",
			address:   "cosmos1t2uflqwqe0fsj0shcfkrvpukewcw40yjj6hdc0",
			chainID:   "cosmos:cosmoshub-3",
		},
		{
			input:     "bip122:000000000019d6689c085ae165831e93:128Lkh3S7CkDTBZ8W7BbpsN3YYizJMp8p6",
			namespace: "bip122",
			reference: "000000000019d6689c085ae165831e93",
			address:   "128Lkh3S7CkDTBZ8W7BbpsN3YYizJMp8p6",
			chainID:   "bip122:000000000019d6689c085ae165831e93",
		},
		{
			input:     "polkadot:b0a8d493285c2df73290dfb7e61f870f:5hmuyxw9xdgbpptgypokw4thfyoe3ryenebr381z9iaegmfy",
			namespace: "polkadot",
			reference: "b0a8d493285c2df73290dfb7e61f870f",
			address:   "5hmuyxw9xdgbpptgypokw4thfyoe3ryenebr381z9iaegmfy",
			chainID:   "polkadot:b0a8d493285c2df73290dfb7e61f870f",
		},
		{
			input:     "hedera:mainnet:0.0.1234567890-zbhlt",
			namespace: "hedera",
			reference: "mainnet",
			address:   "0.0.1234567890-zbhlt",
			chainID:   "hedera:mainnet",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			a, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) failed: %v", tc.input, err)
			}
			if a.Namespace() != Namespace(tc.namespace) {
				t.Errorf("Namespace: got %q, want %q", a.Namespace(), tc.namespace)
			}
			if a.Reference() != tc.reference {
				t.Errorf("Reference: got %q, want %q", a.Reference(), tc.reference)
			}
			if a.Address() != tc.address {
				t.Errorf("Address: got %q, want %q", a.Address(), tc.address)
			}
			if a.ChainID().String() != (tc.chainID) {
				t.Errorf("ChainID: got %q, want %q", a.ChainID(), tc.chainID)
			}
			if a.String() != tc.input {
				t.Errorf("String: got %q, want %q", a.String(), tc.input)
			}
		})
	}
}

func TestParseInvalid(t *testing.T) {
	invalidTestCases := []string{
		"",                      // empty
		"eip155",                // missing parts
		"eip155:1",              // missing address
		"EIP155:1:0xabc",        // namespace must be lowercase (generic)
		"ab:1:0xabc",            // namespace too short
		"abcdefghi:1:0xabc",     // namespace too long
		"cosmos::addr",          // empty reference
		"cosmos:ref:",           // empty address
		"cosmos:abc!def:addr",   // invalid character in reference
		"cosmos:ref:addr/path",  // slash not allowed in address
		"cosmos:ref:addr\\back", // backslash not allowed
	}

	for _, tc := range invalidTestCases {
		t.Run(tc, func(t *testing.T) {
			_, err := Parse(tc)
			if err == nil {
				t.Errorf("Parse(%q) should have failed", tc)
			}
		})
	}
}

func TestGenericAccountID(t *testing.T) {
	a, err := NewGeneric("cosmos", "cosmoshub-3", "cosmos1abc")
	if err != nil {
		t.Fatalf("NewGeneric failed: %v", err)
	}

	if a.Namespace() != "cosmos" {
		t.Errorf("Namespace: got %q", a.Namespace())
	}
	if a.Reference() != "cosmoshub-3" {
		t.Errorf("Reference: got %q", a.Reference())
	}
	if a.Address() != "cosmos1abc" {
		t.Errorf("Address: got %q", a.Address())
	}
	if a.String() != "cosmos:cosmoshub-3:cosmos1abc" {
		t.Errorf("String: got %q", a.String())
	}
}

func TestGenericJSON(t *testing.T) {
	a := MustNewGeneric("cosmos", "cosmoshub-3", "cosmos1abc")

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	expected := `"cosmos:cosmoshub-3:cosmos1abc"`
	if string(data) != expected {
		t.Errorf("Marshal: got %s, want %s", data, expected)
	}

	var b GenericAccountID
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if !a.Equal(&b) {
		t.Errorf("Unmarshal: got %v, want %v", b, a)
	}
}

func TestGenericBinary(t *testing.T) {
	a := MustNewGeneric("cosmos", "cosmoshub-3", "cosmos1abc")

	data, err := a.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary failed: %v", err)
	}

	var b GenericAccountID
	if err := b.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary failed: %v", err)
	}

	if !a.Equal(&b) {
		t.Errorf("Binary roundtrip: got %v, want %v", b, a)
	}
}

func TestGenericCBOR(t *testing.T) {
	a := MustNewGeneric("cosmos", "cosmoshub-3", "cosmos1abc")

	data, err := cbor.Marshal(a)
	if err != nil {
		t.Fatalf("cbor.Marshal failed: %v", err)
	}

	var b GenericAccountID
	if err := cbor.Unmarshal(data, &b); err != nil {
		t.Fatalf("cbor.Unmarshal failed: %v", err)
	}

	if !a.Equal(&b) {
		t.Errorf("CBOR roundtrip: got %v, want %v", b, a)
	}
}

func TestGenericDatabase(t *testing.T) {
	a := MustNewGeneric("cosmos", "cosmoshub-3", "cosmos1abc")

	// Value
	v, err := a.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}
	if v != "cosmos:cosmoshub-3:cosmos1abc" {
		t.Errorf("Value: got %v", v)
	}

	// Scan string
	var b GenericAccountID
	if err := b.Scan("cosmos:cosmoshub-3:cosmos1abc"); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if !a.Equal(&b) {
		t.Errorf("Scan: got %v, want %v", b, a)
	}

	// Scan nil
	var c GenericAccountID
	if err := c.Scan(nil); err != nil {
		t.Fatalf("Scan nil failed: %v", err)
	}
	if !c.IsZero() {
		t.Error("Scan nil should be zero")
	}
}

func TestAccountIDColumns(t *testing.T) {
	a := MustNewGeneric("cosmos", "cosmoshub-3", "cosmos1abc")

	cols := a.ToColumns()
	if cols.Namespace != "cosmos" {
		t.Errorf("Namespace: got %q", cols.Namespace)
	}
	if cols.Reference != "cosmoshub-3" {
		t.Errorf("Reference: got %q", cols.Reference)
	}
	if cols.Address != "cosmos1abc" {
		t.Errorf("Address: got %q", cols.Address)
	}

	// Convert back
	b, err := cols.ToAccountID()
	if err != nil {
		t.Fatalf("ToAccountID failed: %v", err)
	}
	if !a.Equal(b) {
		t.Errorf("roundtrip failed: got %v, want %v", b, a)
	}
}

func TestAccountIDColumnsCompact(t *testing.T) {
	a := MustNewGeneric("cosmos", "cosmoshub-3", "cosmos1abc")

	// Test ToColumnsCompact
	compact := a.ToColumnsCompact()
	if compact.ChainID != "cosmos:cosmoshub-3" {
		t.Errorf("ChainID: got %q, want %q", compact.ChainID, "cosmos:cosmoshub-3")
	}
	if compact.Address != "cosmos1abc" {
		t.Errorf("Address: got %q, want %q", compact.Address, "cosmos1abc")
	}

	// Test String
	if compact.String() != "cosmos:cosmoshub-3:cosmos1abc" {
		t.Errorf("String: got %q", compact.String())
	}

	// Test ToAccountID
	b, err := compact.ToAccountID()
	if err != nil {
		t.Fatalf("ToAccountID failed: %v", err)
	}
	if !a.Equal(b) {
		t.Errorf("roundtrip failed: got %v, want %v", b, a)
	}

	// Test ToFull
	full, err := compact.ToFull()
	if err != nil {
		t.Fatalf("ToFull failed: %v", err)
	}
	if full.Namespace != "cosmos" {
		t.Errorf("ToFull Namespace: got %q", full.Namespace)
	}
	if full.Reference != "cosmoshub-3" {
		t.Errorf("ToFull Reference: got %q", full.Reference)
	}
	if full.Address != "cosmos1abc" {
		t.Errorf("ToFull Address: got %q", full.Address)
	}

	// Test ToCompact from AccountIDColumns
	cols := a.ToColumns()
	compact2 := cols.ToCompact()
	if compact2.ChainID != compact.ChainID {
		t.Errorf("ToCompact ChainID mismatch: got %q, want %q", compact2.ChainID, compact.ChainID)
	}
	if compact2.Address != compact.Address {
		t.Errorf("ToCompact Address mismatch: got %q, want %q", compact2.Address, compact.Address)
	}
}

func TestAccountIDColumnsCompactZero(t *testing.T) {
	var compact AccountIDColumnsCompact
	if !compact.IsZero() {
		t.Error("zero value should be IsZero")
	}
	if compact.String() != "" {
		t.Errorf("zero String should be empty, got %q", compact.String())
	}

	// ToAccountID should fail for zero value
	_, err := compact.ToAccountID()
	if err == nil {
		t.Error("ToAccountID should fail for zero value")
	}

	// ToFull should return zero columns
	full, err := compact.ToFull()
	if err != nil {
		t.Fatalf("ToFull failed: %v", err)
	}
	if !full.IsZero() {
		t.Error("ToFull of zero should be zero")
	}
}

func TestAccountIDColumnsCompactSpecializedTypes(t *testing.T) {
	// 测试 EIP155 类型还原
	t.Run("EIP155", func(t *testing.T) {
		eip := NewEIP155FromHex(1, "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")
		compact := eip.ToColumnsCompact()

		recovered, err := compact.ToAccountID()
		if err != nil {
			t.Fatalf("ToAccountID failed: %v", err)
		}

		eipRecovered, ok := recovered.(EIP155AccountID)
		if !ok {
			t.Fatalf("expected EIP155AccountID, got %T", recovered)
		}
		if eipRecovered.EIP155ChainID().Cmp(eip.EIP155ChainID()) != 0 {
			t.Errorf("ChainID mismatch: got %s, want %s", eipRecovered.ChainID(), eip.ChainID())
		}
	})

	// 测试 Solana 类型还原
	t.Run("Solana", func(t *testing.T) {
		sol := MustNewSolanaFromBase58(SolanaMainnet, "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")
		compact := sol.ToColumnsCompact()

		recovered, err := compact.ToAccountID()
		if err != nil {
			t.Fatalf("ToAccountID failed: %v", err)
		}

		solRecovered, ok := recovered.(SolanaAccountID)
		if !ok {
			t.Fatalf("expected SolanaAccountID, got %T", recovered)
		}
		if solRecovered.Account().String() != sol.Account().String() {
			t.Errorf("Account mismatch")
		}
	})

	// 测试 BIP122 类型还原
	t.Run("BIP122", func(t *testing.T) {
		btc := NewBitcoinMainnet("bc1qwz2lhc40s8ty3l5jg3plpve3y3l82x9l42q7fk")
		compact := btc.ToColumnsCompact()

		recovered, err := compact.ToAccountID()
		if err != nil {
			t.Fatalf("ToAccountID failed: %v", err)
		}

		btcRecovered, ok := recovered.(BIP122AccountID)
		if !ok {
			t.Fatalf("expected BIP122AccountID, got %T", recovered)
		}
		if btcRecovered.Network() != BitcoinMainnet {
			t.Errorf("Network mismatch: got %s, want %s", btcRecovered.Network(), BitcoinMainnet)
		}
	})
}

func TestSplitCAIP2(t *testing.T) {
	tests := []struct {
		input     string
		namespace string
		reference string
		wantErr   bool
	}{
		{
			input:     "eip155:1",
			namespace: "eip155",
			reference: "1",
			wantErr:   false,
		},
		{
			input:     "solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp",
			namespace: "solana",
			reference: "5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp",
			wantErr:   false,
		},
		{
			input:     "bip122:000000000019d6689c085ae165831e93",
			namespace: "bip122",
			reference: "000000000019d6689c085ae165831e93",
			wantErr:   false,
		},
		{
			input:   "",
			wantErr: true,
		},
		{
			input:   "eip155",
			wantErr: true,
		},
		{
			input:   "eip155:1:0xabc", // This is CAIP-10, not CAIP-2
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			ns, ref, err := SplitCAIP2(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("SplitCAIP2(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
				return
			}
			if err == nil {
				if ns != tc.namespace {
					t.Errorf("namespace: got %q, want %q", ns, tc.namespace)
				}
				if ref != tc.reference {
					t.Errorf("reference: got %q, want %q", ref, tc.reference)
				}
			}
		})
	}
}

func TestMustParsePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse should panic on invalid input")
		}
	}()
	MustParse("invalid")
}

func TestZeroValues(t *testing.T) {
	var g GenericAccountID
	if !g.IsZero() {
		t.Error("zero GenericAccountID should be IsZero")
	}
	if g.String() != "" {
		t.Errorf("zero String should be empty, got %q", g.String())
	}
}

func TestEqual(t *testing.T) {
	a1 := MustNewGeneric("cosmos", "hub", "addr1")
	a2 := MustNewGeneric("cosmos", "hub", "addr1")
	a3 := MustNewGeneric("cosmos", "hub", "addr2")

	if !Equal(a1, a2) {
		t.Error("identical should be equal")
	}
	if Equal(a1, a3) {
		t.Error("different addresses should not be equal")
	}
	if !Equal(nil, nil) {
		t.Error("nil == nil should be true")
	}
}

func TestGenericAccountID_Validate(t *testing.T) {
	tests := []struct {
		name      string
		accountID *GenericAccountID
		wantErr   bool
		errType   error // 期望的错误类型
	}{
		// nil 测试
		{
			name:      "nil account ID",
			accountID: nil,
			wantErr:   true,
			errType:   ErrEmptyValue,
		},

		// 无效 namespace 测试
		{
			name:      "namespace too short",
			accountID: newGenericUnchecked("ab", "ref", "addr"),
			wantErr:   true,
			errType:   ErrInvalidNamespace,
		},
		{
			name:      "namespace too long",
			accountID: newGenericUnchecked("abcdefghi", "ref", "addr"),
			wantErr:   true,
			errType:   ErrInvalidNamespace,
		},
		{
			name:      "namespace with uppercase",
			accountID: newGenericUnchecked("EIP155", "1", "0xabc"),
			wantErr:   true,
			errType:   ErrInvalidNamespace,
		},
		{
			name:      "namespace with special chars",
			accountID: newGenericUnchecked("abc@def", "ref", "addr"),
			wantErr:   true,
			errType:   ErrInvalidNamespace,
		},
		{
			name:      "empty namespace",
			accountID: newGenericUnchecked("", "ref", "addr"),
			wantErr:   true,
			errType:   ErrInvalidNamespace,
		},

		// EIP155 namespace 测试
		{
			name:      "eip155 valid",
			accountID: newGenericUnchecked(NamespaceEIP155, "1", "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb"),
			wantErr:   false,
		},
		{
			name:      "eip155 invalid chain id",
			accountID: newGenericUnchecked(NamespaceEIP155, "abc", "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb"),
			wantErr:   true,
			errType:   ErrInvalidReference,
		},
		{
			name:      "eip155 invalid address",
			accountID: newGenericUnchecked(NamespaceEIP155, "1", "invalid"),
			wantErr:   true,
		},
		{
			name:      "eip155 large chain id",
			accountID: newGenericUnchecked(NamespaceEIP155, "56", "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb"),
			wantErr:   false,
		},

		// Solana namespace 测试
		{
			name:      "solana valid mainnet",
			accountID: newGenericUnchecked(NamespaceSolana, string(SolanaMainnet), "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv"),
			wantErr:   false,
		},
		{
			name:      "solana valid devnet",
			accountID: newGenericUnchecked(NamespaceSolana, string(SolanaDevnet), "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv"),
			wantErr:   false,
		},
		{
			name:      "solana invalid address",
			accountID: newGenericUnchecked(NamespaceSolana, string(SolanaMainnet), "invalid"),
			wantErr:   true,
			errType:   ErrInvalidAddress,
		},
		{
			name:      "solana address too short",
			accountID: newGenericUnchecked(NamespaceSolana, string(SolanaMainnet), "abc"),
			wantErr:   true,
			errType:   ErrInvalidAddress,
		},

		// BIP122 namespace 测试
		{
			name:      "bip122 valid bitcoin mainnet",
			accountID: newGenericUnchecked(NamespaceBIP122, string(BitcoinMainnet), "bc1qwz2lhc40s8ty3l5jg3plpve3y3l82x9l42q7fk"),
			wantErr:   false,
		},
		{
			name:      "bip122 valid bitcoin testnet",
			accountID: newGenericUnchecked(NamespaceBIP122, string(BitcoinTestnet), "tb1qw508d6qejxtdg4y5r3zarvary0c5xw7kxpjzsx"),
			wantErr:   false,
		},
		{
			name:      "bip122 invalid address",
			accountID: newGenericUnchecked(NamespaceBIP122, string(BitcoinMainnet), "invalid!address"),
			wantErr:   true,
			errType:   ErrInvalidAddress,
		},
		{
			name:      "bip122 empty address",
			accountID: newGenericUnchecked(NamespaceBIP122, string(BitcoinMainnet), ""),
			wantErr:   true,
			errType:   ErrInvalidAddress,
		},

		// 通用 namespace (default case) 测试
		{
			name:      "generic valid cosmos",
			accountID: newGenericUnchecked("cosmos", "cosmoshub-3", "cosmos1t2uflqwqe0fsj0shcfkrvpukewcw40yjj6hdc0"),
			wantErr:   false,
		},
		{
			name:      "generic valid polkadot",
			accountID: newGenericUnchecked("polkadot", "b0a8d493285c2df73290dfb7e61f870f", "5hmuyxw9xdgbpptgypokw4thfyoe3ryenebr381z9iaegmfy"),
			wantErr:   false,
		},
		{
			name:      "generic valid hedera",
			accountID: newGenericUnchecked("hedera", "mainnet", "0.0.1234567890-zbhlt"),
			wantErr:   false,
		},
		{
			name:      "generic invalid reference with special char",
			accountID: newGenericUnchecked("cosmos", "hub@invalid", "addr"),
			wantErr:   true,
			errType:   ErrInvalidReference,
		},
		{
			name:      "generic reference too long (33 chars)",
			accountID: newGenericUnchecked("cosmos", "123456789012345678901234567890123", "addr"),
			wantErr:   true,
			errType:   ErrInvalidReference,
		},
		{
			name:      "generic empty reference",
			accountID: newGenericUnchecked("cosmos", "", "addr"),
			wantErr:   true,
			errType:   ErrInvalidReference,
		},
		{
			name:      "generic invalid address with slash",
			accountID: newGenericUnchecked("cosmos", "hub", "addr/path"),
			wantErr:   true,
			errType:   ErrInvalidAddress,
		},
		{
			name:      "generic invalid address with backslash",
			accountID: newGenericUnchecked("cosmos", "hub", "addr\\back"),
			wantErr:   true,
			errType:   ErrInvalidAddress,
		},
		{
			name:      "generic address too long (129 chars)",
			accountID: newGenericUnchecked("cosmos", "hub", "a"+string(make([]byte, 128))),
			wantErr:   true,
			errType:   ErrInvalidAddress,
		},
		{
			name:      "generic empty address",
			accountID: newGenericUnchecked("cosmos", "hub", ""),
			wantErr:   true,
			errType:   ErrInvalidAddress,
		},

		// 边界值测试
		{
			name:      "namespace min length (3 chars)",
			accountID: newGenericUnchecked("abc", "ref", "addr"),
			wantErr:   false,
		},
		{
			name:      "namespace max length (8 chars)",
			accountID: newGenericUnchecked("abcdefgh", "ref", "addr"),
			wantErr:   false,
		},
		{
			name:      "reference min length (1 char)",
			accountID: newGenericUnchecked("cosmos", "a", "addr"),
			wantErr:   false,
		},
		{
			name:      "reference max length (32 chars)",
			accountID: newGenericUnchecked("cosmos", "12345678901234567890123456789012", "addr"),
			wantErr:   false,
		},
		{
			name:      "address min length (1 char)",
			accountID: newGenericUnchecked("cosmos", "hub", "a"),
			wantErr:   false,
		},
		{
			name:      "address max length (128 chars)",
			accountID: newGenericUnchecked("cosmos", "hub", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			wantErr:   false,
		},
		{
			name:      "namespace with hyphen",
			accountID: newGenericUnchecked("abc-def", "ref", "addr"),
			wantErr:   false,
		},
		{
			name:      "namespace with number",
			accountID: newGenericUnchecked("abc123", "ref", "addr"),
			wantErr:   false,
		},
		{
			name:      "reference with underscore",
			accountID: newGenericUnchecked("cosmos", "hub_test", "addr"),
			wantErr:   false,
		},
		{
			name:      "reference with hyphen",
			accountID: newGenericUnchecked("cosmos", "hub-test", "addr"),
			wantErr:   false,
		},
		{
			name:      "address with dot",
			accountID: newGenericUnchecked("hedera", "mainnet", "0.0.12345"),
			wantErr:   false,
		},
		{
			name:      "address with hyphen",
			accountID: newGenericUnchecked("hedera", "mainnet", "0.0.1234567890-zbhlt"),
			wantErr:   false,
		},
		{
			name:      "address with percent",
			accountID: newGenericUnchecked("cosmos", "hub", "addr%20test"),
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.accountID.Validate()
			if tc.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
				} else if tc.errType != nil && !errors.Is(err, tc.errType) {
					t.Errorf("Validate() error = %v, want error type %v", err, tc.errType)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestToNative(t *testing.T) {
	// Test EIP155 conversion
	t.Run("eip155", func(t *testing.T) {
		var g GenericAccountID
		err := g.UnmarshalText([]byte("eip155:1:0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb"))
		if err != nil {
			t.Fatalf("UnmarshalText failed: %v", err)
		}

		native := g.ToNative()
		eip, ok := native.(EIP155AccountID)
		if !ok {
			t.Fatalf("expected EIP155AccountID, got %T", native)
		}
		if eip.Namespace() != NamespaceEIP155 {
			t.Errorf("Namespace: got %q", eip.Namespace())
		}
		if eip.Account().String() != "0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb" {
			t.Errorf("Account: got %q", eip.Account().String())
		}
	})

	// Test Solana conversion
	t.Run("solana", func(t *testing.T) {
		var g GenericAccountID
		err := g.UnmarshalText([]byte("solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp:7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv"))
		if err != nil {
			t.Fatalf("UnmarshalText failed: %v", err)
		}

		native := g.ToNative()
		sol, ok := native.(SolanaAccountID)
		if !ok {
			t.Fatalf("expected SolanaAccountID, got %T", native)
		}
		if sol.Namespace() != NamespaceSolana {
			t.Errorf("Namespace: got %q", sol.Namespace())
		}
		if sol.Account().String() != "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv" {
			t.Errorf("Account: got %q", sol.Account().String())
		}
	})

	// Test unknown namespace returns self
	t.Run("unknown namespace", func(t *testing.T) {
		g := MustNewGeneric("cosmos", "cosmoshub-3", "cosmos1abc")
		native := g.ToNative()
		gen, ok := native.(*GenericAccountID)
		if !ok {
			t.Fatalf("expected *GenericAccountID, got %T", native)
		}
		if gen != g {
			t.Error("should return same pointer for unknown namespace")
		}
	})

	// Test nil receiver
	t.Run("nil", func(t *testing.T) {
		var g *GenericAccountID
		native := g.ToNative()
		if native != nil {
			t.Errorf("nil receiver should return nil, got %T", native)
		}
	})
}
