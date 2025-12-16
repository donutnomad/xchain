package caip10

import (
	"encoding/json"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestBIP122Parse(t *testing.T) {
	tests := []struct {
		input   string
		network BIP122Network
		address string
	}{
		{
			input:   "bip122:000000000019d6689c085ae165831e93:35PBEaofpUeH8VnnNSorM1QZsadrZoQp4N",
			network: BitcoinMainnet,
			address: "35PBEaofpUeH8VnnNSorM1QZsadrZoQp4N",
		},
		{
			input:   "bip122:000000000019d6689c085ae165831e93:bc1qwz2lhc40s8ty3l5jg3plpve3y3l82x9l42q7fk",
			network: BitcoinMainnet,
			address: "bc1qwz2lhc40s8ty3l5jg3plpve3y3l82x9l42q7fk",
		},
		{
			input:   "bip122:000000000019d6689c085ae165831e93:bc1pmzfrwwndsqmk5yh69yjr5lfgfg4ev8c0tsc06e",
			network: BitcoinMainnet,
			address: "bc1pmzfrwwndsqmk5yh69yjr5lfgfg4ev8c0tsc06e",
		},
		{
			input:   "bip122:1a91e3dace36e2be3bf030a65679fe82:DBcZSePDaMMduBMLymWHXhkE5ArFEvkagU",
			network: DogecoinMainnet,
			address: "DBcZSePDaMMduBMLymWHXhkE5ArFEvkagU",
		},
		{
			input:   "bip122:12a765e31ffd4059bada1e25190f6e98:ltc1q8c6fshw2dlwun7ekn9qwf37cu2rn755u9ym7p0",
			network: LitecoinMainnet,
			address: "ltc1q8c6fshw2dlwun7ekn9qwf37cu2rn755u9ym7p0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			a, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) failed: %v", tc.input, err)
			}

			// Should implement BIP122AccountID interface
			bip, ok := a.(BIP122AccountID)
			if !ok {
				t.Fatalf("expected BIP122AccountID, got %T", a)
			}

			if bip.Namespace() != NamespaceBIP122 {
				t.Errorf("Namespace: got %q, want %q", bip.Namespace(), NamespaceBIP122)
			}
			if bip.Network() != tc.network {
				t.Errorf("Network: got %q, want %q", bip.Network(), tc.network)
			}
			if bip.Address() != tc.address {
				t.Errorf("Address: got %q, want %q", bip.Address(), tc.address)
			}
		})
	}
}

func TestBIP122NewBitcoinMainnet(t *testing.T) {
	address := "bc1qwz2lhc40s8ty3l5jg3plpve3y3l82x9l42q7fk"
	a := NewBitcoinMainnet(address)

	if a.Namespace() != NamespaceBIP122 {
		t.Errorf("Namespace: got %q", a.Namespace())
	}
	if a.Network() != BitcoinMainnet {
		t.Errorf("Network: got %q", a.Network())
	}
	if a.Address() != address {
		t.Errorf("Address: got %q", a.Address())
	}
	if a.ChainID() != ChainIDBitcoinMainnet {
		t.Errorf("CAIP2: got %q", a.ChainID())
	}
}

func TestBIP122NetworkTypes(t *testing.T) {
	tests := []struct {
		name    string
		create  func() BIP122AccountID
		network BIP122Network
	}{
		{
			name:    "Bitcoin mainnet",
			create:  func() BIP122AccountID { return NewBitcoinMainnet("bc1qtest") },
			network: BitcoinMainnet,
		},
		{
			name:    "Bitcoin testnet",
			create:  func() BIP122AccountID { return NewBitcoinTestnet("tb1qtest") },
			network: BitcoinTestnet,
		},
		{
			name:    "Litecoin mainnet",
			create:  func() BIP122AccountID { return NewLitecoinMainnet("ltc1qtest") },
			network: LitecoinMainnet,
		},
		{
			name:    "Dogecoin mainnet",
			create:  func() BIP122AccountID { return NewDogecoinMainnet("Dtest") },
			network: DogecoinMainnet,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.create()
			if a.Network() != tc.network {
				t.Errorf("Network: got %q, want %q", a.Network(), tc.network)
			}
		})
	}
}

func TestBIP122SetAddress(t *testing.T) {
	a := NewBitcoinMainnet("bc1qold")
	b := a.SetAddress("bc1qnew")

	if a.Address() != "bc1qold" {
		t.Error("original address should not change")
	}
	if b.Address() != "bc1qnew" {
		t.Errorf("new address: got %q, want %q", b.Address(), "bc1qnew")
	}
	if a.Network() != b.Network() {
		t.Error("network should be the same")
	}
}

func TestBIP122JSON(t *testing.T) {
	a := NewBitcoinMainnet("35PBEaofpUeH8VnnNSorM1QZsadrZoQp4N")

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	expected := `"bip122:000000000019d6689c085ae165831e93:35PBEaofpUeH8VnnNSorM1QZsadrZoQp4N"`
	if string(data) != expected {
		t.Errorf("Marshal: got %s, want %s", data, expected)
	}

	var b GenericAccountID
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if !a.Equal(&b) {
		t.Errorf("JSON roundtrip failed: got %v, want %v", b, a)
	}
}

