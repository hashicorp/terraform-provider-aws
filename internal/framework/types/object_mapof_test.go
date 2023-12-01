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

func TestObjectMapTypeOfEqual(t *testing.T) {
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
			other: fwtypes.NewObjectMapTypeOf[ObjectA](ctx),
			want:  true,
		},
		"other struct type": {
			other: fwtypes.NewObjectMapTypeOf[ObjectB](ctx),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwtypes.NewObjectMapTypeOf[ObjectA](ctx).Equal(testCase.other)

			if got != testCase.want {
				t.Errorf("got = %v, want = %v", got, testCase.want)
			}
		})
	}
}

func TestObjectMapTypeOfValueFromTerraform(t *testing.T) {
	t.Parallel()

	objectA := ObjectA{
		Name: types.StringValue("test"),
	}
	objectAType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name": tftypes.String,
		},
	}
	objectAMapType := tftypes.Map{ElementType: objectAType}
	objectAValue := tftypes.NewValue(objectAType, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "test"),
	})
	objectAMapValue := tftypes.NewValue(tftypes.Map{ElementType: objectAType}, map[string]tftypes.Value{"fred": objectAValue})
	objectBType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"length": tftypes.Number,
		},
	}
	objectBValue := tftypes.NewValue(objectBType, map[string]tftypes.Value{
		"length": tftypes.NewValue(tftypes.Number, 42),
	})
	objectBMapValue := tftypes.NewValue(tftypes.Map{ElementType: objectBType}, map[string]tftypes.Value{"fred": objectBValue})

	ctx := context.Background()
	testCases := map[string]struct {
		tfVal   tftypes.Value
		wantVal attr.Value
		wantErr bool
	}{
		"null value": {
			tfVal:   tftypes.NewValue(objectAMapType, nil),
			wantVal: fwtypes.NewObjectMapNullValueMapOf[ObjectA](ctx),
		},
		"unknown value": {
			tfVal:   tftypes.NewValue(objectAMapType, tftypes.UnknownValue),
			wantVal: fwtypes.NewObjectMapUnknownValueMapOf[ObjectA](ctx),
		},
		"valid value": {
			tfVal:   objectAMapValue,
			wantVal: fwtypes.NewObjectMapValueMapOf[ObjectA](ctx, map[string]ObjectA{"fred": objectA}),
		},
		"invalid Terraform value": {
			tfVal:   objectBMapValue,
			wantVal: fwtypes.NewObjectMapValueMapOf[ObjectA](ctx, map[string]ObjectA{"fred": objectA}),
			wantErr: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotVal, err := fwtypes.NewObjectMapTypeOf[ObjectA](ctx).ValueFromTerraform(ctx, testCase.tfVal)
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

func TestObjectMapValueOfEqual(t *testing.T) {
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
			other: fwtypes.NewObjectMapValueMapOf(ctx, map[string]ObjectA{"fred": objectA}),
			want:  true,
		},
		"struct not equal value": {
			other: fwtypes.NewObjectMapValueMapOf(ctx, map[string]ObjectA{"fred": objectA2}),
		},
		"other struct value": {
			other: fwtypes.NewObjectMapValueMapOf(ctx, map[string]ObjectB{"fred": objectB}),
		},
		"null value": {
			other: fwtypes.NewObjectMapNullValueMapOf[ObjectA](ctx),
		},
		"unknown value": {
			other: fwtypes.NewObjectMapUnknownValueMapOf[ObjectA](ctx),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwtypes.NewObjectMapValueMapOf(ctx, map[string]ObjectA{"fred": objectA}).Equal(testCase.other)

			if got != testCase.want {
				t.Errorf("got = %v, want = %v", got, testCase.want)
			}
		})
	}
}
