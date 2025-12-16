package caip10

import (
	"fmt"
	"regexp"
)

const NamespaceBIP122 Namespace = "bip122"

// BIP122Network represents a BIP122 network (chain reference).
// The reference is the first 32 characters of the genesis block hash.
// https://github.com/ChainAgnostic/namespaces/blob/main/bip122/caip10.md
type BIP122Network string

// Common BIP122 networks (genesis block hash prefix)
const (
	// Bitcoin
	BitcoinMainnet BIP122Network = "000000000019d6689c085ae165831e93" // Bitcoin mainnet
	BitcoinTestnet BIP122Network = "000000000933ea01ad0ee984209779ba" // Bitcoin testnet

	// Bitcoin Cash (forked from Bitcoin at block 478559)
	BitcoinCashMainnet BIP122Network = "000000000000000000651ef99cb9fcbe" // Bitcoin Cash mainnet

	// Litecoin
	LitecoinMainnet BIP122Network = "12a765e31ffd4059bada1e25190f6e98" // Litecoin mainnet
	LitecoinTestnet BIP122Network = "4966625a4b2851d9fdee139e56211a0d" // Litecoin testnet

	// Dogecoin
	DogecoinMainnet BIP122Network = "1a91e3dace36e2be3bf030a65679fe82" // Dogecoin mainnet
	DogecoinTestnet BIP122Network = "bb0a78264637406b6360aad926284d54" // Dogecoin testnet

	// Dash
	DashMainnet BIP122Network = "00000ffd590b1485b3caadc19b22e637" // Dash mainnet
)

// String returns the network reference string.
// If the reference is longer than 32 characters, it will be truncated.
func (n BIP122Network) String() string {
	s := string(n)
	if len(s) > 32 {
		return s[:32]
	}
	return s
}

// BIP122 address validation regexes
var (
	// Bitcoin mainnet addresses:
	// - P2SH: starts with "3", base58btc encoded
	// - P2WPKH (SegWit): starts with "bc1q", bech32 encoded
	// - P2TR (Taproot): starts with "bc1p", bech32m encoded
	bitcoinMainnetAddressRegex = regexp.MustCompile(`^(bc1[qpzry9x8gf2tvdw0s3jn54khce6mua7l]{39,59}|3[a-km-zA-HJ-NP-Z1-9]{25,34})$`)

	// Bitcoin testnet addresses:
	// - P2SH: starts with "2", base58btc encoded
	// - P2WPKH/P2TR: starts with "tb1", bech32/bech32m encoded
	bitcoinTestnetAddressRegex = regexp.MustCompile(`^(tb1[qpzry9x8gf2tvdw0s3jn54khce6mua7l]{39,59}|2[a-km-zA-HJ-NP-Z1-9]{25,34})$`)

	// Bitcoin Cash mainnet addresses:
	// - CashAddr: starts with "q" or "p" (without prefix), or "bitcoincash:q/p"
	// - Legacy: starts with "1" or "3", base58btc encoded (same as Bitcoin)
	bitcoinCashMainnetAddressRegex = regexp.MustCompile(`^(bitcoincash:)?[qp][qpzry9x8gf2tvdw0s3jn54khce6mua7l]{41}$|^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$`)

	// Litecoin mainnet addresses:
	// - P2SH: starts with "M" or "3", base58btc encoded
	// - P2WPKH: starts with "ltc1", bech32 encoded
	litecoinMainnetAddressRegex = regexp.MustCompile(`^(ltc1[qpzry9x8gf2tvdw0s3jn54khce6mua7l]{39,59}|[M3][a-km-zA-HJ-NP-Z1-9]{25,34})$`)

	// Litecoin testnet addresses:
	// - P2WPKH: starts with "tltc1", bech32 encoded
	litecoinTestnetAddressRegex = regexp.MustCompile(`^(tltc1[qpzry9x8gf2tvdw0s3jn54khce6mua7l]{39,59}|[mn2][a-km-zA-HJ-NP-Z1-9]{25,34})$`)

	// Dogecoin mainnet addresses:
	// - P2PKH: starts with "D", base58 encoded
	// - P2SH: starts with "9" or "A", base58 encoded
	dogecoinMainnetAddressRegex = regexp.MustCompile(`^[D9A][a-km-zA-HJ-NP-Z1-9]{25,34}$`)

	// Dogecoin testnet addresses:
	// - P2PKH: starts with "n", base58 encoded
	dogecoinTestnetAddressRegex = regexp.MustCompile(`^[nm][a-km-zA-HJ-NP-Z1-9]{25,34}$`)

	// Dash mainnet addresses:
	// - P2PKH: starts with "X", base58 encoded
	// - P2SH: starts with "7", base58 encoded
	dashMainnetAddressRegex = regexp.MustCompile(`^[X7][a-km-zA-HJ-NP-Z1-9]{25,34}$`)

	// Generic BIP122 address regex (loose validation)
	// Covers base58btc addresses and bech32/bech32m addresses
	genericBIP122AddressRegex = regexp.MustCompile(`^([a-km-zA-HJ-NP-Z1-9]{25,35}|[a-z]{1,12}:?[qpzry9x8gf2tvdw0s3jn54khce6mua7l]{39,64})$`)
)

