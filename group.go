package pjson

import (
	"context"
	"errors"
	"reflect"
)

// GroupMarshaler is the interface implemented by types that can
// be marshaled as groups. MarshalJSON() also needs to be implemented
// and should return data suitable for cases where readyness hasn't been
// achieved.
type GroupMarshaler interface {
	GroupMarshalerJSON(ctx context.Context, st *GroupState) ([]byte, error)
}

var ErrRetryNeeded = errors.New("this value needs state resolution before it can be returned") // this error can only be returned by GroupMarshalerJSON

type groupResolveState struct {
	data any
	err  error
}

type GroupResolveFunc func(context.Context, []string) ([]any, error)

type groupCallObj struct {
	group    string
	key      string
	resolver GroupResolveFunc
}

func GroupCall(group, key string, resolver GroupResolveFunc) GroupMarshaler {
	return &groupCallObj{group, key, resolver}
}

func (g *groupCallObj) GroupMarshalerJSON(ctx context.Context, st *GroupState) ([]byte, error) {
	res, err := st.Fetch(g.group, g.key, g.resolver)
	if err != nil {
		return nil, err
	}
	return MarshalContext(ctx, res)
}

// GroupState is a struct holding various state information useful during the
// current encoding
type GroupState struct {
	needRetry int
	data      map[string]*pendingGroupInfo
}

func newGroupState() *GroupState {
	return &GroupState{data: make(map[string]*pendingGroupInfo)}
}

type pendingGroupInfo struct {
	fn       GroupResolveFunc
	err      error
	pending  map[string]bool
	resolved map[string]any
}

// retry will return bool if execution should be retried
func (g *GroupState) retry(ctx context.Context) bool {
	if g == nil {
		return false
	}
	if g.needRetry == 0 {
		return false
	}
	g.needRetry = 0
	needRetry := false
	for _, obj := range g.data {
		if obj.resolve(ctx) {
			needRetry = true
		}
	}
	return needRetry
}

func (g *GroupState) bumpRetry() {
	g.needRetry += 1
}

func (g *GroupState) Fetch(group, key string, resolver GroupResolveFunc) (any, error) {
	ginfo, ok := g.data[group]
	if !ok {
		g.data[group] = &pendingGroupInfo{
			fn:       resolver,
			pending:  map[string]bool{key: true},
			resolved: make(map[string]any),
		}
		return nil, ErrRetryNeeded
	}
	if v, ok := ginfo.resolved[key]; ok {
		return v, nil
	}
	if ginfo.err != nil {
		return nil, ginfo.err
	}
	ginfo.pending[key] = true
	return nil, ErrRetryNeeded
}

// resolve returns true if new stuff has been resolved
func (g *pendingGroupInfo) resolve(ctx context.Context) bool {
	if g.err != nil {
		return false
	}
	// generate list
	pendinglst := make([]string, 0, len(g.pending))
	for k := range g.pending {
		pendinglst = append(pendinglst, k)
	}
	g.pending = make(map[string]bool) // TODO use clear(g.pending) with go1.21
	vals, err := g.fn(ctx, pendinglst)
	if err != nil {
		g.err = err
		return true
	}
	if len(vals) == 0 {
		return false
	}
	if len(vals) > len(pendinglst) {
		// this is not good
		panic("too many responses from group resolver")
	}
	for n, v := range vals {
		g.resolved[pendinglst[n]] = v
	}
	return true
}

// internal encoding methods
func groupMarshalerEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	if v.Kind() == reflect.Pointer && v.IsNil() {
		e.WriteString("null")
		return
	}
	m, ok := v.Interface().(GroupMarshaler)
	if !ok {
		e.WriteString("null")
		return
	}
	if e.groupSt == nil {
		e.groupSt = newGroupState()
	}
	b, err := m.GroupMarshalerJSON(e.ctx, e.groupSt)
	if err == nil {
		e.Grow(len(b))
		out := e.AvailableBuffer()
		out, err = appendCompact(out, b, opts.escapeHTML)
		e.Buffer.Write(out)
	}
	if err != nil {
		e.error(&MarshalerError{v.Type(), err, "MarshalJSON"})
	}
}

func addrGroupMarshalerEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	va := v.Addr()
	if va.IsNil() {
		e.WriteString("null")
		return
	}
	m := va.Interface().(GroupMarshaler)
	if e.groupSt == nil {
		e.groupSt = newGroupState()
	}
	b, err := m.GroupMarshalerJSON(e.ctx, e.groupSt)
	if err == nil {
		e.Grow(len(b))
		out := e.AvailableBuffer()
		out, err = appendCompact(out, b, opts.escapeHTML)
		e.Buffer.Write(out)
	}
	if err != nil {
		e.error(&MarshalerError{v.Type(), err, "MarshalJSON"})
	}
}
