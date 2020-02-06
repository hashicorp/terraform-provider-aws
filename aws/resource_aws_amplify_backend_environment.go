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

func resourceAwsAmplifyBackendEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAmplifyBackendEnvironmentCreate,
		Read:   resourceAwsAmplifyBackendEnvironmentRead,
		Delete: resourceAwsAmplifyBackendEnvironmentDelete,
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
			"deployment_artifacts": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]+$`), "should not contain special characters"),
				),
			},
			"environment_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 10),
					validation.StringMatch(regexp.MustCompile(`^[a-z]+$`), "should only contain lowercase alphabets"),
				),
			},
			"stack_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]+$`), "should not contain special characters"),
				),
			},
		},
	}
}

func resourceAwsAmplifyBackendEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Print("[DEBUG] Creating Amplify BackendEnvironment")

	params := &amplify.CreateBackendEnvironmentInput{
		AppId:           aws.String(d.Get("app_id").(string)),
		EnvironmentName: aws.String(d.Get("environment_name").(string)),
	}

	if v, ok := d.GetOk("deployment_artifacts"); ok {
		params.DeploymentArtifacts = aws.String(v.(string))
	}

	if v, ok := d.GetOk("stack_name"); ok {
		params.StackName = aws.String(v.(string))
	}

	resp, err := conn.CreateBackendEnvironment(params)
	if err != nil {
		return fmt.Errorf("Error creating Amplify BackendEnvironment: %s", err)
	}

	arn := *resp.BackendEnvironment.BackendEnvironmentArn
	d.SetId(arn[strings.Index(arn, ":apps/")+len(":apps/"):])

	return resourceAwsAmplifyBackendEnvironmentRead(d, meta)
}

func resourceAwsAmplifyBackendEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Reading Amplify BackendEnvironment: %s", d.Id())

	s := strings.Split(d.Id(), "/")
	app_id := s[0]
	environment_name := s[2]

	resp, err := conn.GetBackendEnvironment(&amplify.GetBackendEnvironmentInput{
		AppId:           aws.String(app_id),
		EnvironmentName: aws.String(environment_name),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == amplify.ErrCodeNotFoundException {
			log.Printf("[WARN] Amplify BackendEnvironment (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("app_id", app_id)
	d.Set("arn", resp.BackendEnvironment.BackendEnvironmentArn)
	d.Set("deployment_artifacts", resp.BackendEnvironment.DeploymentArtifacts)
	d.Set("environment_name", resp.BackendEnvironment.EnvironmentName)
	d.Set("stack_name", resp.BackendEnvironment.StackName)

	return nil
}

func resourceAwsAmplifyBackendEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Deleting Amplify BackendEnvironment: %s", d.Id())

	s := strings.Split(d.Id(), "/")
	app_id := s[0]
	environment_name := s[2]

	params := &amplify.DeleteBackendEnvironmentInput{
		AppId:           aws.String(app_id),
		EnvironmentName: aws.String(environment_name),
	}

	_, err := conn.DeleteBackendEnvironment(params)
	if err != nil {
		return fmt.Errorf("Error deleting Amplify BackendEnvironment: %s", err)
	}

	return nil
}
