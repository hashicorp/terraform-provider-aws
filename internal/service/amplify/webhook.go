// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify

import (
	"context"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/amplify"
	"github.com/aws/aws-sdk-go-v2/service/amplify/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_amplify_webhook", name="Webhook")
func resourceWebhook() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWebhookCreate,
		ReadWithoutTimeout:   resourceWebhookRead,
		UpdateWithoutTimeout: resourceWebhookUpdate,
		DeleteWithoutTimeout: resourceWebhookDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"branch_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z/_.-]{1,255}$`), "should be not be more than 255 letters, numbers, and the symbols /_.-"),
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrURL: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	input := &amplify.CreateWebhookInput{
		AppId:      aws.String(d.Get("app_id").(string)),
		BranchName: aws.String(d.Get("branch_name").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateWebhook(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Amplify Webhook: %s", err)
	}

	d.SetId(aws.ToString(output.Webhook.WebhookId))

	return append(diags, resourceWebhookRead(ctx, d, meta)...)
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	webhook, err := findWebhookByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Amplify Webhook (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Amplify Webhook (%s): %s", d.Id(), err)
	}

	webhookArn := aws.ToString(webhook.WebhookArn)
	arn, err := arn.Parse(webhookArn)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// arn:${Partition}:amplify:${Region}:${Account}:apps/${AppId}/webhooks/${WebhookId}
	parts := strings.Split(arn.Resource, "/")

	if len(parts) != 4 {
		return sdkdiag.AppendErrorf(diags, "unexpected format for ARN resource (%s)", arn.Resource)
	}

	d.Set("app_id", parts[1])
	d.Set(names.AttrARN, webhookArn)
	d.Set("branch_name", webhook.BranchName)
	d.Set(names.AttrDescription, webhook.Description)
	d.Set(names.AttrURL, webhook.WebhookUrl)

	return diags
}

func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	input := &amplify.UpdateWebhookInput{
		WebhookId: aws.String(d.Id()),
	}

	if d.HasChange("branch_name") {
		input.BranchName = aws.String(d.Get("branch_name").(string))
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	_, err := conn.UpdateWebhook(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Amplify Webhook (%s): %s", d.Id(), err)
	}

	return append(diags, resourceWebhookRead(ctx, d, meta)...)
}

func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	log.Printf("[DEBUG] Deleting Amplify Webhook: %s", d.Id())
	_, err := conn.DeleteWebhook(ctx, &amplify.DeleteWebhookInput{
		WebhookId: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Amplify Webhook (%s): %s", d.Id(), err)
	}

	return diags
}

func findWebhookByID(ctx context.Context, conn *amplify.Client, id string) (*types.Webhook, error) {
	input := &amplify.GetWebhookInput{
		WebhookId: aws.String(id),
	}

	output, err := conn.GetWebhook(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Webhook == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Webhook, nil
}
