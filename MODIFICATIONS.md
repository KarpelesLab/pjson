# PJSON Modifications from Go encoding/json

This document tracks all modifications made to the standard Go `encoding/json` library to create the `pjson` package.

**Current upstream version:** Go 1.23.4's encoding/json package (upgraded from Go 1.20.5)

## Summary of Custom Features

### 1. Context Support

Added context support throughout the encoding and decoding process, allowing context-aware JSON marshaling and unmarshaling.

#### New Types and Interfaces

- **`MarshalerContext`** interface (encode.go:245-249):
  ```go
  type MarshalerContext interface {
      MarshalContextJSON(ctx context.Context) ([]byte, error)
  }
  ```

- **`UnmarshalerContext`** interface (decode.go:130-132):
  ```go
  type UnmarshalerContext interface {
      UnmarshalContextJSON(context.Context, []byte) error
  }
  ```

#### New Functions

- **`MarshalContext(ctx context.Context, v any) ([]byte, error)`** (encode.go:174-186)
  - Same as `Marshal` but accepts a context that is passed to `MarshalerContext` implementers

- **`UnmarshalContext(ctx context.Context, data []byte, v any) error`** (decode.go:103-116)
  - Same as `Unmarshal` but accepts a context that is passed to `UnmarshalerContext` implementers

- **`NewDecoderContext(ctx context.Context, r io.Reader) *Decoder`** (stream.go:35-39)
  - Creates a decoder with context support

- **`NewEncoderContext(ctx context.Context, w io.Writer) *Encoder`** (stream.go:205-207)
  - Creates an encoder with context support

- **`(*Encoder).SetContext(ctx context.Context)`** (stream.go:274-276)
  - Sets the encoder's context

#### Modified Internal Structures

- **`encodeState`** (encode.go:309-325):
  - Added `ctx context.Context` field
  - Added `needRetry int` field for group marshaling support
  - Added `groupSt *GroupState` field for group state tracking
  - Added `public bool` field for protected field filtering

- **`decodeState`** (decode.go:221-232):
  - Added `ctx context.Context` field

- **`Encoder`** (stream.go:187-197):
  - Added `ctx context.Context` field
  - Added `public bool` field

### 2. Group Marshaling System

A powerful feature that allows batching multiple values that need external resolution (like database lookups) into single calls, avoiding N+1 query problems during JSON encoding.

#### New File: group.go

- **`GroupMarshaler`** interface:
  ```go
  type GroupMarshaler interface {
      GroupMarshalerJSON(ctx context.Context, st *GroupState) ([]byte, error)
  }
  ```

- **`GroupState`** struct:
  - Manages state for batched resolution during encoding
  - Provides `Fetch(group, key string, resolver GroupResolveFunc) (any, error)` method

- **`GroupResolveFunc`** type:
  ```go
  type GroupResolveFunc func(context.Context, []string) ([]any, error)
  ```

- **`GroupCall(group, key string, resolver GroupResolveFunc) GroupMarshaler`**:
  - Helper function to create dynamic group-resolved values

- **`ErrRetryNeeded`** error:
  - Signals that the encoding needs to retry after group resolution

#### Encoding Changes for Groups

- The `marshal` function in encode.go now includes a retry loop (encode.go:367-378) that:
  1. Attempts to encode the value
  2. Collects all group fetch requests
  3. Resolves batched requests
  4. Retries encoding with resolved values

- Added `groupMarshalerEncoder` and `addrGroupMarshalerEncoder` functions (group.go:134-175)

### 3. Public/Protect Tag Option

Allows marking struct fields as "protected" so they can be hidden when encoding for public consumption.

#### New File: context.go

- **`ContextPublic(parent context.Context) context.Context`**:
  - Returns a context that marks the encoding as "public"

- **`isPublic(ctx context.Context) bool`**:
  - Checks if a context is marked as public

#### Field Tag Support

- Added `protect` tag option (encode.go:1408):
  ```go
  Field string `json:"field,protect"`
  ```

- Added `protect bool` field to `field` struct (encode.go:1292)

- Modified `structEncoder.encode()` to skip protected fields when encoding in public mode (encode.go:854-856):
  ```go
  if e.public && f.protect {
      continue
  }
  ```

### 4. RawMessage Type Alias

#### New File: raw.go

- **`RawMessage`** is now an alias for `json.RawMessage`:
  ```go
  type RawMessage = json.RawMessage
  ```
  This ensures compatibility with the standard library's `RawMessage` type.

## Files Modified from Original

| File | Description of Changes |
|------|----------------------|
| `encode.go` | Added context fields to encodeState, MarshalerContext interface, MarshalContext function, group marshaler support, protect tag handling, ctxMarshalerEncoder functions |
| `decode.go` | Added context field to decodeState, UnmarshalerContext interface, UnmarshalContext function, modified indirect() to support UnmarshalerContext |
| `stream.go` | Added context fields to Encoder/Decoder, NewDecoderContext, NewEncoderContext, SetContext functions |

## Files Added

| File | Description |
|------|-------------|
| `context.go` | Public context helpers for protect tag functionality |
| `group.go` | Group marshaling system implementation |
| `group_test.go` | Tests for group marshaling |
| `raw.go` | RawMessage type alias |

## API Summary

### Encoding with Context
```go
// With MarshalerContext interface
type MyType struct {
    Data string
}

func (m *MyType) MarshalContextJSON(ctx context.Context) ([]byte, error) {
    // Access context values during marshaling
    return pjson.Marshal(m.Data)
}

data, err := pjson.MarshalContext(ctx, myValue)
```

### Group Marshaling
```go
type UserRef struct {
    UserID string
}

func (u *UserRef) GroupMarshalerJSON(ctx context.Context, st *pjson.GroupState) ([]byte, error) {
    user, err := st.Fetch("users", u.UserID, batchLoadUsers)
    if err != nil {
        return nil, err
    }
    return pjson.MarshalContext(ctx, user)
}

func batchLoadUsers(ctx context.Context, userIDs []string) ([]any, error) {
    // Load all users in a single database query
    return users, nil
}
```

### Protected Fields
```go
type User struct {
    Name     string `json:"name"`
    Password string `json:"password,protect"` // Hidden in public mode
}

// Normal encoding includes password
data, _ := pjson.Marshal(user)

// Public encoding excludes protected fields
ctx := pjson.ContextPublic(context.Background())
publicData, _ := pjson.MarshalContext(ctx, user)
```

## Upgrade Notes

When upgrading from upstream Go encoding/json:

1. The following custom additions must be preserved:
   - `context.go` - entire file
   - `group.go` - entire file
   - `raw.go` - entire file
   - Context support in `encodeState`, `decodeState`, `Encoder`, `Decoder`
   - `MarshalerContext` and `UnmarshalerContext` interfaces
   - `MarshalContext`, `UnmarshalContext` functions
   - `NewDecoderContext`, `NewEncoderContext` functions
   - Group marshaler encoder functions and retry logic in `marshal()`
   - `protect` tag handling in `field` struct and `structEncoder.encode()`
   - `ctxMarshalerEncoder` and `addrCtxMarshalerEncoder` functions

2. The `newTypeEncoder` function needs to check for custom interfaces in order:
   - `GroupMarshaler` (checked first)
   - `MarshalerContext`
   - `Marshaler`
   - `encoding.TextMarshaler`
