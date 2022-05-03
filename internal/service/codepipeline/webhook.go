package codepipeline

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceWebhookCreate,
		Read:   resourceWebhookRead,
		Update: resourceWebhookUpdate,
		Delete: resourceWebhookDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
					validation.StringMatch(regexp.MustCompile(`[A-Za-z0-9.@\-_]+`), ""),
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWebhookCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodePipelineConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
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
		Tags: Tags(tags.IgnoreAWS()),
	}

	webhook, err := conn.PutWebhook(request)
	if err != nil {
		return fmt.Errorf("Error creating webhook: %w", err)
	}

	d.SetId(aws.StringValue(webhook.Webhook.Arn))

	return resourceWebhookRead(d, meta)
}

func resourceWebhookRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodePipelineConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn := d.Id()
	webhook, err := GetWebhook(conn, arn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		names.LogNotFoundRemoveState(names.CodePipeline, names.ErrActionReading, ResWebhook, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CodePipeline, names.ErrActionReading, ResWebhook, d.Id(), err)
	}

	webhookDef := webhook.Definition

	name := aws.StringValue(webhookDef.Name)
	if name == "" {
		return fmt.Errorf("Webhook not found: %s", arn)
	}

	d.Set("name", name)
	d.Set("url", webhook.Url)
	d.Set("target_action", webhookDef.TargetAction)
	d.Set("target_pipeline", webhookDef.TargetPipeline)
	d.Set("authentication", webhookDef.Authentication)
	d.Set("arn", webhook.Arn)

	if err := d.Set("authentication_configuration", flattenWebhookAuthenticationConfiguration(webhookDef.AuthenticationConfiguration)); err != nil {
		return fmt.Errorf("error setting authentication_configuration: %w", err)
	}

	if err := d.Set("filter", flattenWebhookFilters(webhookDef.Filters)); err != nil {
		return fmt.Errorf("error setting filter: %w", err)
	}

	tags := KeyValueTags(webhook.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceWebhookUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodePipelineConn

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

		_, err := conn.PutWebhook(request)
		if err != nil {
			return fmt.Errorf("Error updating webhook: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating CodePipeline Webhook (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceWebhookRead(d, meta)
}

func resourceWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodePipelineConn
	name := d.Get("name").(string)

	input := codepipeline.DeleteWebhookInput{
		Name: &name,
	}
	_, err := conn.DeleteWebhook(&input)

	if err != nil {
		return fmt.Errorf("Could not delete webhook: %w", err)
	}

	return nil
}

func GetWebhook(conn *codepipeline.CodePipeline, arn string) (*codepipeline.ListWebhookItem, error) {
	var nextToken string

	for {
		input := &codepipeline.ListWebhooksInput{
			MaxResults: aws.Int64(60),
		}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}

		out, err := conn.ListWebhooks(input)
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

	return nil, &resource.NotFoundError{
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