func TestBIP122Binary(t *testing.T) {
	a := NewBitcoinMainnet("35PBEaofpUeH8VnnNSorM1QZsadrZoQp4N")

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

func TestBIP122CBOR(t *testing.T) {
	a := NewBitcoinMainnet("35PBEaofpUeH8VnnNSorM1QZsadrZoQp4N")

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

func TestBIP122Database(t *testing.T) {
	a := NewBitcoinMainnet("35PBEaofpUeH8VnnNSorM1QZsadrZoQp4N")

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

func TestBIP122ToColumns(t *testing.T) {
	a := NewBitcoinMainnet("35PBEaofpUeH8VnnNSorM1QZsadrZoQp4N")

	cols := a.ToColumns()
	if Namespace(cols.Namespace) != NamespaceBIP122 {
		t.Errorf("Namespace: got %q", cols.Namespace)
	}
	if cols.Reference != BitcoinMainnet.String() {
		t.Errorf("Reference: got %q", cols.Reference)
	}

	// Convert back - should use BIP122 parser
	b, err := cols.ToAccountID()
	if err != nil {
		t.Fatalf("ToAccountID failed: %v", err)
	}
	if _, ok := b.(BIP122AccountID); !ok {
		t.Errorf("expected BIP122AccountID, got %T", b)
	}
}

func TestBIP122ZeroValue(t *testing.T) {
	var a *bip122AccountID
	if !a.IsZero() {
		t.Error("nil pointer should be IsZero")
	}
}

func TestBIP122Equal(t *testing.T) {
	a1 := NewBitcoinMainnet("35PBEaofpUeH8VnnNSorM1QZsadrZoQp4N")
	a2 := NewBitcoinMainnet("35PBEaofpUeH8VnnNSorM1QZsadrZoQp4N")
	a3 := NewBitcoinMainnet("bc1qwz2lhc40s8ty3l5jg3plpve3y3l82x9l42q7fk")

	if !a1.Equal(a2) {
		t.Error("identical addresses should be equal")
	}
	if a1.Equal(a3) {
		t.Error("different addresses should not be equal")
	}
}

func TestBIP122ParserRegistered(t *testing.T) {
	p, ok := GetParser(NamespaceBIP122)
	if !ok {
		t.Fatal("BIP122 parser not registered")
	}
	if p.Namespace() != NamespaceBIP122 {
		t.Errorf("Namespace: got %q", p.Namespace())
	}
}

func TestValidateBIP122Address(t *testing.T) {
	tests := []struct {
		name    string
		network BIP122Network
		address string
		wantErr bool
	}{
		// Bitcoin mainnet
		{
			name:    "Bitcoin mainnet P2SH",
			network: BitcoinMainnet,
			address: "35PBEaofpUeH8VnnNSorM1QZsadrZoQp4N",
			wantErr: false,
		},
		{
			name:    "Bitcoin mainnet SegWit",
			network: BitcoinMainnet,
			address: "bc1qwz2lhc40s8ty3l5jg3plpve3y3l82x9l42q7fk",
			wantErr: false,
		},
		{
			name:    "Bitcoin mainnet Taproot",
			network: BitcoinMainnet,
			address: "bc1pmzfrwwndsqmk5yh69yjr5lfgfg4ev8c0tsc06e",
			wantErr: false,
		},
		{
			name:    "Bitcoin mainnet invalid",
			network: BitcoinMainnet,
			address: "invalid",
			wantErr: true,
		},
		// Dogecoin mainnet
		{
			name:    "Dogecoin mainnet P2PKH",
			network: DogecoinMainnet,
			address: "DBcZSePDaMMduBMLymWHXhkE5ArFEvkagU",
			wantErr: false,
		},
		// Litecoin mainnet
		{
			name:    "Litecoin mainnet SegWit",
			network: LitecoinMainnet,
			address: "ltc1q8c6fshw2dlwun7ekn9qwf37cu2rn755u9ym7p0",
			wantErr: false,
		},
		// Empty address
		{
			name:    "empty address",
			network: BitcoinMainnet,
			address: "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateBIP122Address(tc.network, tc.address)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateBIP122Address(%q, %q) error = %v, wantErr %v", tc.network, tc.address, err, tc.wantErr)
			}
		})
	}
}

func TestBIP122WithValidation(t *testing.T) {
	// Valid address
	a, err := NewBIP122WithValidation(BitcoinMainnet, "35PBEaofpUeH8VnnNSorM1QZsadrZoQp4N")
	if err != nil {
		t.Fatalf("NewBIP122WithValidation failed: %v", err)
	}
	if a.Address() != "35PBEaofpUeH8VnnNSorM1QZsadrZoQp4N" {
		t.Errorf("Address mismatch")
	}

	// Invalid address
	_, err = NewBIP122WithValidation(BitcoinMainnet, "invalid")
	if err == nil {
		t.Error("expected error for invalid address")
	}
}

func TestBIP122NilReceiver(t *testing.T) {
	var a *bip122AccountID

	if a.Network() != "" {
		t.Error("nil receiver Network should return empty string")
	}
	if a.SetAddress("test") != nil {
		t.Error("nil receiver SetAddress should return nil")
	}
}
