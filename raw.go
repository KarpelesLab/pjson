package pjson

import "encoding/json"

// RawMessage is a raw encoded JSON value. We simply redirect to encoding/json.RawMessage
type RawMessage json.RawMessage
