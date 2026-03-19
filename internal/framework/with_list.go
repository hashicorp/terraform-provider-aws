// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/listresource"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// Lister is an interface for resources that support List operations
type Lister[T listresource.InterceptorParams | listresource.InterceptorParamsSDK] interface {
	AppendResultInterceptor(listresource.ListResultInterceptor[T])
}

var _ Lister[listresource.InterceptorParams] = &WithList{}

// WithList provides common functionality for ListResources
type WithList struct {
	withListResourceConfigSchema
	interceptors []listresource.ListResultInterceptor[listresource.InterceptorParams]
}

type FlattenFunc func()

func (w *WithList) AppendResultInterceptor(interceptor listresource.ListResultInterceptor[listresource.InterceptorParams]) {
	w.interceptors = append(w.interceptors, interceptor)
}

func (w WithList) ResultInterceptors() []listresource.ListResultInterceptor[listresource.InterceptorParams] {
	return w.interceptors
}

func (w *WithList) runResultInterceptors(ctx context.Context, when listresource.When, awsClient *conns.AWSClient, includeResource bool, data any, result *list.ListResult) diag.Diagnostics {
	var diags diag.Diagnostics
	params := listresource.InterceptorParams{
		C:               awsClient,
		IncludeResource: includeResource,
		Data:            data,
		Result:          result,
		When:            when,
	}

	switch when {
	case listresource.Before:
		for interceptor := range slices.Values(w.interceptors) {
			diags.Append(interceptor.Read(ctx, params)...)
		}
	case listresource.After:
		for interceptor := range tfslices.BackwardValues(w.interceptors) {
			diags.Append(interceptor.Read(ctx, params)...)
		}
	}

	return diags
}

func (w *WithList) SetResult(ctx context.Context, awsClient *conns.AWSClient, includeResource bool, data any, result *list.ListResult, f FlattenFunc) {
	var diags diag.Diagnostics

	diags.Append(w.runResultInterceptors(ctx, listresource.Before, awsClient, includeResource, data, result)...)
	if diags.HasError() {
		result.Diagnostics.Append(diags...)
		return
	}

	f()
	if result.Diagnostics.HasError() {
		return
	}

	diags.Append(setZeroValueAttrFieldsToNull(ctx, data)...)
	if diags.HasError() {
		result.Diagnostics.Append(diags...)
		return
	}

	diags.Append(result.Resource.Set(ctx, data)...)
	if diags.HasError() {
		result.Diagnostics.Append(diags...)
		return
	}

	diags.Append(w.runResultInterceptors(ctx, listresource.After, awsClient, includeResource, data, result)...)
	if diags.HasError() {
		result.Diagnostics.Append(diags...)
		return
	}
}

func setZeroValueAttrFieldsToNull(ctx context.Context, target any) diag.Diagnostics {
	var diags diag.Diagnostics

	value := reflect.ValueOf(target)
	if !value.IsValid() || value.Kind() != reflect.Ptr || value.IsNil() {
		return diags
	}

	walkStructSetZeroAttrNull(ctx, value.Elem(), &diags)

	return diags
}

func walkStructSetZeroAttrNull(ctx context.Context, value reflect.Value, diags *diag.Diagnostics) {
	if diags.HasError() || !value.IsValid() || value.Kind() != reflect.Struct {
		return
	}

	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		if !field.CanSet() {
			continue
		}

		if field.Kind() != reflect.Struct {
			continue
		}

		if attrValue, ok := field.Interface().(attr.Value); ok {
			if field.IsZero() {
				nullValue, err := fwtypes.NullValueOf(ctx, attrValue)
				if err != nil {
					diags.AddError("Normalizing List Result", err.Error())
					return
				}

				if nullValue == nil {
					continue
				}

				nullValueReflect := reflect.ValueOf(nullValue)
				switch {
				case nullValueReflect.Type().AssignableTo(field.Type()):
					field.Set(nullValueReflect)
				case nullValueReflect.Type().ConvertibleTo(field.Type()):
					field.Set(nullValueReflect.Convert(field.Type()))
				default:
					diags.AddError("Normalizing List Result", fmt.Sprintf("cannot assign null value of type %T to field type %s", nullValue, field.Type()))
					return
				}
			}

			continue
		}

		walkStructSetZeroAttrNull(ctx, field, diags)
		if diags.HasError() {
			return
		}
	}
}
