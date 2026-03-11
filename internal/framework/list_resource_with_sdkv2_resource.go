// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"slices"
	"unique"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/listresource"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	tfunique "github.com/hashicorp/terraform-provider-aws/internal/unique"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type WithRegionSpec interface {
	SetRegionSpec(regionSpec unique.Handle[inttypes.ServicePackageResourceRegion])
}

var _ Lister[listresource.InterceptorParamsSDK] = &ListResourceWithSDKv2Resource{}

type ListResourceWithSDKv2Resource struct {
	withListResourceConfigSchema
	ResourceWithConfigure
	resourceSchema *schema.Resource
	identitySpec   inttypes.Identity
	identitySchema *schema.ResourceIdentity
	regionSpec     unique.Handle[inttypes.ServicePackageResourceRegion]
	interceptors   []listresource.ListResultInterceptor[listresource.InterceptorParamsSDK]
}

func (l *ListResourceWithSDKv2Resource) AppendResultInterceptor(interceptor listresource.ListResultInterceptor[listresource.InterceptorParamsSDK]) {
	l.interceptors = append(l.interceptors, interceptor)
}

func (l *ListResourceWithSDKv2Resource) SetRegionSpec(regionSpec unique.Handle[inttypes.ServicePackageResourceRegion]) {
	l.regionSpec = regionSpec

	var isRegionOverrideEnabled bool
	if !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		isRegionOverrideEnabled = true
	}

	if isRegionOverrideEnabled {
		if _, ok := l.resourceSchema.SchemaMap()[names.AttrRegion]; !ok {
			// TODO: Use standard shared `region` attribute
			l.resourceSchema.SchemaMap()[names.AttrRegion] = &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			}
		}
	}
}

func identityAttributeSchemaType(it inttypes.IdentityType) schema.ValueType {
	switch it {
	case inttypes.BoolIdentityType:
		return schema.TypeBool
	case inttypes.IntIdentityType, inttypes.Int64IdentityType:
		return schema.TypeInt
	case inttypes.FloatIdentityType, inttypes.Float64IdentityType:
		return schema.TypeFloat
	default:
		return schema.TypeString
	}
}

func (l *ListResourceWithSDKv2Resource) SetIdentitySpec(identitySpec inttypes.Identity) {
	out := make(map[string]*schema.Schema)
	for _, v := range identitySpec.Attributes {
		out[v.Name()] = &schema.Schema{
			Type: identityAttributeSchemaType(v.IdentityType()),
		}
		if v.Required() {
			out[v.Name()].Required = true
		} else {
			out[v.Name()].Optional = true
		}
	}

	identitySchema := schema.ResourceIdentity{
		SchemaFunc: func() map[string]*schema.Schema {
			return out
		},
	}

	l.identitySchema = &identitySchema
	l.resourceSchema.Identity = &identitySchema
	l.identitySpec = identitySpec
}

func (l *ListResourceWithSDKv2Resource) runResultInterceptors(ctx context.Context, when listresource.When, awsClient *conns.AWSClient, d *schema.ResourceData, includeResource bool) diag.Diagnostics {
	var diags diag.Diagnostics
	params := listresource.InterceptorParamsSDK{
		C:               awsClient,
		IncludeResource: includeResource,
		ResourceData:    d,
		When:            when,
	}

	switch when {
	case listresource.Before:
		for interceptor := range slices.Values(l.interceptors) {
			diags.Append(interceptor.Read(ctx, params)...)
			if diags.HasError() {
				return diags
			}
		}
	case listresource.After:
		for interceptor := range tfslices.BackwardValues(l.interceptors) {
			diags.Append(interceptor.Read(ctx, params)...)
			if diags.HasError() {
				return diags
			}
		}
	}

	return diags
}

