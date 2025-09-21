// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_sns_publish, name="Publish")
func newPublishAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &publishAction{}, nil
}

var (
	_ action.Action = (*publishAction)(nil)
)

type publishAction struct {
	framework.ActionWithModel[publishActionModel]
}

type publishActionModel struct {
	framework.WithRegionModel
	TopicArn          types.String                                           `tfsdk:"topic_arn"`
	Message           types.String                                           `tfsdk:"message"`
	Subject           types.String                                           `tfsdk:"subject"`
	MessageStructure  types.String                                           `tfsdk:"message_structure"`
	MessageAttributes fwtypes.ListNestedObjectValueOf[messageAttributeModel] `tfsdk:"message_attributes"`
}

type messageAttributeModel struct {
	MapBlockKey types.String `tfsdk:"map_block_key"`
	DataType    types.String `tfsdk:"data_type"`
	StringValue types.String `tfsdk:"string_value"`
}

func (a *publishAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Publishes a message to an Amazon SNS topic. This action allows for imperative message publishing with full control over message attributes and structure.",
		Attributes: map[string]schema.Attribute{
			names.AttrMessage: schema.StringAttribute{
				Description: "The message to publish. For JSON message structure, this should be a JSON object with protocol-specific messages.",
				Required:    true,
			},
			names.AttrTopicARN: schema.StringAttribute{
				Description: "The ARN of the SNS topic to publish the message to.",
				Required:    true,
			},
			"message_structure": schema.StringAttribute{
				Description: "Set to 'json' if you want to send different messages for each protocol. If not specified, the message will be sent as-is to all protocols.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(names.AttrJSON),
				},
			},
			"subject": schema.StringAttribute{
				Description: "Optional subject for the message. Only used for email and email-json protocols.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"message_attributes": schema.ListNestedBlock{
				Description: "Message attributes to include with the message. Each block represents one attribute where map_block_key becomes the attribute name.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[messageAttributeModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{ // nosemgrep:ci.semgrep.framework.map_block_key-meaningful-names
						"data_type": schema.StringAttribute{
							Description: "The data type of the message attribute. Valid values are String, Number, and Binary.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("String", "Number", "Binary"),
							},
						},
						"map_block_key": schema.StringAttribute{
							Description: "The name of the message attribute (used as map key).",
							Required:    true,
						},
						"string_value": schema.StringAttribute{
							Description: "The value of the message attribute.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (a *publishAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config publishActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().SNSClient(ctx)

	topicArn := config.TopicArn.ValueString()
	message := config.Message.ValueString()

	tflog.Info(ctx, "Starting SNS publish message action", map[string]any{
		names.AttrTopicARN: topicArn,
		"message_length":   len(message),
		"has_subject":      !config.Subject.IsNull(),
	})

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Publishing message to SNS topic %s...", topicArn),
	})

	input := &sns.PublishInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, config, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure required fields are set (AutoFlex should handle these, but being explicit)
	input.TopicArn = aws.String(topicArn)
	input.Message = aws.String(message)

	output, err := conn.Publish(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Publish SNS Message",
			fmt.Sprintf("Could not publish message to SNS topic %s: %s", topicArn, err),
		)
		return
	}

	messageId := aws.ToString(output.MessageId)
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Message published successfully to SNS topic %s (Message ID: %s)", topicArn, messageId),
	})

	tflog.Info(ctx, "SNS publish message action completed successfully", map[string]any{
		names.AttrTopicARN: topicArn,
		"message_id":       messageId,
	})
}
