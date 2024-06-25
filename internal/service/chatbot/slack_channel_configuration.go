// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chatbot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chatbot/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Slack Channel Configuration")
// @Tags(identifierAttribute="resource_arn")
func newSlackChannelConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &slackChannelConfigurationResource{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultUpdateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

type slackChannelConfigurationResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *slackChannelConfigurationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_chatbot_slack_channel_configuration"
}

func (r *slackChannelConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"chat_configuration_arn": schema.StringAttribute{
				Computed: true,
			},
			"configuration_name": schema.StringAttribute{
				Required: true,
			},
			"guardrail_policy_arns": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"iam_role_arn": schema.StringAttribute{
				Required: true,
			},
			"logging_level": schema.StringAttribute{
				Optional: true,
			},
			"slack_channel_id": schema.StringAttribute{
				Required: true,
			},
			"slack_channel_name": schema.StringAttribute{
				Computed: true,
			},
			"slack_team_id": schema.StringAttribute{
				Required: true,
			},
			"slack_team_name": schema.StringAttribute{
				Computed: true,
			},
			"sns_topic_arns": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"user_authorization_required": schema.BoolAttribute{
				Optional: true,
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

func (r *slackChannelConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data slackChannelConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ChatbotClient(ctx)

	input := &chatbot.CreateSlackChannelConfigurationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	_, err := conn.CreateSlackChannelConfiguration(ctx, input)
	if err != nil {
		response.Diagnostics.AddError("creating Chatbot Slack Channel Configuration", err.Error())

		return
	}

	output, err := waitSlackChannelConfigurationAvailable(ctx, conn, data.ChatConfigurationARN.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Chatbot Slack Channel Configuration (%s) create", data.ChatConfigurationARN.ValueString()), err.Error())

		return
	}

	// Set values for unknowns
	data.ChatConfigurationARN = fwflex.StringToFramework(ctx, output.ChatConfigurationArn)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *slackChannelConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data slackChannelConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().ChatbotClient(ctx)

	output, err := findSlackChannelConfigurationByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Chatbot, create.ErrActionReading, ResNameSlackChannelConfiguration, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *slackChannelConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new slackChannelConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ChatbotClient(ctx)

	if slackChannelConfigurationHasChanges(ctx, new, old) {
		input := &chatbot.UpdateSlackChannelConfigurationInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateSlackChannelConfiguration(ctx, input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Chatbot Slack Channel Configuration (%s)", new.ChatConfigurationARN.ValueString()), err.Error())

			return
		}

		if _, err := waitSlackChannelConfigurationAvailable(ctx, conn, old.ChatConfigurationARN.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Chatbot Slack Channel Configuration (%s) update", new.ChatConfigurationARN.ValueString()), err.Error())

			return
		}
	}

	output, err := findSlackChannelConfigurationByID(ctx, conn, old.ChatConfigurationARN.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Chatbot Slack Channel Configuration (%s)", old.ChatConfigurationARN.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *slackChannelConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data slackChannelConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ChatbotClient(ctx)

	tflog.Debug(ctx, "deleting Chatbot Slack Channel Configuration", map[string]interface{}{
		names.AttrID: data.ID.ValueString(),
	})

	input := &chatbot.DeleteSlackChannelConfigurationInput{
		ChatConfigurationArn: aws.String(data.ChatConfigurationARN.ValueString()),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, r.DeleteTimeout(ctx, data.Timeouts), func() (interface{}, error) {
		return conn.DeleteSlackChannelConfiguration(ctx, input)
	}, "DependencyViolation")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Chatbot Slack Channel Configuration (%s)", data.ChatConfigurationARN.ValueString()), err.Error())

		return
	}

	if _, err := waitSlackChannelConfigurationDeleted(ctx, conn, data.ChatConfigurationARN.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Chatbot Slack Channel Configuration (%s) delete", data.ChatConfigurationARN.ValueString()), err.Error())

		return
	}
}

func (r *slackChannelConfigurationResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findSlackChannelConfiguration(ctx context.Context, conn *chatbot.Client, input *chatbot.DescribeSlackChannelConfigurationsInput) (*awstypes.SlackChannelConfiguration, error) {
	output, err := findSlackChannelConfigurations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSlackChannelConfigurations(ctx context.Context, conn *chatbot.Client, input *chatbot.DescribeSlackChannelConfigurationsInput) ([]awstypes.SlackChannelConfiguration, error) {
	var output []awstypes.SlackChannelConfiguration

	pages := chatbot.NewDescribeSlackChannelConfigurationsPaginator(conn, input)
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

		output = append(output, page.SlackChannelConfigurations...)
	}

	return output, nil
}

func findSlackChannelConfigurationByID(ctx context.Context, conn *chatbot.Client, arn string) (*awstypes.SlackChannelConfiguration, error) {
	input := &chatbot.DescribeSlackChannelConfigurationsInput{
		ChatConfigurationArn: aws.String(arn),
	}

	return findSlackChannelConfiguration(ctx, conn, input)
}

const (
	slackChannelConfigurationAvailable = "AVAILABLE"
)

func statusSlackChannelConfiguration(ctx context.Context, conn *chatbot.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findSlackChannelConfigurationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, slackChannelConfigurationAvailable, nil
	}
}

func waitSlackChannelConfigurationAvailable(ctx context.Context, conn *chatbot.Client, id string, timeout time.Duration) (*awstypes.SlackChannelConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{},
		Target:     []string{slackChannelConfigurationAvailable},
		Refresh:    statusSlackChannelConfiguration(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SlackChannelConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitSlackChannelConfigurationDeleted(ctx context.Context, conn *chatbot.Client, id string, timeout time.Duration) (*awstypes.SlackChannelConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{slackChannelConfigurationAvailable},
		Target:     []string{},
		Refresh:    statusSlackChannelConfiguration(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SlackChannelConfiguration); ok {
		return output, err
	}

	return nil, err
}

type slackChannelConfigurationResourceModel struct {
	ChatConfigurationARN      types.String   `tfsdk:"chat_configuration_arn"`
	ConfigurationName         types.String   `tfsdk:"configuration_name"`
	GuardrailPolicyARNs       types.List     `tfsdk:"guardrail_policy_arns"`
	IAMRoleARN                types.String   `tfsdk:"iam_role_arn"`
	ID                        types.String   `tfsdk:"id"`
	LoggingLevel              types.String   `tfsdk:"logging_level"`
	SlackChannelID            types.String   `tfsdk:"slack_channel_id"`
	SlackChannelName          types.String   `tfsdk:"slack_channel_name"`
	SlackTeamID               types.String   `tfsdk:"slack_team_id"`
	SlackTeamName             types.String   `tfsdk:"slack_team_name"`
	SNSTopicARNs              types.List     `tfsdk:"sns_topic_arns"`
	Tags                      types.Map      `tfsdk:"tags"`
	TagsAll                   types.Map      `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value `tfsdk:"timeouts"`
	UserAuthorizationRequired types.Bool     `tfsdk:"user_authorization_required"`
}

func (data *slackChannelConfigurationResourceModel) InitFromID() error {
	data.ChatConfigurationARN = data.ID

	return nil
}

func (data *slackChannelConfigurationResourceModel) setID() {
	data.ID = data.ChatConfigurationARN
}

func slackChannelConfigurationHasChanges(_ context.Context, plan, state slackChannelConfigurationResourceModel) bool {
	return !plan.ChatConfigurationARN.Equal(state.ChatConfigurationARN) ||
		!plan.ConfigurationName.Equal(state.ConfigurationName) ||
		!plan.GuardrailPolicyARNs.Equal(state.GuardrailPolicyARNs) ||
		!plan.IAMRoleARN.Equal(state.IAMRoleARN) ||
		!plan.LoggingLevel.Equal(state.LoggingLevel) ||
		!plan.SlackChannelID.Equal(state.SlackChannelID) ||
		!plan.SlackChannelName.Equal(state.SlackChannelName) ||
		!plan.SlackTeamID.Equal(state.SlackTeamID) ||
		!plan.SlackTeamName.Equal(state.SlackTeamName) ||
		!plan.SNSTopicARNs.Equal(state.SNSTopicARNs) ||
		!plan.UserAuthorizationRequired.Equal(state.UserAuthorizationRequired)
}
