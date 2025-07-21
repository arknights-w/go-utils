package context

import (
	"context"
	"fmt"
	"testing"
)

func TestContext(t *testing.T) {
	var (
		ctx_key, ctx_val     = "a", "b"
		src_key, src_val     = "c", "d"
		child_key, child_val = "e", "f"
	)
	type str string
	var (
		child_key2, child_val2 = str("e"), str("ff")
	)
	ctx := context.Background()
	ctx = context.WithValue(ctx, ctx_key, ctx_val)
	srcCtx := NewContext(ctx)
	srcCtx.Set(src_key, src_val)
	childCtx := srcCtx.Child()
	childCtx.Set(child_key, child_val)
	childCtx.Set(child_key2, child_val2)

	fmt.Printf("ctx_val: %v\n", childCtx.Value(ctx_key))
	fmt.Printf("src_val: %v\n", childCtx.Value(src_key))
	fmt.Printf("child_val: %v\n", childCtx.Value(child_key))
	fmt.Printf("src_val2: %v\n", childCtx.Value(child_key2))
}

func TestSyncContext(t *testing.T) {
	var (
		ctx_key, ctx_val     = "a", "b"
		src_key, src_val     = "c", "d"
		child_key, child_val = "e", "f"
	)
	type str string
	var (
		src_key2, src_val2 = str("e"), str("ff")
	)
	ctx := context.Background()
	ctx = context.WithValue(ctx, ctx_key, ctx_val)
	srcCtx := NewSyncContext(ctx)
	srcCtx.Set(src_key, src_val)
	childCtx := srcCtx.Child()
	childCtx.Set(child_key, child_val)
	childCtx.Set(src_key2, src_val2)

	fmt.Printf("ctx_val: %v\n", childCtx.Value(ctx_key))
	fmt.Printf("src_val: %v\n", childCtx.Value(src_key))
	fmt.Printf("child_val: %v\n", childCtx.Value(child_key))
	fmt.Printf("src_val2: %v\n", childCtx.Value(src_key2))
	if childCtx.Err() != nil {
		t.Fatalf("childCtx.Err()!=nil")
	}
}
