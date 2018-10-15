package aws

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
)

func resourceAwsCodePipelineWebhook() *schema.Resource {
	return &schema.Resource{
		Create:        resourceAwsCodePipelineWebhookCreate,
		Read:          resourceAwsCodePipelineWebhookRead,
		Update:        nil,
		Delete:        resourceAwsCodePipelineWebhookDelete,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"auth": {
				Type:     schema.TypeList,
				MaxItems: 1,
				MinItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								codepipeline.WebhookAuthenticationTypeGithubHmac,
								codepipeline.WebhookAuthenticationTypeIp,
								codepipeline.WebhookAuthenticationTypeUnauthenticated,
							}, false),
						},
						"secret_token": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"allowed_ip_range": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.CIDRNetwork(0, 32),
						},
					},
				},
			},
			"filter": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"json_path": {
							Type:     schema.TypeString,
							Required: true,
						},

						"match_equals": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"target": {
				Type:     schema.TypeList,
				MaxItems: 1,
				MinItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeString,
							Required: true,
						},
						"pipeline": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func extractCodePipelineWebhookAttr(d *schema.ResourceData, attr string) (map[string]interface{}, error) {
	if v, ok := d.GetOk(attr); ok {
		l := v.([]interface{})
		if len(l) <= 0 {
			return nil, fmt.Errorf("Attribute %s is missing", attr)
		}

		data := l[0].(map[string]interface{})
		return data, nil
	}

	return nil, fmt.Errorf("Could not find attribute %s", attr)
}

func extractCodePipelineWebhookRules(filters *schema.Set) ([]*codepipeline.WebhookFilterRule, error) {
	var rules []*codepipeline.WebhookFilterRule

	for _, f := range filters.List() {
		r := f.(map[string]interface{})
		filter := codepipeline.WebhookFilterRule{
			JsonPath:    aws.String(r["json_path"].(string)),
			MatchEquals: aws.String(r["match_equals"].(string)),
		}

		rules = append(rules, &filter)
	}

	return rules, nil
}

func extractCodePipelineWebhookAuthConfig(auth map[string]interface{}) (*codepipeline.WebhookAuthConfiguration, error) {
	var authConfig codepipeline.WebhookAuthConfiguration
	switch auth["type"].(string) {
	case codepipeline.WebhookAuthenticationTypeIp:
		ipRange := auth["allowed_ip_range"].(string)
		if ipRange == "" {
			return nil, fmt.Errorf("An IP range must be set when using IP-based auth")
		}

		authConfig.AllowedIPRange = &ipRange

		break
	case codepipeline.WebhookAuthenticationTypeGithubHmac:
		secretToken := auth["secret_token"].(string)
		if secretToken == "" {
			return nil, fmt.Errorf("Secret token must be set when using GITHUB_HMAC")
		}

		authConfig.SecretToken = &secretToken
		break
	case codepipeline.WebhookAuthenticationTypeUnauthenticated:
		break
	default:
		return nil, fmt.Errorf("Invalid authentication type %s", auth["type"])
	}

	return &authConfig, nil
}

func resourceAwsCodePipelineWebhookCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codepipelineconn

	auth, err := extractCodePipelineWebhookAttr(d, "auth")
	if err != nil {
		return err
	}

	target, err := extractCodePipelineWebhookAttr(d, "target")
	if err != nil {
		return err
	}

	rules, err := extractCodePipelineWebhookRules(d.Get("filter").(*schema.Set))
	if err != nil {
		return err
	}

	authConfig, err := extractCodePipelineWebhookAuthConfig(auth)

	request := &codepipeline.PutWebhookInput{
		Webhook: &codepipeline.WebhookDefinition{
			Authentication:              aws.String(auth["type"].(string)),
			Filters:                     rules,
			Name:                        aws.String(d.Get("name").(string)),
			TargetAction:                aws.String(target["action"].(string)),
			TargetPipeline:              aws.String(target["pipeline"].(string)),
			AuthenticationConfiguration: authConfig,
		},
	}

	webhook, err := conn.PutWebhook(request)
	if err != nil {
		return fmt.Errorf("Error creating webhook: %s", err)
	}

	arn := *webhook.Webhook.Arn
	d.SetId(arn)

	url := *webhook.Webhook.Url
	d.Set("url", url)

	return resourceAwsCodePipelineWebhookRead(d, meta)
}

