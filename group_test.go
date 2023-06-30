package pjson_test

import (
	"context"
	"strings"
	"testing"

	"github.com/KarpelesLab/pjson"
)

func resolverA(ctx context.Context, keys []string) ([]any, error) {
	res := make([]any, len(keys))

	for n, k := range keys {
		res[n] = strings.ToUpper(k)
	}
	return res, nil
}

type objectA struct {
	key string
}

type objectB struct {
	A []any
	B map[string]any
	C any
	D any
	E any
}

func (o *objectA) GroupMarshalerJSON(ctx context.Context, st *pjson.GroupState) ([]byte, error) {
	v, err := st.Fetch("resolverA", o.key, resolverA)
	if err != nil {
		return nil, err
	}
	return pjson.MarshalContext(ctx, v)
}

func TestGroups(t *testing.T) {
	tst := []*objectA{
		&objectA{key: "foo"},
		&objectA{key: "bar"},
	}

	res, err := pjson.Marshal(tst)
	if err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}
	// [,]
	// ["FOO","BAR"]
	if string(res) != `["FOO","BAR"]` {
		t.Errorf("unexpected result, expected [\"FOO\",\"BAR\"] but got %s", res)
	}

	tst2 := &objectB{
		A: []any{&objectA{key: "foo"}, "not foo"},
		B: map[string]any{"key": &objectA{key: "test"}, "a": "b", "z": "x"},
		C: "hello",
		D: &objectA{key: "world"},
		E: pjson.GroupCall("resolverA", "keyVal", resolverA),
	}

	res, err = pjson.Marshal(tst2)
	if err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}

	if string(res) != `{"A":["FOO","not foo"],"B":{"a":"b","key":"TEST","z":"x"},"C":"hello","D":"WORLD","E":"KEYVAL"}` {
		t.Errorf(`unexpected result - expected {"A":["FOO","not foo"],"B":{"a":"b","key":"TEST","z":"x"},"C":"hello","D":"WORLD","E":"KEYVAL"} but got: %s`, res)
	}
}
