// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package lexv2models

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lexv2models_bot_version", name="Bot Version")
func newBotVersionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &botVersionResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type botVersionResource struct {
	framework.ResourceWithConfigure
	// Doesn't work with map[string]Object.
	// framework.ResourceWithModel[botVersionResourceModel]
	framework.WithNoUpdate
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *botVersionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bot_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bot_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"locale_specification": schema.MapAttribute{
				Required: true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[botVersionLocaleDetailsModel](ctx),
				},
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

const (
	botVersionResourceIDPartCount = 2
)

func (r *botVersionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data botVersionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	var input lexmodelsv2.CreateBotVersionInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.BotVersionLocaleSpecification = tfmaps.ApplyToAllValues(data.BotVersionLocaleSpecification.Elements(), func(v attr.Value) awstypes.BotVersionLocaleDetails {
		return awstypes.BotVersionLocaleDetails{
			SourceBotVersion: fwflex.StringFromFramework(ctx, v.(types.Object).Attributes()["source_bot_version"].(types.String)),
		}
	})

	output, err := conn.CreateBotVersion(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Lex v2 Bot Version", err.Error())

		return
	}

	// Set values for unknowns.
	botID, botVersion := aws.ToString(output.BotId), aws.ToString(output.BotVersion)
	id, _ := intflex.FlattenResourceId([]string{botID, botVersion}, botVersionResourceIDPartCount, false)
	data.BotVersion = fwflex.StringValueToFramework(ctx, botVersion)
	data.ID = fwflex.StringValueToFramework(ctx, id)

	if _, err := waitBotVersionCreated(ctx, conn, botID, botVersion, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.BotID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Lex v2 Bot Locale (%s) create", id), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *botVersionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data botVersionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	parts, err := intflex.ExpandResourceId(id, botVersionResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	botID, botVersion := parts[0], parts[1]
	output, err := findBotVersionByTwoPartKey(ctx, conn, botID, botVersion)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Lex v2 Bot Version (%s)", id), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *botVersionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data botVersionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	parts, err := intflex.ExpandResourceId(id, botVersionResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	botID, botVersion := parts[0], parts[1]
	input := lexmodelsv2.DeleteBotVersionInput{
		BotId:      aws.String(botID),
		BotVersion: aws.String(botVersion),
	}
	_, err = conn.DeleteBotVersion(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
		errs.IsAErrorMessageContains[*awstypes.PreconditionFailedException](err, "does not exist") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Lex v2 Bot Version (%s)", id), err.Error())

		return
	}

	if _, err := waitBotVersionDeleted(ctx, conn, botID, botVersion, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Lex v2 Bot Version (%s) delete", id), err.Error())

		return
	}
}

func findBotVersionByTwoPartKey(ctx context.Context, conn *lexmodelsv2.Client, botID, botVersion string) (*lexmodelsv2.DescribeBotVersionOutput, error) {
	input := lexmodelsv2.DescribeBotVersionInput{
		BotId:      aws.String(botID),
		BotVersion: aws.String(botVersion),
	}
	output, err := conn.DescribeBotVersion(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.BotVersion == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusBotVersion(conn *lexmodelsv2.Client, botID, botVersion string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findBotVersionByTwoPartKey(ctx, conn, botID, botVersion)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.BotStatus), nil
	}
}

func waitBotVersionCreated(ctx context.Context, conn *lexmodelsv2.Client, botID, botVersion string, timeout time.Duration) (*lexmodelsv2.DescribeBotVersionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.BotStatusCreating, awstypes.BotStatusVersioning),
		Target:                    enum.Slice(awstypes.BotStatusAvailable),
		Refresh:                   statusBotVersion(conn, botID, botVersion),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lexmodelsv2.DescribeBotVersionOutput); ok {
		retry.SetLastError(err, botFailureReasons(output.FailureReasons))

		return output, err
	}

	return nil, err
}

func waitBotVersionDeleted(ctx context.Context, conn *lexmodelsv2.Client, botID, botVersion string, timeout time.Duration) (*lexmodelsv2.DescribeBotVersionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BotStatusDeleting),
		Target:  []string{},
		Refresh: statusBotVersion(conn, botID, botVersion),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lexmodelsv2.DescribeBotVersionOutput); ok {
		retry.SetLastError(err, botFailureReasons(output.FailureReasons))

		return output, err
	}

	return nil, err
}

type botVersionResourceModel struct {
	framework.WithRegionModel
	BotID      types.String `tfsdk:"bot_id"`
	BotVersion types.String `tfsdk:"bot_version"`
	//BotVersionLocaleSpecification fwtypes.MapOfObject `tfsdk:"locale_specification" autoflex:"-"`
	BotVersionLocaleSpecification types.Map      `tfsdk:"locale_specification" autoflex:"-"`
	Description                   types.String   `tfsdk:"description"`
	ID                            types.String   `tfsdk:"id"`
	Timeouts                      timeouts.Value `tfsdk:"timeouts"`
}

type botVersionLocaleDetailsModel struct {
	SourceBotVersion types.String `tfsdk:"source_bot_version"`
}
