package pjson_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/KarpelesLab/pjson"
)

// Test MarshalerContext interface

type ctxKey string

type contextAwareType struct {
	Value string
}

func (c *contextAwareType) MarshalContextJSON(ctx context.Context) ([]byte, error) {
	// Get value from context and include it in output
	prefix := ""
	if v := ctx.Value(ctxKey("prefix")); v != nil {
		prefix = v.(string)
	}
	return pjson.Marshal(prefix + c.Value)
}

func TestMarshalerContext(t *testing.T) {
	obj := &contextAwareType{Value: "hello"}

	// Without context value
	res, err := pjson.Marshal(obj)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if string(res) != `"hello"` {
		t.Errorf("Expected \"hello\", got %s", res)
	}

	// With context value
	ctx := context.WithValue(context.Background(), ctxKey("prefix"), "ctx:")
	res, err = pjson.MarshalContext(ctx, obj)
	if err != nil {
		t.Fatalf("MarshalContext failed: %v", err)
	}
	if string(res) != `"ctx:hello"` {
		t.Errorf("Expected \"ctx:hello\", got %s", res)
	}
}

// Test UnmarshalerContext interface

type contextAwareUnmarshal struct {
	Value   string
	FromCtx string
}

func (c *contextAwareUnmarshal) UnmarshalContextJSON(ctx context.Context, data []byte) error {
	// Get value from context
	if v := ctx.Value(ctxKey("injected")); v != nil {
		c.FromCtx = v.(string)
	}
	return json.Unmarshal(data, &c.Value)
}

func TestUnmarshalerContext(t *testing.T) {
	// Without context value
	var obj1 contextAwareUnmarshal
	err := pjson.Unmarshal([]byte(`"test"`), &obj1)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if obj1.Value != "test" {
		t.Errorf("Expected Value=\"test\", got %q", obj1.Value)
	}
	if obj1.FromCtx != "" {
		t.Errorf("Expected FromCtx=\"\", got %q", obj1.FromCtx)
	}

	// With context value
	var obj2 contextAwareUnmarshal
	ctx := context.WithValue(context.Background(), ctxKey("injected"), "from-context")
	err = pjson.UnmarshalContext(ctx, []byte(`"test2"`), &obj2)
	if err != nil {
		t.Fatalf("UnmarshalContext failed: %v", err)
	}
	if obj2.Value != "test2" {
		t.Errorf("Expected Value=\"test2\", got %q", obj2.Value)
	}
	if obj2.FromCtx != "from-context" {
		t.Errorf("Expected FromCtx=\"from-context\", got %q", obj2.FromCtx)
	}
}

// Test protect tag

type userWithProtectedFields struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password,protect"`
	APIKey   string `json:"api_key,protect"`
	IsAdmin  bool   `json:"is_admin,protect"`
}

func TestProtectTag(t *testing.T) {
	user := &userWithProtectedFields{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "secret123",
		APIKey:   "key-abc-123",
		IsAdmin:  true,
	}

	// Normal marshal - should include all fields
	res, err := pjson.Marshal(user)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var normalResult map[string]any
	if err := json.Unmarshal(res, &normalResult); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if normalResult["name"] != "John Doe" {
		t.Errorf("Expected name=\"John Doe\", got %v", normalResult["name"])
	}
	if normalResult["password"] != "secret123" {
		t.Errorf("Expected password=\"secret123\", got %v", normalResult["password"])
	}
	if normalResult["api_key"] != "key-abc-123" {
		t.Errorf("Expected api_key=\"key-abc-123\", got %v", normalResult["api_key"])
	}
	if normalResult["is_admin"] != true {
		t.Errorf("Expected is_admin=true, got %v", normalResult["is_admin"])
	}

	// Public marshal - should exclude protected fields
	ctx := pjson.ContextPublic(context.Background())
	res, err = pjson.MarshalContext(ctx, user)
	if err != nil {
		t.Fatalf("MarshalContext failed: %v", err)
	}

	var publicResult map[string]any
	if err := json.Unmarshal(res, &publicResult); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if publicResult["name"] != "John Doe" {
		t.Errorf("Expected name=\"John Doe\", got %v", publicResult["name"])
	}
	if publicResult["email"] != "john@example.com" {
		t.Errorf("Expected email=\"john@example.com\", got %v", publicResult["email"])
	}
	if _, exists := publicResult["password"]; exists {
		t.Errorf("Expected password to be excluded in public mode, but got %v", publicResult["password"])
	}
	if _, exists := publicResult["api_key"]; exists {
		t.Errorf("Expected api_key to be excluded in public mode, but got %v", publicResult["api_key"])
	}
	if _, exists := publicResult["is_admin"]; exists {
		t.Errorf("Expected is_admin to be excluded in public mode, but got %v", publicResult["is_admin"])
	}
}

// Test protect tag with omitempty

type mixedTagsStruct struct {
	Name     string `json:"name"`
	Secret   string `json:"secret,protect"`
	Optional string `json:"optional,omitempty"`
	Both     string `json:"both,protect,omitempty"`
}

