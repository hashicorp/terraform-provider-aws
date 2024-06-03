// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type ObjectA struct {
	Name types.String `tfsdk:"name"`
}

type ObjectB struct {
	Length types.Int64 `tfsdk:"length"`
}

func TestObjectTypeOfEqual(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := map[string]struct {
		other attr.Type
		want  bool
	}{
		"string type": {
			other: types.StringType,
		},
		"equal type": {
			other: fwtypes.NewObjectTypeOf[ObjectA](ctx),
			want:  true,
		},
		"other struct type": {
			other: fwtypes.NewObjectTypeOf[ObjectB](ctx),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwtypes.NewObjectTypeOf[ObjectA](ctx).Equal(testCase.other)

			if got != testCase.want {
				t.Errorf("got = %v, want = %v", got, testCase.want)
			}
		})
	}
}

func TestObjectTypeOfValueFromTerraform(t *testing.T) {
	t.Parallel()

	objectA := ObjectA{
		Name: types.StringValue("test"),
	}
	objectAType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name": tftypes.String,
		},
	}
	objectAValue := tftypes.NewValue(objectAType, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "test"),
	})
	objectBType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"length": tftypes.Number,
		},
	}
	objectBValue := tftypes.NewValue(objectBType, map[string]tftypes.Value{
		"length": tftypes.NewValue(tftypes.Number, 42),
	})

	ctx := context.Background()
	testCases := map[string]struct {
		tfVal   tftypes.Value
		wantVal attr.Value
		wantErr bool
	}{
		"null value": {
			tfVal:   tftypes.NewValue(objectAType, nil),
			wantVal: fwtypes.NewObjectValueOfNull[ObjectA](ctx),
		},
		"unknown value": {
			tfVal:   tftypes.NewValue(objectAType, tftypes.UnknownValue),
			wantVal: fwtypes.NewObjectValueOfUnknown[ObjectA](ctx),
		},
		"valid value": {
			tfVal:   objectAValue,
			wantVal: fwtypes.NewObjectValueOfMust[ObjectA](ctx, &objectA),
		},
		"invalid Terraform value": {
			tfVal:   objectBValue,
			wantVal: fwtypes.NewObjectValueOfMust[ObjectA](ctx, &objectA),
			wantErr: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotVal, err := fwtypes.NewObjectTypeOf[ObjectA](ctx).ValueFromTerraform(ctx, testCase.tfVal)
			gotErr := err != nil

			if gotErr != testCase.wantErr {
				t.Errorf("gotErr = %v, wantErr = %v", gotErr, testCase.wantErr)
			}

			if gotErr {
				if !testCase.wantErr {
					t.Errorf("err = %q", err)
				}
			} else if diff := cmp.Diff(gotVal, testCase.wantVal); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestObjectValueOfEqual(t *testing.T) {
	t.Parallel()

	objectA := ObjectA{
		Name: types.StringValue("test"),
	}
	objectB := ObjectB{
		Length: types.Int64Value(42),
	}
	objectA2 := ObjectA{
		Name: types.StringValue("test2"),
	}

	ctx := context.Background()
	testCases := map[string]struct {
		other attr.Value
		want  bool
	}{
		"string value": {
			other: types.StringValue("test"),
		},
		"equal value": {
			other: fwtypes.NewObjectValueOfMust(ctx, &objectA),
			want:  true,
		},
		"struct not equal value": {
			other: fwtypes.NewObjectValueOfMust(ctx, &objectA2),
		},
		"other struct value": {
			other: fwtypes.NewObjectValueOfMust(ctx, &objectB),
		},
		"null value": {
			other: fwtypes.NewObjectValueOfNull[ObjectA](ctx),
		},
		"unknown value": {
			other: fwtypes.NewObjectValueOfUnknown[ObjectA](ctx),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwtypes.NewObjectValueOfMust(ctx, &objectA).Equal(testCase.other)

			if got != testCase.want {
				t.Errorf("got = %v, want = %v", got, testCase.want)
			}
		})
	}
}

func TestNullOutObjectPtrFields(t *testing.T) {
	t.Parallel()

	type A struct {
		F1 types.Bool                        `tfsdk:"f1"`
		F2 types.String                      `tfsdk:"f2"`
		F3 fwtypes.ListValueOf[types.String] `tfsdk:"f3"`
		F4 fwtypes.SetValueOf[types.Int64]   `tfsdk:"f4"`
	}

	ctx := context.Background()
	a := new(A)
	a.F1 = types.BoolValue(true)
	a.F2 = types.StringValue("test")
	a.F3 = fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{types.StringValue("test")})
	a.F4 = fwtypes.NewSetValueOfMust[types.Int64](ctx, []attr.Value{types.Int64Value(-1)})
	diags := fwtypes.NullOutObjectPtrFields(ctx, a)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if !a.F1.IsNull() {
		t.Errorf("expected F1 to be null")
	}
	if !a.F2.IsNull() {
		t.Errorf("expected F2 to be null")
	}
	if !a.F3.IsNull() {
		t.Errorf("expected F3 to be null")
	}
	if !a.F4.IsNull() {
		t.Errorf("expected F4 to be null")
	}
}
