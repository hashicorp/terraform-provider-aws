// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chatbot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chatbot/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_chatbot_slack_channel_configuration", name="Slack Channel Configuration")
// @Tags(identifierAttribute="chat_configuration_arn")
func newDataSourceSlackChannelConfiguration(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceSlackChannelConfiguration{}, nil
}

const (
	DSNameSlackChannelConfiguration = "Slack Channel Configuration Data Source"
)

type dataSourceSlackChannelConfiguration struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceSlackChannelConfiguration) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for managing an AWS Chatbot Slack Channel Configuration.",
		Attributes: map[string]schema.Attribute{
			"chat_configuration_arn": schema.StringAttribute{
				Description: "ARN of the Slack channel configuration.",
				CustomType:  fwtypes.ARNType,
				Required:    true,
			},
			"configuration_name": schema.StringAttribute{
				Description: "Name of the Slack channel configuration.",
				Computed:    true,
			},
			names.AttrIAMRoleARN: schema.StringAttribute{
				Description: "ARN of the IAM role that defines the permissions for AWS Chatbot.",
				Computed:    true,
			},
			"logging_level": schema.StringAttribute{
				Description: "Specifies the logging level for this configuration.",
				CustomType:  fwtypes.StringEnumType[loggingLevel](),
				Computed:    true,
			},
			"slack_channel_id": schema.StringAttribute{
				Description: "ID of the Slack channel.",
				Computed:    true,
			},
			"slack_channel_name": schema.StringAttribute{
				Description: "Name of the Slack channel.",
				Computed:    true,
			},
			"slack_team_id": schema.StringAttribute{
				Description: "ID of the Slack workspace authorized with AWS Chatbot.",
				Computed:    true,
			},
			"slack_team_name": schema.StringAttribute{
				Description: "Name of the Slack workspace.",
				Computed:    true,
			},
			"sns_topic_arns": schema.SetAttribute{
				Description: "ARNs of the SNS topics that deliver notifications to AWS Chatbot.",
				CustomType:  fwtypes.SetOfStringType,
				Computed:    true,
			},
			names.AttrState: schema.StringAttribute{
				Description: "State of the configuration.",
				CustomType:  fwtypes.StringEnumType[slackChannelConfigurationState](),
				Computed:    true,
			},
			names.AttrTags:    tftags.TagsAttributeComputedOnly(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"user_authorization_required": schema.BoolAttribute{
				Description: "Enables use of a user role requirement in your chat configuration.",
				Computed:    true,
			},
		},
	}
}

func (d *dataSourceSlackChannelConfiguration) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ChatbotClient(ctx)

	var data dataSourceSlackChannelConfigurationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findSlackChannelConfigurationByARNForDataSource(ctx, conn, data.ChatConfigurationArn.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Chatbot, create.ErrActionReading, DSNameSlackChannelConfiguration, data.ChatConfigurationArn.ValueString(), err),
			err.Error(),
		)
		return
	}

	// Using a field name prefix allows mapping fields such as `SlackChannelConfigurationId` to `ID`
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("SlackChannelConfiguration"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findSlackChannelConfigurationByARNForDataSource(ctx context.Context, conn *chatbot.Client, chatConfigurationArn string) (*awstypes.SlackChannelConfiguration, error) {
	input := &chatbot.DescribeSlackChannelConfigurationsInput{
		ChatConfigurationArn: aws.String(chatConfigurationArn),
	}

	output, err := conn.DescribeSlackChannelConfigurations(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(output.SlackChannelConfigurations) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return &output.SlackChannelConfigurations[0], nil
}

type dataSourceSlackChannelConfigurationModel struct {
	ChatConfigurationArn      fwtypes.ARN                                        `tfsdk:"chat_configuration_arn"`
	ConfigurationName         types.String                                       `tfsdk:"configuration_name"`
	IamRoleArn                types.String                                       `tfsdk:"iam_role_arn"`
	LoggingLevel              fwtypes.StringEnum[loggingLevel]                   `tfsdk:"logging_level"`
	SlackChannelId            types.String                                       `tfsdk:"slack_channel_id"`
	SlackChannelName          types.String                                       `tfsdk:"slack_channel_name"`
	SlackTeamId               types.String                                       `tfsdk:"slack_team_id"`
	SlackTeamName             types.String                                       `tfsdk:"slack_team_name"`
	SnsTopicArns              fwtypes.SetValueOf[types.String]                   `tfsdk:"sns_topic_arns"`
	State                     fwtypes.StringEnum[slackChannelConfigurationState] `tfsdk:"state"`
	Tags                      tftags.Map                                         `tfsdk:"tags"`
	TagsAll                   tftags.Map                                         `tfsdk:"tags_all"`
	UserAuthorizationRequired types.Bool                                         `tfsdk:"user_authorization_required"`
}
