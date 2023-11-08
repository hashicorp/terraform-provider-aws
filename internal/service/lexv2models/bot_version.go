// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"

	// "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Bot Version")
func newResourceBotVersion(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceBotVersion{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameBotVersion = "Bot Version"
)

type resourceBotVersion struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceBotVersion) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_lexv2models_bot_version"
}

func (r *resourceBotVersion) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bot_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"idle_session_ttl_in_seconds": schema.Int64Attribute{
				Computed: true,
			},
			"locale_specification": schema.MapAttribute{
				Required:    true,
				ElementType: types.ObjectType{AttrTypes: botVersionLocaleDetails},
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"type": schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.BotType](),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// "version": schema.StringAttribute{
			// 	Computed: true,
			// 	PlanModifiers: []planmodifier.String{
			// 		stringplanmodifier.RequiresReplace(),
			// 	},
			// },
		},
		Blocks: map[string]schema.Block{
			"members": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alias_id": schema.StringAttribute{
							Computed: true,
						},
						"alias_name": schema.StringAttribute{
							Computed: true,
						},
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"version": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"data_privacy": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"child_directed": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceBotVersion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var plan resourceBotVersionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var localeSpec map[string]versionLocaleDetailsData
	resp.Diagnostics.Append(plan.LocaleSpecification.ElementsAs(ctx, &localeSpec, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lexmodelsv2.CreateBotVersionInput{
		BotId:                         aws.String(plan.BotID.String()),
		BotVersionLocaleSpecification: expandLocalSpecification(ctx, localeSpec),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	out, err := conn.CreateBotVersion(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameBotVersion, plan.BotID.ValueString(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.BotVersion == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameBotVersion, plan.BotID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.BotID = flex.StringToFramework(ctx, out.BotId)
	state := plan

	createTimeout := r.CreateTimeout(ctx, state.Timeouts)
	_, err = waitBotVersionCreated(ctx, conn, plan.BotID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForCreation, ResNameBotVersion, state.BotID.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceBotVersion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceBotVersionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := FindBotVersionByID(ctx, conn, state.BotID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionSetting, ResNameBotVersion, state.BotID.String(), err),
			err.Error(),
		)
		return
	}

	state.Description = flex.StringToFramework(ctx, out.Description)
	state.BotID = flex.StringToFramework(ctx, out.BotId)
	state.Name = flex.StringToFramework(ctx, out.BotName)
	state.Type = flex.StringToFramework(ctx, (*string)(&out.BotType))
	state.IdleSessionTTLInSeconds = flex.Int32ToFramework(ctx, out.IdleSessionTTLInSeconds)
	state.RoleARN = flex.StringToFrameworkARN(ctx, out.RoleArn)

	members, _ := flattenMembers(ctx, out.BotMembers)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Members = members

	datap, _ := flattenDataPrivacy(out.DataPrivacy)
	if resp.Diagnostics.HasError() {
		return
	}

	state.DataPrivacy = datap
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceBotVersion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No-op update

}

func (r *resourceBotVersion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var state resourceBotVersionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lexmodelsv2.DeleteBotVersionInput{
		BotId: aws.String(state.BotID.ValueString()),
	}

	_, err := conn.DeleteBotVersion(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionDeleting, ResNameBotVersion, state.BotID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitBotVersionDeleted(ctx, conn, state.BotID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForDeletion, ResNameBotVersion, state.BotID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceBotVersion) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitBotVersionCreated(ctx context.Context, conn *lexmodelsv2.Client, id string, timeout time.Duration) (*lexmodelsv2.DescribeBotVersionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.BotStatusVersioning),
		Target:                    enum.Slice(awstypes.BotStatusAvailable),
		Refresh:                   statusBotVersion(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexmodelsv2.DescribeBotVersionOutput); ok {
		return out, err
	}

	return nil, err
}

func waitBotVersionDeleted(ctx context.Context, conn *lexmodelsv2.Client, id string, timeout time.Duration) (*lexmodelsv2.DescribeBotVersionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BotStatusDeleting),
		Target:  []string{},
		Refresh: statusBotVersion(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexmodelsv2.DescribeBotVersionOutput); ok {
		return out, err
	}

	return nil, err
}

func statusBotVersion(ctx context.Context, conn *lexmodelsv2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindBotVersionByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.BotStatus)), nil
	}
}

func FindBotVersionByID(ctx context.Context, conn *lexmodelsv2.Client, id string) (*lexmodelsv2.DescribeBotVersionOutput, error) {
	in := &lexmodelsv2.DescribeBotVersionInput{
		BotId: aws.String(id),
	}

	out, err := conn.DescribeBotVersion(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.BotVersion == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func expandLocalSpecification(ctx context.Context, tfMap map[string]versionLocaleDetailsData) map[string]awstypes.BotVersionLocaleDetails {
	if len(tfMap) == 0 {
		return nil
	}

	tfObj := make(map[string]awstypes.BotVersionLocaleDetails)
	for key, value := range tfMap {
		tfObj[key] = awstypes.BotVersionLocaleDetails{SourceBotVersion: aws.String(value.SourceBotVersion.ValueString())}
	}

	return tfObj
}

type resourceBotVersionData struct {
	LocaleSpecification     types.Map      `tfsdk:"locale_specification"`
	DataPrivacy             types.List     `tfsdk:"data_privacy"`
	Description             types.String   `tfsdk:"description"`
	BotID                   types.String   `tfsdk:"bot_id"`
	IdleSessionTTLInSeconds types.Int64    `tfsdk:"idle_session_ttl_in_seconds"`
	Members                 types.List     `tfsdk:"members"`
	Name                    types.String   `tfsdk:"name"`
	RoleARN                 fwtypes.ARN    `tfsdk:"role_arn"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
	Type                    types.String   `tfsdk:"type"`
}

type versionLocaleDetailsData struct {
	SourceBotVersion types.String `tfsdk:"source_bot_version"`
}

var botVersionLocaleDetails = map[string]attr.Type{
	"source_bot_version": types.StringType,
}
