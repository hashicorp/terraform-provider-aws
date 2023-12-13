// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codepipeline

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codepipeline_webhook", name="Webhook")
// @Tags(identifierAttribute="id")
func ResourceWebhook() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWebhookCreate,
		ReadWithoutTimeout:   resourceWebhookRead,
		UpdateWithoutTimeout: resourceWebhookUpdate,
		DeleteWithoutTimeout: resourceWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice(codepipeline.WebhookAuthenticationType_Values(), false),
			},
			"authentication_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				MinItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"secret_token": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringLenBetween(1, 100),
						},
						"allowed_ip_range": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsCIDRNetwork(0, 32),
						},
					},
				},
			},
			"filter": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"json_path": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 150),
						},

						"match_equals": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 150),
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_.@-]+`), ""),
				),
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_action": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"target_pipeline": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	authType := d.Get("authentication").(string)
	var authConfig map[string]interface{}
	if v, ok := d.GetOk("authentication_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		authConfig = v.([]interface{})[0].(map[string]interface{})
	}

	request := &codepipeline.PutWebhookInput{
		Webhook: &codepipeline.WebhookDefinition{
			Authentication:              aws.String(authType),
			Filters:                     extractWebhookRules(d.Get("filter").(*schema.Set)),
			Name:                        aws.String(d.Get("name").(string)),
			TargetAction:                aws.String(d.Get("target_action").(string)),
			TargetPipeline:              aws.String(d.Get("target_pipeline").(string)),
			AuthenticationConfiguration: extractWebhookAuthConfig(authType, authConfig),
		},
		Tags: getTagsIn(ctx),
	}

	webhook, err := conn.PutWebhookWithContext(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating webhook: %s", err)
	}

	d.SetId(aws.StringValue(webhook.Webhook.Arn))

	return append(diags, resourceWebhookRead(ctx, d, meta)...)
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	arn := d.Id()
	webhook, err := GetWebhook(ctx, conn, arn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.CodePipeline, create.ErrActionReading, ResNameWebhook, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CodePipeline, create.ErrActionReading, ResNameWebhook, d.Id(), err)
	}

	webhookDef := webhook.Definition

	name := aws.StringValue(webhookDef.Name)
	if name == "" {
		return sdkdiag.AppendErrorf(diags, "Webhook not found: %s", arn)
	}

	d.Set("name", name)
	d.Set("url", webhook.Url)
	d.Set("target_action", webhookDef.TargetAction)
	d.Set("target_pipeline", webhookDef.TargetPipeline)
	d.Set("authentication", webhookDef.Authentication)
	d.Set("arn", webhook.Arn)

	if err := d.Set("authentication_configuration", flattenWebhookAuthenticationConfiguration(webhookDef.AuthenticationConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting authentication_configuration: %s", err)
	}

	if err := d.Set("filter", flattenWebhookFilters(webhookDef.Filters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting filter: %s", err)
	}

	setTagsOut(ctx, webhook.Tags)

	return diags
}

func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	if d.HasChangesExcept("tags_all", "tags", "register_with_third_party") {
		authType := d.Get("authentication").(string)

		var authConfig map[string]interface{}
		if v, ok := d.GetOk("authentication_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			authConfig = v.([]interface{})[0].(map[string]interface{})
		}

		request := &codepipeline.PutWebhookInput{
			Webhook: &codepipeline.WebhookDefinition{
				Authentication:              aws.String(authType),
				Filters:                     extractWebhookRules(d.Get("filter").(*schema.Set)),
				Name:                        aws.String(d.Get("name").(string)),
				TargetAction:                aws.String(d.Get("target_action").(string)),
				TargetPipeline:              aws.String(d.Get("target_pipeline").(string)),
				AuthenticationConfiguration: extractWebhookAuthConfig(authType, authConfig),
			},
		}

		_, err := conn.PutWebhookWithContext(ctx, request)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating webhook: %s", err)
		}
	}

	return append(diags, resourceWebhookRead(ctx, d, meta)...)
}

func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)
	name := d.Get("name").(string)

	input := codepipeline.DeleteWebhookInput{
		Name: &name,
	}
	_, err := conn.DeleteWebhookWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Could not delete webhook: %s", err)
	}

	return diags
}

func GetWebhook(ctx context.Context, conn *codepipeline.CodePipeline, arn string) (*codepipeline.ListWebhookItem, error) {
	var nextToken string

	for {
		input := &codepipeline.ListWebhooksInput{
			MaxResults: aws.Int64(60),
		}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}

		out, err := conn.ListWebhooksWithContext(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, w := range out.Webhooks {
			if arn == aws.StringValue(w.Arn) {
				return w, nil
			}
		}

		if out.NextToken == nil {
			break
		}

		nextToken = aws.StringValue(out.NextToken)
	}

	return nil, &retry.NotFoundError{
		Message: fmt.Sprintf("No webhook with ARN %s found", arn),
	}
}

func flattenWebhookFilters(filters []*codepipeline.WebhookFilterRule) []interface{} {
	results := []interface{}{}
	for _, filter := range filters {
		f := map[string]interface{}{
			"json_path":    aws.StringValue(filter.JsonPath),
			"match_equals": aws.StringValue(filter.MatchEquals),
		}
		results = append(results, f)
	}

	return results
}

func flattenWebhookAuthenticationConfiguration(authConfig *codepipeline.WebhookAuthConfiguration) []interface{} {
	conf := map[string]interface{}{}
	if authConfig.AllowedIPRange != nil {
		conf["allowed_ip_range"] = aws.StringValue(authConfig.AllowedIPRange)
	}

	if authConfig.SecretToken != nil {
		conf["secret_token"] = aws.StringValue(authConfig.SecretToken)
	}

	var results []interface{}
	if len(conf) > 0 {
		results = append(results, conf)
	}

	return results
}

func extractWebhookRules(filters *schema.Set) []*codepipeline.WebhookFilterRule {
	var rules []*codepipeline.WebhookFilterRule

	for _, f := range filters.List() {
		r := f.(map[string]interface{})
		filter := codepipeline.WebhookFilterRule{
			JsonPath:    aws.String(r["json_path"].(string)),
			MatchEquals: aws.String(r["match_equals"].(string)),
		}

		rules = append(rules, &filter)
	}

	return rules
}

func extractWebhookAuthConfig(authType string, authConfig map[string]interface{}) *codepipeline.WebhookAuthConfiguration {
	var conf codepipeline.WebhookAuthConfiguration

	switch authType {
	case codepipeline.WebhookAuthenticationTypeIp:
		conf.AllowedIPRange = aws.String(authConfig["allowed_ip_range"].(string))
	case codepipeline.WebhookAuthenticationTypeGithubHmac:
		conf.SecretToken = aws.String(authConfig["secret_token"].(string))
	}

	return &conf
}
