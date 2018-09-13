package aws

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"

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
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"secret_token": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"allowed_ip_range": {
							Type:     schema.TypeList,
							Default:  "0.0.0.0/0",
							Optional: true,
						},
					},
				},
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
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
			"target": {
				Type:     schema.TypeMap,
				Optional: true,
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

func resourceAwsCodePipelineWebhookCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codepipelineconn

	auth := d.Get("auth").(map[string]interface{})
	target := d.Get("target").(map[string]interface{})
	filters := d.Get("filter").(*schema.Set)

	var rules []*codepipeline.WebhookFilterRule

	for _, f := range filters.List() {
		r := f.(map[string]interface{})
		filter := codepipeline.WebhookFilterRule{
			JsonPath:    aws.String(r["json_path"].(string)),
			MatchEquals: aws.String(r["match_equals"].(string)),
		}

		rules = append(rules, &filter)
	}

	request := &codepipeline.PutWebhookInput{
		Webhook: &codepipeline.WebhookDefinition{
			Authentication: aws.String(auth["type"].(string)),
			Filters:        rules,
			Name:           aws.String(d.Get("name").(string)),
			TargetAction:   aws.String(target["action"].(string)),
			TargetPipeline: aws.String(target["pipeline"].(string)),
		},
	}

	var authConfig codepipeline.WebhookAuthConfiguration
	switch auth["type"] {
	case "IP":
		ipRange := auth["allowed_ip_range"].(string)
		authConfig.AllowedIPRange = &ipRange
		break
	case "GITHUB_HMAC":
		secretToken := auth["secret_token"].(string)
		authConfig.SecretToken = &secretToken
		break
	case "UNAUTHENTICATED":
		break
	default:
		return fmt.Errorf("Invalid authentication type %s", auth["type"])
	}

	request.Webhook.AuthenticationConfiguration = &authConfig
	webhook, err := conn.PutWebhook(request)
	if err != nil {
		return fmt.Errorf("Error creating webhook: %s", err)
	}

	arn := *webhook.Webhook.Arn
	d.SetId(arn)

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

func setFilters(webhook codepipeline.WebhookDefinition, d *schema.ResourceData) error {
	filters := []interface{}{}
	for _, filter := range webhook.Filters {
		f := map[string]interface{}{
			"json_path":    filter.JsonPath,
			"match_equals": filter.MatchEquals,
		}
		filters = append(filters, f)
	}

	if err := d.Set("filter", filters); err != nil {
		return err
	}

	return nil
}

func setAuthentication(webhook codepipeline.WebhookDefinition, d *schema.ResourceData) error {
	auth := map[string]interface{}{
		"type": webhook.Authentication,
	}

	ipRange := *webhook.AuthenticationConfiguration.AllowedIPRange
	if ipRange != "" {
		auth["allowed_ip_range"] = ipRange
	}

	secretToken := *webhook.AuthenticationConfiguration.SecretToken
	if secretToken != "" {
		auth["secret_token"] = secretToken
	}

	if err := d.Set("auth", auth); err != nil {
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

	if err = setAuthentication(found, d); err != nil {
		return err
	}

	if err = setFilters(found, d); err != nil {
		return err
	}

	target := map[string]interface{}{
		"action":   found.TargetAction,
		"pipeline": found.TargetPipeline,
	}

	if err := d.Set("target", target); err != nil {
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

	return nil
}