// ValidateBIP122Address validates a BIP122 address string for a specific network.
// Returns nil if valid, error otherwise.
func ValidateBIP122Address(network BIP122Network, address string) error {
	if len(address) == 0 {
		return fmt.Errorf("%w: empty address", ErrInvalidAddress)
	}

	var regex *regexp.Regexp
	switch network {
	case BitcoinMainnet:
		regex = bitcoinMainnetAddressRegex
	case BitcoinTestnet:
		regex = bitcoinTestnetAddressRegex
	case BitcoinCashMainnet:
		regex = bitcoinCashMainnetAddressRegex
	case LitecoinMainnet:
		regex = litecoinMainnetAddressRegex
	case LitecoinTestnet:
		regex = litecoinTestnetAddressRegex
	case DogecoinMainnet:
		regex = dogecoinMainnetAddressRegex
	case DogecoinTestnet:
		regex = dogecoinTestnetAddressRegex
	case DashMainnet:
		regex = dashMainnetAddressRegex
	default:
		// Use generic validation for unknown networks
		regex = genericBIP122AddressRegex
	}

	if !regex.MatchString(address) {
		return fmt.Errorf("%w: invalid address format for network %s", ErrInvalidAddress, network)
	}

	return nil
}

// ValidateBIP122AddressLoose validates a BIP122 address with loose validation.
// This accepts any address that matches the generic BIP122 address pattern.
func ValidateBIP122AddressLoose(address string) error {
	if len(address) == 0 {
		return fmt.Errorf("%w: empty address", ErrInvalidAddress)
	}

	if !genericBIP122AddressRegex.MatchString(address) {
		return fmt.Errorf("%w: invalid BIP122 address format", ErrInvalidAddress)
	}

	return nil
}

// BIP122AccountID is the interface for BIP122 account IDs.
// https://github.com/ChainAgnostic/namespaces/blob/main/bip122/caip10.md
type BIP122AccountID interface {
	AccountID
	// Network returns the BIP122 network.
	Network() BIP122Network
	// SetAddress returns a new BIP122AccountID with the specified address.
	SetAddress(address string) BIP122AccountID
}

// Ensure bip122AccountID implements BIP122AccountID at compile time
var _ BIP122AccountID = (*bip122AccountID)(nil)

func init() {
	RegisterParser(&bip122Parser{})
}

// bip122AccountID represents a BIP122 account ID per CAIP-10.
type bip122AccountID struct {
	*GenericAccountID // embedded, inherits all serialization methods
	network           BIP122Network
}

// NewBIP122 creates a new BIP122AccountID.
func NewBIP122(network BIP122Network, address string) BIP122AccountID {
	base := newGenericUnchecked(NamespaceBIP122, network.String(), address)
	return &bip122AccountID{
		GenericAccountID: base,
		network:          network,
	}
}

// NewBIP122WithValidation creates a new BIP122AccountID with address validation.
func NewBIP122WithValidation(network BIP122Network, address string) (BIP122AccountID, error) {
	if err := ValidateBIP122Address(network, address); err != nil {
		return nil, err
	}
	return NewBIP122(network, address), nil
}

// NewBitcoinMainnet creates a BIP122AccountID for Bitcoin mainnet.
func NewBitcoinMainnet(address string) BIP122AccountID {
	return NewBIP122(BitcoinMainnet, address)
}

// NewBitcoinTestnet creates a BIP122AccountID for Bitcoin testnet.
func NewBitcoinTestnet(address string) BIP122AccountID {
	return NewBIP122(BitcoinTestnet, address)
}

// NewBitcoinCashMainnet creates a BIP122AccountID for Bitcoin Cash mainnet.
func NewBitcoinCashMainnet(address string) BIP122AccountID {
	return NewBIP122(BitcoinCashMainnet, address)
}

// NewLitecoinMainnet creates a BIP122AccountID for Litecoin mainnet.
func NewLitecoinMainnet(address string) BIP122AccountID {
	return NewBIP122(LitecoinMainnet, address)
}

// NewDogecoinMainnet creates a BIP122AccountID for Dogecoin mainnet.
func NewDogecoinMainnet(address string) BIP122AccountID {
	return NewBIP122(DogecoinMainnet, address)
}

// NewDashMainnet creates a BIP122AccountID for Dash mainnet.
func NewDashMainnet(address string) BIP122AccountID {
	return NewBIP122(DashMainnet, address)
}

// Network returns the BIP122 network.
func (a *bip122AccountID) Network() BIP122Network {
	if a == nil {
		return ""
	}
	return a.network
}

// SetAddress returns a new BIP122AccountID with the specified address.
func (a *bip122AccountID) SetAddress(address string) BIP122AccountID {
	if a == nil {
		return nil
	}
	return NewBIP122(a.network, address)
}

// IsZero reports whether the AccountID is the zero value.
func (a *bip122AccountID) IsZero() bool {
	return a == nil || a.GenericAccountID == nil || a.GenericAccountID.IsZero()
}

// Equal reports whether two AccountIDs are equal.
func (a *bip122AccountID) Equal(other AccountID) bool {
	if a.IsZero() && (other == nil || other.IsZero()) {
		return true
	}
	if a.IsZero() || other == nil || other.IsZero() {
		return false
	}
	return a.GenericAccountID.Equal(other)
}

// --- bip122Parser ---

type bip122Parser struct{}

func (p *bip122Parser) Namespace() Namespace {
	return NamespaceBIP122
}

func (p *bip122Parser) Parse(s string) (AccountID, error) {
	ns, ref, addr, err := SplitCAIP10(s)
	if err != nil {
		return nil, err
	}
	if ns != NamespaceBIP122 {
		return nil, fmt.Errorf("%w: expected %q, got %q", ErrInvalidNamespace, NamespaceBIP122, ns)
	}
	return NewBIP122(BIP122Network(ref), addr), nil
}

func (p *bip122Parser) ParseAddress(reference, address string) (AccountID, error) {
	return NewBIP122(BIP122Network(reference), address), nil
}
