// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Instance Metadata Defaults")
func newResourceInstanceMetadataDefaults(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEC2InstanceMetadataDefaults{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameInstanceMetadataDefaults = "EC2 Instance Metadata Defaults"
)

type resourceEC2InstanceMetadataDefaults struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

type resourceEC2InstanceMetadataDefaultsData struct {
	HttpEndpoint            types.String `tfsdk:"http_endpoint"`
	HttpPutResponseHopLimit types.Int64  `tfsdk:"http_put_response_hop_limit"`
	HttpTokens              types.String `tfsdk:"http_tokens"`
	InstanceMetadataTags    types.String `tfsdk:"instance_metadata_tags"`
}

func (r *resourceEC2InstanceMetadataDefaults) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_ec2_instance_metadata_defaults"
}

func (r *resourceEC2InstanceMetadataDefaults) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"http_endpoint": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.DefaultInstanceMetadataEndpointState](),
				},
			},
			"http_put_response_hop_limit": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(-1, 64),
				},
			},
			"http_tokens": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.MetadataDefaultHttpTokensState](),
				},
			},
			"instance_metadata_tags": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.DefaultInstanceMetadataTagsState](),
				},
			},
		},
		Blocks: map[string]schema.Block{},
	}
}

func (r *resourceEC2InstanceMetadataDefaults) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	meta := r.Meta()
	conn := meta.EC2Client(ctx)

	var plan resourceEC2InstanceMetadataDefaultsData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ec2.ModifyInstanceMetadataDefaultsInput{}

	if !plan.HttpEndpoint.IsNull() {
		in.HttpEndpoint = awstypes.DefaultInstanceMetadataEndpointState(plan.HttpEndpoint.ValueString())
	}
	if !plan.HttpPutResponseHopLimit.IsNull() {
		// When the value is -1, it means "no preference"
		// In which case we don't set the argument and leave it to null
		if httpPutResponseHopLimit := int32(plan.HttpPutResponseHopLimit.ValueInt64()); httpPutResponseHopLimit != -1 {
			in.HttpPutResponseHopLimit = aws.Int32(httpPutResponseHopLimit)
		}
	}
	if !plan.HttpTokens.IsNull() {
		in.HttpTokens = awstypes.MetadataDefaultHttpTokensState(plan.HttpTokens.ValueString())
	}
	if !plan.InstanceMetadataTags.IsNull() {
		in.InstanceMetadataTags = awstypes.DefaultInstanceMetadataTagsState(plan.InstanceMetadataTags.ValueString())
	}

	out, err := conn.ModifyInstanceMetadataDefaults(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameInstanceMetadataDefaults, meta.Region, err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Return == nil || *out.Return == false {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameInstanceMetadataDefaults, meta.Region, nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceEC2InstanceMetadataDefaults) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	meta := r.Meta()
	conn := meta.EC2Client(ctx)

	var state resourceEC2InstanceMetadataDefaultsData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.GetInstanceMetadataDefaults(ctx, &ec2.GetInstanceMetadataDefaultsInput{})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNameInstanceMetadataDefaults, meta.Region, err),
			err.Error(),
		)
		return
	}

	if out.AccountLevel.HttpEndpoint != "" {
		state.HttpEndpoint = types.StringValue(string(out.AccountLevel.HttpEndpoint))
	}

	if out.AccountLevel.HttpPutResponseHopLimit == nil {
		state.HttpPutResponseHopLimit = types.Int64Value(-1)
	} else {
		state.HttpPutResponseHopLimit = types.Int64Value(int64(*out.AccountLevel.HttpPutResponseHopLimit))
	}

	if out.AccountLevel.HttpTokens != "" {
		state.HttpTokens = types.StringValue(string(out.AccountLevel.HttpTokens))
	}

	if out.AccountLevel.InstanceMetadataTags != "" {
		state.InstanceMetadataTags = types.StringValue(string(out.AccountLevel.InstanceMetadataTags))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEC2InstanceMetadataDefaults) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceEC2InstanceMetadataDefaultsData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// TODO: deduplicate
	meta := r.Meta()
	conn := meta.EC2Client(ctx)

	in := &ec2.ModifyInstanceMetadataDefaultsInput{}

	if !plan.HttpEndpoint.IsNull() {
		in.HttpEndpoint = awstypes.DefaultInstanceMetadataEndpointState(plan.HttpEndpoint.ValueString())
	}

	if plan.HttpPutResponseHopLimit.IsNull() {
		in.HttpPutResponseHopLimit = aws.Int32(-1)
	} else {
		in.HttpPutResponseHopLimit = aws.Int32(int32(plan.HttpPutResponseHopLimit.ValueInt64()))
	}

	if !plan.HttpTokens.IsNull() {
		in.HttpTokens = awstypes.MetadataDefaultHttpTokensState(plan.HttpTokens.ValueString())
	}

	if !plan.InstanceMetadataTags.IsNull() {
		in.InstanceMetadataTags = awstypes.DefaultInstanceMetadataTagsState(plan.InstanceMetadataTags.ValueString())
	}

	out, err := conn.ModifyInstanceMetadataDefaults(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameInstanceMetadataDefaults, meta.Region, err),
			err.Error(),
		)
		return
	}
	if out == nil || aws.ToBool(out.Return) == false {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameInstanceMetadataDefaults, meta.Region, nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// END todo
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceEC2InstanceMetadataDefaults) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	meta := r.Meta()
	conn := meta.EC2Client(ctx)

	var state resourceEC2InstanceMetadataDefaultsData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	out, err := conn.ModifyInstanceMetadataDefaults(ctx, &ec2.ModifyInstanceMetadataDefaultsInput{
		HttpEndpoint:            awstypes.DefaultInstanceMetadataEndpointStateNoPreference,
		HttpPutResponseHopLimit: aws.Int32(-1), // -1 means "no preference"
		HttpTokens:              awstypes.MetadataDefaultHttpTokensStateNoPreference,
		InstanceMetadataTags:    awstypes.DefaultInstanceMetadataTagsStateNoPreference,
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameInstanceMetadataDefaults, meta.Region, err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Return == nil || *out.Return == false {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameInstanceMetadataDefaults, meta.Region, nil),
			errors.New("empty output").Error(),
		)
		return
	}
}
