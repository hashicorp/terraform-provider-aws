// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chatbot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chatbot/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_chatbot_teams_channel_configuration", name="Teams Channel Configuration")
// @Tags(identifierAttribute="chat_configuration_arn")
func newTeamsChannelConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &teamsChannelConfigurationResource{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultUpdateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

type teamsChannelConfigurationResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *teamsChannelConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"channel_id": schema.StringAttribute{
				Required: true,
			},
			"channel_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"chat_configuration_arn": framework.ARNAttributeComputedOnly(),
			"configuration_name": schema.StringAttribute{
				Required: true,
			},
			"guardrail_policy_arns": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrIAMRoleARN: schema.StringAttribute{
				Required: true,
			},
			"logging_level": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[loggingLevel](),
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString(string(loggingLevelNone)),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sns_topic_arns": schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"team_id": schema.StringAttribute{
				Required: true,
			},
			"team_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Required: true,
			},
			"user_authorization_required": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *teamsChannelConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data teamsChannelConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ChatbotClient(ctx)

	input := &chatbot.CreateMicrosoftTeamsChannelConfigurationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateMicrosoftTeamsChannelConfiguration(ctx, input)
	if err != nil {
		create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionCreating, ResNameTeamsChannelConfiguration, data.TeamID.ValueString(), err)

		return
	}

	output, err := waitTeamsChannelConfigurationAvailable(ctx, conn, aws.ToString(out.ChannelConfiguration.TeamId), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionWaitingForCreation, ResNameTeamsChannelConfiguration, aws.ToString(out.ChannelConfiguration.TeamId), err)

		return
	}

	// Set values for unknowns
	data.ChatConfigurationARN = fwflex.StringToFramework(ctx, output.ChatConfigurationArn)
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *teamsChannelConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data teamsChannelConfigurationResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ChatbotClient(ctx)

	output, err := findTeamsChannelConfigurationByTeamID(ctx, conn, data.TeamID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionReading, ResNameTeamsChannelConfiguration, data.TeamID.ValueString(), err)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *teamsChannelConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new teamsChannelConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ChatbotClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := &chatbot.UpdateMicrosoftTeamsChannelConfigurationInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateMicrosoftTeamsChannelConfiguration(ctx, input)
		if err != nil {
			create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionUpdating, ResNameTeamsChannelConfiguration, new.TeamID.ValueString(), err)

			return
		}

		if _, err := waitTeamsChannelConfigurationAvailable(ctx, conn, old.TeamID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionWaitingForUpdate, ResNameTeamsChannelConfiguration, new.TeamID.ValueString(), err)

			return
		}
	}

	output, err := findTeamsChannelConfigurationByTeamID(ctx, conn, old.TeamID.ValueString())
	if err != nil {
		create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionReading, ResNameTeamsChannelConfiguration, old.TeamID.ValueString(), err)

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *teamsChannelConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data teamsChannelConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ChatbotClient(ctx)

	tflog.Debug(ctx, "deleting Chatbot Teams Channel Configuration", map[string]any{
		"team_id":                data.TeamID.ValueString(),
		"chat_configuration_arn": data.ChatConfigurationARN.ValueString(),
	})

	input := &chatbot.DeleteMicrosoftTeamsChannelConfigurationInput{
		ChatConfigurationArn: data.ChatConfigurationARN.ValueStringPointer(),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, r.DeleteTimeout(ctx, data.Timeouts), func() (any, error) {
		return conn.DeleteMicrosoftTeamsChannelConfiguration(ctx, input)
	}, "DependencyViolation")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionDeleting, ResNameTeamsChannelConfiguration, data.TeamID.ValueString(), err)
		return
	}

	if _, err := waitTeamsChannelConfigurationDeleted(ctx, conn, data.TeamID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionWaitingForDeletion, ResNameTeamsChannelConfiguration, data.TeamID.ValueString(), err)
		return
	}
}

func (r *teamsChannelConfigurationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("team_id"), request, response)
}

func findTeamsChannelConfiguration(ctx context.Context, conn *chatbot.Client, input *chatbot.ListMicrosoftTeamsChannelConfigurationsInput) (*awstypes.TeamsChannelConfiguration, error) {
	output, err := findTeamsChannelConfigurations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTeamsChannelConfigurations(ctx context.Context, conn *chatbot.Client, input *chatbot.ListMicrosoftTeamsChannelConfigurationsInput) ([]awstypes.TeamsChannelConfiguration, error) {
	var output []awstypes.TeamsChannelConfiguration

	pages := chatbot.NewListMicrosoftTeamsChannelConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TeamChannelConfigurations...)
	}

	return output, nil
}

func findTeamsChannelConfigurationByTeamID(ctx context.Context, conn *chatbot.Client, teamID string) (*awstypes.TeamsChannelConfiguration, error) {
	input := &chatbot.ListMicrosoftTeamsChannelConfigurationsInput{
		TeamId: aws.String(teamID),
	}

	return findTeamsChannelConfiguration(ctx, conn, input)
}

const (
	teamsChannelConfigurationAvailable = "AVAILABLE"
)

func statusTeamsChannelConfiguration(ctx context.Context, conn *chatbot.Client, teamID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findTeamsChannelConfigurationByTeamID(ctx, conn, teamID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, teamsChannelConfigurationAvailable, nil
	}
}

func waitTeamsChannelConfigurationAvailable(ctx context.Context, conn *chatbot.Client, teamID string, timeout time.Duration) (*awstypes.TeamsChannelConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{},
		Target:     []string{teamsChannelConfigurationAvailable},
		Refresh:    statusTeamsChannelConfiguration(ctx, conn, teamID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TeamsChannelConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitTeamsChannelConfigurationDeleted(ctx context.Context, conn *chatbot.Client, teamID string, timeout time.Duration) (*awstypes.TeamsChannelConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{teamsChannelConfigurationAvailable},
		Target:     []string{},
		Refresh:    statusTeamsChannelConfiguration(ctx, conn, teamID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TeamsChannelConfiguration); ok {
		return output, err
	}

	return nil, err
}

type teamsChannelConfigurationResourceModel struct {
	ChannelID                 types.String                      `tfsdk:"channel_id"`
	ChannelName               types.String                      `tfsdk:"channel_name"`
	ChatConfigurationARN      types.String                      `tfsdk:"chat_configuration_arn"`
	ConfigurationName         types.String                      `tfsdk:"configuration_name"`
	GuardrailPolicyARNs       fwtypes.ListValueOf[types.String] `tfsdk:"guardrail_policy_arns"`
	IAMRoleARN                types.String                      `tfsdk:"iam_role_arn"`
	LoggingLevel              fwtypes.StringEnum[loggingLevel]  `tfsdk:"logging_level"`
	SNSTopicARNs              fwtypes.SetValueOf[types.String]  `tfsdk:"sns_topic_arns"`
	Tags                      tftags.Map                        `tfsdk:"tags"`
	TagsAll                   tftags.Map                        `tfsdk:"tags_all"`
	TeamID                    types.String                      `tfsdk:"team_id"`
	TeamName                  types.String                      `tfsdk:"team_name"`
	TenantID                  types.String                      `tfsdk:"tenant_id"`
	Timeouts                  timeouts.Value                    `tfsdk:"timeouts"`
	UserAuthorizationRequired types.Bool                        `tfsdk:"user_authorization_required"`
}
