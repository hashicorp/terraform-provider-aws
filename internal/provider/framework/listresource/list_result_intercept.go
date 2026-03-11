// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package listresource

import (
	"context"
	"fmt"
	"reflect"
	"unique"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/interceptors"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// when represents the point in the CRUD request lifecycle that an interceptor is run.
// Multiple values can be ORed together.
type when uint16

type When = when

const (
	Before  when = 1 << iota // Interceptor is invoked before call to method in schema
	After                    // Interceptor is invoked after successful call to method in schema
	OnError                  // Interceptor is invoked after unsuccessful call to method in schema
	Finally                  // Interceptor is invoked after After or OnError
)

type InterceptorParams struct {
	C               *conns.AWSClient
	IncludeResource bool
	Data            any
	Result          *list.ListResult
	When            when
}

type ListResultInterceptor[T InterceptorParams | InterceptorParamsSDK] interface {
	Read(ctx context.Context, params T) diag.Diagnostics
}

// TODO: this could be unique as well
type tagsInterceptor struct {
	interceptors.HTags
}

func TagsInterceptor(tags unique.Handle[inttypes.ServicePackageResourceTags]) tagsInterceptor {
	return tagsInterceptor{
		HTags: interceptors.HTags(tags),
	}
}

// Copied from tagsResourceInterceptor.read()
func (r tagsInterceptor) Read(ctx context.Context, params InterceptorParams) diag.Diagnostics {
	var diags diag.Diagnostics

	if !params.IncludeResource {
		return diags
	}

	sp, serviceName, resourceName, _, tagsInContext, ok := interceptors.InfoFromContext(ctx, params.C)
	if !ok {
		return diags
	}

	switch params.When {
	case After:
		// If the R handler didn't set tags, try and read them from the service API.
		if tagsInContext.TagsOut.IsNone() {
			// Some old resources may not have the required attribute set after Read:
			// https://github.com/hashicorp/terraform-provider-aws/issues/31180
			if identifier := r.GetIdentifierFramework(ctx, params.Result.Resource); identifier != "" {
				if err := r.ListTags(ctx, sp, params.C, identifier); err != nil {
					diags.AddError(fmt.Sprintf("listing tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())

					return diags
				}
			}
		}

		apiTags := tagsInContext.TagsOut.UnwrapOrDefault()

		// AWS APIs often return empty lists of tags when none have been configured.
		var stateTags tftags.Map
		params.Result.Resource.GetAttribute(ctx, path.Root(names.AttrTags), &stateTags)
		// Remove any provider configured ignore_tags and system tags from those returned from the service API.
		// The resource's configured tags do not include any provider configured default_tags.
		if v := apiTags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(params.C.IgnoreTagsConfig(ctx)).ResolveDuplicatesFramework(ctx, params.C.DefaultTagsConfig(ctx), params.C.IgnoreTagsConfig(ctx), stateTags, &diags).Map(); len(v) > 0 {
			stateTags = tftags.NewMapFromMapValue(fwflex.FlattenFrameworkStringValueMapLegacy(ctx, v))
		}
		diags.Append(params.Result.Resource.SetAttribute(ctx, path.Root(names.AttrTags), &stateTags)...)
		if diags.HasError() {
			return diags
		}

		// Computed tags_all do.
		stateTagsAll := fwflex.FlattenFrameworkStringValueMapLegacy(ctx, apiTags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(params.C.IgnoreTagsConfig(ctx)).Map())
		diags.Append(params.Result.Resource.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.NewMapFromMapValue(stateTagsAll))...)
		if diags.HasError() {
			return diags
		}
	}

	tagsInContext.TagsOut = nil

	return diags
}

type identityInterceptor struct {
	attributes []inttypes.IdentityAttribute
}

func IdentityInterceptor(attributes []inttypes.IdentityAttribute) identityInterceptor {
	return identityInterceptor{
		attributes: attributes,
	}
}

func (r identityInterceptor) Read(ctx context.Context, params InterceptorParams) diag.Diagnostics {
	var diags diag.Diagnostics

	awsClient := params.C

	switch params.When {
	// The Before step is not needed if Framework pre-populates the Identity as it does with CRUD operations
	case Before:
		identityType := params.Result.Identity.Schema.Type()

		obj, d := newEmptyObject(identityType)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		diags.Append(params.Result.Identity.Set(ctx, obj)...)
		if diags.HasError() {
			return diags
		}

	case After:
		for _, att := range r.attributes {
			switch att.Name() {
			case names.AttrAccountID:
				diags.Append(params.Result.Identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.AccountID(ctx))...)
				if diags.HasError() {
					return diags
				}

			case names.AttrRegion:
				diags.Append(params.Result.Identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.Region(ctx))...)
				if diags.HasError() {
					return diags
				}

			default:
				var attrVal attr.Value
				diags.Append(params.Result.Resource.GetAttribute(ctx, path.Root(att.ResourceAttributeName()), &attrVal)...)
				if diags.HasError() {
					return diags
				}

				diags.Append(params.Result.Identity.SetAttribute(ctx, path.Root(att.Name()), attrVal)...)
				if diags.HasError() {
					return diags
				}
			}
		}
	}

	return diags
}

