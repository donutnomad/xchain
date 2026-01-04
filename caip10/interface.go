// Package caip10 implements the CAIP-10 Account ID Specification.
// See: https://github.com/ChainAgnostic/CAIPs/blob/main/CAIPs/caip-10.md
package caip10

import (
	"database/sql/driver"
	"encoding"
	"encoding/json"
)

// AccountID is the base interface for CAIP-10 account identifiers.
// Format: namespace:reference:address
type AccountID interface {
	// Core accessors

	Namespace() Namespace
	Reference() string
	Address() string
	ChainID() ChainID // CAIP-2 chain ID (namespace:reference)

	// State

	IsZero() bool
	Equal(other AccountID) bool
	Validate() error

	// fmt.Stringer

	String() string

	// Serialization interfaces

	encoding.TextMarshaler
	encoding.TextUnmarshaler
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	json.Marshaler
	json.Unmarshaler

	// Database interfaces

	driver.Valuer
	Scan(src any) error

	// CBOR serialization

	MarshalCBOR() ([]byte, error)
	UnmarshalCBOR(data []byte) error

	// Conversion

	ToColumns() AccountIDColumns
	ToColumnsCompact() AccountIDColumnsCompact
}

// Parser is the interface for namespace-specific parsers.
type Parser interface {
	Namespace() Namespace
	Parse(s string) (AccountID, error)
	ParseAddress(reference, address string) (AccountID, error)
}

// registry holds namespace-specific parsers
var registry = make(map[Namespace]Parser)

// RegisterParser registers a parser for a namespace.
func RegisterParser(p Parser) {
	registry[p.Namespace()] = p
}

// GetParser returns the parser for a namespace.
func GetParser(namespace Namespace) (Parser, bool) {
	p, ok := registry[namespace]
	return p, ok
}

// Parse parses a CAIP-10 string into an AccountID.
// It automatically selects the appropriate parser based on namespace.
func Parse(s string) (AccountID, error) {
	ns, ref, addr, err := SplitCAIP10(s)
	if err != nil {
		return nil, err
	}

	if p, ok := registry[ns]; ok {
		return p.ParseAddress(ref, addr)
	}

	return NewGeneric(ns, ref, addr)
}

// MustParse parses a CAIP-10 string and panics if invalid.
func MustParse(s string) AccountID {
	a, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return a
}

// ParseWithNamespace parses using a specific namespace parser.
func ParseWithNamespace(namespace Namespace, reference, address string) (AccountID, error) {
	if p, ok := registry[namespace]; ok {
		return p.ParseAddress(reference, address)
	}
	return NewGeneric(namespace, reference, address)
}

// ParseWithChainID parses using a specific chainId parser.
func ParseWithChainID(chainID string, address string) (AccountID, error) {
	return Parse(chainID + ":" + address)
}

// Equal compares two AccountIDs for equality.
func Equal(a, b AccountID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(b)
}

// AccountIDColumns is a helper struct for storing AccountID as separate database columns.
type AccountIDColumns struct {
	Namespace string `json:"namespace" db:"namespace" gorm:"column:namespace;type:varchar(8);not null"`
	Reference string `json:"reference" db:"reference" gorm:"column:reference;type:varchar(32);not null"`
	Address   string `json:"address" db:"address" gorm:"column:address;type:varchar(128);not null"`
}

// ToAccountID converts AccountIDColumns back to AccountID with validation.
func (c AccountIDColumns) ToAccountID() (AccountID, error) {
	return ParseWithNamespace(Namespace(c.Namespace), c.Reference, c.Address)
}

// MustToAccountID converts AccountIDColumns to AccountID and panics if invalid.
func (c AccountIDColumns) MustToAccountID() AccountID {
	a, err := c.ToAccountID()
	if err != nil {
		panic(err)
	}
	return a
}

// IsZero reports whether all fields are empty.
func (c AccountIDColumns) IsZero() bool {
	return c.Namespace == "" && c.Reference == "" && c.Address == ""
}

// String returns the CAIP-10 string representation.
func (c AccountIDColumns) String() string {
	if c.IsZero() {
		return ""
	}
	return c.Namespace + ":" + c.Reference + ":" + c.Address
}

// Validate checks if the columns are valid per CAIP-10 spec.
func (c AccountIDColumns) Validate() error {
	_, err := c.ToAccountID()
	return err
}

// ToCompact converts to the compact two-field format.
func (c AccountIDColumns) ToCompact() AccountIDColumnsCompact {
	if c.IsZero() {
		return AccountIDColumnsCompact{}
	}
	return AccountIDColumnsCompact{
		ChainID: c.Namespace + ":" + c.Reference,
		Address: c.Address,
	}
}

// AccountIDColumnsCompact is a compact two-field format for storing AccountID.
// ChainID is the CAIP-2 chain identifier (namespace:reference).
type AccountIDColumnsCompact struct {
	ChainID string `json:"chain_id" db:"chain_id" gorm:"column:chain_id;type:varchar(41);not null"` // namespace:reference (max 8+1+32=41)
	Address string `json:"address" db:"address" gorm:"column:address;type:varchar(128);not null"`
}

// ToAccountID converts AccountIDColumnsCompact back to AccountID with validation.
func (c AccountIDColumnsCompact) ToAccountID() (AccountID, error) {
	if c.IsZero() {
		return nil, ErrEmptyValue
	}
	return Parse(c.ChainID + ":" + c.Address)
}

// MustToAccountID converts AccountIDColumnsCompact to AccountID and panics if invalid.
func (c AccountIDColumnsCompact) MustToAccountID() AccountID {
	a, err := c.ToAccountID()
	if err != nil {
		panic(err)
	}
	return a
}

// IsZero reports whether all fields are empty.
func (c AccountIDColumnsCompact) IsZero() bool {
	return c.ChainID == "" && c.Address == ""
}

// String returns the CAIP-10 string representation.
func (c AccountIDColumnsCompact) String() string {
	if c.IsZero() {
		return ""
	}
	return c.ChainID + ":" + c.Address
}

// Validate checks if the columns are valid per CAIP-10 spec.
func (c AccountIDColumnsCompact) Validate() error {
	_, err := c.ToAccountID()
	return err
}

// ToFull converts to the full three-field format.
func (c AccountIDColumnsCompact) ToFull() (AccountIDColumns, error) {
	if c.IsZero() {
		return AccountIDColumns{}, nil
	}
	// Split ChainID into namespace and reference
	ns, ref, err := SplitCAIP2(c.ChainID)
	if err != nil {
		return AccountIDColumns{}, err
	}
	return AccountIDColumns{
		Namespace: ns,
		Reference: ref,
		Address:   c.Address,
	}, nil
}
