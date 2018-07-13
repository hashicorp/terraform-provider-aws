package aws

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
				Elem: &schema.Schema{
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

	var rules []aws.WebhookFilterRule

	for _, f := range filters.List() {
		r = f.(map[string]interface{})
		rules = append(rules, &aws.WebhookFilterRule{
			JsonPath:    aws.String(r["json_path"]),
			MatchEquals: aws.String(r["match_equals"]),
		})
	}

	request = &codepipeline.PutWebhookInput{
		Webhook: &codepipeline.WebhookDefinition{
			Authentication: &aws.String(auth["type"]),
			Filters:        &rules,
			Name:           &aws.String(d.Get("name").(string)),
			TargetAction:   &aws.String(target["action"]),
			TargetPipeline: &aws.String(target["pipeline"]),
		},
	}

	var authConfig WebhookAuthConfiguration
	switch auth["type"] {
	case "IP":
		authConfig.AllowedIPRange = auth["allowed_ip_range"]
		break
	case "GITHUB_HMAC":
		authConfig.SecretToken = auth["allowed_ip_range"]
		break
	case "UNAUTHENTICATED":
		break
	default:
		return fmt.Errorf("Invalid authentication type %s", auth["type"])
	}

	request.Webhook.AuthenticationConfiguration = &authConfig
	webhook, err := codepipeline.PutWebhook(request)
	if err != nil {
		return fmt.Errorf("Error creating webhook: %s", err)
	}

	d.SetId(webhook.Arn)

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
			return pools, err
		}
		pools = append(pools, out.Webhooks...)

		if out.NextToken == nil {
			break
		}
		nextToken = aws.StringValue(out.NextToken)
	}

	return webhooks, nil
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

	found * codepipeline.ListWebhookItem
	for _, w := range webhooks {
		if w.Arn == arn {
			found = w
			break
		}
	}

	if found.Arn == "" {
		return fmt.Errorf("Webhook not found: %s", arn)
	}

	d.Set("name", found.Definition.Name)
}

func resourceAwsCodePipelineWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codepipelineconn
	return nil
}
