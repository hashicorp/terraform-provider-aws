// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Bot")
func newResourceBot(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceBot{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameBot = "Bot"
)

type resourceBot struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceBot) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_lexv2models_bot"
}

func (r *resourceBot) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
			"idle_session_ttl_in_seconds": schema.Int64Attribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"role_arn": schema.StringAttribute{
				Required: true,
			},
			"test_bot_alias_tags": schema.StringAttribute{
				Optional: true,
			},
			"type": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"members": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alias_id": schema.StringAttribute{
							Required: true,
						},
						"alias_name": schema.StringAttribute{
							Required: true,
						},
						"id": schema.StringAttribute{
							Required: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"version": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"data_privacy": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"child_directed": schema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceBot) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var plan resourceBotData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var dp []dataPrivacyData
	resp.Diagnostics.Append(plan.DataPrivacy.ElementsAs(ctx, &dp, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dpInput, d := expandDataPrivacy(ctx, dp)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := lexmodelsv2.CreateBotInput{
		BotName:                 aws.String(plan.Name.ValueString()),
		DataPrivacy:             dpInput,
		IdleSessionTTLInSeconds: aws.Int32(int32(plan.IdleSessionTTLInSeconds.ValueInt64())),
		RoleArn:                 aws.String(plan.RoleARN.ValueString()),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	out, err := conn.CreateBot(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameBot, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.BotId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameBot, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.BotId)
	state := plan
	// resp.Diagnostics.Append(state.refreshFromOutput(ctx, out.BotId)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceBot) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var state resourceBotData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindBotByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionSetting, ResNameBot, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.RoleARN = flex.StringToFramework(ctx, out.RoleArn)
	state.ID = flex.StringToFramework(ctx, out.BotId)
	state.Name = flex.StringToFramework(ctx, out.BotName)
	state.Type = flex.StringToFramework(ctx, (*string)(&out.BotType))
	state.Description = flex.StringToFramework(ctx, out.Description)
	state.IdleSessionTTLInSeconds = flex.Int32ToFramework(ctx, out.IdleSessionTTLInSeconds)

	datap, _ := flattenDataPrivacy(ctx, out.DataPrivacy)

	state.DataPrivacy = datap
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceBot) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var plan, state resourceBotData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.IdleSessionTTLInSeconds.Equal(state.IdleSessionTTLInSeconds) ||
		!plan.RoleARN.Equal(state.RoleARN) ||
		!plan.TestBotAliasTags.Equal(state.TestBotAliasTags) ||
		!plan.DataPrivacy.Equal(state.DataPrivacy) ||
		!plan.Type.Equal(state.Type) {

		var dp []dataPrivacyData
		resp.Diagnostics.Append(plan.DataPrivacy.ElementsAs(ctx, &dp, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		dpInput, d := expandDataPrivacy(ctx, dp)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		in := lexmodelsv2.UpdateBotInput{
			BotId:                   aws.String(plan.ID.ValueString()),
			BotName:                 aws.String(plan.Name.ValueString()),
			IdleSessionTTLInSeconds: aws.Int32(int32(plan.IdleSessionTTLInSeconds.ValueInt64())),
			DataPrivacy:             dpInput,
			RoleArn:                 aws.String(plan.RoleARN.ValueString()),
		}

		if !plan.Description.IsNull() {
			in.Description = aws.String(plan.Description.ValueString())
		}

		if !plan.Members.IsNull() {
			var tfList []membersData
			resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}

		out, err := conn.UpdateBot(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameBot, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.BotId == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameBot, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		state.refreshFromOutput(ctx, out)
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitBotUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForUpdate, ResNameBot, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceBot) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var state resourceBotData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lexmodelsv2.DeleteBotInput{
		BotId: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteBot(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionDeleting, ResNameBot, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitBotDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForDeletion, ResNameBot, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceBot) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func waitBotCreated(ctx context.Context, conn *lexmodelsv2.Client, id string, timeout time.Duration) (*lexmodelsv2.DescribeBotOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusBot(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexmodelsv2.DescribeBotOutput); ok {
		return out, err
	}

	return nil, err
}

func waitBotUpdated(ctx context.Context, conn *lexmodelsv2.Client, id string, timeout time.Duration) (*lexmodelsv2.DescribeBotOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusBot(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexmodelsv2.DescribeBotOutput); ok {
		return out, err
	}

	return nil, err
}

func waitBotDeleted(ctx context.Context, conn *lexmodelsv2.Client, id string, timeout time.Duration) (*lexmodelsv2.DescribeBotOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusBot(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexmodelsv2.DescribeBotOutput); ok {
		return out, err
	}

	return nil, err
}

func statusBot(ctx context.Context, conn *lexmodelsv2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindBotByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.BotStatus)), nil
	}
}

func FindBotByID(ctx context.Context, conn *lexmodelsv2.Client, id string) (*lexmodelsv2.DescribeBotOutput, error) {
	in := &lexmodelsv2.DescribeBotInput{
		BotId: aws.String(id),
	}

	out, err := conn.DescribeBot(ctx, in)
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

	if out == nil || out.BotId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenDataPrivacy(ctx context.Context, apiObject *awstypes.DataPrivacy) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: dataPrivacyAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"child_directed": types.BoolValue(aws.ToBool(&apiObject.ChildDirected)),
	}
	objVal, d := types.ObjectValue(dataPrivacyAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func expandDataPrivacy(ctx context.Context, tfList []dataPrivacyData) (*awstypes.DataPrivacy, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(tfList) == 0 {
		return nil, diags
	}

	dp := tfList[0]
	cdBool, _ := strconv.ParseBool(dp.ChildDirected.ValueString())

	return &awstypes.DataPrivacy{
		ChildDirected: aws.ToBool(&cdBool),
	}, diags
}

func expandMembers(ctx context.Context, tfList []membersData) (*awstypes.BotMember, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	mb := tfList[0]
	return &awstypes.BotMember{
		BotMemberAliasId:   aws.String(mb.AliasID.ValueString()),
		BotMemberAliasName: aws.String(mb.AliasName.ValueString()),
		BotMemberId:        aws.String(mb.ID.ValueString()),
		BotMemberName:      aws.String(mb.Name.ValueString()),
		BotMemberVersion:   aws.String(mb.Version.ValueString()),
	}, diags
}

func (rd *resourceBotData) refreshFromOutput(ctx context.Context, out *lexmodelsv2.UpdateBotOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}
	rd.RoleARN = flex.StringToFramework(ctx, out.RoleArn)
	rd.ID = flex.StringToFramework(ctx, out.BotId)
	rd.Name = flex.StringToFramework(ctx, out.BotName)
	rd.Type = flex.StringToFramework(ctx, (*string)(&out.BotType))
	rd.Description = flex.StringToFramework(ctx, out.Description)
	rd.IdleSessionTTLInSeconds = flex.Int32ToFramework(ctx, out.IdleSessionTTLInSeconds)

	// TIP: Setting a complex type.
	datap, d := flattenDataPrivacy(ctx, out.DataPrivacy)
	diags.Append(d...)
	rd.DataPrivacy = datap

	return diags
}

type resourceBotData struct {
	DataPrivacy             types.List     `tfsdk:"data_privacy"`
	Description             types.String   `tfsdk:"description"`
	ID                      types.String   `tfsdk:"id"`
	IdleSessionTTLInSeconds types.Int64    `tfsdk:"idle_session_ttl_in_seconds"`
	Name                    types.String   `tfsdk:"name"`
	Members                 types.List     `tfsdk:"members"`
	RoleARN                 types.String   `tfsdk:"role_arn"`
	Tags                    types.Map      `tfsdk:"tags"`
	TagsAll                 types.Map      `tfsdk:"tags_all"`
	TestBotAliasTags        types.Map      `tfsdk:"test_bot_alias_tags"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
	Type                    types.String   `tfsdk:"type"`
}

type dataPrivacyData struct {
	ChildDirected types.String `tfsdk:"child_directed"`
}

type membersData struct {
	AliasID   types.String `tfsdk:"alias_id"`
	AliasName types.String `tfsdk:"alias_name"`
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Version   types.String `tfsdk:"version"`
}

var dataPrivacyAttrTypes = map[string]attr.Type{
	"child_directed": types.StringType,
}
