package caip10

import (
	"fmt"
	"regexp"

	"filippo.io/edwards25519"
	"github.com/donutnomad/solana-web3/web3"
)

const NamespaceSolana = "solana"

// SolanaNetwork represents a Solana network (chain reference).
// https://github.com/ChainAgnostic/namespaces/blob/main/solana/caip10.md
type SolanaNetwork string

// Common Solana networks
const (
	SolanaMainnet SolanaNetwork = "5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp" // mainnet-beta genesis hash
	SolanaDevnet  SolanaNetwork = "EtWTRABZaYq6iMfeYKouRu166VU2xqa1" // devnet genesis hash
	SolanaTestnet SolanaNetwork = "4uhcVJyU9pJkvQyS88uRDiswHXSCkY3z" // testnet genesis hash
)

// String returns the network reference string.
func (n SolanaNetwork) String() string {
	s := string(n)
	if len(s) > 32 {
		return s[:32]
	}
	return s
}

// SolanaAddressLength is the expected length of a decoded Solana public key.
const SolanaAddressLength = 32

// solanaAddressRegex validates the base58 format of Solana addresses.
// Solana addresses are 32-byte arrays encoded with bitcoin base58 alphabet.
// https://solana.com/developers/guides/advanced/exchange#validating-user-supplied-account-addresses-for-withdrawals
var solanaAddressRegex = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{32,44}$`)

// ValidateSolanaAddress validates a Solana address string.
// Returns nil if valid, error otherwise.
//
// Validation steps:
//  1. Format check: must match [1-9A-HJ-NP-Za-km-z]{32,44}
//  2. Base58 decode succeeds
//  3. Decoded length is exactly 32 bytes
//  4. Valid ed25519 public key (point on curve)
//
// Note: Solana addresses are not checksummed, so some typos may still pass validation.
func ValidateSolanaAddress(base58Address string) error {
	if !solanaAddressRegex.MatchString(base58Address) {
		return fmt.Errorf("%w: invalid base58 format", ErrInvalidAddress)
	}

	addr, err := web3.NewPublicKey(base58Address)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidAddress, err)
	}

	addrBytes := addr.Bytes()
	if len(addrBytes) != SolanaAddressLength {
		return fmt.Errorf("%w: decoded address must be %d bytes, got %d",
			ErrInvalidAddress, SolanaAddressLength, len(addrBytes))
	}

	// Validate ed25519 public key (check if point is on curve)
	if _, err := new(edwards25519.Point).SetBytes(addrBytes); err != nil {
		return fmt.Errorf("%w: not a valid ed25519 public key", ErrInvalidAddress)
	}

	return nil
}

// ValidateSolanaAddressLoose validates a Solana address without ed25519 curve check.
// Use this for PDAs (Program Derived Addresses) which are off-curve by design.
func ValidateSolanaAddressLoose(base58Address string) error {
	if !solanaAddressRegex.MatchString(base58Address) {
		return fmt.Errorf("%w: invalid base58 format", ErrInvalidAddress)
	}

	addr, err := web3.NewPublicKey(base58Address)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidAddress, err)
	}

	if len(addr.Bytes()) != SolanaAddressLength {
		return fmt.Errorf("%w: decoded address must be %d bytes, got %d",
			ErrInvalidAddress, SolanaAddressLength, len(addr.Bytes()))
	}

	return nil
}

// IsOnCurve checks if a Solana public key is on the ed25519 curve.
// Normal account addresses are on curve, while PDAs are off curve.
func IsOnCurve(pubkey web3.PublicKey) bool {
	_, err := new(edwards25519.Point).SetBytes(pubkey.Bytes())
	return err == nil
}

// SolanaAccountID is the interface for Solana account IDs.
// https://solana.com/developers/guides/advanced/exchange#validating-user-supplied-account-addresses-for-withdrawals
type SolanaAccountID interface {
	AccountID
	// Account returns the native web3.PublicKey.
	Account() web3.PublicKey
	// SetAccount returns a new SolanaAccountID with the specified account.
	SetAccount(account web3.PublicKey) SolanaAccountID
	// IsOnCurve returns true if the address is on the ed25519 curve.
	// Normal accounts are on curve, PDAs (Program Derived Addresses) are off curve.
	IsOnCurve() bool
	// IsMainnet returns true if this is a mainnet account.
	IsMainnet() bool
	// IsDevnet returns true if this is a devnet account.
	IsDevnet() bool
	// IsTestnet returns true if this is a testnet account.
	IsTestnet() bool
}

// Ensure solanaAccountID implements SolanaAccountID at compile time
var _ SolanaAccountID = (*solanaAccountID)(nil)

func init() {
	RegisterParser(&solanaParser{})
}

// solanaAccountID represents a Solana account ID per CAIP-10.
type solanaAccountID struct {
	*GenericAccountID                // embedded, inherits all serialization methods
	pubkey            web3.PublicKey // native Solana public key
}

// NewSolana creates a new SolanaAccountID.
func NewSolana(network SolanaNetwork, address web3.PublicKey) SolanaAccountID {
	addrStr := address.String()
	base := newGenericUnchecked(NamespaceSolana, network.String(), addrStr)
	return &solanaAccountID{
		GenericAccountID: base,
		pubkey:           address,
	}
}

// NewSolanaFromBase58 creates a new SolanaAccountID from a base58 address string.
// Validation:
//  1. Format check: must match [1-9A-HJ-NP-Za-km-z]{32,44}
//  2. Base58 decode and verify 32-byte length
//
// Note: Solana addresses are not checksummed, so typos cannot be fully detected.
func NewSolanaFromBase58(network SolanaNetwork, base58Address string) (SolanaAccountID, error) {
	// Step 1: Basic format validation
	if !solanaAddressRegex.MatchString(base58Address) {
		return nil, fmt.Errorf("%w: invalid base58 format", ErrInvalidAddress)
	}

	// Step 2: Decode and verify length (web3.NewPublicKey handles this)
	addr, err := web3.NewPublicKey(base58Address)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAddress, err)
	}

	// Step 3: Verify decoded length is exactly 32 bytes
	if len(addr.Bytes()) != SolanaAddressLength {
		return nil, fmt.Errorf("%w: decoded address must be %d bytes, got %d",
			ErrInvalidAddress, SolanaAddressLength, len(addr.Bytes()))
	}

	return NewSolana(network, addr), nil
}

// MustNewSolanaFromBase58 creates a new SolanaAccountID from base58 and panics if invalid.
func MustNewSolanaFromBase58(network SolanaNetwork, base58Address string) SolanaAccountID {
	a, err := NewSolanaFromBase58(network, base58Address)
	if err != nil {
		panic(err)
	}
	return a
}

// NewSolanaMainnet creates a SolanaAccountID for mainnet.
func NewSolanaMainnet(address web3.PublicKey) SolanaAccountID {
	return NewSolana(SolanaMainnet, address)
}

// NewSolanaDevnet creates a SolanaAccountID for devnet.
func NewSolanaDevnet(address web3.PublicKey) SolanaAccountID {
	return NewSolana(SolanaDevnet, address)
}

// NewSolanaTestnet creates a SolanaAccountID for testnet.
func NewSolanaTestnet(address web3.PublicKey) SolanaAccountID {
	return NewSolana(SolanaTestnet, address)
}

// Account returns the native web3.PublicKey.
func (a *solanaAccountID) Account() web3.PublicKey {
	if a == nil {
		return web3.PublicKey{}
	}
	return a.pubkey
}

// SetAccount returns a new SolanaAccountID with the specified account.
func (a *solanaAccountID) SetAccount(account web3.PublicKey) SolanaAccountID {
	if a == nil {
		return nil
	}
	return NewSolana(SolanaNetwork(a.Reference()), account)
}

// IsOnCurve returns true if the address is on the ed25519 curve.
// Normal accounts are on curve, PDAs (Program Derived Addresses) are off curve.
func (a *solanaAccountID) IsOnCurve() bool {
	if a == nil {
		return false
	}
	return IsOnCurve(a.pubkey)
}

// IsMainnet returns true if this is a mainnet account.
func (a *solanaAccountID) IsMainnet() bool {
	return a != nil && a.GenericAccountID != nil && a.Reference() == SolanaMainnet.String()
}

// IsDevnet returns true if this is a devnet account.
func (a *solanaAccountID) IsDevnet() bool {
	return a != nil && a.GenericAccountID != nil && a.Reference() == SolanaDevnet.String()
}

// IsTestnet returns true if this is a testnet account.
func (a *solanaAccountID) IsTestnet() bool {
	return a != nil && a.GenericAccountID != nil && a.Reference() == SolanaTestnet.String()
}

// IsZero reports whether the AccountID is the zero value.
func (a *solanaAccountID) IsZero() bool {
	return a == nil || a.GenericAccountID == nil || a.GenericAccountID.IsZero()
}

// Equal reports whether two AccountIDs are equal.
func (a *solanaAccountID) Equal(other AccountID) bool {
	if a.IsZero() && (other == nil || other.IsZero()) {
		return true
	}
	if a.IsZero() || other == nil || other.IsZero() {
		return false
	}
	return a.GenericAccountID.Equal(other)
}

// --- solanaParser ---

type solanaParser struct{}

func (p *solanaParser) Namespace() string {
	return NamespaceSolana
}

func (p *solanaParser) Parse(s string) (AccountID, error) {
	ns, ref, addr, err := SplitCAIP10(s)
	if err != nil {
		return nil, err
	}
	if ns != NamespaceSolana {
		return nil, fmt.Errorf("%w: expected %q, got %q", ErrInvalidNamespace, NamespaceSolana, ns)
	}
	return NewSolanaFromBase58(SolanaNetwork(ref), addr)
}

func (p *solanaParser) ParseAddress(reference, address string) (AccountID, error) {
	return NewSolanaFromBase58(SolanaNetwork(reference), address)
}
