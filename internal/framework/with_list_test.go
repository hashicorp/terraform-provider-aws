// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type testNestedModel struct {
	Field1 types.String `tfsdk:"field1"`
	Field2 types.Int32  `tfsdk:"field2"`
}

type testListModel struct {
	Name   types.String                                     `tfsdk:"name"`
	Tags   fwtypes.MapValueOf[types.String]                 `tfsdk:"tags"`
	Config fwtypes.ObjectValueOf[testNestedModel]           `tfsdk:"config"`
	Rules  fwtypes.ListNestedObjectValueOf[testNestedModel] `tfsdk:"rules"`
	Nested testNestedModel                                  `tfsdk:"nested"`
}

type testMixedFieldModel struct {
	Name   types.String `tfsdk:"name"`
	Count  int          `tfsdk:"count"`
	hidden types.String `tfsdk:"hidden"`
}

type testDeepLeafModel struct {
	Field1 types.String `tfsdk:"field1"`
	Field2 types.Int32  `tfsdk:"field2"`
}

type testDeepMiddleModel struct {
	Leaf testDeepLeafModel `tfsdk:"leaf"`
}

type testDeepRootModel struct {
	Name   types.String        `tfsdk:"name"`
	Middle testDeepMiddleModel `tfsdk:"middle"`
}

func TestSetZeroValueAttrFieldsToNull(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := map[string]struct {
		data     *testListModel
		expected *testListModel
	}{
		"zero attr fields become null": {
			data: &testListModel{
				Name: types.StringValue("example"),
			},
			expected: expectedTestListModel(ctx, types.StringValue("example")),
		},
		"null values remain null": {
			data: &testListModel{
				Name: types.StringNull(),
			},
			expected: expectedTestListModel(ctx, types.StringNull()),
		},
		"unknown values remain unknown": {
			data: &testListModel{
				Name: types.StringUnknown(),
			},
			expected: expectedTestListModel(ctx, types.StringUnknown()),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if diags := setZeroValueAttrFieldsToNull(ctx, test.data); diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			assertTestListModelEqual(t, test.data, test.expected)
		})
	}
}

func TestSetZeroValueAttrFieldsToNull_IgnoresInvalidTargets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := map[string]struct {
		target any
	}{
		"nil pointer": {
			target: (*testListModel)(nil),
		},
		"pointer to non-struct": {
			target: func() any {
				v := 42
				return &v
			}(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if diags := setZeroValueAttrFieldsToNull(ctx, test.target); diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}
		})
	}
}

func TestSetZeroValueAttrFieldsToNull_AdditionalCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := map[string]struct {
		target               any
		assert               func(*testing.T, any)
		expectError          bool
		expectedErrorSummary string
		expectedErrorDetail  string
	}{
		"returns error for non-pointer": {
			target:               testListModel{},
			expectError:          true,
			expectedErrorSummary: "Normalizing List Result",
			expectedErrorDetail:  "target must be a pointer, got framework.testListModel",
		},
		"skips unsettable and non-struct fields": {
			target: &testMixedFieldModel{
				Count:  42,
				hidden: types.String{},
			},
			assert: func(t *testing.T, target any) {
				t.Helper()

				data := target.(*testMixedFieldModel)

				if !data.Name.IsNull() {
					t.Fatalf("expected Name to be normalized to null")
				}
				if got, want := data.Count, 42; got != want {
					t.Fatalf("expected Count to remain %d, got %d", want, got)
				}
				if !data.hidden.Equal(types.String{}) {
					t.Fatalf("expected hidden field to remain zero value")
				}
			},
		},
		"recurses two levels deep": {
			target: &testDeepRootModel{
				Name: types.StringValue("example"),
			},
			assert: func(t *testing.T, target any) {
				t.Helper()

				data := target.(*testDeepRootModel)

				if !data.Name.Equal(types.StringValue("example")) {
					t.Fatalf("expected Name to remain unchanged")
				}
				if !data.Middle.Leaf.Field1.IsNull() {
					t.Fatalf("expected Middle.Leaf.Field1 to be normalized to null")
				}
				if !data.Middle.Leaf.Field2.IsNull() {
					t.Fatalf("expected Middle.Leaf.Field2 to be normalized to null")
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := setZeroValueAttrFieldsToNull(ctx, test.target)

			if test.expectError {
				if !diags.HasError() {
					t.Fatal("expected diagnostics")
				}

				if got, want := diags.ErrorsCount(), 1; got != want {
					t.Fatalf("expected %d errors, got %d", want, got)
				}

				err := diags.Errors()[0]
				if got, want := err.Summary(), test.expectedErrorSummary; got != want {
					t.Fatalf("expected summary %q, got %q", want, got)
				}
				if got, want := err.Detail(), test.expectedErrorDetail; got != want {
					t.Fatalf("expected detail %q, got %q", want, got)
				}

				return
			}

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			test.assert(t, test.target)
		})
	}
}

func expectedTestListModel(ctx context.Context, name types.String) *testListModel {
	return &testListModel{
		Name:   name,
		Tags:   fwtypes.NewMapValueOfNull[types.String](ctx),
		Config: fwtypes.NewObjectValueOfNull[testNestedModel](ctx),
		Rules:  fwtypes.NewListNestedObjectValueOfNull[testNestedModel](ctx),
		Nested: testNestedModel{
			Field1: types.StringNull(),
			Field2: types.Int32Null(),
		},
	}
}

func assertTestListModelEqual(t *testing.T, got, want *testListModel) {
	t.Helper()

	if !got.Name.Equal(want.Name) {
		t.Fatalf("expected Name to equal %s, got %s", want.Name, got.Name)
	}
	if !got.Tags.Equal(want.Tags) {
		t.Fatalf("expected Tags to equal %s, got %s", want.Tags, got.Tags)
	}
	if !got.Config.Equal(want.Config) {
		t.Fatalf("expected Config to equal %s, got %s", want.Config, got.Config)
	}
	if !got.Rules.Equal(want.Rules) {
		t.Fatalf("expected Rules to equal %s, got %s", want.Rules, got.Rules)
	}
	if !got.Nested.Field1.Equal(want.Nested.Field1) {
		t.Fatalf("expected Nested.Field1 to equal %s, got %s", want.Nested.Field1, got.Nested.Field1)
	}
	if !got.Nested.Field2.Equal(want.Nested.Field2) {
		t.Fatalf("expected Nested.Field2 to equal %s, got %s", want.Nested.Field2, got.Nested.Field2)
	}
}