func getAllCodePipelineWebhooks(conn *codepipeline.CodePipeline) ([]*codepipeline.ListWebhookItem, error) {
	var webhooks []*codepipeline.ListWebhookItem
	var nextToken string

	for {
		input := &codepipeline.ListWebhooksInput{
			MaxResults: aws.Int64(int64(60)),
		}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}

		out, err := conn.ListWebhooks(input)
		if err != nil {
			return webhooks, err
		}

		webhooks = append(webhooks, out.Webhooks...)

		if out.NextToken == nil {
			break
		}

		nextToken = aws.StringValue(out.NextToken)
	}

	return webhooks, nil
}

func setCodePipelineWebhookFilters(webhook codepipeline.WebhookDefinition, d *schema.ResourceData) error {
	filters := []interface{}{}
	for _, filter := range webhook.Filters {
		f := map[string]interface{}{
			"json_path":    *filter.JsonPath,
			"match_equals": *filter.MatchEquals,
		}
		filters = append(filters, f)
	}

	if err := d.Set("filter", filters); err != nil {
		return err
	}

	return nil
}

func setCodePipelineWebhookAuthentication(webhook codepipeline.WebhookDefinition, d *schema.ResourceData) error {
	var result []interface{}

	auth := map[string]interface{}{}

	authType := *webhook.Authentication
	auth["type"] = authType

	if webhook.AuthenticationConfiguration.AllowedIPRange != nil {
		ipRange := *webhook.AuthenticationConfiguration.AllowedIPRange
		auth["allowed_ip_range"] = ipRange
	}

	if webhook.AuthenticationConfiguration.SecretToken != nil {
		secretToken := *webhook.AuthenticationConfiguration.SecretToken
		auth["secret_token"] = secretToken
	}

	result = append(result, auth)
	if err := d.Set("auth", result); err != nil {
		return err
	}

	return nil
}

func setCodePipelineWebhookTarget(webhook codepipeline.WebhookDefinition, d *schema.ResourceData) error {
	var t []interface{}
	target := map[string]interface{}{
		"action":   *webhook.TargetAction,
		"pipeline": *webhook.TargetPipeline,
	}
	t = append(t, target)

	if err := d.Set("target", t); err != nil {
		return err
	}

	return nil
}

func resourceAwsCodePipelineWebhookRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codepipelineconn
	arn := d.Id()
	webhooks, err := getAllCodePipelineWebhooks(conn)
	if err != nil {
		return fmt.Errorf("Error fetching webhooks: %s", err)
	}

	if len(webhooks) == 0 {
		return fmt.Errorf("No webhooks returned!")
	}

	var found codepipeline.WebhookDefinition
	for _, w := range webhooks {
		a := *w.Arn
		if a == arn {
			found = *w.Definition
			break
		}
	}

	name := *found.Name
	if name == "" {
		return fmt.Errorf("Webhook not found: %s", arn)
	}

	d.Set("name", name)

	if err = setCodePipelineWebhookAuthentication(found, d); err != nil {
		return err
	}

	if err = setCodePipelineWebhookFilters(found, d); err != nil {
		return err
	}

	if err = setCodePipelineWebhookTarget(found, d); err != nil {
		return err
	}

	return nil
}

func resourceAwsCodePipelineWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codepipelineconn
	name := d.Get("name").(string)

	input := codepipeline.DeleteWebhookInput{
		Name: &name,
	}
	_, err := conn.DeleteWebhook(&input)

	if err != nil {
		return fmt.Errorf("Could not delete webhook: %s", err)
	}

	d.SetId("")

	return nil
}