func newEmptyObject(typ attr.Type) (obj basetypes.ObjectValue, diags diag.Diagnostics) {
	i, ok := typ.(attr.TypeWithAttributeTypes)
	if !ok {
		diags.AddError(
			"Internal Error",
			"An unexpected error occurred. "+
				"This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Expected value type to implement attr.TypeWithAttributeTypes, got: %T", typ),
		)
		return
	}

	attrTypes := i.AttributeTypes()
	attrValues := make(map[string]attr.Value, len(attrTypes))
	// TODO: only handles string types
	for attrName := range attrTypes {
		attrValues[attrName] = types.StringNull()
	}
	obj, d := basetypes.NewObjectValue(attrTypes, attrValues)
	diags.Append(d...)
	if d.HasError() {
		return basetypes.ObjectValue{}, diags
	}

	return obj, diags
}

type setRegionInterceptor struct{}

func SetRegionInterceptor() setRegionInterceptor {
	return setRegionInterceptor{}
}

// Copied from resourceSetRegionInStateInterceptor.read()
func (r setRegionInterceptor) Read(ctx context.Context, params InterceptorParams) diag.Diagnostics {
	var diags diag.Diagnostics

	switch params.When {
	case After:
		diags.Append(params.Result.Resource.SetAttribute(ctx, path.Root(names.AttrRegion), params.C.Region(ctx))...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

type defaultObjectInterceptor struct{}

func DefaultObjectInterceptor() defaultObjectInterceptor {
	return defaultObjectInterceptor{}
}

func (r defaultObjectInterceptor) Read(ctx context.Context, params InterceptorParams) diag.Diagnostics {
	var diags diag.Diagnostics
	switch params.When {
	case Before:
		if reflect.ValueOf(params.Data).Kind() != reflect.Ptr {
			diags.AddError(
				"Internal Error",
				"data object must be a pointer")
			return diags
		}

		objData := dereferencePointer(reflect.ValueOf(params.Data))

		for tfName, fieldName := range tfFieldToStructFieldMap() {
			field := objData.FieldByName(fieldName)
			if !field.IsValid() {
				continue
			}

			if !implementsAttrValue(field) {
				diags.AddError(
					"Internal Error",
					"An unexpected error occurred. "+
						"This is always an error in the provider. "+
						"Please report the following to the provider developer:\n\n"+
						fmt.Sprintf("Expected field %s to implement attr.Value, got: %T", fieldName, objData.FieldByName(fieldName).Interface()),
				)
				return diags
			}

			switch field.Interface().(attr.Value).Type(ctx).(type) {
			case basetypes.MapTypable:
				if field.Type() == reflect.TypeFor[tftags.Map]() {
					field.Set(reflect.ValueOf(tftags.NewMapValueNull()))
				}
			case basetypes.ObjectTypable:
				if field.Type() == reflect.TypeFor[timeouts.Value]() {
					timeoutsType, d := params.Result.Resource.Schema.TypeAtPath(ctx, path.Root(tfName))
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}
					nullObj, d := newNullObject(timeoutsType)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					t := timeouts.Value{}
					t.Object = nullObj
					field.Set(reflect.ValueOf(t))
				}
			}
		}

		return diags
	}

	return diags
}

func dereferencePointer(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Ptr {
		return value.Elem()
	}
	return value
}

func implementsAttrValue(field reflect.Value) bool {
	return field.Type().Implements(reflect.TypeFor[attr.Value]())
}

func tfFieldToStructFieldMap() map[string]string {
	return map[string]string{
		names.AttrTags:     "Tags",
		names.AttrTagsAll:  "TagsAll",
		names.AttrTimeouts: "Timeouts",
	}
}

func newNullObject(typ attr.Type) (obj basetypes.ObjectValue, diags diag.Diagnostics) {
	i, ok := typ.(attr.TypeWithAttributeTypes)
	if !ok {
		diags.AddError(
			"Internal Error",
			"An unexpected error occurred. "+
				"This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Expected value type to implement attr.TypeWithAttributeTypes, got: %T", typ),
		)
		return
	}

	attrTypes := i.AttributeTypes()

	obj = basetypes.NewObjectNull(attrTypes)

	return obj, diags
}

type InterceptorParamsSDK struct {
	C               *conns.AWSClient
	IncludeResource bool
	ResourceData    *schema.ResourceData
	When            when
}

type tagsInterceptorSDK struct {
	interceptors.HTags
}

func TagsInterceptorSDK(tags unique.Handle[inttypes.ServicePackageResourceTags]) tagsInterceptorSDK {
	return tagsInterceptorSDK{
		HTags: interceptors.HTags(tags),
	}
}

func (r tagsInterceptorSDK) Read(ctx context.Context, params InterceptorParamsSDK) diag.Diagnostics {
	var diags diag.Diagnostics

	if !params.IncludeResource {
		return diags
	}

	sp, _, _, _, tagsInContext, ok := interceptors.InfoFromContext(ctx, params.C)
	if !ok {
		return diags
	}

	switch params.When {
	case After:
		// If the R handler didn't set tags, try and read them from the service API.
		if tagsInContext.TagsOut.IsNone() {
			// Some old resources may not have the required attribute set after Read:
			// https://github.com/hashicorp/terraform-provider-aws/issues/31180
			if identifier := r.GetIdentifierSDKv2(ctx, params.ResourceData); identifier != "" {
				if err := r.ListTags(ctx, sp, params.C, identifier); err != nil {
					diags.Append(diag.NewErrorDiagnostic(
						"Error Listing Tags",
						fmt.Sprintf("An error occurred while listing tags for %s: %s", sp.ServicePackageName(), err.Error()),
					))
					return diags
				}
			}
		}

		// Remove any provider configured ignore_tags and system tags from those returned from the service API.
		tags := tagsInContext.TagsOut.UnwrapOrDefault().IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(params.C.IgnoreTagsConfig(ctx))

		// The resource's configured tags can now include duplicate tags that have been configured on the provider.
		if err := params.ResourceData.Set(names.AttrTags, tags.ResolveDuplicates(ctx, params.C.DefaultTagsConfig(ctx), params.C.IgnoreTagsConfig(ctx), params.ResourceData, names.AttrTags, nil).Map()); err != nil {
			diags.Append(diag.NewErrorDiagnostic(
				"Error Setting Tags",
				fmt.Sprintf("An error occurred while listing tags for %s: %s", sp.ServicePackageName(), err.Error()),
			))
			return diags
		}

		// Computed tags_all do.
		if err := params.ResourceData.Set(names.AttrTagsAll, tags.Map()); err != nil {
			diags.Append(diag.NewErrorDiagnostic(
				"Error Listing TagsAll",
				fmt.Sprintf("An error occurred while listing tags for %s: %s", sp.ServicePackageName(), err.Error()),
			))
			return diags
		}

		// reset tags in context for next resource
		tagsInContext.TagsOut = nil

		return diags
	}

	return diags
}
