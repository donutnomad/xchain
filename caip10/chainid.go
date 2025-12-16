package caip10

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Namespace string

// Ethereum
var (
	ChainIDEthereumMainnet = NewChainIDByEIP155(1)
	ChainIDEthereumSepolia = NewChainIDByEIP155(11155111)
	ChainIDEthereumHoodi   = NewChainIDByEIP155(560048)
)

// Arbitrum
var (
	ChainIDArbitrumOne     = NewChainIDByEIP155(42161)
	ChainIDArbitrumNova    = NewChainIDByEIP155(42170)
	ChainIDArbitrumSepolia = NewChainIDByEIP155(421614)
)

// Optimism
var (
	ChainIDOptimism        = NewChainIDByEIP155(10)
	ChainIDOptimismSepolia = NewChainIDByEIP155(11155420)
)

// Base
var (
	ChainIDBase        = NewChainIDByEIP155(8453)
	ChainIDBaseSepolia = NewChainIDByEIP155(84532)
)

// Polygon
var (
	ChainIDPolygon      = NewChainIDByEIP155(137)
	ChainIDPolygonAmoy  = NewChainIDByEIP155(80002)
	ChainIDPolygonZkEVM = NewChainIDByEIP155(1101)
)

// zkSync Era
var (
	ChainIDZkSyncEra        = NewChainIDByEIP155(324)
	ChainIDZkSyncEraSepolia = NewChainIDByEIP155(300)
)

// Linea
var (
	ChainIDLinea        = NewChainIDByEIP155(59144)
	ChainIDLineaSepolia = NewChainIDByEIP155(59141)
)

// Scroll
var (
	ChainIDScroll        = NewChainIDByEIP155(534352)
	ChainIDScrollSepolia = NewChainIDByEIP155(534351)
)

// BNB Smart Chain
var (
	ChainIDBSC        = NewChainIDByEIP155(56)
	ChainIDBSCTestnet = NewChainIDByEIP155(97)
)

// opBNB
var (
	ChainIDOpBNB        = NewChainIDByEIP155(204)
	ChainIDOpBNBTestnet = NewChainIDByEIP155(5611)
)

// Avalanche
var (
	ChainIDAvalanche     = NewChainIDByEIP155(43114)
	ChainIDAvalancheFuji = NewChainIDByEIP155(43113)
)

// Fantom
var (
	ChainIDFantom = NewChainIDByEIP155(250)
)

// Gnosis
var (
	ChainIDGnosis = NewChainIDByEIP155(100)
)

// Celo
var (
	ChainIDCelo = NewChainIDByEIP155(42220)
)

// Solana
var (
	ChainIDSolanaMainnet = NewChainIDBySolana(SolanaMainnet)
	ChainIDSolanaDevnet  = NewChainIDBySolana(SolanaDevnet)
	ChainIDSolanaTestnet = NewChainIDBySolana(SolanaTestnet)
)

// Bitcoin
var (
	ChainIDBitcoinMainnet = MustNewChainIDByBIP122(BitcoinMainnet)
	ChainIDBitcoinTestnet = MustNewChainIDByBIP122(BitcoinTestnet)
)

// bip122ReferenceRegex validates BIP122 chain reference.
// The reference is the first 32 characters of the genesis block hash (hex encoded).
var bip122ReferenceRegex = regexp.MustCompile(`^[a-f0-9]{32}$`)

