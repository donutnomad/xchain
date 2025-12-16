package caip10

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time interface checks
var (
	_ encoding.TextMarshaler     = ChainID{}
	_ encoding.TextUnmarshaler   = (*ChainID)(nil)
	_ encoding.BinaryMarshaler   = ChainID{}
	_ encoding.BinaryUnmarshaler = (*ChainID)(nil)
	_ json.Marshaler             = ChainID{}
	_ json.Unmarshaler           = (*ChainID)(nil)
	_ driver.Valuer              = ChainID{}
	_ sql.Scanner                = (*ChainID)(nil)
)

// --- Constructor Tests ---

func TestNewChainIDByEIP155(t *testing.T) {
	tests := []struct {
		chainID uint64
		want    string
	}{
		{1, "eip155:1"},
		{137, "eip155:137"},
		{11155111, "eip155:11155111"},
	}

	for _, tt := range tests {
		c := NewEIP155ChainID(tt.chainID)
		assert.Equal(t, NamespaceEIP155, c.Namespace)
		assert.Equal(t, tt.want, c.String())
	}
}

func TestNewChainIDBySolana(t *testing.T) {
	c := NewSolanaChainID(SolanaMainnet)
	assert.Equal(t, NamespaceSolana, c.Namespace)
	assert.Equal(t, "solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp", c.String())

	c2 := NewSolanaChainID(SolanaDevnet)
	assert.Equal(t, "solana:EtWTRABZaYq6iMfeYKouRu166VU2xqa1", c2.String())
}

func TestNewChainIDByBIP122(t *testing.T) {
	// Valid block hash (32 hex chars)
	c, err := NewBIP122ChainID("000000000019d6689c085ae165831e93")
	require.NoError(t, err)
	assert.Equal(t, NamespaceBIP122, c.Namespace)
	assert.Equal(t, "bip122:000000000019d6689c085ae165831e93", c.String())

	// Invalid: too short
	_, err = NewBIP122ChainID("000000000019d668")
	assert.Error(t, err)

	// Invalid: uppercase
	_, err = NewBIP122ChainID("000000000019D6689C085AE165831E93")
	assert.Error(t, err)

	// Invalid: non-hex characters
	_, err = NewBIP122ChainID("000000000019d6689c085ae165831xyz")
	assert.Error(t, err)
}

func TestMustNewChainIDByBIP122(t *testing.T) {
	// Valid
	c := MustNewBIP122ChainID("000000000019d6689c085ae165831e93")
	assert.Equal(t, "bip122:000000000019d6689c085ae165831e93", c.String())

	// Invalid should panic
	assert.Panics(t, func() {
		MustNewBIP122ChainID("invalid")
	})
}

func TestNewChainIDFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNs  Namespace
		wantRef string
		wantErr bool
	}{
		{"eip155 mainnet", "eip155:1", NamespaceEIP155, "1", false},
		{"eip155 polygon", "eip155:137", NamespaceEIP155, "137", false},
		{"solana mainnet", "solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp", NamespaceSolana, "5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp", false},
		{"solana devnet", "solana:EtWTRABZaYq6iMfeYKouRu166VU2xqa1", NamespaceSolana, "EtWTRABZaYq6iMfeYKouRu166VU2xqa1", false},
		{"bip122 bitcoin", "bip122:000000000019d6689c085ae165831e93", NamespaceBIP122, "000000000019d6689c085ae165831e93", false},
		{"invalid format", "invalid", "", "", true},
		{"invalid eip155 ref", "eip155:abc", "", "", true},
		{"invalid solana ref short", "solana:invalid", "", "", true},
		{"invalid solana ref long", "solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdpKnmDNgT76nhR", "", "", true},
		{"invalid bip122 ref", "bip122:invalid", "", "", true},
		{"unknown namespace", "unknown:1", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ParseChainID(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantNs, c.Namespace)
			assert.Equal(t, tt.wantRef, c.Reference)
		})
	}
}

func TestMustNewChainIDFromString(t *testing.T) {
	// Valid
	c := MustParseChainID("eip155:1")
	assert.Equal(t, "eip155:1", c.String())

	// Invalid should panic
	assert.Panics(t, func() {
		MustParseChainID("invalid")
	})
}

// --- IsZero, Equal, Validate Tests ---

func TestChainID_IsZero(t *testing.T) {
	assert.True(t, ChainID{}.IsZero())
	assert.True(t, ChainID{Namespace: "", Reference: ""}.IsZero())
	assert.False(t, ChainID{Namespace: NamespaceEIP155, Reference: "1"}.IsZero())
	assert.False(t, ChainID{Namespace: NamespaceEIP155, Reference: ""}.IsZero())
	assert.False(t, ChainID{Namespace: "", Reference: "1"}.IsZero())
}

func TestChainID_Equal(t *testing.T) {
	c1 := NewEIP155ChainID(1)
	c2 := NewEIP155ChainID(1)
	c3 := NewEIP155ChainID(137)

	assert.True(t, c1.Equal(c2))
	assert.False(t, c1.Equal(c3))
	assert.True(t, ChainID{}.Equal(ChainID{}))
}