func (l *ListResourceWithSDKv2Resource) RawV5Schemas(ctx context.Context, _ list.RawV5SchemaRequest, response *list.RawV5SchemaResponse) {
	response.ProtoV5Schema = l.resourceSchema.ProtoSchema(ctx)()
	response.ProtoV5IdentitySchema = l.resourceSchema.ProtoIdentitySchema(ctx)()
}

func (l *ListResourceWithSDKv2Resource) SetResourceSchema(resource *schema.Resource) {
	l.resourceSchema = resource
}

func (l *ListResourceWithSDKv2Resource) ResourceData() *schema.ResourceData {
	return l.resourceSchema.Data(&terraform.InstanceState{})
}

func (l *ListResourceWithSDKv2Resource) setResourceIdentity(ctx context.Context, client *conns.AWSClient, d *schema.ResourceData) error {
	identity, err := d.Identity()
	if err != nil {
		return err
	}

	for _, attr := range l.identitySpec.Attributes {
		switch attr.Name() {
		case names.AttrAccountID:
			if err := identity.Set(attr.Name(), client.AccountID(ctx)); err != nil {
				return err
			}

		case names.AttrRegion:
			if err := identity.Set(attr.Name(), client.Region(ctx)); err != nil {
				return err
			}

		default:
			val, ok := getAttributeOk(d, attr.ResourceAttributeName())
			if !ok {
				continue
			}
			if err := identity.Set(attr.Name(), val); err != nil {
				return err
			}
		}
	}

	return nil
}

type resourceData interface {
	Id() string
	GetOk(string) (any, bool)
}

func getAttributeOk(d resourceData, name string) (any, bool) {
	if name == "id" {
		return d.Id(), true
	}
	if v, ok := d.GetOk(name); !ok {
		return nil, false
	} else {
		return v, true
	}
}

// TODO modify to accept func() as parameter
// will allow to use before interceptors as well
func (l *ListResourceWithSDKv2Resource) SetResult(ctx context.Context, awsClient *conns.AWSClient, includeResource bool, result *list.ListResult, rd *schema.ResourceData) {
	if err := l.runResultInterceptors(ctx, listresource.After, awsClient, rd, includeResource); err.HasError() {
		result.Diagnostics.Append(err...)
		return
	}

	err := l.setResourceIdentity(ctx, awsClient, rd)
	if err != nil {
		result.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Error Listing Remote Resources",
			"An unexpected error occurred setting resource identity. "+
				"This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+
				"Error: "+err.Error(),
		))
		return
	}

	tfTypeIdentity, err := rd.TfTypeIdentityState()
	if err != nil {
		result.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Error Listing Remote Resources",
			"An unexpected error occurred converting identity state. "+
				"This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+
				"Error: "+err.Error(),
		))
		return
	}

	result.Diagnostics.Append(result.Identity.Set(ctx, *tfTypeIdentity)...)
	if result.Diagnostics.HasError() {
		return
	}

	if includeResource {
		if !tfunique.IsHandleNil(l.regionSpec) && l.regionSpec.Value().IsOverrideEnabled {
			if err := rd.Set(names.AttrRegion, awsClient.Region(ctx)); err != nil {
				result.Diagnostics.Append(diag.NewErrorDiagnostic(
					"Error Listing Remote Resources",
					"An unexpected error occurred. "+
						"This is always an error in the provider. "+
						"Please report the following to the provider developer:\n\n"+
						"Error: "+err.Error(),
				))
				return
			}
		}

		tfTypeResource, err := rd.TfTypeResourceState()
		if err != nil {
			result.Diagnostics.Append(diag.NewErrorDiagnostic(
				"Error Listing Remote Resources",
				"An unexpected error occurred converting resource state. "+
					"This is always an error in the provider. "+
					"Please report the following to the provider developer:\n\n"+
					"Error: "+err.Error(),
			))
			return
		}

		result.Diagnostics.Append(result.Resource.Set(ctx, *tfTypeResource)...)
		if result.Diagnostics.HasError() {
			return
		}
	}
}
