package caip10

import (
	"fmt"
	"math/big"

	"github.com/donutnomad/eths/ecommon"
)

const NamespaceEIP155 Namespace = "eip155"

// maxEIP155ChainID is the maximum chain ID allowed (32 decimal digits: 10^32 - 1).
// This ensures the reference string fits within CAIP-10's 32-character limit.
var maxEIP155ChainID = func() *big.Int {
	maxVal := new(big.Int)
	maxVal.Exp(big.NewInt(10), big.NewInt(32), nil) // 10^32
	maxVal.Sub(maxVal, big.NewInt(1))               // 10^32 - 1
	return maxVal
}()

// EIP155AccountID is the interface for EIP-155 (Ethereum) account IDs.
type EIP155AccountID interface {
	AccountID
	// Account returns the native ecommon.Address.
	Account() ecommon.Address
	// EIP155ChainID returns the chain ID as *big.Int.
	EIP155ChainID() *big.Int
	// SetChainID returns a new EIP155AccountID with the specified chain ID.
	SetChainID(chainID *big.Int) EIP155AccountID
	// SetAddress returns a new EIP155AccountID with the specified address.
	SetAddress(address ecommon.Address) EIP155AccountID
}

// Ensure eip155AccountID implements EIP155AccountID at compile time
var _ EIP155AccountID = (*eip155AccountID)(nil)

func init() {
	RegisterParser(&eip155Parser{})
}

// eip155AccountID represents an Ethereum account ID per CAIP-10.
type eip155AccountID struct {
	*GenericAccountID                 // embedded, inherits all serialization methods
	ethAddr           ecommon.Address // native Ethereum address
	chainID           *big.Int        // EIP-155 chain ID
}

// NewEIP155 creates a new EIP155AccountID.
// If chainID exceeds maxEIP155ChainID (10^32 - 1), it will be capped to that value.
func NewEIP155[C eip155ChainID](chainID C, address ecommon.Address) EIP155AccountID {
	_chainID := bigIntOrIntToBigInt(chainID)
	// Cap chainID to maximum allowed value (10^32 - 1)
	if _chainID.Cmp(maxEIP155ChainID) > 0 {
		_chainID = new(big.Int).Set(maxEIP155ChainID)
	}
	return &eip155AccountID{
		GenericAccountID: newGenericUnchecked(NamespaceEIP155, _chainID.String(), address.Hex()),
		ethAddr:          address,
		chainID:          new(big.Int).Set(_chainID),
	}
}

// NewEIP155FromHex creates a new EIP155AccountID from a chain ID and hex address string.
func NewEIP155FromHex[C eip155ChainID](chainID C, hexAddress string) EIP155AccountID {
	addr := ecommon.HexToAddress(hexAddress)
	return NewEIP155(chainID, addr)
}

// newEIP155FromReference creates EIP155AccountID from string reference (used by parser).
func newEIP155FromReference(reference, hexAddress string) (EIP155AccountID, error) {
	chainID, ok := new(big.Int).SetString(reference, 10)
	if !ok {
		return nil, fmt.Errorf("%w: invalid chain ID %q", ErrInvalidReference, reference)
	}
	return NewEIP155FromHex(chainID, hexAddress), nil
}

// Account returns the native ecommon.Address.
func (a *eip155AccountID) Account() ecommon.Address {
	if a == nil {
		return ecommon.Address{}
	}
	return a.ethAddr
}

// EIP155ChainID returns the chain ID as *big.Int.
func (a *eip155AccountID) EIP155ChainID() *big.Int {
	if a == nil || a.chainID == nil {
		return nil
	}
	return new(big.Int).Set(a.chainID)
}

// SetChainID returns a new EIP155AccountID with the specified chain ID.
func (a *eip155AccountID) SetChainID(chainID *big.Int) EIP155AccountID {
	if a == nil {
		return nil
	}
	return NewEIP155(chainID, a.ethAddr)
}

// SetAddress returns a new EIP155AccountID with the specified address.
func (a *eip155AccountID) SetAddress(address ecommon.Address) EIP155AccountID {
	if a == nil {
		return nil
	}
	return NewEIP155(a.chainID, address)
}

// IsZero reports whether the AccountID is the zero value.
func (a *eip155AccountID) IsZero() bool {
	return a == nil || a.GenericAccountID == nil || a.GenericAccountID.IsZero()
}

// Equal reports whether two AccountIDs are equal.
func (a *eip155AccountID) Equal(other AccountID) bool {
	if a.IsZero() && (other == nil || other.IsZero()) {
		return true
	}
	if a.IsZero() || other == nil || other.IsZero() {
		return false
	}
	// For EIP155, use embedded GenericAccountID's Equal
	return a.GenericAccountID.Equal(other)
}

// --- eip155Parser ---

type eip155Parser struct{}

func (p *eip155Parser) Namespace() Namespace {
	return NamespaceEIP155
}

func (p *eip155Parser) Parse(s string) (AccountID, error) {
	ns, ref, addr, err := SplitCAIP10(s)
	if err != nil {
		return nil, err
	}
	if ns != NamespaceEIP155 {
		return nil, fmt.Errorf("%w: expected %q, got %q", ErrInvalidNamespace, NamespaceEIP155, ns)
	}
	return newEIP155FromReference(ref, addr)
}

func (p *eip155Parser) ParseAddress(reference, address string) (AccountID, error) {
	return newEIP155FromReference(reference, address)
}