func TestChainID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		chainID ChainID
		wantErr bool
	}{
		{"valid eip155", NewEIP155ChainID(1), false},
		{"valid solana mainnet", NewSolanaChainID(SolanaMainnet), false},
		{"valid solana devnet", NewSolanaChainID(SolanaDevnet), false},
		{"valid bip122", MustNewBIP122ChainID("000000000019d6689c085ae165831e93"), false},
		{"zero value", ChainID{}, true},
		{"invalid eip155 ref", ChainID{Namespace: NamespaceEIP155, Reference: "abc"}, true},
		{"invalid solana ref short", ChainID{Namespace: NamespaceSolana, Reference: "invalid"}, true},
		{"invalid solana ref long", ChainID{Namespace: NamespaceSolana, Reference: "5eykt4UsFv8P8NJdTREpY1vzqKqZKvdpKnmDNgT76nhR"}, true},
		{"invalid bip122 ref", ChainID{Namespace: NamespaceBIP122, Reference: "invalid"}, true},
		{"unknown namespace", ChainID{Namespace: "unknown", Reference: "1"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.chainID.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// --- String Tests ---

func TestChainID_String(t *testing.T) {
	assert.Equal(t, "", ChainID{}.String())
	assert.Equal(t, "eip155:1", NewEIP155ChainID(1).String())
	assert.Equal(t, "bip122:000000000019d6689c085ae165831e93", MustNewBIP122ChainID("000000000019d6689c085ae165831e93").String())
}

// --- ToAccountID Tests ---

func TestChainID_ToAccountID(t *testing.T) {
	c := NewEIP155ChainID(1)
	acc, err := c.ToAccountID("0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")
	require.NoError(t, err)
	assert.Equal(t, "eip155:1:0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb", acc.String())

	// Test with Solana
	sc := NewSolanaChainID(SolanaMainnet)
	sacc, err := sc.ToAccountID("7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")
	require.NoError(t, err)
	assert.Equal(t, "solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp:7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv", sacc.String())
}

func TestChainID_MustToAccountID(t *testing.T) {
	c := NewEIP155ChainID(1)

	// Valid
	acc := c.MustToAccountID("0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb")
	assert.Equal(t, "eip155:1:0xab16a96D359eC26a11e2C2b3d8f8B8942d5Bfcdb", acc.String())
}

// --- Serialization Tests ---

func TestChainID_TextMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		chainID ChainID
		want    string
	}{
		{
			name:    "ethereum mainnet",
			chainID: ChainID{Namespace: NamespaceEIP155, Reference: "1"},
			want:    "eip155:1",
		},
		{
			name:    "ethereum sepolia",
			chainID: ChainID{Namespace: NamespaceEIP155, Reference: "11155111"},
			want:    "eip155:11155111",
		},
		{
			name:    "solana mainnet",
			chainID: ChainID{Namespace: NamespaceSolana, Reference: "5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp"},
			want:    "solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp",
		},
		{
			name:    "bip122 bitcoin",
			chainID: ChainID{Namespace: NamespaceBIP122, Reference: "000000000019d6689c085ae165831e93"},
			want:    "bip122:000000000019d6689c085ae165831e93",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test MarshalText
			data, err := tt.chainID.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(data))

			// Test UnmarshalText
			var got ChainID
			err = got.UnmarshalText(data)
			require.NoError(t, err)
			assert.Equal(t, tt.chainID, got)
		})
	}
}

func TestChainID_TextMarshalZeroValue(t *testing.T) {
	var c ChainID
	data, err := c.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "", string(data))
}

func TestChainID_TextUnmarshalEmpty(t *testing.T) {
	var c ChainID
	err := c.UnmarshalText([]byte(""))
	require.NoError(t, err)
	assert.True(t, c.IsZero())
}

func TestChainID_BinaryMarshalUnmarshal(t *testing.T) {
	original := ChainID{Namespace: NamespaceEIP155, Reference: "1"}

	// Test MarshalBinary
	data, err := original.MarshalBinary()
	require.NoError(t, err)
	assert.Equal(t, "eip155:1", string(data))

	// Test UnmarshalBinary
	var got ChainID
	err = got.UnmarshalBinary(data)
	require.NoError(t, err)
	assert.Equal(t, original, got)
}

