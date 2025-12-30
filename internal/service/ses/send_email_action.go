// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_ses_send_email, name="Send Email")
func newSendEmailAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &sendEmailAction{}, nil
}

var (
	_ action.Action = (*sendEmailAction)(nil)
)

type sendEmailAction struct {
	framework.ActionWithModel[sendEmailActionModel]
}

type sendEmailActionModel struct {
	framework.WithRegionModel
	Source           types.String                      `tfsdk:"source"`
	ToAddresses      fwtypes.ListValueOf[types.String] `tfsdk:"to_addresses"`
	CcAddresses      fwtypes.ListValueOf[types.String] `tfsdk:"cc_addresses"`
	BccAddresses     fwtypes.ListValueOf[types.String] `tfsdk:"bcc_addresses"`
	Subject          types.String                      `tfsdk:"subject"`
	TextBody         types.String                      `tfsdk:"text_body"`
	HtmlBody         types.String                      `tfsdk:"html_body"`
	ReplyToAddresses fwtypes.ListValueOf[types.String] `tfsdk:"reply_to_addresses"`
	ReturnPath       types.String                      `tfsdk:"return_path"`
}

func (a *sendEmailAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sends an email using Amazon SES. This action allows for imperative email sending with full control over recipients, content, and formatting.",
		Attributes: map[string]schema.Attribute{
			names.AttrSource: schema.StringAttribute{
				Description: "The email address that is sending the email. This address must be either individually verified with Amazon SES, or from a domain that has been verified with Amazon SES.",
				Required:    true,
			},
			"to_addresses": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Description: "The To: field(s) of the message.",
				Optional:    true,
			},
			"cc_addresses": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Description: "The CC: field(s) of the message.",
				Optional:    true,
			},
			"bcc_addresses": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Description: "The BCC: field(s) of the message.",
				Optional:    true,
			},
			"subject": schema.StringAttribute{
				Description: "The subject of the message: A short summary of the content, which will appear in the recipient's inbox.",
				Required:    true,
			},
			"text_body": schema.StringAttribute{
				Description: "The message body in text format. Either text_body or html_body must be specified.",
				Optional:    true,
			},
			"html_body": schema.StringAttribute{
				Description: "The message body in HTML format. Either text_body or html_body must be specified.",
				Optional:    true,
			},
			"reply_to_addresses": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Description: "The reply-to email address(es) for the message. If the recipient replies to the message, each reply-to address will receive the reply.",
				Optional:    true,
			},
			"return_path": schema.StringAttribute{
				Description: "The email address that bounces and complaints will be forwarded to when feedback forwarding is enabled.",
				Optional:    true,
			},
		},
	}
}

func (a *sendEmailAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config sendEmailActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that at least one body type is provided
	if config.TextBody.IsNull() && config.HtmlBody.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Email Body",
			"Either text_body or html_body must be specified",
		)
		return
	}

	conn := a.Meta().SESClient(ctx)

	source := config.Source.ValueString()
	subject := config.Subject.ValueString()

	tflog.Info(ctx, "Starting SES send email action", map[string]any{
		names.AttrSource: source,
		"subject":        subject,
		"has_text_body":  !config.TextBody.IsNull(),
		"has_html_body":  !config.HtmlBody.IsNull(),
	})

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Sending email from %s...", source),
	})

	// Build destination
	destination := &awstypes.Destination{}
	if !config.ToAddresses.IsNull() {
		destination.ToAddresses = fwflex.ExpandFrameworkStringValueList(ctx, config.ToAddresses)
	}
	if !config.CcAddresses.IsNull() {
		destination.CcAddresses = fwflex.ExpandFrameworkStringValueList(ctx, config.CcAddresses)
	}
	if !config.BccAddresses.IsNull() {
		destination.BccAddresses = fwflex.ExpandFrameworkStringValueList(ctx, config.BccAddresses)
	}

	// Build message
	message := &awstypes.Message{
		Subject: &awstypes.Content{
			Data: aws.String(subject),
		},
		Body: &awstypes.Body{},
	}

	if !config.TextBody.IsNull() {
		message.Body.Text = &awstypes.Content{
			Data: config.TextBody.ValueStringPointer(),
		}
	}
	if !config.HtmlBody.IsNull() {
		message.Body.Html = &awstypes.Content{
			Data: config.HtmlBody.ValueStringPointer(),
		}
	}

	// Build input
	input := &ses.SendEmailInput{
		Source:      aws.String(source),
		Destination: destination,
		Message:     message,
	}

	if !config.ReplyToAddresses.IsNull() {
		input.ReplyToAddresses = fwflex.ExpandFrameworkStringValueList(ctx, config.ReplyToAddresses)
	}

	if !config.ReturnPath.IsNull() {
		input.ReturnPath = config.ReturnPath.ValueStringPointer()
	}

	// Send email
	output, err := conn.SendEmail(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Send Email",
			fmt.Sprintf("Could not send email from %s: %s", source, err),
		)
		return
	}

	messageId := aws.ToString(output.MessageId)
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Email sent successfully (Message ID: %s)", messageId),
	})

	tflog.Info(ctx, "SES send email action completed successfully", map[string]any{
		names.AttrSource: source,
		"message_id":     messageId,
	})
}
