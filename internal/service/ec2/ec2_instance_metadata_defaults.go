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

	// Convert what we read from the AWS API to our format
	state = r.instanceMetadataDefaultsToPlan(out)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEC2InstanceMetadataDefaults) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceEC2InstanceMetadataDefaultsData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := r.planToModifyInstanceMetadataDefaultsInput(&plan)
	if err := r.updateDefaultInstanceMetadataDefaults(ctx, in, create.ErrActionCreating); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Update is very similar to Create as AWS has a single API call ModifyInstanceMetadataDefaults
func (r *resourceEC2InstanceMetadataDefaults) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceEC2InstanceMetadataDefaultsData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	in := r.planToModifyInstanceMetadataDefaultsInput(&plan)
	if err := r.updateDefaultInstanceMetadataDefaults(ctx, in, create.ErrActionUpdating); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceEC2InstanceMetadataDefaults) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceEC2InstanceMetadataDefaultsData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	in := &ec2.ModifyInstanceMetadataDefaultsInput{
		HttpEndpoint:            awstypes.DefaultInstanceMetadataEndpointStateNoPreference,
		HttpPutResponseHopLimit: aws.Int32(-1), // -1 means "no preference"
		HttpTokens:              awstypes.MetadataDefaultHttpTokensStateNoPreference,
		InstanceMetadataTags:    awstypes.DefaultInstanceMetadataTagsStateNoPreference,
	}

	if err := r.updateDefaultInstanceMetadataDefaults(ctx, in, create.ErrActionDeleting); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
	}
}

// utility function to avoid duplicating code, since Create, Update and Delete all use the same AWS API call ModifyInstanceMetadataDefaults
func (r *resourceEC2InstanceMetadataDefaults) updateDefaultInstanceMetadataDefaults(ctx context.Context, in *ec2.ModifyInstanceMetadataDefaultsInput, action string) error {
	meta := r.Meta()
	conn := meta.EC2Client(ctx)
	region := meta.Region

	out, err := conn.ModifyInstanceMetadataDefaults(ctx, in)
	if err != nil {
		return errors.New(create.ProblemStandardMessage(names.EC2, action, ResNameInstanceMetadataDefaults, region, err))
	}
	if out == nil || !aws.ToBool(out.Return) {
		return errors.New(create.ProblemStandardMessage(names.EC2, action, ResNameInstanceMetadataDefaults, region, errors.New("empty output")))
	}
	return nil
}

// converts the plan to the input for the ModifyInstanceMetadataDefaults API call
func (r *resourceEC2InstanceMetadataDefaults) planToModifyInstanceMetadataDefaultsInput(plan *resourceEC2InstanceMetadataDefaultsData) *ec2.ModifyInstanceMetadataDefaultsInput {
	in := &ec2.ModifyInstanceMetadataDefaultsInput{}

	// When an attribute is not explicily set, we don't populate it in the ModifyInstanceMetadataDefaultsInput object

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

	return in
}

// converts a result from the AWS API call to the state
func (r *resourceEC2InstanceMetadataDefaults) instanceMetadataDefaultsToPlan(out *ec2.GetInstanceMetadataDefaultsOutput) resourceEC2InstanceMetadataDefaultsData {
	state := resourceEC2InstanceMetadataDefaultsData{}

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

	return state
}