// solanaReferenceRegex validates Solana chain reference.
// The reference is the first 32 characters of the genesis hash (base58 encoded).
var solanaReferenceRegex = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{32}$`)

type ChainID struct {
	Namespace Namespace `json:"namespace"`
	Reference string    `json:"reference"`
}

// validateReference validates the reference for a given namespace.
func validateReference(ns Namespace, reference string) error {
	switch ns {
	case NamespaceEIP155:
		if _, err := strconv.ParseUint(reference, 10, 64); err != nil {
			return fmt.Errorf("%w: invalid EIP155 chain id %q", ErrInvalidReference, reference)
		}
	case NamespaceSolana:
		if !solanaReferenceRegex.MatchString(reference) {
			return fmt.Errorf("%w: invalid Solana reference, must be 32 base58 characters, got %q", ErrInvalidReference, reference)
		}
	case NamespaceBIP122:
		if !bip122ReferenceRegex.MatchString(reference) {
			return fmt.Errorf("%w: invalid BIP122 block hash, must be 32 lowercase hex characters, got %q", ErrInvalidReference, reference)
		}
	default:
		return fmt.Errorf("%w: unknown namespace %q", ErrInvalidNamespace, ns)
	}
	return nil
}

func NewChainIDByEIP155(chainID uint64) ChainID {
	return ChainID{Namespace: NamespaceEIP155, Reference: strconv.FormatUint(chainID, 10)}
}

func NewChainIDBySolana(network SolanaNetwork) ChainID {
	return ChainID{Namespace: NamespaceSolana, Reference: network.String()}
}

// NewChainIDByBIP122 creates a ChainID for BIP122 namespace.
// blockHash should be the first 32 characters of the genesis block hash (hex encoded).
func NewChainIDByBIP122(blockHash BIP122Network) (ChainID, error) {
	if err := validateReference(NamespaceBIP122, string(blockHash)); err != nil {
		return ChainID{}, err
	}
	return ChainID{Namespace: NamespaceBIP122, Reference: string(blockHash)}, nil
}

// MustNewChainIDByBIP122 creates a ChainID for BIP122 namespace and panics if invalid.
func MustNewChainIDByBIP122(blockHash BIP122Network) ChainID {
	c, err := NewChainIDByBIP122(blockHash)
	if err != nil {
		panic(err)
	}
	return c
}

func MustNewChainIDFromString(s string) ChainID {
	c, err := NewChainIDFromString(s)
	if err != nil {
		panic(err)
	}
	return c
}

func NewChainIDFromString(s string) (ChainID, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return ChainID{}, fmt.Errorf("%w: invalid format %q", ErrInvalidFormat, s)
	}
	ns := Namespace(parts[0])
	reference := parts[1]
	if err := validateReference(ns, reference); err != nil {
		return ChainID{}, err
	}
	return ChainID{Namespace: ns, Reference: reference}, nil
}

// IsZero reports whether the ChainID is the zero value.
func (c ChainID) IsZero() bool {
	return c.Namespace == "" && c.Reference == ""
}

// Equal reports whether two ChainIDs are equal.
func (c ChainID) Equal(other ChainID) bool {
	return c.Namespace == other.Namespace && c.Reference == other.Reference
}

// Validate checks if the ChainID is valid.
func (c ChainID) Validate() error {
	if c.IsZero() {
		return ErrEmptyValue
	}
	return validateReference(c.Namespace, c.Reference)
}

func (c ChainID) String() string {
	if c.IsZero() {
		return ""
	}
	return string(c.Namespace) + ":" + c.Reference
}

// ToAccountID creates an AccountID from this ChainID with the given address.
func (c ChainID) ToAccountID(address string) (AccountID, error) {
	return ParseWithNamespace(c.Namespace, c.Reference, address)
}

// MustToAccountID creates an AccountID from this ChainID and panics if invalid.
func (c ChainID) MustToAccountID(address string) AccountID {
	a, err := c.ToAccountID(address)
	if err != nil {
		panic(err)
	}
	return a
}

// MarshalText implements encoding.TextMarshaler.
func (c ChainID) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (c *ChainID) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*c = ChainID{}
		return nil
	}
	parsed, err := NewChainIDFromString(string(text))
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (c ChainID) MarshalBinary() ([]byte, error) {
	return []byte(c.String()), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (c *ChainID) UnmarshalBinary(data []byte) error {
	return c.UnmarshalText(data)
}

// MarshalJSON implements json.Marshaler.
func (c ChainID) MarshalJSON() ([]byte, error) {
	if c.IsZero() {
		return []byte(`""`), nil
	}
	return json.Marshal(c.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ChainID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*c = ChainID{}
		return nil
	}
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid JSON string for ChainID")
	}
	s := string(data[1 : len(data)-1])
	if s == "" {
		*c = ChainID{}
		return nil
	}
	return c.UnmarshalText([]byte(s))
}

// Value implements driver.Valuer.
func (c ChainID) Value() (driver.Value, error) {
	if c.IsZero() {
		return nil, nil
	}
	return c.String(), nil
}

// Scan implements sql.Scanner.
func (c *ChainID) Scan(src any) error {
	switch v := src.(type) {
	case string:
		if v == "" {
			*c = ChainID{}
			return nil
		}
		return c.UnmarshalText([]byte(v))
	case []byte:
		if len(v) == 0 {
			*c = ChainID{}
			return nil
		}
		return c.UnmarshalText(v)
	case nil:
		*c = ChainID{}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into ChainID", src)
	}
}
