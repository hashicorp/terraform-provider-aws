package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsAmplifyWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAmplifyWebhookCreate,
		Read:   resourceAwsAmplifyWebhookRead,
		Update: resourceAwsAmplifyWebhookUpdate,
		Delete: resourceAwsAmplifyWebhookDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"app_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"branch_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9/_.-]+$`), "should only contain letters, numbers, and the symbols /_.-"),
				),
			},
			"description": {
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

func resourceAwsAmplifyWebhookCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Print("[DEBUG] Creating Amplify Webhook")

	params := &amplify.CreateWebhookInput{
		AppId:      aws.String(d.Get("app_id").(string)),
		BranchName: aws.String(d.Get("branch_name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	resp, err := conn.CreateWebhook(params)
	if err != nil {
		return fmt.Errorf("Error creating Amplify Webhook: %s", err)
	}

	d.SetId(*resp.Webhook.WebhookId)

	return resourceAwsAmplifyWebhookRead(d, meta)
}

func resourceAwsAmplifyWebhookRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Reading Amplify Webhook: %s", d.Id())

	resp, err := conn.GetWebhook(&amplify.GetWebhookInput{
		WebhookId: aws.String(d.Id()),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == amplify.ErrCodeNotFoundException {
			log.Printf("[WARN] Amplify Webhook (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	s := strings.Split(aws.StringValue(resp.Webhook.WebhookArn), "/")
	app_id := s[1]

	d.Set("app_id", app_id)
	d.Set("branch_name", resp.Webhook.BranchName)
	d.Set("description", resp.Webhook.Description)
	d.Set("arn", resp.Webhook.WebhookArn)
	d.Set("url", resp.Webhook.WebhookUrl)

	return nil
}

func resourceAwsAmplifyWebhookUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Updating Amplify Webhook: %s", d.Id())

	params := &amplify.UpdateWebhookInput{
		WebhookId: aws.String(d.Id()),
	}

	if d.HasChange("branch_name") {
		params.BranchName = aws.String(d.Get("branch_name").(string))
	}

	if d.HasChange("description") {
		params.Description = aws.String(d.Get("description").(string))
	}

	_, err := conn.UpdateWebhook(params)
	if err != nil {
		return fmt.Errorf("Error updating Amplify Webhook: %s", err)
	}

	return resourceAwsAmplifyWebhookRead(d, meta)
}

func resourceAwsAmplifyWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Deleting Amplify Webhook: %s", d.Id())

	params := &amplify.DeleteWebhookInput{
		WebhookId: aws.String(d.Id()),
	}

	_, err := conn.DeleteWebhook(params)
	if err != nil {
		return fmt.Errorf("Error deleting Amplify Webhook: %s", err)
	}

	return nil
}
