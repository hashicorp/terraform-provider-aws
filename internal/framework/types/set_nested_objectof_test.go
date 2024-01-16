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

func TestSetNestedObjectTypeOfEqual(t *testing.T) {
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
			other: fwtypes.NewSetNestedObjectTypeOf[ObjectA](ctx),
			want:  true,
		},
		"other struct type": {
			other: fwtypes.NewSetNestedObjectTypeOf[ObjectB](ctx),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwtypes.NewSetNestedObjectTypeOf[ObjectA](ctx).Equal(testCase.other)

			if got != testCase.want {
				t.Errorf("got = %v, want = %v", got, testCase.want)
			}
		})
	}
}

func TestSetNestedObjectTypeOfValueFromTerraform(t *testing.T) {
	t.Parallel()

	objectA := ObjectA{
		Name: types.StringValue("test"),
	}
	objectAType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name": tftypes.String,
		},
	}
	objectASetType := tftypes.Set{ElementType: objectAType}
	objectAValue := tftypes.NewValue(objectAType, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "test"),
	})
	objectASetValue := tftypes.NewValue(tftypes.Set{ElementType: objectAType}, []tftypes.Value{objectAValue})
	objectBType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"length": tftypes.Number,
		},
	}
	objectBValue := tftypes.NewValue(objectBType, map[string]tftypes.Value{
		"length": tftypes.NewValue(tftypes.Number, 42),
	})
	objectBSetValue := tftypes.NewValue(tftypes.Set{ElementType: objectBType}, []tftypes.Value{objectBValue})

	ctx := context.Background()
	testCases := map[string]struct {
		tfVal   tftypes.Value
		wantVal attr.Value
		wantErr bool
	}{
		"null value": {
			tfVal:   tftypes.NewValue(objectASetType, nil),
			wantVal: fwtypes.NewSetNestedObjectValueOfNull[ObjectA](ctx),
		},
		"unknown value": {
			tfVal:   tftypes.NewValue(objectASetType, tftypes.UnknownValue),
			wantVal: fwtypes.NewSetNestedObjectValueOfUnknown[ObjectA](ctx),
		},
		"valid value": {
			tfVal:   objectASetValue,
			wantVal: fwtypes.NewSetNestedObjectValueOfPtr[ObjectA](ctx, &objectA),
		},
		"invalid Terraform value": {
			tfVal:   objectBSetValue,
			wantVal: fwtypes.NewSetNestedObjectValueOfPtr[ObjectA](ctx, &objectA),
			wantErr: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotVal, err := fwtypes.NewSetNestedObjectTypeOf[ObjectA](ctx).ValueFromTerraform(ctx, testCase.tfVal)
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

func TestSetNestedObjectValueOfEqual(t *testing.T) {
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
			other: fwtypes.NewSetNestedObjectValueOfPtr(ctx, &objectA),
			want:  true,
		},
		"struct not equal value": {
			other: fwtypes.NewSetNestedObjectValueOfPtr(ctx, &objectA2),
		},
		"other struct value": {
			other: fwtypes.NewSetNestedObjectValueOfPtr(ctx, &objectB),
		},
		"null value": {
			other: fwtypes.NewSetNestedObjectValueOfNull[ObjectA](ctx),
		},
		"unknown value": {
			other: fwtypes.NewSetNestedObjectValueOfUnknown[ObjectA](ctx),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwtypes.NewSetNestedObjectValueOfPtr(ctx, &objectA).Equal(testCase.other)

			if got != testCase.want {
				t.Errorf("got = %v, want = %v", got, testCase.want)
			}
		})
	}
}
