package caip10

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// Ensure GenericAccountID implements AccountID at compile time
var _ AccountID = (*GenericAccountID)(nil)

// GenericAccountID is the base implementation of AccountID.
// It can be embedded by namespace-specific implementations to inherit serialization methods.
type GenericAccountID struct {
	// https://github.com/ChainAgnostic/namespaces
	namespace Namespace
	reference string
	address   string
}

// NewGeneric creates a new GenericAccountID with validation.
func NewGeneric(namespace Namespace, reference, address string) (*GenericAccountID, error) {
	a := &GenericAccountID{
		namespace: namespace,
		reference: reference,
		address:   address,
	}
	if err := a.Validate(); err != nil {
		return nil, err
	}
	return a, nil
}

// MustNewGeneric creates a new GenericAccountID and panics if invalid.
func MustNewGeneric(namespace Namespace, reference, address string) *GenericAccountID {
	a, err := NewGeneric(namespace, reference, address)
	if err != nil {
		panic(err)
	}
	return a
}

// newGenericUnchecked creates without validation (for internal use by embedders).
func newGenericUnchecked(namespace Namespace, reference, address string) *GenericAccountID {
	return &GenericAccountID{
		namespace: namespace,
		reference: reference,
		address:   address,
	}
}

// Namespace returns the blockchain namespace.
func (a *GenericAccountID) Namespace() Namespace {
	if a == nil {
		return ""
	}
	return a.namespace
}

// Reference returns the chain reference.
func (a *GenericAccountID) Reference() string {
	if a == nil {
		return ""
	}
	return a.reference
}

// Address returns the account address.
func (a *GenericAccountID) Address() string {
	if a == nil {
		return ""
	}
	return a.address
}

// ChainID returns the CAIP-2 chain ID (namespace:reference).
func (a *GenericAccountID) ChainID() ChainID {
	if a == nil {
		return ChainID{}
	}
	return ChainID{
		Namespace: a.Namespace(),
		Reference: a.Reference(),
	}
}

// String returns the full CAIP-10 string representation.
func (a *GenericAccountID) String() string {
	if a.IsZero() {
		return ""
	}
	return string(a.namespace) + ":" + a.reference + ":" + a.address
}

// IsZero reports whether the AccountID is the zero value.
func (a *GenericAccountID) IsZero() bool {
	return a == nil || (a.namespace == "" && a.reference == "" && a.address == "")
}

// Equal reports whether two AccountIDs are equal.
func (a *GenericAccountID) Equal(other AccountID) bool {
	if a.IsZero() && (other == nil || other.IsZero()) {
		return true
	}
	if a.IsZero() || other == nil || other.IsZero() {
		return false
	}
	return a.namespace == other.Namespace() &&
		a.reference == other.Reference() &&
		a.address == other.Address()
}

// Validate checks if the AccountID is valid per CAIP-10 spec.
func (a *GenericAccountID) Validate() error {
	if a == nil {
		return ErrEmptyValue
	}
	if !NamespaceRegex.MatchString(string(a.namespace)) {
		return fmt.Errorf("%w: must match [-a-z0-9]{3,8}, got %q", ErrInvalidNamespace, a.namespace)
	}
	switch a.namespace {
	case NamespaceEIP155:
		_, err := newEIP155FromReference(a.reference, a.address)
		if err != nil {
			return err
		}
	case NamespaceSolana:
		_, err := NewSolanaFromBase58(SolanaNetwork(a.reference), a.address)
		if err != nil {
			return err
		}
	case NamespaceBIP122:
		_, err := NewBIP122WithValidation(BIP122Network(a.reference), a.address)
		if err != nil {
			return err
		}
	default:
		if !ReferenceRegex.MatchString(a.reference) {
			return fmt.Errorf("%w: must match [-_a-zA-Z0-9]{1,32}, got %q", ErrInvalidReference, a.reference)
		}
		if !AddressRegex.MatchString(a.address) {
			return fmt.Errorf("%w: must match [-.%%a-zA-Z0-9]{1,128}, got %q", ErrInvalidAddress, a.address)
		}
	}
	return nil
}

// ToColumns converts to AccountIDColumns for database storage.
func (a *GenericAccountID) ToColumns() AccountIDColumns {
	if a == nil {
		return AccountIDColumns{}
	}
	return AccountIDColumns{
		Namespace: string(a.namespace),
		Reference: a.reference,
		Address:   a.address,
	}
}

// ToColumnsCompact converts to AccountIDColumnsCompact for database storage.
func (a *GenericAccountID) ToColumnsCompact() AccountIDColumnsCompact {
	if a == nil {
		return AccountIDColumnsCompact{}
	}
	return AccountIDColumnsCompact{
		ChainID: a.ChainID().String(),
		Address: a.address,
	}
}

