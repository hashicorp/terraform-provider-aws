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

func (l *ListResourceWithSDKv2Resource) runResultInterceptors(ctx context.Context, when listresource.When, awsClient *conns.AWSClient, includeResource bool, d *schema.ResourceData, result *list.ListResult) diag.Diagnostics {
	var diags diag.Diagnostics
	params := listresource.InterceptorParamsSDK{
		C:               awsClient,
		IncludeResource: includeResource,
		ResourceData:    d,
		Result:          result,
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

// TODO modify to accept func() as parameter
// will allow to use before interceptors as well
func (l *ListResourceWithSDKv2Resource) SetResult(ctx context.Context, awsClient *conns.AWSClient, includeResource bool, rd *schema.ResourceData, result *list.ListResult) {
	if err := l.runResultInterceptors(ctx, listresource.After, awsClient, includeResource, rd, result); err.HasError() {
		result.Diagnostics.Append(err...)
		return
	}
}
