package gomerk

import "errors"

var (
	ErrEmptyTree         = errors.New("expected non-zero number of leaves")
	ErrInvalidNodeLength = errors.New("expected 32 bytes")
	ErrNotALeaf          = errors.New("index is not a leaf")
	ErrLeafNotInTree     = errors.New("leaf is not in tree")
	ErrDuplicatedIndex   = errors.New("cannot prove duplicated index")
	ErrIndexOutOfBounds  = errors.New("index out of bounds")
	ErrInvalidFormat     = errors.New("invalid tree format")
	ErrInvariant         = errors.New("invariant violation")
	ErrInvalidHex        = errors.New("invalid hex string")
	ErrAbiEncode         = errors.New("abi encoding error")
	ErrUnsupportedType   = errors.New("unsupported type")
	ErrMismatchedCount   = errors.New("mismatched leaf encoding count")
)
