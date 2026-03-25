// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package privatestate_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/privatestate"
)

func TestWriteOnlyValueStore_HasValue(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	store1 := privatestate.NewWriteOnlyValueStore(&privateState{}, "key")
	store2 := privatestate.NewWriteOnlyValueStore(&privateState{}, "key")
	store2.SetValue(ctx, types.StringValue("value1"))

	testCases := []struct {
		testName  string
		store     *privatestate.WriteOnlyValueStore
		wantValue bool
	}{
		{
			testName: "empty state",
			store:    store1,
		},
		{
			testName:  "has value",
			store:     store2,
			wantValue: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()
			gotValue, diags := testCase.store.HasValue(ctx)
			if diags.HasError() {
				t.Fatal("unexpected error")
			}
			if got, want := gotValue, testCase.wantValue; !cmp.Equal(got, want) {
				t.Errorf("got %t, want %t", got, want)
			}
		})
	}
}

func TestWriteOnlyValueStore_EqualValue(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	store1 := privatestate.NewWriteOnlyValueStore(&privateState{}, "key")
	store1.SetValue(ctx, types.StringValue("value1"))
	store2 := privatestate.NewWriteOnlyValueStore(&privateState{}, "key")
	store2.SetValue(ctx, types.StringValue("value2"))

	testCases := []struct {
		testName  string
		store     *privatestate.WriteOnlyValueStore
		wantEqual bool
	}{
		{
			testName:  "equal",
			store:     store1,
			wantEqual: true,
		},
		{
			testName: "not equal",
			store:    store2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()
			gotEqual, diags := testCase.store.EqualValue(ctx, types.StringValue("value1"))
			if diags.HasError() {
				t.Fatal("unexpected error")
			}
			if got, want := gotEqual, testCase.wantEqual; !cmp.Equal(got, want) {
				t.Errorf("got %t, want %t", got, want)
			}
		})
	}
}

type privateState struct {
	data map[string][]byte
}

func (p *privateState) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics
	bytes := p.data[key]
	return bytes, diags
}

func (p *privateState) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	var diags diag.Diagnostics

	if p.data == nil {
		p.data = make(map[string][]byte)
	}

	p.data[key] = value
	return diags
}
