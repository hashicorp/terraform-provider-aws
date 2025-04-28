// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_inference_profile", name="Inference Profile")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrock;bedrock.GetInferenceProfileOutput")
// @Testing(importIgnore="model_source.#;model_source.0.%;model_source.0.copy_from")
func newResourceInferenceProfile(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceInferenceProfile{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameInferenceProfile = "Inference Profile"
)

type resourceInferenceProfile struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceInferenceProfile) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	modelsAttribute := framework.ResourceComputedListOfObjectsAttribute[resourceInferenceProfileModelModel](ctx)
	modelsAttribute.PlanModifiers = []planmodifier.List{
		listplanmodifier.UseStateForUnknown(),
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"models": modelsAttribute,
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.InferenceProfileStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.InferenceProfileType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"model_source": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceProfileModelModelSource](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"copy_from": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceInferenceProfile) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var plan resourceInferenceProfileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrock.CreateInferenceProfileInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("InferenceProfile"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := tfresource.RetryGWhen(ctx, 2*time.Minute, func() (*bedrock.CreateInferenceProfileOutput, error) {
		return conn.CreateInferenceProfile(ctx, &input)
	}, func(err error) (bool, error) {
		if errs.IsA[*awstypes.ConflictException](err) {
			return true, err
		}
		return false, err
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionCreating, ResNameInferenceProfile, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan, flex.WithFieldNamePrefix("InferenceProfile"))...)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	profile, err := waitInferenceProfileCreated(ctx, conn, *out.InferenceProfileArn, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForCreation, ResNameInferenceProfile, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// Populate the rest of the fields from the describe call, create only returns the ARN and Status
	resp.Diagnostics.Append(flex.Flatten(ctx, profile, &plan, flex.WithFieldNamePrefix("InferenceProfile"))...)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceInferenceProfile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state resourceInferenceProfileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findInferenceProfileByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionSetting, ResNameInferenceProfile, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("InferenceProfile"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceInferenceProfile) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var new resourceInferenceProfileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Tags only.

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceInferenceProfile) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state resourceInferenceProfileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrock.DeleteInferenceProfileInput{
		InferenceProfileIdentifier: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteInferenceProfile(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionDeleting, ResNameInferenceProfile, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitInferenceProfileDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForDeletion, ResNameInferenceProfile, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceInferenceProfile) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitInferenceProfileCreated(ctx context.Context, conn *bedrock.Client, id string, timeout time.Duration) (*bedrock.GetInferenceProfileOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(string(awstypes.InferenceProfileStatusActive)),
		Refresh:                   statusInferenceProfile(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrock.GetInferenceProfileOutput); ok {
		return out, err
	}

	return nil, err
}

func waitInferenceProfileDeleted(ctx context.Context, conn *bedrock.Client, id string, timeout time.Duration) (*bedrock.GetInferenceProfileOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(string(awstypes.InferenceProfileStatusActive)),
		Target:  []string{},
		Refresh: statusInferenceProfile(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrock.GetInferenceProfileOutput); ok {
		return out, err
	}

	return nil, err
}

func statusInferenceProfile(ctx context.Context, conn *bedrock.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findInferenceProfileByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

type resourceInferenceProfileModel struct {
	ARN         types.String                                                        `tfsdk:"arn"`
	ID          types.String                                                        `tfsdk:"id"`
	ModelSource fwtypes.ListNestedObjectValueOf[inferenceProfileModelModelSource]   `tfsdk:"model_source"`
	Description types.String                                                        `tfsdk:"description"`
	Name        types.String                                                        `tfsdk:"name"`
	Models      fwtypes.ListNestedObjectValueOf[resourceInferenceProfileModelModel] `tfsdk:"models"`
	Status      fwtypes.StringEnum[awstypes.InferenceProfileStatus]                 `tfsdk:"status"`
	Type        fwtypes.StringEnum[awstypes.InferenceProfileType]                   `tfsdk:"type"`
	CreatedAt   timetypes.RFC3339                                                   `tfsdk:"created_at"`
	UpdatedAt   timetypes.RFC3339                                                   `tfsdk:"updated_at"`
	Timeouts    timeouts.Value                                                      `tfsdk:"timeouts"`
	Tags        tftags.Map                                                          `tfsdk:"tags"`
	TagsAll     tftags.Map                                                          `tfsdk:"tags_all"`
}

type inferenceProfileModelModelSource struct {
	CopyFrom types.String `tfsdk:"copy_from"`
}

type resourceInferenceProfileModelModel struct {
	ModelARN types.String `tfsdk:"model_arn"`
}

var (
	_ flex.Expander  = inferenceProfileModelModelSource{}
	_ flex.Flattener = &inferenceProfileModelModelSource{}
)

func (m inferenceProfileModelModelSource) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return &awstypes.InferenceProfileModelSourceMemberCopyFrom{
		Value: flex.StringValueFromFramework(ctx, m.CopyFrom),
	}, nil
}

func (m *inferenceProfileModelModelSource) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch val := v.(type) {
	case *awstypes.InferenceProfileModelSourceMemberCopyFrom:
		m.CopyFrom = flex.StringValueToFramework(ctx, val.Value)
		return diags

	default:
		return diags
	}
}
