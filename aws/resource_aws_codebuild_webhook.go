package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCodeBuildWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeBuildWebhookCreate,
		Read:   resourceAwsCodeBuildWebhookRead,
		Delete: resourceAwsCodeBuildWebhookDelete,

		Schema: map[string]*schema.Schema{
			"branch_filter": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"project_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"payload_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCodeBuildWebhookCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	params := &codebuild.CreateWebhookInput{
		ProjectName: aws.String(d.Get("project_name").(string)),
	}

	if v, ok := d.GetOk("branch_filter"); ok {
		params.BranchFilter = aws.String(v.(string))
	}

	resp, err := conn.CreateWebhook(params)

	if err != nil {
		return fmt.Errorf("[ERROR] Error creating CodeBuild Webhook: %s", err)
	}

	// Limit on one webhook per project
	d.SetId(d.Get("project_name").(string))

	d.Set("payload_url", *resp.Webhook.PayloadUrl)
	d.Set("secret", *resp.Webhook.Secret)
	d.Set("url", *resp.Webhook.Url)

	return nil
}

func resourceAwsCodeBuildWebhookRead(d *schema.ResourceData, meta interface{}) error {
	// Codebuild doesn't support getting webhooks
	return nil
}

func resourceAwsCodeBuildWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	params := &codebuild.DeleteWebhookInput{
		ProjectName: aws.String(d.Get("project_name").(string)),
	}

	_, err := conn.DeleteWebhook(params)

	if err != nil {
		return fmt.Errorf("[ERROR] Error deleting CodeBuild Webhook: %s", err)
	}

	return nil
}
