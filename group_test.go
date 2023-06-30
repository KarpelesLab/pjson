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
		t.Errorf("failed to marshal: %s", err)
	}
	// [,]
	// ["FOO","BAR"]
	if string(res) != `["FOO","BAR"]` {
		t.Errorf("unexpected result, expected [\"FOO\",\"BAR\"] but got %s", res)
	}
}
