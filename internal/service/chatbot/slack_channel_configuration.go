// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chatbot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chatbot/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_chatbot_slack_channel_configuration", name="Slack Channel Configuration")
// @Tags(identifierAttribute="chat_configuration_arn")
func newResourceSlackChannelConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceSlackChannelConfiguration{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameSlackChannelConfiguration = "Slack Channel Configuration"
)

type resourceSlackChannelConfiguration struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceSlackChannelConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_chatbot_slack_channel_configuration"
}

func (r *resourceSlackChannelConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"configuration_name": schema.StringAttribute{
				Required: true,
			},
			"chat_configuration_arn": framework.ARNAttributeComputedOnly(),
			"guardrail_policy_arns": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: fwtypes.ARNType,
				Optional:    true,
				Computed:    true,
			},
			names.AttrIAMRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrID: framework.IDAttribute(),
			"logging_level": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"ERROR", "INFO", "NONE"}...),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"slack_channel_id": schema.StringAttribute{
				Required: true,
			},
			"slack_channel_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"slack_team_id": schema.StringAttribute{
				Required: true,
			},
			"slack_team_name": schema.StringAttribute{
				Computed: true,
			},
			"sns_topic_arns": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: fwtypes.ARNType,
				Optional:    true,
				Computed:    true,
			},
			"user_authorization_required": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (r *resourceSlackChannelConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceSlackChannelConfigurationData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ChatbotClient(ctx)

	input := &chatbot.CreateSlackChannelConfigurationInput{
		Tags: getTagsIn(ctx),
	}
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outputRaw, err := conn.CreateSlackChannelConfiguration(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Chatbot, create.ErrActionCreating, ResNameSlackChannelConfiguration, plan.ConfigurationName.String(), err),
			err.Error(),
		)
		return
	}
	if outputRaw == nil || outputRaw.ChannelConfiguration == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Chatbot, create.ErrActionCreating, ResNameSlackChannelConfiguration, plan.ConfigurationName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	output, err := findSlackChannelConfigurationByID(ctx, conn, *outputRaw.ChannelConfiguration.ChatConfigurationArn)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Chatbot, create.ErrActionCreating, ResNameSlackChannelConfiguration, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.setID()
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceSlackChannelConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ChatbotClient(ctx)

	var state resourceSlackChannelConfigurationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := state.InitFromID(); err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	output, err := findSlackChannelConfigurationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Chatbot, create.ErrActionReading, ResNameSlackChannelConfiguration, state.ChatConfigurationArn.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func findSlackChannelConfigurationByID(ctx context.Context, conn *chatbot.Client, ID string) (*awstypes.SlackChannelConfiguration, error) {
	input := &chatbot.DescribeSlackChannelConfigurationsInput{
		ChatConfigurationArn: aws.String(ID),
	}

	if ID == "" {
		return nil, tfresource.NewEmptyResultError(input)
	}

	output, err := conn.DescribeSlackChannelConfigurations(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.SlackChannelConfigurations == nil || len(output.SlackChannelConfigurations) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return &output.SlackChannelConfigurations[0], nil
}

func (r *resourceSlackChannelConfiguration) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (r *resourceSlackChannelConfiguration) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan resourceSlackChannelConfigurationData

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ChatbotClient(ctx)

	if !plan.GuardrailPolicyARNs.Equal(state.GuardrailPolicyARNs) ||
		!plan.LoggingLevel.Equal(state.LoggingLevel) ||
		!plan.SlackChannelName.Equal(state.SlackChannelName) ||
		!plan.SlackTeamID.Equal(state.SlackTeamID) ||
		!plan.SNSTopicARNs.Equal(state.SNSTopicARNs) ||
		!plan.UserAuthorizationRequired.Equal(state.UserAuthorizationRequired) {
		input := &chatbot.UpdateSlackChannelConfigurationInput{
			ChatConfigurationArn: aws.String(state.ChatConfigurationArn.ValueString()),
			SlackChannelId:       plan.SlackChannelID.ValueStringPointer(),
		}
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		outputRaw, err := conn.UpdateSlackChannelConfiguration(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Chatbot, create.ErrActionUpdating, ResNameSlackChannelConfiguration, plan.ChatConfigurationArn.String(), err),
				err.Error(),
			)
			return
		}

		output, err := findSlackChannelConfigurationByID(ctx, conn, *outputRaw.ChannelConfiguration.ChatConfigurationArn)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Chatbot, create.ErrActionReading, ResNameSlackChannelConfiguration, plan.ChatConfigurationArn.String(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan)...)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceSlackChannelConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceSlackChannelConfigurationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ChatbotClient(ctx)

	in := &chatbot.DeleteSlackChannelConfigurationInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, state, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteSlackChannelConfiguration(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Chatbot, create.ErrActionDeleting, ResNameSlackChannelConfiguration, state.ChatConfigurationArn.String(), err),
			err.Error(),
		)
		return
	}
}

type resourceSlackChannelConfigurationData struct {
	ChatConfigurationArn      types.String                     `tfsdk:"chat_configuration_arn"`
	ConfigurationName         types.String                     `tfsdk:"configuration_name"`
	GuardrailPolicyARNs       fwtypes.SetValueOf[types.String] `tfsdk:"guardrail_policy_arns"`
	IAMRoleARN                fwtypes.ARN                      `tfsdk:"iam_role_arn"`
	ID                        types.String                     `tfsdk:"id"`
	LoggingLevel              types.String                     `tfsdk:"logging_level"`
	SlackChannelID            types.String                     `tfsdk:"slack_channel_id"`
	SlackChannelName          types.String                     `tfsdk:"slack_channel_name"`
	SlackTeamID               types.String                     `tfsdk:"slack_team_id"`
	SlackTeamName             types.String                     `tfsdk:"slack_team_name"`
	SNSTopicARNs              fwtypes.SetValueOf[types.String] `tfsdk:"sns_topic_arns"`
	Tags                      types.Map                        `tfsdk:"tags"`
	TagsAll                   types.Map                        `tfsdk:"tags_all"`
	UserAuthorizationRequired types.Bool                       `tfsdk:"user_authorization_required"`
}

func (data *resourceSlackChannelConfigurationData) InitFromID() error {
	data.ChatConfigurationArn = data.ID
	return nil
}

func (data *resourceSlackChannelConfigurationData) setID() {
	data.ID = data.ChatConfigurationArn
}
