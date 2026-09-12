// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type attributeTypesTestStruct1 struct{}

type attributeTypesTestStruct2 struct {
	ARN             types.String `tfsdk:"arn"`
	ID              types.Int64  `tfsdk:"id"`
	IncludeProperty types.Bool   `tfsdk:"include_property"`
}

type attributeTypesTestStruct3 struct {
	F1 types.String `tfsdk:"f1"`
	attributeTypesTestStruct2
	F2 types.Int32 `tfsdk:"f2"`
}

type attributeTypesTestStructExcludedField struct {
	Name     types.String `tfsdk:"name"`
	Excluded types.String `tfsdk:"-"`
}

type attributeTypesTestStructMissingTag struct {
	Name types.String
}

type attributeTypesTestStructNonAttrValue struct {
	Name    types.String `tfsdk:"name"`
	NotAttr string       `tfsdk:"not_attr"`
}

type attributeTypesTestInnerModel struct {
	Value types.String `tfsdk:"value"`
}

type attributeTypesTestNestedModel struct {
	Name   types.String                                                  `tfsdk:"name"`
	Nested fwtypes.ObjectValueOf[attributeTypesTestInnerModel]           `tfsdk:"nested"`
	List   fwtypes.ListNestedObjectValueOf[attributeTypesTestInnerModel] `tfsdk:"list"`
}

func TestAttributeTypes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attributeTypes func(context.Context) (map[string]attr.Type, diag.Diagnostics)
		expectErr      bool
		expected       map[string]attr.Type
	}{
		"empty struct": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestStruct1],
			expected:       map[string]attr.Type{},
		},
		"non-struct type": {
			attributeTypes: fwtypes.AttributeTypes[int],
			expectErr:      true,
		},
		"missing tfsdk tag": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestStructMissingTag],
			expectErr:      true,
		},
		"flat struct": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestStruct2],
			expected: map[string]attr.Type{
				"arn":              types.StringType,
				"id":               types.Int64Type,
				"include_property": types.BoolType,
			},
		},
		"pointer to struct": {
			attributeTypes: fwtypes.AttributeTypes[*attributeTypesTestStruct2],
			expected: map[string]attr.Type{
				"arn":              types.StringType,
				"id":               types.Int64Type,
				"include_property": types.BoolType,
			},
		},
		"embedded struct": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestStruct3],
			expected: map[string]attr.Type{
				"f1":               types.StringType,
				"arn":              types.StringType,
				"id":               types.Int64Type,
				"include_property": types.BoolType,
				"f2":               types.Int32Type,
			},
		},
		"excluded field": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestStructExcludedField],
			expected: map[string]attr.Type{
				"name": types.StringType,
			},
		},
		"non-attr.Value field skipped": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestStructNonAttrValue],
			expected: map[string]attr.Type{
				"name": types.StringType,
			},
		},
		"nested framework value fields": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestNestedModel],
			expected: map[string]attr.Type{
				"name":   types.StringType,
				"nested": fwtypes.NewObjectTypeOf[attributeTypesTestInnerModel](context.Background()),
				"list":   fwtypes.NewListNestedObjectTypeOf[attributeTypesTestInnerModel](context.Background()),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			got, diags := testCase.attributeTypes(ctx)

			if got, want := diags.HasError(), testCase.expectErr; got != want {
				t.Fatalf("expectErr = %t, got diagnostics: %v", want, diags)
			}
			if testCase.expectErr {
				return
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected diff (+expected, -got): %s", diff)
			}
		})
	}
}

// Concurrency test models. Kept unique to TestAttributeTypesConcurrent so their cache entries
// are cold when the test runs, exercising the concurrent LoadOrStore path.
type attributeTypesConcurrentModelA struct {
	Name types.String `tfsdk:"name"`
	ID   types.Int64  `tfsdk:"id"`
}

type attributeTypesConcurrentModelB struct {
	ARN types.String `tfsdk:"arn"`
}

// TestAttributeTypesConcurrent exercises concurrent access to the cache: many goroutines call
// AttributeTypes simultaneously for both the same type (racing on one key) and distinct types.
// Run with -race -count=N to observe multiple interleavings.
func TestAttributeTypesConcurrent(t *testing.T) {
	t.Parallel()

	const goroutines = 50

	ctx := context.Background()
	start := make(chan struct{})

	var wg sync.WaitGroup
	resultsA := make([]map[string]attr.Type, goroutines)
	wg.Add(goroutines)
	for i := range goroutines {
		go func() {
			defer wg.Done()
			<-start // Release all goroutines together to maximize contention.

			// Alternate between two types so the cache sees concurrent same-key and
			// distinct-key access.
			if i%2 == 0 {
				m, diags := fwtypes.AttributeTypes[attributeTypesConcurrentModelA](ctx)
				if diags.HasError() {
					t.Errorf("unexpected diagnostics: %v", diags)
					return
				}
				resultsA[i] = m
			} else {
				if _, diags := fwtypes.AttributeTypes[attributeTypesConcurrentModelB](ctx); diags.HasError() {
					t.Errorf("unexpected diagnostics: %v", diags)
				}
			}
		}()
	}

	close(start)
	wg.Wait()

	// All goroutines that requested type A must observe the same cached map instance.
	var first map[string]attr.Type
	for i := 0; i < goroutines; i += 2 {
		if first == nil {
			first = resultsA[i]
			continue
		}
		if reflect.ValueOf(resultsA[i]).Pointer() != reflect.ValueOf(first).Pointer() {
			t.Errorf("concurrent calls for the same type returned different map instances")
		}
	}
}

// TestAttributeTypesCacheIdentity confirms that AttributeTypes memoizes its result: repeated
// calls for the same type, and calls for both T and *T, return the same cached map instance.
func TestAttributeTypesCacheIdentity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	value, diags := fwtypes.AttributeTypes[attributeTypesTestStruct2](ctx)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	valueAgain, diags := fwtypes.AttributeTypes[attributeTypesTestStruct2](ctx)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	pointer, diags := fwtypes.AttributeTypes[*attributeTypesTestStruct2](ctx)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	// Map identity is compared by the underlying map header pointer.
	if reflect.ValueOf(value).Pointer() != reflect.ValueOf(valueAgain).Pointer() {
		t.Errorf("repeated calls for the same type returned different map instances")
	}
	if reflect.ValueOf(value).Pointer() != reflect.ValueOf(pointer).Pointer() {
		t.Errorf("T and *T returned different map instances")
	}
}

func BenchmarkAttributeTypes(b *testing.B) {
	ctx := b.Context()

	b.Run("flat", func(b *testing.B) {
		for b.Loop() {
			fwtypes.AttributeTypesMust[attributeTypesTestStruct2](ctx)
		}
	})

	b.Run("nested", func(b *testing.B) {
		for b.Loop() {
			fwtypes.AttributeTypesMust[attributeTypesTestNestedModel](ctx)
		}
	})
}