func TestProtectWithOmitempty(t *testing.T) {
	obj := &mixedTagsStruct{
		Name:     "test",
		Secret:   "hidden",
		Optional: "",
		Both:     "both-value",
	}

	// Normal marshal
	res, err := pjson.Marshal(obj)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var normalResult map[string]any
	if err := json.Unmarshal(res, &normalResult); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if _, exists := normalResult["optional"]; exists {
		t.Errorf("Expected optional to be omitted (empty), but it exists")
	}
	if normalResult["secret"] != "hidden" {
		t.Errorf("Expected secret=\"hidden\", got %v", normalResult["secret"])
	}
	if normalResult["both"] != "both-value" {
		t.Errorf("Expected both=\"both-value\", got %v", normalResult["both"])
	}

	// Public marshal
	ctx := pjson.ContextPublic(context.Background())
	res, err = pjson.MarshalContext(ctx, obj)
	if err != nil {
		t.Fatalf("MarshalContext failed: %v", err)
	}

	var publicResult map[string]any
	if err := json.Unmarshal(res, &publicResult); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if _, exists := publicResult["secret"]; exists {
		t.Errorf("Expected secret to be excluded in public mode")
	}
	if _, exists := publicResult["both"]; exists {
		t.Errorf("Expected both to be excluded in public mode")
	}
}

// Test nested structs with protect

type innerSecret struct {
	Public string `json:"public"`
	Secret string `json:"secret,protect"`
}

type outerStruct struct {
	Name  string       `json:"name"`
	Inner *innerSecret `json:"inner"`
}

func TestProtectNested(t *testing.T) {
	obj := &outerStruct{
		Name: "outer",
		Inner: &innerSecret{
			Public: "visible",
			Secret: "hidden",
		},
	}

	// Public marshal - nested protected fields should also be excluded
	ctx := pjson.ContextPublic(context.Background())
	res, err := pjson.MarshalContext(ctx, obj)
	if err != nil {
		t.Fatalf("MarshalContext failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(res, &result); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	inner, ok := result["inner"].(map[string]any)
	if !ok {
		t.Fatalf("Expected inner to be an object, got %T", result["inner"])
	}

	if inner["public"] != "visible" {
		t.Errorf("Expected inner.public=\"visible\", got %v", inner["public"])
	}
	if _, exists := inner["secret"]; exists {
		t.Errorf("Expected inner.secret to be excluded in public mode")
	}
}

// Test RawMessage compatibility with standard library

func TestRawMessageCompatibility(t *testing.T) {
	// pjson.RawMessage should be usable where json.RawMessage is expected
	var pjsonRaw pjson.RawMessage = []byte(`{"test": true}`)
	var jsonRaw json.RawMessage = pjsonRaw // This should compile

	if string(jsonRaw) != `{"test": true}` {
		t.Errorf("RawMessage conversion failed")
	}

	// And vice versa
	jsonRaw2 := json.RawMessage(`[1,2,3]`)
	pjsonRaw2 := pjson.RawMessage(jsonRaw2)

	if string(pjsonRaw2) != `[1,2,3]` {
		t.Errorf("RawMessage conversion failed")
	}
}

// Test MarshalerContext in slices and maps

type ctxAwareItem struct {
	ID string
}

func (c *ctxAwareItem) MarshalContextJSON(ctx context.Context) ([]byte, error) {
	multiplier := 1
	if v := ctx.Value(ctxKey("multiplier")); v != nil {
		multiplier = v.(int)
	}
	return pjson.Marshal(c.ID + ":" + string(rune('0'+multiplier)))
}

func TestMarshalerContextInCollections(t *testing.T) {
	items := []*ctxAwareItem{
		{ID: "a"},
		{ID: "b"},
	}

	ctx := context.WithValue(context.Background(), ctxKey("multiplier"), 5)
	res, err := pjson.MarshalContext(ctx, items)
	if err != nil {
		t.Fatalf("MarshalContext failed: %v", err)
	}

	if string(res) != `["a:5","b:5"]` {
		t.Errorf("Expected [\"a:5\",\"b:5\"], got %s", res)
	}

	// Test in map
	itemMap := map[string]*ctxAwareItem{
		"first":  {ID: "x"},
		"second": {ID: "y"},
	}

	res, err = pjson.MarshalContext(ctx, itemMap)
	if err != nil {
		t.Fatalf("MarshalContext failed: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(res, &result); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	if result["first"] != "x:5" {
		t.Errorf("Expected first=\"x:5\", got %v", result["first"])
	}
	if result["second"] != "y:5" {
		t.Errorf("Expected second=\"y:5\", got %v", result["second"])
	}
}

// Test UnmarshalerContext with objects and arrays

type ctxAwareContainer struct {
	Items []string
	Ctx   string
}

func (c *ctxAwareContainer) UnmarshalContextJSON(ctx context.Context, data []byte) error {
	if v := ctx.Value(ctxKey("container-ctx")); v != nil {
		c.Ctx = v.(string)
	}
	return json.Unmarshal(data, &c.Items)
}

func TestUnmarshalerContextWithArray(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxKey("container-ctx"), "array-context")

	var container ctxAwareContainer
	err := pjson.UnmarshalContext(ctx, []byte(`["a","b","c"]`), &container)
	if err != nil {
		t.Fatalf("UnmarshalContext failed: %v", err)
	}

	if len(container.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(container.Items))
	}
	if container.Ctx != "array-context" {
		t.Errorf("Expected Ctx=\"array-context\", got %q", container.Ctx)
	}
}
