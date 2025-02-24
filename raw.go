package pjson

import (
	"encoding/json"
)

// RawMessage is a raw encoded JSON value.
// It implements Marshaler and Unmarshaler and can
// be used to delay JSON decoding or precompute a JSON encoding.
type RawMessage = json.RawMessage

var _ Marshaler = (*RawMessage)(nil)
var _ Unmarshaler = (*RawMessage)(nil)
