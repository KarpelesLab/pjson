package pjson

import "context"

type jsonContextOption int

const (
	jsonOptionPublic jsonContextOption = iota
)

func ContextPublic(parent context.Context) context.Context {
	return context.WithValue(parent, jsonOptionPublic, true)
}

func isPublic(ctx context.Context) bool {
	v, ok := ctx.Value(jsonOptionPublic).(bool)
	return ok && v
}
