// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	awstypes "github.com/aws/aws-sdk-go-v2/service/location/types"
	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Key")
// @Tags(identifierAttribute="key_arn")
func newResourceKey(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceKey{}, nil
}

const (
	ResNameKey = "Key"
)

type resourceKey struct {
	framework.ResourceWithConfigure
}

func (r *resourceKey) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_location_key"
}

func (r *resourceKey) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"create_time": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1000),
				},
				Default: stringdefault.StaticString("Managed By Terraform"),
			},
			"expire_time": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("no_expiry")),
				},
			},
			"force_update": schema.BoolAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
			"key": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"no_expiry": schema.BoolAttribute{
				Optional: true,
				Validators: []validator.Bool{
					boolvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("expire_time"),
					),
				},
			},
			"update_time": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"restrictions": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[restrictionsData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"allow_actions": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
						"allow_referers": schema.ListAttribute{
							Computed: true,
							Default: listdefault.StaticValue(
								types.ListValueMust(types.StringType, []attr.Value{types.StringValue("*")}),
							),
							ElementType: types.StringType,
							Optional:    true,
						},
						"allow_resources": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *resourceKey) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LocationClient(ctx)

	var plan resourceKeyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &location.CreateKeyInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 'ExpireTime' must not be set when 'NoExpiry' has value true.
	if !plan.NoExpiry.IsNull() {
		in.NoExpiry = aws.Bool(plan.NoExpiry.ValueBool())
	} else if !plan.ExpireTime.IsNull() {
		expireTime, _ := time.Parse(time.RFC3339, plan.ExpireTime.ValueString())
		in.ExpireTime = aws.Time(expireTime)
	}
	in.Tags = getTagsInV2(ctx)

	out, err := conn.CreateKey(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Location, create.ErrActionCreating, ResNameKey, plan.KeyName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Location, create.ErrActionCreating, ResNameKey, plan.KeyName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.KeyName)
	plan.KeyName = flex.StringToFramework(ctx, out.KeyName)

	// Read after create to get computed attributes omitted from the create response
	readOut, err := findKeyByID(ctx, conn, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Location, create.ErrActionSetting, ResNameKey, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	plan.CreateTime = flex.StringValueToFramework(ctx, readOut.CreateTime.Format(time.RFC3339))
	plan.Key = flex.StringToFramework(ctx, readOut.Key)
	plan.KeyARN = flex.StringToFrameworkARN(ctx, readOut.KeyArn)
	plan.UpdateTime = flex.StringValueToFramework(ctx, readOut.UpdateTime.Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceKey) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LocationClient(ctx)

	var state resourceKeyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findKeyByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Location, create.ErrActionSetting, ResNameKey, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.CreateTime = flex.StringValueToFramework(ctx, out.CreateTime.Format(time.RFC3339))
	state.UpdateTime = flex.StringValueToFramework(ctx, out.UpdateTime.Format(time.RFC3339))
	if out.ExpireTime == nil {
		state.NoExpiry = flex.BoolToFramework(ctx, ptr.Bool(true))
	} else {
		state.ExpireTime = flex.StringValueToFramework(ctx, out.ExpireTime.Format(time.RFC3339))
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceKey) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LocationClient(ctx)

	var plan, state resourceKeyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) ||
		!plan.ExpireTime.Equal(state.ExpireTime) ||
		!plan.NoExpiry.Equal(state.NoExpiry) ||
		!plan.Restrictions.Equal(state.Restrictions) {

		in := &location.UpdateKeyInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// 'ExpireTime' must not be set when 'NoExpiry' has value true.
		if !plan.NoExpiry.IsNull() {
			in.NoExpiry = aws.Bool(plan.NoExpiry.ValueBool())
		} else if !plan.ExpireTime.IsNull() {
			expireTime, _ := time.Parse(time.RFC3339, plan.ExpireTime.ValueString())
			in.ExpireTime = aws.Time(expireTime)
		}

		out, err := conn.UpdateKey(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Location, create.ErrActionUpdating, ResNameKey, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Location, create.ErrActionUpdating, ResNameKey, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceKey) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LocationClient(ctx)

	var state resourceKeyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &location.DeleteKeyInput{
		KeyName: aws.String(state.KeyName.ValueString()),
	}

	_, err := conn.DeleteKey(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Location, create.ErrActionDeleting, ResNameKey, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceKey) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceKey) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func findKeyByID(ctx context.Context, conn *location.Client, id string) (*location.DescribeKeyOutput, error) {
	in := &location.DescribeKeyInput{
		KeyName: aws.String(id),
	}

	out, err := conn.DescribeKey(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceKeyData struct {
	CreateTime   types.String                                      `tfsdk:"create_time"`
	Description  types.String                                      `tfsdk:"description"`
	ExpireTime   types.String                                      `tfsdk:"expire_time"`
	ForceUpdate  types.Bool                                        `tfsdk:"force_update"`
	ID           types.String                                      `tfsdk:"id"`
	Key          types.String                                      `tfsdk:"key"`
	KeyARN       fwtypes.ARN                                       `tfsdk:"key_arn"`
	KeyName      types.String                                      `tfsdk:"key_name"`
	NoExpiry     types.Bool                                        `tfsdk:"no_expiry"`
	Restrictions fwtypes.ListNestedObjectValueOf[restrictionsData] `tfsdk:"restrictions"`
	Tags         types.Map                                         `tfsdk:"tags"`
	TagsAll      types.Map                                         `tfsdk:"tags_all"`
	UpdateTime   types.String                                      `tfsdk:"update_time"`
}

type restrictionsData struct {
	AllowActions   fwtypes.ListValueOf[types.String] `tfsdk:"allow_actions"`
	AllowReferers  fwtypes.ListValueOf[types.String] `tfsdk:"allow_referers"`
	AllowResources fwtypes.ListValueOf[types.String] `tfsdk:"allow_resources"`
}
