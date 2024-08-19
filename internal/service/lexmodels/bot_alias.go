// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	botAliasCreateTimeout = 1 * time.Minute
	botAliasUpdateTimeout = 1 * time.Minute
	botAliasDeleteTimeout = 5 * time.Minute
)

// @SDKResource("aws_lex_bot_alias")
func ResourceBotAlias() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBotAliasCreate,
		ReadWithoutTimeout:   resourceBotAliasRead,
		UpdateWithoutTimeout: resourceBotAliasUpdate,
		DeleteWithoutTimeout: resourceBotAliasDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceBotAliasImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(botAliasCreateTimeout),
			Update: schema.DefaultTimeout(botAliasUpdateTimeout),
			Delete: schema.DefaultTimeout(botAliasDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bot_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validBotName,
			},
			"bot_version": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validBotVersion,
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"conversation_logs": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrIAMRoleARN: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(20, 2048),
								verify.ValidARN,
							),
						},
						// Currently the API docs do not list a min and max for this list.
						// https://docs.aws.amazon.com/lex/latest/dg/API_PutBotAlias.html#lex-PutBotAlias-request-conversationLogs
						"log_settings": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     logSettings,
						},
					},
				},
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			names.AttrLastUpdatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validBotAliasName,
			},
		},
	}
}

var validBotAliasName = validation.All(
	validation.StringLenBetween(1, 100),
	validation.StringMatch(regexache.MustCompile(`^([A-Za-z]_?)+$`), ""),
)

func resourceBotAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn(ctx)

	botName := d.Get("bot_name").(string)
	botAliasName := d.Get(names.AttrName).(string)
	id := fmt.Sprintf("%s:%s", botName, botAliasName)

	input := &lexmodelbuildingservice.PutBotAliasInput{
		BotName:     aws.String(botName),
		BotVersion:  aws.String(d.Get("bot_version").(string)),
		Description: aws.String(d.Get(names.AttrDescription).(string)),
		Name:        aws.String(botAliasName),
	}

	if v, ok := d.GetOk("conversation_logs"); ok {
		conversationLogs, err := expandConversationLogs(v)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Lex Model Bot Alias (%s): %s", id, err)
		}
		input.ConversationLogs = conversationLogs
	}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		output, err := conn.PutBotAliasWithContext(ctx, input)

		input.Checksum = output.Checksum
		// IAM eventual consistency
		if tfawserr.ErrMessageContains(err, lexmodelbuildingservice.ErrCodeBadRequestException, "Lex can't access your IAM role") {
			return retry.RetryableError(err)
		}
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return retry.RetryableError(fmt.Errorf("%q bot alias still creating, another operation is pending: %w", id, err))
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) { // nosemgrep:ci.helper-schema-TimeoutError-check-doesnt-return-output
		_, err = conn.PutBotAliasWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lex Model Bot Alias (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceBotAliasRead(ctx, d, meta)...)
}

func resourceBotAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn(ctx)

	resp, err := conn.GetBotAliasWithContext(ctx, &lexmodelbuildingservice.GetBotAliasInput{
		BotName: aws.String(d.Get("bot_name").(string)),
		Name:    aws.String(d.Get(names.AttrName).(string)),
	})
	if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
		log.Printf("[WARN] Bot alias (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting bot alias '%s': %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "lex",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("bot:%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())

	d.Set("bot_name", resp.BotName)
	d.Set("bot_version", resp.BotVersion)
	d.Set("checksum", resp.Checksum)
	d.Set(names.AttrCreatedDate, resp.CreatedDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, resp.Description)
	d.Set(names.AttrLastUpdatedDate, resp.LastUpdatedDate.Format(time.RFC3339))
	d.Set(names.AttrName, resp.Name)

	if resp.ConversationLogs != nil {
		d.Set("conversation_logs", flattenConversationLogs(resp.ConversationLogs))
	}

	return diags
}

func resourceBotAliasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn(ctx)

	input := &lexmodelbuildingservice.PutBotAliasInput{
		BotName:    aws.String(d.Get("bot_name").(string)),
		BotVersion: aws.String(d.Get("bot_version").(string)),
		Checksum:   aws.String(d.Get("checksum").(string)),
		Name:       aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("conversation_logs"); ok {
		conversationLogs, err := expandConversationLogs(v)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Lex Model Bot Alias (%s): %s", d.Id(), err)
		}
		input.ConversationLogs = conversationLogs
	}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *retry.RetryError {
		_, err := conn.PutBotAliasWithContext(ctx, input)

		// IAM eventual consistency
		if tfawserr.ErrMessageContains(err, lexmodelbuildingservice.ErrCodeBadRequestException, "Lex can't access your IAM role") {
			return retry.RetryableError(err)
		}
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return retry.RetryableError(fmt.Errorf("%q bot alias still updating", d.Id()))
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutBotAliasWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Lex Model Bot Alias (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBotAliasRead(ctx, d, meta)...)
}

func resourceBotAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsConn(ctx)

	botName := d.Get("bot_name").(string)
	botAliasName := d.Get(names.AttrName).(string)

	input := &lexmodelbuildingservice.DeleteBotAliasInput{
		BotName: aws.String(botName),
		Name:    aws.String(botAliasName),
	}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := conn.DeleteBotAliasWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return retry.RetryableError(fmt.Errorf("%q: bot alias still deleting", d.Id()))
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteBotAliasWithContext(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lex Model Bot Alias (%s): %s", d.Id(), err)
	}

	if _, err := waitBotAliasDeleted(ctx, conn, botAliasName, botName); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lex Model Bot Alias (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func resourceBotAliasImport(ctx context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid Lex Bot Alias resource id '%s', expected BOT_NAME:BOT_ALIAS_NAME", d.Id())
	}

	d.Set("bot_name", parts[0])
	d.Set(names.AttrName, parts[1])

	return []*schema.ResourceData{d}, nil
}

var logSettings = &schema.Resource{
	Schema: map[string]*schema.Schema{
		names.AttrDestination: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.Destination_Values(), false),
		},
		names.AttrKMSKeyARN: {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(20, 2048),
				verify.ValidARN,
			),
		},
		"log_type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.LogType_Values(), false),
		},
		names.AttrResourceARN: {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 2048),
				verify.ValidARN,
			),
		},
		"resource_prefix": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

func flattenConversationLogs(response *lexmodelbuildingservice.ConversationLogsResponse) (flattened []map[string]interface{}) {
	return []map[string]interface{}{
		{
			names.AttrIAMRoleARN: aws.StringValue(response.IamRoleArn),
			"log_settings":       flattenLogSettings(response.LogSettings),
		},
	}
}

func expandConversationLogs(rawObject interface{}) (*lexmodelbuildingservice.ConversationLogsRequest, error) {
	request := rawObject.([]interface{})[0].(map[string]interface{})

	logSettings, err := expandLogSettings(request["log_settings"].(*schema.Set).List())
	if err != nil {
		return nil, err
	}
	return &lexmodelbuildingservice.ConversationLogsRequest{
		IamRoleArn:  aws.String(request[names.AttrIAMRoleARN].(string)),
		LogSettings: logSettings,
	}, nil
}

func flattenLogSettings(responses []*lexmodelbuildingservice.LogSettingsResponse) (flattened []map[string]interface{}) {
	for _, response := range responses {
		flattened = append(flattened, map[string]interface{}{
			names.AttrDestination: response.Destination,
			names.AttrKMSKeyARN:   response.KmsKeyArn,
			"log_type":            response.LogType,
			names.AttrResourceARN: response.ResourceArn,
			"resource_prefix":     response.ResourcePrefix,
		})
	}
	return
}

func expandLogSettings(rawValues []interface{}) ([]*lexmodelbuildingservice.LogSettingsRequest, error) {
	requests := make([]*lexmodelbuildingservice.LogSettingsRequest, 0, len(rawValues))

	for _, rawValue := range rawValues {
		value, ok := rawValue.(map[string]interface{})
		if !ok {
			continue
		}
		destination := value[names.AttrDestination].(string)
		request := &lexmodelbuildingservice.LogSettingsRequest{
			Destination: aws.String(destination),
			LogType:     aws.String(value["log_type"].(string)),
			ResourceArn: aws.String(value[names.AttrResourceARN].(string)),
		}

		if v, ok := value[names.AttrKMSKeyARN]; ok && v != "" {
			if destination != lexmodelbuildingservice.DestinationS3 {
				return nil, fmt.Errorf("`kms_key_arn` cannot be specified when `destination` is %q", destination)
			}
			request.KmsKeyArn = aws.String(value[names.AttrKMSKeyARN].(string))
		}

		requests = append(requests, request)
	}

	return requests, nil
}
