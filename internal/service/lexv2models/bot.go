// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Bot")
// @Tags(identifierAttribute="arn")
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
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"idle_session_ttl_in_seconds": schema.Int64Attribute{
				Required: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"test_bot_alias_tags": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			names.AttrType: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.BotType](),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"members": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alias_id": schema.StringAttribute{
							Required: true,
						},
						"alias_name": schema.StringAttribute{
							Required: true,
						},
						names.AttrID: schema.StringAttribute{
							Required: true,
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						names.AttrVersion: schema.StringAttribute{
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
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

	dpInput := expandDataPrivacy(ctx, dp)

	in := lexmodelsv2.CreateBotInput{
		BotName:                 aws.String(plan.Name.ValueString()),
		DataPrivacy:             dpInput,
		IdleSessionTTLInSeconds: aws.Int32(int32(plan.IdleSessionTTLInSeconds.ValueInt64())),
		RoleArn:                 flex.StringFromFramework(ctx, plan.RoleARN),
		BotTags:                 getTagsIn(ctx),
	}

	if !plan.TestBotAliasTags.IsNull() {
		in.TestBotAliasTags = flex.ExpandFrameworkStringValueMap(ctx, plan.TestBotAliasTags)
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	var bm []membersData
	if !plan.Members.IsNull() {
		bmInput := expandMembers(ctx, bm)
		in.BotMembers = bmInput
	}

	if !plan.Type.IsNull() {
		in.BotType = awstypes.BotType(plan.Type.ValueString())
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
	botArn := arn.ARN{
		Partition: r.Meta().Partition,
		Service:   "lex",
		Region:    r.Meta().Region,
		AccountID: r.Meta().AccountID,
		Resource:  fmt.Sprintf("bot/%s", aws.ToString(out.BotId)),
	}.String()
	plan.ID = flex.StringToFramework(ctx, out.BotId)
	state := plan
	state.Type = flex.StringValueToFramework(ctx, out.BotType)
	state.ARN = flex.StringValueToFramework(ctx, botArn)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)

	createTimeout := r.CreateTimeout(ctx, state.Timeouts)
	_, err = waitBotCreated(ctx, conn, state.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForDeletion, ResNameBot, state.ID.String(), err),
			err.Error(),
		)
		return
	}
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

	botArn := arn.ARN{
		Partition: r.Meta().Partition,
		Service:   "lex",
		Region:    r.Meta().Region,
		AccountID: r.Meta().AccountID,
		Resource:  fmt.Sprintf("bot/%s", aws.ToString(out.BotId)),
	}.String()
	state.ARN = flex.StringValueToFramework(ctx, botArn)
	state.RoleARN = flex.StringToFrameworkARN(ctx, out.RoleArn)
	state.ID = flex.StringToFramework(ctx, out.BotId)
	state.Name = flex.StringToFramework(ctx, out.BotName)
	state.Type = flex.StringValueToFramework(ctx, out.BotType)
	state.Description = flex.StringToFramework(ctx, out.Description)
	state.IdleSessionTTLInSeconds = flex.Int32ToFramework(ctx, out.IdleSessionTTLInSeconds)

	members, errDiags := flattenMembers(ctx, out.BotMembers)
	resp.Diagnostics.Append(errDiags...)
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

func (r *resourceBot) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var plan, state resourceBotData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) ||
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

		dpInput := expandDataPrivacy(ctx, dp)

		in := lexmodelsv2.UpdateBotInput{
			BotId:                   flex.StringFromFramework(ctx, plan.ID),
			BotName:                 flex.StringFromFramework(ctx, plan.Name),
			IdleSessionTTLInSeconds: aws.Int32(int32(plan.IdleSessionTTLInSeconds.ValueInt64())),
			DataPrivacy:             dpInput,
			RoleArn:                 flex.StringFromFramework(ctx, plan.RoleARN),
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

		if !plan.Type.IsNull() {
			in.BotType = awstypes.BotType(plan.Type.ValueString())
		}

		_, err := conn.UpdateBot(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameBot, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		out, err := waitBotUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForUpdate, ResNameBot, plan.ID.String(), err),
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
		resp.Diagnostics.Append(plan.refreshFromOutput(ctx, out)...)
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

func (r *resourceBot) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func (r *resourceBot) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitBotCreated(ctx context.Context, conn *lexmodelsv2.Client, id string, timeout time.Duration) (*lexmodelsv2.DescribeBotOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.BotStatusCreating),
		Target:                    enum.Slice(awstypes.BotStatusAvailable),
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
		Pending:                   enum.Slice(awstypes.BotStatusUpdating),
		Target:                    enum.Slice(awstypes.BotStatusAvailable),
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
		Pending: enum.Slice(awstypes.BotStatusDeleting),
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

func flattenDataPrivacy(apiObject *awstypes.DataPrivacy) (types.List, diag.Diagnostics) {
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

func flattenMembers(ctx context.Context, apiObject []awstypes.BotMember) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: botMembersAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, source := range apiObject {
		obj := map[string]attr.Value{
			"alias_name":      flex.StringToFramework(ctx, source.BotMemberAliasName),
			"alias_id":        flex.StringToFramework(ctx, source.BotMemberAliasId),
			names.AttrID:      flex.StringToFramework(ctx, source.BotMemberId),
			names.AttrName:    flex.StringToFramework(ctx, source.BotMemberName),
			names.AttrVersion: flex.StringToFramework(ctx, source.BotMemberVersion),
		}
		objVal, d := types.ObjectValue(botMembersAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func expandDataPrivacy(ctx context.Context, tfList []dataPrivacyData) *awstypes.DataPrivacy {
	if len(tfList) == 0 {
		return nil
	}

	dp := tfList[0]
	cdBool := flex.BoolFromFramework(ctx, dp.ChildDirected)

	return &awstypes.DataPrivacy{
		ChildDirected: aws.ToBool(cdBool),
	}
}

func expandMembers(ctx context.Context, tfList []membersData) []awstypes.BotMember {
	if len(tfList) == 0 {
		return nil
	}
	var mb []awstypes.BotMember

	for _, item := range tfList {
		new := awstypes.BotMember{
			BotMemberAliasId:   flex.StringFromFramework(ctx, item.AliasID),
			BotMemberAliasName: flex.StringFromFramework(ctx, item.AliasName),
			BotMemberId:        flex.StringFromFramework(ctx, item.ID),
			BotMemberName:      flex.StringFromFramework(ctx, item.Name),
			BotMemberVersion:   flex.StringFromFramework(ctx, item.Version),
		}
		mb = append(mb, new)
	}

	return mb
}

func (rd *resourceBotData) refreshFromOutput(ctx context.Context, out *lexmodelsv2.DescribeBotOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}
	rd.RoleARN = flex.StringToFrameworkARN(ctx, out.RoleArn)
	rd.ID = flex.StringToFramework(ctx, out.BotId)
	rd.Name = flex.StringToFramework(ctx, out.BotName)
	rd.Type = flex.StringToFramework(ctx, (*string)(&out.BotType))
	rd.Description = flex.StringToFramework(ctx, out.Description)
	rd.IdleSessionTTLInSeconds = flex.Int32ToFramework(ctx, out.IdleSessionTTLInSeconds)

	datap, d := flattenDataPrivacy(out.DataPrivacy)
	diags.Append(d...)
	rd.DataPrivacy = datap

	return diags
}

type resourceBotData struct {
	ARN                     types.String   `tfsdk:"arn"`
	DataPrivacy             types.List     `tfsdk:"data_privacy"`
	Description             types.String   `tfsdk:"description"`
	ID                      types.String   `tfsdk:"id"`
	IdleSessionTTLInSeconds types.Int64    `tfsdk:"idle_session_ttl_in_seconds"`
	Name                    types.String   `tfsdk:"name"`
	Members                 types.List     `tfsdk:"members"`
	RoleARN                 fwtypes.ARN    `tfsdk:"role_arn"`
	Tags                    types.Map      `tfsdk:"tags"`
	TagsAll                 types.Map      `tfsdk:"tags_all"`
	TestBotAliasTags        types.Map      `tfsdk:"test_bot_alias_tags"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
	Type                    types.String   `tfsdk:"type"`
}

type dataPrivacyData struct {
	ChildDirected types.Bool `tfsdk:"child_directed"`
}

type membersData struct {
	AliasID   types.String `tfsdk:"alias_id"`
	AliasName types.String `tfsdk:"alias_name"`
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Version   types.String `tfsdk:"version"`
}

var dataPrivacyAttrTypes = map[string]attr.Type{
	"child_directed": types.BoolType,
}

var botMembersAttrTypes = map[string]attr.Type{
	"alias_id":        types.StringType,
	"alias_name":      types.StringType,
	names.AttrID:      types.StringType,
	names.AttrName:    types.StringType,
	names.AttrVersion: types.StringType,
}
