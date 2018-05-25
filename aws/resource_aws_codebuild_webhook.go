package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCodeBuildWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeBuildWebhookCreate,
		Read:   resourceAwsCodeBuildWebhookRead,
		Delete: resourceAwsCodeBuildWebhookDelete,
		Update: resourceAwsCodeBuildWebhookUpdate,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"branch_filter": {
				Type:     schema.TypeString,
				Optional: true,
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

	_, err := conn.CreateWebhook(&codebuild.CreateWebhookInput{
		ProjectName:  aws.String(d.Get("project_name").(string)),
		BranchFilter: aws.String(d.Get("branch_filter").(string)),
	})
	if err != nil {
		return err
	}

	d.SetId(d.Get("project_name").(string))

	return resourceAwsCodeBuildWebhookRead(d, meta)
}

func resourceAwsCodeBuildWebhookRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	resp, err := conn.BatchGetProjects(&codebuild.BatchGetProjectsInput{
		Names: []*string{
			aws.String(d.Id()),
		},
	})

	if err != nil {
		return err
	}

	if len(resp.Projects) == 0 {
		log.Printf("[WARN] CodeBuild Project %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	project := resp.Projects[0]
	d.Set("branch_filter", project.Webhook.BranchFilter)
	d.Set("payload_url", project.Webhook.PayloadUrl)
	d.Set("project_name", project.Name)
	d.Set("secret", project.Webhook.Secret)
	d.Set("url", project.Webhook.Url)

	return nil
}

func resourceAwsCodeBuildWebhookUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	_, err := conn.UpdateWebhook(&codebuild.UpdateWebhookInput{
		ProjectName:  aws.String(d.Id()),
		BranchFilter: aws.String(d.Get("branch_filter").(string)),
		RotateSecret: aws.Bool(false),
	})

	if err != nil {
		return err
	}

	return resourceAwsCodeBuildWebhookRead(d, meta)
}

func resourceAwsCodeBuildWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	_, err := conn.DeleteWebhook(&codebuild.DeleteWebhookInput{
		ProjectName: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, codebuild.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}