// ToNative converts GenericAccountID to its namespace-specific type.
// Returns EIP155AccountID for eip155, SolanaAccountID for solana, or *GenericAccountID for others.
func (a *GenericAccountID) ToNative() any {
	if a == nil {
		return nil
	}
	if p, ok := registry[a.namespace]; ok {
		native, err := p.ParseAddress(a.reference, a.address)
		if err == nil {
			return native
		}
	}
	return a
}

// --- encoding.TextMarshaler / encoding.TextUnmarshaler ---

func (a *GenericAccountID) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *GenericAccountID) UnmarshalText(text []byte) error {
	ns, ref, addr, err := SplitCAIP10(string(text))
	if err != nil {
		return err
	}
	parsed, err := NewGeneric(ns, ref, addr)
	if err != nil {
		return err
	}
	*a = *parsed
	return nil
}

// --- encoding.BinaryMarshaler / encoding.BinaryUnmarshaler ---

func (a *GenericAccountID) MarshalBinary() ([]byte, error) {
	if a == nil {
		return []byte{0, 0, 0, 0}, nil
	}
	nsLen := len(a.namespace)
	refLen := len(a.reference)
	addrLen := len(a.address)

	buf := make([]byte, 4+nsLen+refLen+addrLen)
	buf[0] = byte(nsLen)
	buf[1] = byte(refLen)
	binary.BigEndian.PutUint16(buf[2:4], uint16(addrLen))

	offset := 4
	copy(buf[offset:], a.namespace)
	offset += nsLen
	copy(buf[offset:], a.reference)
	offset += refLen
	copy(buf[offset:], a.address)

	return buf, nil
}

func (a *GenericAccountID) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("%w: binary data too short", ErrInvalidFormat)
	}

	nsLen := int(data[0])
	refLen := int(data[1])
	addrLen := int(binary.BigEndian.Uint16(data[2:4]))

	expectedLen := 4 + nsLen + refLen + addrLen
	if len(data) != expectedLen {
		return fmt.Errorf("%w: binary data length mismatch", ErrInvalidFormat)
	}

	offset := 4
	namespace := string(data[offset : offset+nsLen])
	offset += nsLen
	reference := string(data[offset : offset+refLen])
	offset += refLen
	address := string(data[offset : offset+addrLen])

	parsed, err := NewGeneric(Namespace(namespace), reference, address)
	if err != nil {
		return err
	}
	*a = *parsed
	return nil
}

// --- json.Marshaler / json.Unmarshaler ---

func (a *GenericAccountID) MarshalJSON() ([]byte, error) {
	if a.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(`"` + a.String() + `"`), nil
}

func (a *GenericAccountID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*a = GenericAccountID{}
		return nil
	}

	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("%w: expected JSON string", ErrInvalidFormat)
	}

	s := string(data[1 : len(data)-1])
	if s == "" {
		*a = GenericAccountID{}
		return nil
	}

	return a.UnmarshalText([]byte(s))
}

// --- database/sql interfaces ---

func (a *GenericAccountID) Value() (driver.Value, error) {
	if a.IsZero() {
		return nil, nil
	}
	return a.String(), nil
}

func (a *GenericAccountID) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*a = GenericAccountID{}
		return nil
	case string:
		if v == "" {
			*a = GenericAccountID{}
			return nil
		}
		return a.UnmarshalText([]byte(v))
	case []byte:
		if len(v) == 0 {
			*a = GenericAccountID{}
			return nil
		}
		return a.UnmarshalText(v)
	default:
		return fmt.Errorf("caip10: cannot scan type %T into GenericAccountID", src)
	}
}

// --- CBOR ---

func (a *GenericAccountID) MarshalCBOR() ([]byte, error) {
	if a.IsZero() {
		return cbor.Marshal("")
	}
	return cbor.Marshal(a.String())
}

func (a *GenericAccountID) UnmarshalCBOR(data []byte) error {
	var s string
	if err := cbor.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		*a = GenericAccountID{}
		return nil
	}
	return a.UnmarshalText([]byte(s))
}

// --- Generic Parser ---

// GenericParser is the default parser for unknown namespaces.
type GenericParser struct {
	ns Namespace
}

// NewGenericParser creates a parser for a specific namespace.
func NewGenericParser(namespace Namespace) *GenericParser {
	return &GenericParser{ns: namespace}
}

func (p *GenericParser) Namespace() Namespace {
	return p.ns
}

func (p *GenericParser) Parse(s string) (AccountID, error) {
	ns, ref, addr, err := SplitCAIP10(s)
	if err != nil {
		return nil, err
	}
	if ns != p.ns {
		return nil, fmt.Errorf("%w: expected namespace %q, got %q", ErrInvalidNamespace, p.ns, ns)
	}
	return NewGeneric(ns, ref, addr)
}

func (p *GenericParser) ParseAddress(reference, address string) (AccountID, error) {
	return NewGeneric(p.ns, reference, address)
}
