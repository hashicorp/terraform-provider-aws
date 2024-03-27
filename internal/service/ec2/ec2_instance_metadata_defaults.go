// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	httpPutResponseHopLimitNoPreference = -1
)

// @FrameworkResource(name="Instance Metadata Defaults")
func newInstanceMetadataDefaultsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &instanceMetadataDefaultsResource{}

	return r, nil
}

type instanceMetadataDefaultsResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (*instanceMetadataDefaultsResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_ec2_instance_metadata_defaults"
}

func (r *instanceMetadataDefaultsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	httpEndpointType := fwtypes.StringEnumType[awstypes.DefaultInstanceMetadataEndpointState]()
	httpTokensType := fwtypes.StringEnumType[awstypes.MetadataDefaultHttpTokensState]()
	instanceMetadataTagsType := fwtypes.StringEnumType[awstypes.DefaultInstanceMetadataTagsState]()

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"http_endpoint": schema.StringAttribute{
				CustomType: httpEndpointType,
				Optional:   true,
				Computed:   true,
				Default:    httpEndpointType.AttributeDefault(awstypes.DefaultInstanceMetadataEndpointStateNoPreference),
			},
			"http_put_response_hop_limit": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(httpPutResponseHopLimitNoPreference),
				Validators: []validator.Int64{
					int64validator.Between(-1, 64),
				},
			},
			"http_tokens": schema.StringAttribute{
				CustomType: httpTokensType,
				Optional:   true,
				Computed:   true,
				Default:    httpTokensType.AttributeDefault(awstypes.MetadataDefaultHttpTokensStateNoPreference),
			},
			names.AttrID: framework.IDAttribute(),
			"instance_metadata_tags": schema.StringAttribute{
				CustomType: instanceMetadataTagsType,
				Optional:   true,
				Computed:   true,
				Default:    instanceMetadataTagsType.AttributeDefault(awstypes.DefaultInstanceMetadataTagsStateNoPreference),
			},
		},
	}
}

func (r *instanceMetadataDefaultsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data instanceMetadataDefaultsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.ModifyInstanceMetadataDefaultsInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.ModifyInstanceMetadataDefaults(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating EC2 Instance Metadata Defaults", err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *instanceMetadataDefaultsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data instanceMetadataDefaultsModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	output, err := findInstanceMetadataDefaults(ctx, conn)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError("reading EC2 Instance Metadata Defaults", err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Update is very similar to Create as AWS has a single API call ModifyInstanceMetadataDefaults
func (r *instanceMetadataDefaultsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new instanceMetadataDefaultsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.ModifyInstanceMetadataDefaultsInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.ModifyInstanceMetadataDefaults(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("updating EC2 Instance Metadata Defaults", err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *instanceMetadataDefaultsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	input := &ec2.ModifyInstanceMetadataDefaultsInput{
		HttpEndpoint:            awstypes.DefaultInstanceMetadataEndpointStateNoPreference,
		HttpPutResponseHopLimit: aws.Int32(httpPutResponseHopLimitNoPreference),
		HttpTokens:              awstypes.MetadataDefaultHttpTokensStateNoPreference,
		InstanceMetadataTags:    awstypes.DefaultInstanceMetadataTagsStateNoPreference,
	}

	_, err := conn.ModifyInstanceMetadataDefaults(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("deleting EC2 Instance Metadata Defaults", err.Error())

		return
	}
}

func findInstanceMetadataDefaults(ctx context.Context, conn *ec2.Client) (*awstypes.InstanceMetadataDefaultsResponse, error) {
	input := &ec2.GetInstanceMetadataDefaultsInput{}

	output, err := conn.GetInstanceMetadataDefaults(ctx, &ec2.GetInstanceMetadataDefaultsInput{})

	if err != nil {
		return nil, err
	}

	if output == nil || output.AccountLevel == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AccountLevel, nil
}

type instanceMetadataDefaultsModel struct {
	HttpEndpoint            types.String `tfsdk:"http_endpoint"`
	HttpPutResponseHopLimit types.Int64  `tfsdk:"http_put_response_hop_limit"`
	HttpTokens              types.String `tfsdk:"http_tokens"`
	InstanceMetadataTags    types.String `tfsdk:"instance_metadata_tags"`
}
