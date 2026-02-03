// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package lexv2models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lexv2models_bot", name="Bot")
// @Tags(identifierAttribute="arn")
func newBotResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &botResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type botResource struct {
	framework.ResourceWithModel[botResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *botResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.BotType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"data_privacy": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataPrivacyModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
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
			"members": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[botMemberModel](ctx),
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *botResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data botResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.BotName)
	var input lexmodelsv2.CreateBotInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.BotTags = getTagsIn(ctx)

	output, err := conn.CreateBot(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Lex v2 Bot (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	botID := aws.ToString(output.BotId)
	data.ARN = fwflex.StringValueToFramework(ctx, r.botARN(ctx, botID))
	data.BotID = fwflex.StringValueToFramework(ctx, botID)
	data.BotType = fwtypes.StringEnumValue(output.BotType)

	if _, err := waitBotCreated(ctx, conn, botID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.BotID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Lex v2 Bot (%s) create", botID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *botResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data botResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	output, err := findBotByID(ctx, conn, data.BotID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Lex v2 Bot (%s)", data.BotID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.ARN = fwflex.StringValueToFramework(ctx, r.botARN(ctx, data.BotID.ValueString()))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *botResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old botResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	if !new.BotType.Equal(old.BotType) ||
		!new.DataPrivacy.Equal(old.DataPrivacy) ||
		!new.Description.Equal(old.Description) ||
		!new.IdleSessionTTLInSeconds.Equal(old.IdleSessionTTLInSeconds) ||
		!new.RoleARN.Equal(old.RoleARN) ||
		!new.TestBotAliasTags.Equal(old.TestBotAliasTags) {
		botID := fwflex.StringValueFromFramework(ctx, new.BotID)
		var input lexmodelsv2.UpdateBotInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateBot(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Lex v2 Bot (%s)", botID), err.Error())

			return
		}

		if _, err := waitBotUpdated(ctx, conn, botID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Lex v2 Bot (%s) update", botID), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *botResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data botResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	botID := fwflex.StringValueFromFramework(ctx, data.BotID)
	input := lexmodelsv2.DeleteBotInput{
		BotId: aws.String(botID),
	}
	_, err := conn.DeleteBot(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
		errs.IsAErrorMessageContains[*awstypes.PreconditionFailedException](err, "does not exist") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Lex v2 Bot (%s)", botID), err.Error())

		return
	}

	if _, err := waitBotDeleted(ctx, conn, botID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Lex v2 Bot (%s) delete", botID), err.Error())

		return
	}
}

func (r *botResource) botARN(ctx context.Context, botID string) string {
	return r.Meta().RegionalARN(ctx, "lex", "bot/"+botID)
}

func findBotByID(ctx context.Context, conn *lexmodelsv2.Client, id string) (*lexmodelsv2.DescribeBotOutput, error) {
	input := lexmodelsv2.DescribeBotInput{
		BotId: aws.String(id),
	}
	output, err := conn.DescribeBot(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.BotId == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func botFailureReasons(failureReasons []string) error {
	return errors.Join(tfslices.ApplyToAll(failureReasons, errors.New)...)
}

func statusBot(conn *lexmodelsv2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findBotByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.BotStatus), nil
	}
}

func waitBotCreated(ctx context.Context, conn *lexmodelsv2.Client, id string, timeout time.Duration) (*lexmodelsv2.DescribeBotOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.BotStatusCreating),
		Target:                    enum.Slice(awstypes.BotStatusAvailable),
		Refresh:                   statusBot(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lexmodelsv2.DescribeBotOutput); ok {
		retry.SetLastError(err, botFailureReasons(output.FailureReasons))

		return output, err
	}

	return nil, err
}

func waitBotUpdated(ctx context.Context, conn *lexmodelsv2.Client, id string, timeout time.Duration) (*lexmodelsv2.DescribeBotOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.BotStatusUpdating),
		Target:                    enum.Slice(awstypes.BotStatusAvailable),
		Refresh:                   statusBot(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lexmodelsv2.DescribeBotOutput); ok {
		retry.SetLastError(err, botFailureReasons(output.FailureReasons))

		return output, err
	}

	return nil, err
}

func waitBotDeleted(ctx context.Context, conn *lexmodelsv2.Client, id string, timeout time.Duration) (*lexmodelsv2.DescribeBotOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BotStatusDeleting),
		Target:  []string{},
		Refresh: statusBot(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lexmodelsv2.DescribeBotOutput); ok {
		retry.SetLastError(err, botFailureReasons(output.FailureReasons))

		return output, err
	}

	return nil, err
}

type botResourceModel struct {
	framework.WithRegionModel
	ARN                     types.String                                      `tfsdk:"arn"`
	BotID                   types.String                                      `tfsdk:"id"`
	BotMembers              fwtypes.ListNestedObjectValueOf[botMemberModel]   `tfsdk:"members"`
	BotName                 types.String                                      `tfsdk:"name"`
	BotType                 fwtypes.StringEnum[awstypes.BotType]              `tfsdk:"type"`
	DataPrivacy             fwtypes.ListNestedObjectValueOf[dataPrivacyModel] `tfsdk:"data_privacy"`
	Description             types.String                                      `tfsdk:"description"`
	IdleSessionTTLInSeconds types.Int64                                       `tfsdk:"idle_session_ttl_in_seconds"`
	RoleARN                 fwtypes.ARN                                       `tfsdk:"role_arn"`
	Tags                    tftags.Map                                        `tfsdk:"tags"`
	TagsAll                 tftags.Map                                        `tfsdk:"tags_all"`
	TestBotAliasTags        fwtypes.MapOfString                               `tfsdk:"test_bot_alias_tags"`
	Timeouts                timeouts.Value                                    `tfsdk:"timeouts"`
}

type dataPrivacyModel struct {
	ChildDirected types.Bool `tfsdk:"child_directed"`
}

type botMemberModel struct {
	BotMemberAliasID   types.String `tfsdk:"alias_id"`
	BotMemberAliasName types.String `tfsdk:"alias_name"`
	BotMemberID        types.String `tfsdk:"id"`
	BotMemberName      types.String `tfsdk:"name"`
	BotMemberVersion   types.String `tfsdk:"version"`
}
