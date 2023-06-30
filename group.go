package pjson

import "errors"

// GroupMarshaler is the interface implemented by types that can
// be marshaled as groups. MarshalJSON() also needs to be implemented
// and should return data suitable for cases where readyness hasn't been
// achieved.
type GroupMarshaler interface {
	Marshaler
	GroupMarshalerJSON(st *GroupState) ([]byte, error)
}

var ErrRetryNeeded = errors.New("this value needs state resolution before it can be returned") // this error can only be returned by GroupMarshalerJSON

type groupResolveState struct {
	data any
	err  error
}

type GroupResolveFunc func([]string) ([]any, error)

// GroupState is a struct holding various state information useful during the
// current encoding
type GroupState struct {
	data map[string]*pendingGroupInfo
}

type pendingGroupInfo struct {
	fn       GroupResolveFunc
	err      error
	pending  map[string]bool
	resolved map[string]any
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

func (g *pendingGroupInfo) resolve() {
	if g.err != nil {
		return
	}
	// generate list
	pendinglst := make([]string, 0, len(g.pending))
	for k := range g.pending {
		pendinglst = append(pendinglst, k)
	}
	g.pending = make(map[string]bool) // TODO use clear(g.pending) with go1.21
	vals, err := g.fn(pendinglst)
	if err != nil {
		g.err = err
		return
	}
	if len(vals) > len(pendinglst) {
		// this is not good
		panic("too many responses from group resolver")
	}
	for n, v := range vals {
		g.resolved[pendinglst[n]] = v
	}
}
