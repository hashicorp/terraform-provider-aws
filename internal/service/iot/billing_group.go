// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_iot_billing_group", name="Billing Group")
// @Tags(identifierAttribute="arn")
func newResourceBillingGroup(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceBillingGroup{}

	return r, nil
}

const (
	ResNameBillingGroup = "Billing Group"
)

type resourceBillingGroup struct {
	framework.ResourceWithConfigure
}

func (r *resourceBillingGroup) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"metadata": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[metadataModel](ctx),
				ElementType: fwtypes.NewObjectTypeOf[metadataModel](ctx),
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVersion: schema.Int64Attribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrProperties: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[propertiesModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}

	response.Schema = s
}

func (r *resourceBillingGroup) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().IoTClient(ctx)
	var data resourceBillingGroupData
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &iot.CreateBillingGroupInput{
		Tags: getTagsIn(ctx),
	}
	response.Diagnostics.Append(flex.Expand(ctx, data, input, flex.WithFieldNamePrefix("BillingGroup"))...)
	if response.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateBillingGroup(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IoT, create.ErrActionCreating, ResNameBillingGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IoT, create.ErrActionCreating, ResNameBillingGroup, data.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	findOut, err := findBillingGroupByName(ctx, conn, data.Name.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IoT, create.ErrActionCreating, ResNameBillingGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	// To preserve historical behavior of the Plugin-SDKV2 based resource, treat
	// billing group properties with a nested nil description as a null object
	if findOut.BillingGroupProperties != nil && findOut.BillingGroupProperties.BillingGroupDescription == nil {
		findOut.BillingGroupProperties = nil
	}

	// To preserve historical behavior of the Plugin-SDKV2 based resource, the billing
	// group name is copied to the `id` attribute (not the ID generated by AWS)
	data.ID = types.StringValue(data.Name.ValueString())
	response.Diagnostics.Append(flex.Flatten(ctx, findOut, &data, flex.WithFieldNamePrefix("BillingGroup"))...)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceBillingGroup) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().IoTClient(ctx)

	var data resourceBillingGroupData
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	out, err := findBillingGroupByName(ctx, conn, data.ID.ValueString())
	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IoT, create.ErrActionReading, ResNameBillingGroup, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	// To preserve historical behavior of the Plugin-SDKV2 based resource, treat
	// billing group properties with a nested nil description as a null object
	if out.BillingGroupProperties != nil && out.BillingGroupProperties.BillingGroupDescription == nil {
		out.BillingGroupProperties = nil
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("BillingGroup"))...)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceBillingGroup) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().IoTClient(ctx)

	var old, new resourceBillingGroupData
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	// For tag-only updates the version value needs to be copied
	// from state into the proposed plan. This will be overwritten
	// by flex.Flatten below if the properties argument changed.
	new.Version = old.Version
	if !old.Properties.Equal(new.Properties) {
		input := &iot.UpdateBillingGroupInput{}
		response.Diagnostics.Append(flex.Expand(ctx, new, input, flex.WithFieldNamePrefix("BillingGroup"))...)
		if response.Diagnostics.HasError() {
			return
		}

		input.ExpectedVersion = old.Version.ValueInt64Pointer()

		out, err := conn.UpdateBillingGroup(ctx, input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IoT, create.ErrActionUpdating, ResNameBillingGroup, new.Name.String(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(flex.Flatten(ctx, out, &new)...)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceBillingGroup) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().IoTClient(ctx)

	var data resourceBillingGroupData
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteBillingGroup(ctx, &iot.DeleteBillingGroupInput{
		BillingGroupName: data.Name.ValueStringPointer(),
		ExpectedVersion:  data.Version.ValueInt64Pointer(),
	})

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IoT, create.ErrActionDeleting, ResNameBillingGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceBillingGroup) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
}

func findBillingGroupByName(ctx context.Context, conn *iot.Client, name string) (*iot.DescribeBillingGroupOutput, error) {
	input := &iot.DescribeBillingGroupInput{
		BillingGroupName: aws.String(name),
	}

	output, err := conn.DescribeBillingGroup(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type resourceBillingGroupData struct {
	ARN        types.String                                     `tfsdk:"arn"`
	ID         types.String                                     `tfsdk:"id" autoflex:",noflatten"`
	Metadata   fwtypes.ListNestedObjectValueOf[metadataModel]   `tfsdk:"metadata"`
	Name       types.String                                     `tfsdk:"name"`
	Tags       tftags.Map                                       `tfsdk:"tags"`
	TagsAll    tftags.Map                                       `tfsdk:"tags_all"`
	Version    types.Int64                                      `tfsdk:"version"`
	Properties fwtypes.ListNestedObjectValueOf[propertiesModel] `tfsdk:"properties"`
}

type propertiesModel struct {
	Description types.String `tfsdk:"description"`
}

type metadataModel struct {
	CreationDate timetypes.RFC3339 `tfsdk:"creation_date"`
}
