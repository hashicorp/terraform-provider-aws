// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codepipeline

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codepipeline_webhook", name="Webhook")
// @Tags(identifierAttribute="id")
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.WebhookAuthenticationType](),
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
			names.AttrFilter: {
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
			names.AttrName: {
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
			names.AttrURL: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodePipelineClient(ctx)

	authType := types.WebhookAuthenticationType(d.Get("authentication").(string))
	name := d.Get(names.AttrName).(string)
	input := &codepipeline.PutWebhookInput{
		Tags: getTagsIn(ctx),
		Webhook: &types.WebhookDefinition{
			Authentication: authType,
			// "missing required field, PutWebhookInput.Webhook.AuthenticationConfiguration".
			AuthenticationConfiguration: &types.WebhookAuthConfiguration{},
			Filters:                     expandWebhookFilterRules(d.Get(names.AttrFilter).(*schema.Set)),
			Name:                        aws.String(name),
			TargetAction:                aws.String(d.Get("target_action").(string)),
			TargetPipeline:              aws.String(d.Get("target_pipeline").(string)),
		},
	}

	if v, ok := d.GetOk("authentication_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Webhook.AuthenticationConfiguration = expandWebhookAuthConfiguration(authType, v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.PutWebhook(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodePipeline Webhook (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Webhook.Arn))

	return append(diags, resourceWebhookRead(ctx, d, meta)...)
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodePipelineClient(ctx)

	webhook, err := findWebhookByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodePipeline Webhook %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodePipeline Webhook (%s): %s", d.Id(), err)
	}

	webhookDef := webhook.Definition
	d.Set(names.AttrARN, webhook.Arn)
	d.Set("authentication", webhookDef.Authentication)
	if err := d.Set("authentication_configuration", flattenWebhookAuthConfiguration(webhookDef.AuthenticationConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting authentication_configuration: %s", err)
	}
	if err := d.Set(names.AttrFilter, flattenWebhookFilterRules(webhookDef.Filters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting filter: %s", err)
	}
	d.Set(names.AttrName, webhookDef.Name)
	d.Set("target_action", webhookDef.TargetAction)
	d.Set("target_pipeline", webhookDef.TargetPipeline)
	d.Set(names.AttrURL, webhook.Url)

	setTagsOut(ctx, webhook.Tags)

	return diags
}

func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodePipelineClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		authType := types.WebhookAuthenticationType(d.Get("authentication").(string))
		input := &codepipeline.PutWebhookInput{
			Webhook: &types.WebhookDefinition{
				Authentication: authType,
				// "missing required field, PutWebhookInput.Webhook.AuthenticationConfiguration".
				AuthenticationConfiguration: &types.WebhookAuthConfiguration{},
				Filters:                     expandWebhookFilterRules(d.Get(names.AttrFilter).(*schema.Set)),
				Name:                        aws.String(d.Get(names.AttrName).(string)),
				TargetAction:                aws.String(d.Get("target_action").(string)),
				TargetPipeline:              aws.String(d.Get("target_pipeline").(string)),
			},
		}

		if v, ok := d.GetOk("authentication_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Webhook.AuthenticationConfiguration = expandWebhookAuthConfiguration(authType, v.([]interface{})[0].(map[string]interface{}))
		}

		_, err := conn.PutWebhook(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodePipeline Webhook (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWebhookRead(ctx, d, meta)...)
}

func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodePipelineClient(ctx)

	log.Printf("[INFO] Deleting CodePipeline Webhook: %s", d.Id())
	_, err := conn.DeleteWebhook(ctx, &codepipeline.DeleteWebhookInput{
		Name: aws.String(d.Get(names.AttrName).(string)),
	})

	if errs.IsA[*types.WebhookNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodePipeline Webhook (%s): %s", d.Id(), err)
	}

	return diags
}

func findWebhookByARN(ctx context.Context, conn *codepipeline.Client, arn string) (*types.ListWebhookItem, error) {
	input := &codepipeline.ListWebhooksInput{}

	return findWebhook(ctx, conn, input, func(v *types.ListWebhookItem) bool {
		return arn == aws.ToString(v.Arn)
	})
}

func findWebhook(ctx context.Context, conn *codepipeline.Client, input *codepipeline.ListWebhooksInput, filter tfslices.Predicate[*types.ListWebhookItem]) (*types.ListWebhookItem, error) {
	output, err := findWebhooks(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findWebhooks(ctx context.Context, conn *codepipeline.Client, input *codepipeline.ListWebhooksInput, filter tfslices.Predicate[*types.ListWebhookItem]) ([]*types.ListWebhookItem, error) {
	var output []*types.ListWebhookItem

	pages := codepipeline.NewListWebhooksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Webhooks {
			v := v
			if v := &v; filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func flattenWebhookFilterRules(filters []types.WebhookFilterRule) []interface{} {
	results := []interface{}{}
	for _, filter := range filters {
		f := map[string]interface{}{
			"json_path":    aws.ToString(filter.JsonPath),
			"match_equals": aws.ToString(filter.MatchEquals),
		}
		results = append(results, f)
	}

	return results
}

func flattenWebhookAuthConfiguration(authConfig *types.WebhookAuthConfiguration) []interface{} {
	conf := map[string]interface{}{}
	if authConfig.AllowedIPRange != nil {
		conf["allowed_ip_range"] = aws.ToString(authConfig.AllowedIPRange)
	}

	if authConfig.SecretToken != nil {
		conf["secret_token"] = aws.ToString(authConfig.SecretToken)
	}

	var results []interface{}
	if len(conf) > 0 {
		results = append(results, conf)
	}

	return results
}

func expandWebhookFilterRules(filters *schema.Set) []types.WebhookFilterRule {
	var rules []types.WebhookFilterRule

	for _, f := range filters.List() {
		r := f.(map[string]interface{})
		filter := types.WebhookFilterRule{
			JsonPath:    aws.String(r["json_path"].(string)),
			MatchEquals: aws.String(r["match_equals"].(string)),
		}

		rules = append(rules, filter)
	}

	return rules
}

func expandWebhookAuthConfiguration(authType types.WebhookAuthenticationType, authConfig map[string]interface{}) *types.WebhookAuthConfiguration {
	var conf types.WebhookAuthConfiguration

	switch authType {
	case types.WebhookAuthenticationTypeIp:
		conf.AllowedIPRange = aws.String(authConfig["allowed_ip_range"].(string))
	case types.WebhookAuthenticationTypeGithubHmac:
		conf.SecretToken = aws.String(authConfig["secret_token"].(string))
	}

	return &conf
}
