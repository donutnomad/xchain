package caip10

import (
	"math/big"

	"github.com/holiman/uint256"
)

type eip155ChainID interface {
	*big.Int | *uint256.Int | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

func bigIntOrIntToBigInt[N *big.Int | *uint256.Int | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](num N) *big.Int {
	switch v := any(num).(type) {
	case *big.Int:
		return v
	case *uint256.Int:
		return v.ToBig()
	case uint:
		return new(big.Int).SetUint64(uint64(v))
	case uint8:
		return new(big.Int).SetUint64(uint64(v))
	case uint16:
		return new(big.Int).SetUint64(uint64(v))
	case uint32:
		return new(big.Int).SetUint64(uint64(v))
	case uint64:
		return new(big.Int).SetUint64(v)
	case int:
		return new(big.Int).SetInt64(int64(v))
	case int8:
		return new(big.Int).SetInt64(int64(v))
	case int16:
		return new(big.Int).SetInt64(int64(v))
	case int32:
		return new(big.Int).SetInt64(int64(v))
	case int64:
		return new(big.Int).SetInt64(v)
	default:
		panic("unreachable")
	}
}
