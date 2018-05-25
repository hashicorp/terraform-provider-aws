package aws

import (
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

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"branch_filter": {
				Type:     schema.TypeString,
				Optional: true,
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

	resp, err := conn.CreateWebhook(&codebuild.CreateWebhookInput{
		ProjectName:  aws.String(d.Get("name").(string)),
		BranchFilter: aws.String(d.Get("branch_filter").(string)),
	})
	if err != nil {
		return err
	}

	d.SetId(d.Get("name").(string))
	d.Set("branch_filter", d.Get("branch_filter").(string))
	d.Set("url", resp.Webhook.Url)
	return nil
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
		d.SetId("")
		return nil
	}

	project := resp.Projects[0]
	d.Set("branch_filter", project.Webhook.BranchFilter)
	d.Set("url", project.Webhook.Url)
	return nil
}

func resourceAwsCodeBuildWebhookUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	if !d.HasChange("branch_filter") {
		return nil
	}

	_, err := conn.UpdateWebhook(&codebuild.UpdateWebhookInput{
		ProjectName:  aws.String(d.Id()),
		BranchFilter: aws.String(d.Get("branch_filter").(string)),
		RotateSecret: aws.Bool(false),
	})

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsCodeBuildWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	_, err := conn.DeleteWebhook(&codebuild.DeleteWebhookInput{
		ProjectName: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, codebuild.ErrCodeResourceNotFoundException, "") {
			d.SetId("")
			return nil
		}
		return err
	}

	d.SetId("")
	return nil
}
