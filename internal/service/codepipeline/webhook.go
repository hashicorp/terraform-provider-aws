// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codepipeline

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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
						"allowed_ip_range": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsCIDRNetwork(0, 32),
						},
						"secret_token": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringLenBetween(1, 100),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	authType := d.Get("authentication").(string)
	name := d.Get("name").(string)
	input := &codepipeline.PutWebhookInput{
		Tags: getTagsIn(ctx),
		Webhook: &codepipeline.WebhookDefinition{
			Authentication: aws.String(authType),
			Filters:        expandWebhookFilterRules(d.Get("filter").(*schema.Set)),
			Name:           aws.String(name),
			TargetAction:   aws.String(d.Get("target_action").(string)),
			TargetPipeline: aws.String(d.Get("target_pipeline").(string)),
		},
	}

	if v, ok := d.GetOk("authentication_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Webhook.AuthenticationConfiguration = expandWebhookAuthConfiguration(authType, v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.PutWebhookWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodePipeline Webhook (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Webhook.Arn))

	return append(diags, resourceWebhookRead(ctx, d, meta)...)
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	webhook, err := FindWebhookByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodePipeline Webhook %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodePipeline Webhook (%s): %s", d.Id(), err)
	}

	webhookDef := webhook.Definition
	d.Set("arn", webhook.Arn)
	d.Set("authentication", webhookDef.Authentication)
	if err := d.Set("authentication_configuration", flattenWebhookAuthConfiguration(webhookDef.AuthenticationConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting authentication_configuration: %s", err)
	}
	if err := d.Set("filter", flattenWebhookFilterRules(webhookDef.Filters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting filter: %s", err)
	}
	d.Set("name", webhookDef.Name)
	d.Set("target_action", webhookDef.TargetAction)
	d.Set("target_pipeline", webhookDef.TargetPipeline)
	d.Set("url", webhook.Url)

	setTagsOut(ctx, webhook.Tags)

	return diags
}

func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		authType := d.Get("authentication").(string)
		input := &codepipeline.PutWebhookInput{
			Webhook: &codepipeline.WebhookDefinition{
				Authentication: aws.String(authType),
				Filters:        expandWebhookFilterRules(d.Get("filter").(*schema.Set)),
				Name:           aws.String(d.Get("name").(string)),
				TargetAction:   aws.String(d.Get("target_action").(string)),
				TargetPipeline: aws.String(d.Get("target_pipeline").(string)),
			},
		}

		if v, ok := d.GetOk("authentication_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Webhook.AuthenticationConfiguration = expandWebhookAuthConfiguration(authType, v.([]interface{})[0].(map[string]interface{}))
		}

		_, err := conn.PutWebhookWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodePipeline Webhook (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWebhookRead(ctx, d, meta)...)
}

func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	log.Printf("[INFO] Deleting CodePipeline Webhook: %s", d.Id())
	_, err := conn.DeleteWebhookWithContext(ctx, &codepipeline.DeleteWebhookInput{
		Name: aws.String(d.Get("name").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodePipeline Webhook (%s): %s", d.Id(), err)
	}

	return diags
}

func FindWebhookByARN(ctx context.Context, conn *codepipeline.CodePipeline, arn string) (*codepipeline.ListWebhookItem, error) {
	input := &codepipeline.ListWebhooksInput{}

	return findWebhook(ctx, conn, input, func(v *codepipeline.ListWebhookItem) bool {
		return arn == aws.StringValue(v.Arn)
	})
}

func findWebhook(ctx context.Context, conn *codepipeline.CodePipeline, input *codepipeline.ListWebhooksInput, filter tfslices.Predicate[*codepipeline.ListWebhookItem]) (*codepipeline.ListWebhookItem, error) {
	output, err := findWebhooks(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findWebhooks(ctx context.Context, conn *codepipeline.CodePipeline, input *codepipeline.ListWebhooksInput, filter tfslices.Predicate[*codepipeline.ListWebhookItem]) ([]*codepipeline.ListWebhookItem, error) {
	var output []*codepipeline.ListWebhookItem

	err := conn.ListWebhooksPagesWithContext(ctx, input, func(page *codepipeline.ListWebhooksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Webhooks {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func flattenWebhookFilterRules(filters []*codepipeline.WebhookFilterRule) []interface{} {
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

func flattenWebhookAuthConfiguration(authConfig *codepipeline.WebhookAuthConfiguration) []interface{} {
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

func expandWebhookFilterRules(filters *schema.Set) []*codepipeline.WebhookFilterRule {
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

func expandWebhookAuthConfiguration(authType string, authConfig map[string]interface{}) *codepipeline.WebhookAuthConfiguration {
	var conf codepipeline.WebhookAuthConfiguration

	switch authType {
	case codepipeline.WebhookAuthenticationTypeIp:
		conf.AllowedIPRange = aws.String(authConfig["allowed_ip_range"].(string))
	case codepipeline.WebhookAuthenticationTypeGithubHmac:
		conf.SecretToken = aws.String(authConfig["secret_token"].(string))
	}

	return &conf
}
