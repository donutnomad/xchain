package caip10

import (
	"errors"
	"fmt"
	"regexp"
)

// Validation constraints per CAIP-10 spec
const (
	NamespaceMinLen = 3
	NamespaceMaxLen = 8
	ReferenceMinLen = 1
	ReferenceMaxLen = 32
	AddressMinLen   = 1
	AddressMaxLen   = 128
)

// Validation regex patterns per CAIP-10/CAIP-2 spec
var (
	NamespaceRegex = regexp.MustCompile(`^[-a-z0-9]{3,8}$`)
	ReferenceRegex = regexp.MustCompile(`^[-_a-zA-Z0-9]{1,32}$`)
	AddressRegex   = regexp.MustCompile(`^[-.%a-zA-Z0-9]{1,128}$`)
)

// Common errors
var (
	ErrInvalidFormat    = errors.New("caip10: invalid account ID format")
	ErrInvalidNamespace = errors.New("caip10: invalid namespace")
	ErrInvalidReference = errors.New("caip10: invalid reference")
	ErrInvalidAddress   = errors.New("caip10: invalid address")
	ErrEmptyValue       = errors.New("caip10: empty value")
)

// SplitCAIP2 splits a CAIP-2 chain ID string into namespace and reference.
// Format: namespace:reference
func SplitCAIP2(s string) (namespace, reference string, err error) {
	if len(s) == 0 {
		return "", "", ErrEmptyValue
	}

	// Find colon separator
	i := 0
	for i < len(s) && s[i] != ':' {
		i++
	}
	if i >= len(s) {
		return "", "", fmt.Errorf("%w: missing namespace separator", ErrInvalidFormat)
	}
	namespace = s[:i]
	reference = s[i+1:]

	// Ensure reference doesn't contain another colon (that would be CAIP-10)
	for j := range reference {
		if reference[j] == ':' {
			return "", "", fmt.Errorf("%w: unexpected colon in reference", ErrInvalidFormat)
		}
	}

	return namespace, reference, nil
}

// SplitCAIP10 splits a CAIP-10 string into namespace, reference, and address.
func SplitCAIP10(s string) (namespace Namespace, reference, address string, err error) {
	if len(s) == 0 {
		return "", "", "", ErrEmptyValue
	}

	// Find first colon for namespace
	i := 0
	for i < len(s) && s[i] != ':' {
		i++
	}
	if i >= len(s) {
		return "", "", "", fmt.Errorf("%w: missing namespace separator", ErrInvalidFormat)
	}
	namespace = Namespace(s[:i])

	// Find second colon for reference
	j := i + 1
	for j < len(s) && s[j] != ':' {
		j++
	}
	if j >= len(s) {
		return "", "", "", fmt.Errorf("%w: missing reference separator", ErrInvalidFormat)
	}
	reference = s[i+1 : j]
	address = s[j+1:]

	return namespace, reference, address, nil
}
