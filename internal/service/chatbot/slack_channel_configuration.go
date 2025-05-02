// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
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

// @FrameworkResource("aws_chatbot_slack_channel_configuration", name="Slack Channel Configuration")
// @Tags(identifierAttribute="chat_configuration_arn")
func newSlackChannelConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &slackChannelConfigurationResource{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultUpdateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

type slackChannelConfigurationResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *slackChannelConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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

	cout, err := conn.CreateSlackChannelConfiguration(ctx, input)
	if err != nil {
		create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionCreating, ResNameSlackChannelConfiguration, data.ChatConfigurationARN.ValueString(), err)

		return
	}

	output, err := waitSlackChannelConfigurationAvailable(ctx, conn, aws.ToString(cout.ChannelConfiguration.ChatConfigurationArn), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionWaitingForCreation, ResNameSlackChannelConfiguration, aws.ToString(cout.ChannelConfiguration.ChatConfigurationArn), err)

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

func (r *slackChannelConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data slackChannelConfigurationResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionExpandingResourceId, ResNameSlackChannelConfiguration, data.ChatConfigurationARN.ValueString(), err)

		return
	}

	conn := r.Meta().ChatbotClient(ctx)

	output, err := findSlackChannelConfigurationByARN(ctx, conn, data.ChatConfigurationARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionReading, ResNameSlackChannelConfiguration, data.ChatConfigurationARN.ValueString(), err)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

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

	diff, d := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := &chatbot.UpdateSlackChannelConfigurationInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateSlackChannelConfiguration(ctx, input)
		if err != nil {
			create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionUpdating, ResNameSlackChannelConfiguration, new.ChatConfigurationARN.ValueString(), err)

			return
		}

		if _, err := waitSlackChannelConfigurationAvailable(ctx, conn, old.ChatConfigurationARN.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionWaitingForUpdate, ResNameSlackChannelConfiguration, new.ChatConfigurationARN.ValueString(), err)

			return
		}
	}

	output, err := findSlackChannelConfigurationByARN(ctx, conn, old.ChatConfigurationARN.ValueString())
	if err != nil {
		create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionReading, ResNameSlackChannelConfiguration, old.ChatConfigurationARN.ValueString(), err)

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

	tflog.Debug(ctx, "deleting Chatbot Slack Channel Configuration", map[string]any{
		"chat_configuration_arn": data.ChatConfigurationARN.ValueString(),
	})

	input := &chatbot.DeleteSlackChannelConfigurationInput{
		ChatConfigurationArn: data.ChatConfigurationARN.ValueStringPointer(),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, r.DeleteTimeout(ctx, data.Timeouts), func() (any, error) {
		return conn.DeleteSlackChannelConfiguration(ctx, input)
	}, "DependencyViolation")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionDeleting, ResNameSlackChannelConfiguration, data.ChatConfigurationARN.ValueString(), err)
		return
	}

	if _, err := waitSlackChannelConfigurationDeleted(ctx, conn, data.ChatConfigurationARN.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		create.AddError(&response.Diagnostics, names.Chatbot, create.ErrActionWaitingForDeletion, ResNameSlackChannelConfiguration, data.ChatConfigurationARN.ValueString(), err)
		return
	}
}

func (r *slackChannelConfigurationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("chat_configuration_arn"), request, response)
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

func findSlackChannelConfigurationByARN(ctx context.Context, conn *chatbot.Client, arn string) (*awstypes.SlackChannelConfiguration, error) {
	input := &chatbot.DescribeSlackChannelConfigurationsInput{
		ChatConfigurationArn: aws.String(arn),
	}

	return findSlackChannelConfiguration(ctx, conn, input)
}

const (
	slackChannelConfigurationAvailable = "AVAILABLE"
)

func statusSlackChannelConfiguration(ctx context.Context, conn *chatbot.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findSlackChannelConfigurationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, slackChannelConfigurationAvailable, nil
	}
}

func waitSlackChannelConfigurationAvailable(ctx context.Context, conn *chatbot.Client, arn string, timeout time.Duration) (*awstypes.SlackChannelConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{},
		Target:     []string{slackChannelConfigurationAvailable},
		Refresh:    statusSlackChannelConfiguration(ctx, conn, arn),
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

func waitSlackChannelConfigurationDeleted(ctx context.Context, conn *chatbot.Client, arn string, timeout time.Duration) (*awstypes.SlackChannelConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{slackChannelConfigurationAvailable},
		Target:     []string{},
		Refresh:    statusSlackChannelConfiguration(ctx, conn, arn),
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
	ChatConfigurationARN      types.String                      `tfsdk:"chat_configuration_arn"`
	ConfigurationName         types.String                      `tfsdk:"configuration_name"`
	GuardrailPolicyARNs       fwtypes.ListValueOf[types.String] `tfsdk:"guardrail_policy_arns"`
	IAMRoleARN                types.String                      `tfsdk:"iam_role_arn"`
	LoggingLevel              fwtypes.StringEnum[loggingLevel]  `tfsdk:"logging_level"`
	SlackChannelID            types.String                      `tfsdk:"slack_channel_id"`
	SlackChannelName          types.String                      `tfsdk:"slack_channel_name"`
	SlackTeamID               types.String                      `tfsdk:"slack_team_id"`
	SlackTeamName             types.String                      `tfsdk:"slack_team_name"`
	SNSTopicARNs              fwtypes.SetValueOf[types.String]  `tfsdk:"sns_topic_arns"`
	Tags                      tftags.Map                        `tfsdk:"tags"`
	TagsAll                   tftags.Map                        `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                    `tfsdk:"timeouts"`
	UserAuthorizationRequired types.Bool                        `tfsdk:"user_authorization_required"`
}

func (data *slackChannelConfigurationResourceModel) InitFromID() error {
	_, err := arn.Parse(data.ChatConfigurationARN.ValueString())
	return err
}

type loggingLevel string

const (
	loggingLevelError loggingLevel = "ERROR"
	loggingLevelInfo  loggingLevel = "INFO"
	loggingLevelNone  loggingLevel = "NONE"
)

func (loggingLevel) Values() []loggingLevel {
	return []loggingLevel{
		loggingLevelError,
		loggingLevelInfo,
		loggingLevelNone,
	}
}
