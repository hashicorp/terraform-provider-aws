// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	httpPutResponseHopLimitNoPreference = -1
)

// @FrameworkResource("aws_ec2_instance_metadata_defaults", name="Instance Metadata Defaults")
func newInstanceMetadataDefaultsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &instanceMetadataDefaultsResource{}

	return r, nil
}

type instanceMetadataDefaultsResource struct {
	framework.ResourceWithModel[instanceMetadataDefaultsResourceModel]
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
					int64validator.Any(
						int64validator.Between(1, 64),
						int64validator.OneOf(httpPutResponseHopLimitNoPreference),
					),
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

func (r *instanceMetadataDefaultsResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("http_endpoint"),
			path.MatchRoot("http_put_response_hop_limit"),
			path.MatchRoot("http_tokens"),
			path.MatchRoot("instance_metadata_tags"),
		),
	}
}

func (r *instanceMetadataDefaultsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data instanceMetadataDefaultsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := ec2.ModifyInstanceMetadataDefaultsInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.ModifyInstanceMetadataDefaults(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating EC2 Instance Metadata Defaults", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = types.StringValue(r.Meta().AccountID(ctx))

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *instanceMetadataDefaultsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data instanceMetadataDefaultsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	output, err := findInstanceMetadataDefaults(ctx, conn)

	switch {
	case err == nil && inttypes.IsZero(output):
		err = tfresource.NewEmptyResultError()
		fallthrough
	case retry.NotFound(err):
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	case err != nil:
		response.Diagnostics.AddError("reading EC2 Instance Metadata Defaults", err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Defaults.
	if data.HttpEndpoint.IsNull() {
		data.HttpEndpoint = fwtypes.StringEnumValue(awstypes.DefaultInstanceMetadataEndpointStateNoPreference)
	}
	if data.HttpPutResponseHopLimit.IsNull() || data.HttpPutResponseHopLimit.ValueInt64() == 0 {
		data.HttpPutResponseHopLimit = types.Int64Value(httpPutResponseHopLimitNoPreference)
	}
	if data.HttpTokens.IsNull() {
		data.HttpTokens = fwtypes.StringEnumValue(awstypes.MetadataDefaultHttpTokensStateNoPreference)
	}
	if data.InstanceMetadataTags.IsNull() {
		data.InstanceMetadataTags = fwtypes.StringEnumValue(awstypes.DefaultInstanceMetadataTagsStateNoPreference)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Update is very similar to Create as AWS has a single API call ModifyInstanceMetadataDefaults
func (r *instanceMetadataDefaultsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new instanceMetadataDefaultsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := ec2.ModifyInstanceMetadataDefaultsInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.ModifyInstanceMetadataDefaults(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("updating EC2 Instance Metadata Defaults", err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *instanceMetadataDefaultsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	input := ec2.ModifyInstanceMetadataDefaultsInput{
		HttpEndpoint:            awstypes.DefaultInstanceMetadataEndpointStateNoPreference,
		HttpPutResponseHopLimit: aws.Int32(httpPutResponseHopLimitNoPreference),
		HttpTokens:              awstypes.MetadataDefaultHttpTokensStateNoPreference,
		InstanceMetadataTags:    awstypes.DefaultInstanceMetadataTagsStateNoPreference,
	}

	_, err := conn.ModifyInstanceMetadataDefaults(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("deleting EC2 Instance Metadata Defaults", err.Error())

		return
	}
}

type instanceMetadataDefaultsResourceModel struct {
	framework.WithRegionModel
	HttpEndpoint            fwtypes.StringEnum[awstypes.DefaultInstanceMetadataEndpointState] `tfsdk:"http_endpoint"`
	HttpPutResponseHopLimit types.Int64                                                       `tfsdk:"http_put_response_hop_limit"`
	HttpTokens              fwtypes.StringEnum[awstypes.MetadataDefaultHttpTokensState]       `tfsdk:"http_tokens"`
	ID                      types.String                                                      `tfsdk:"id"`
	InstanceMetadataTags    fwtypes.StringEnum[awstypes.DefaultInstanceMetadataTagsState]     `tfsdk:"instance_metadata_tags"`
}