func TestChainID_JSONMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		chainID ChainID
		want    string
	}{
		{
			name:    "ethereum",
			chainID: ChainID{Namespace: NamespaceEIP155, Reference: "1"},
			want:    `"eip155:1"`,
		},
		{
			name:    "solana",
			chainID: ChainID{Namespace: NamespaceSolana, Reference: "5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp"},
			want:    `"solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp"`,
		},
		{
			name:    "bip122",
			chainID: ChainID{Namespace: NamespaceBIP122, Reference: "000000000019d6689c085ae165831e93"},
			want:    `"bip122:000000000019d6689c085ae165831e93"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test MarshalJSON
			data, err := json.Marshal(tt.chainID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(data))

			// Test UnmarshalJSON
			var got ChainID
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)
			assert.Equal(t, tt.chainID, got)
		})
	}
}

func TestChainID_JSONMarshalZeroValue(t *testing.T) {
	var c ChainID
	data, err := json.Marshal(c)
	require.NoError(t, err)
	assert.Equal(t, `""`, string(data))
}

func TestChainID_JSONUnmarshalNull(t *testing.T) {
	var c ChainID
	err := json.Unmarshal([]byte("null"), &c)
	require.NoError(t, err)
	assert.True(t, c.IsZero())
}

func TestChainID_JSONUnmarshalEmptyString(t *testing.T) {
	var c ChainID
	err := json.Unmarshal([]byte(`""`), &c)
	require.NoError(t, err)
	assert.True(t, c.IsZero())
}

func TestChainID_JSONInStruct(t *testing.T) {
	type Wrapper struct {
		Chain ChainID `json:"chain"`
		Name  string  `json:"name"`
	}

	original := Wrapper{
		Chain: ChainID{Namespace: NamespaceEIP155, Reference: "1"},
		Name:  "test",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Equal(t, `{"chain":"eip155:1","name":"test"}`, string(data))

	var got Wrapper
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)
	assert.Equal(t, original, got)
}

func TestChainID_JSONInStructZeroValue(t *testing.T) {
	type Wrapper struct {
		Chain ChainID `json:"chain"`
		Name  string  `json:"name"`
	}

	original := Wrapper{
		Chain: ChainID{},
		Name:  "test",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Equal(t, `{"chain":"","name":"test"}`, string(data))

	var got Wrapper
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)
	assert.Equal(t, original, got)
}

// --- Database Tests ---

func TestChainID_DatabaseValueScan(t *testing.T) {
	original := ChainID{Namespace: NamespaceEIP155, Reference: "1"}

	// Test Value
	val, err := original.Value()
	require.NoError(t, err)
	assert.Equal(t, "eip155:1", val)

	// Test Scan from string
	var got1 ChainID
	err = got1.Scan("eip155:1")
	require.NoError(t, err)
	assert.Equal(t, original, got1)

	// Test Scan from []byte
	var got2 ChainID
	err = got2.Scan([]byte("eip155:1"))
	require.NoError(t, err)
	assert.Equal(t, original, got2)

	// Test Scan from nil
	var got3 ChainID
	err = got3.Scan(nil)
	require.NoError(t, err)
	assert.Equal(t, ChainID{}, got3)
}

func TestChainID_DatabaseValueZero(t *testing.T) {
	var c ChainID
	val, err := c.Value()
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestChainID_DatabaseScanEmpty(t *testing.T) {
	var c ChainID
	err := c.Scan("")
	require.NoError(t, err)
	assert.True(t, c.IsZero())

	err = c.Scan([]byte{})
	require.NoError(t, err)
	assert.True(t, c.IsZero())
}

func TestChainID_ScanInvalidType(t *testing.T) {
	var c ChainID
	err := c.Scan(123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot scan int into ChainID")
}

func TestChainID_UnmarshalJSONInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"not a string", `123`},
		{"invalid format", `"invalid"`},
		{"missing reference", `"eip155:"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c ChainID
			err := json.Unmarshal([]byte(tt.input), &c)
			assert.Error(t, err)
		})
	}
}

func TestChainID_RoundTrip(t *testing.T) {
	// Test complete round-trip: string -> ChainID -> all formats -> ChainID
	original := MustParseChainID("eip155:1")

	// Text round-trip
	textData, _ := original.MarshalText()
	var fromText ChainID
	require.NoError(t, fromText.UnmarshalText(textData))
	assert.Equal(t, original, fromText)

	// Binary round-trip
	binaryData, _ := original.MarshalBinary()
	var fromBinary ChainID
	require.NoError(t, fromBinary.UnmarshalBinary(binaryData))
	assert.Equal(t, original, fromBinary)

	// JSON round-trip
	jsonData, _ := json.Marshal(original)
	var fromJSON ChainID
	require.NoError(t, json.Unmarshal(jsonData, &fromJSON))
	assert.Equal(t, original, fromJSON)

	// Database round-trip
	dbVal, _ := original.Value()
	var fromDB ChainID
	require.NoError(t, fromDB.Scan(dbVal))
	assert.Equal(t, original, fromDB)
}

func TestChainID_RoundTripBIP122(t *testing.T) {
	original := MustParseChainID("bip122:000000000019d6689c085ae165831e93")

	// Text round-trip
	textData, _ := original.MarshalText()
	var fromText ChainID
	require.NoError(t, fromText.UnmarshalText(textData))
	assert.Equal(t, original, fromText)

	// JSON round-trip
	jsonData, _ := json.Marshal(original)
	var fromJSON ChainID
	require.NoError(t, json.Unmarshal(jsonData, &fromJSON))
	assert.Equal(t, original, fromJSON)
}
