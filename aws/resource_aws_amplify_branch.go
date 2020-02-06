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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAmplifyBranch() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAmplifyBranchCreate,
		Read:   resourceAwsAmplifyBranchRead,
		Update: resourceAwsAmplifyBranchUpdate,
		Delete: resourceAwsAmplifyBranchDelete,
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
			"associated_resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"backend_environment_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"basic_auth_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_basic_auth": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"password": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"username": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"branch_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9/_.-]+$`), "should only contain letters, numbers, and the symbols /_.-"),
				),
			},
			"build_spec": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"custom_domains": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination_branch": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-z0-9-]+$`), "should only contain lowercase alphabets, numbers, and -"),
				),
			},
			"enable_auto_build": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"enable_notification": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"enable_pull_request_preview": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"environment_variables": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"framework": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pull_request_environment_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"source_branch": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stage": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "NONE",
				ValidateFunc: validation.StringInSlice([]string{
					amplify.StageProduction,
					amplify.StageBeta,
					amplify.StageDevelopment,
					amplify.StageExperimental,
					amplify.StagePullRequest,
				}, false),
			},
			"ttl": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// ttl is set to "5" by default
					if old == "5" && new == "" {
						return true
					}
					return false
				},
			},
			// non-API
			"sns_topic_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsAmplifyBranchCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Print("[DEBUG] Creating Amplify Branch")

	params := &amplify.CreateBranchInput{
		AppId:      aws.String(d.Get("app_id").(string)),
		BranchName: aws.String(d.Get("branch_name").(string)),
	}

	if v, ok := d.GetOk("backend_environment_arn"); ok {
		params.BackendEnvironmentArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("basic_auth_config"); ok {
		enable, credentials := expandAmplifyBasicAuthConfig(v.([]interface{}))
		params.EnableBasicAuth = enable
		params.BasicAuthCredentials = credentials
	}

	if v, ok := d.GetOk("build_spec"); ok {
		params.BuildSpec = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("display_name"); ok {
		params.DisplayName = aws.String(v.(string))
	}

	// Note: don't use GetOk here because enable_auto_build can be false
	if v := d.Get("enable_auto_build"); v != nil {
		params.EnableAutoBuild = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("enable_notification"); ok {
		params.EnableNotification = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("enable_pull_request_preview"); ok {
		params.EnablePullRequestPreview = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("environment_variables"); ok {
		params.EnvironmentVariables = stringMapToPointers(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("framework"); ok {
		params.Framework = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pull_request_environment_name"); ok {
		params.PullRequestEnvironmentName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("stage"); ok {
		params.Stage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ttl"); ok {
		params.Ttl = aws.String(v.(string))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		params.Tags = keyvaluetags.New(v).IgnoreAws().AmplifyTags()
	}

	resp, err := conn.CreateBranch(params)
	if err != nil {
		return fmt.Errorf("Error creating Amplify Branch: %s", err)
	}

	arn := *resp.Branch.BranchArn
	d.SetId(arn[strings.Index(arn, ":apps/")+len(":apps/"):])

	return resourceAwsAmplifyBranchRead(d, meta)
}

func resourceAwsAmplifyBranchRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Reading Amplify Branch: %s", d.Id())

	s := strings.Split(d.Id(), "/")
	app_id := s[0]
	branch_name := s[2]

	resp, err := conn.GetBranch(&amplify.GetBranchInput{
		AppId:      aws.String(app_id),
		BranchName: aws.String(branch_name),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == amplify.ErrCodeNotFoundException {
			log.Printf("[WARN] Amplify Branch (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("app_id", app_id)
	d.Set("associated_resources", resp.Branch.AssociatedResources)
	d.Set("backend_environment_arn", resp.Branch.BackendEnvironmentArn)
	d.Set("arn", resp.Branch.BranchArn)
	if err := d.Set("basic_auth_config", flattenAmplifyBasicAuthConfig(resp.Branch.EnableBasicAuth, resp.Branch.BasicAuthCredentials)); err != nil {
		return fmt.Errorf("error setting basic_auth_config: %s", err)
	}
	d.Set("branch_name", resp.Branch.BranchName)
	d.Set("build_spec", resp.Branch.BuildSpec)
	d.Set("custom_domains", resp.Branch.CustomDomains)
	d.Set("description", resp.Branch.Description)
	d.Set("destination_branch", resp.Branch.DestinationBranch)
	d.Set("display_name", resp.Branch.DisplayName)
	d.Set("enable_auto_build", resp.Branch.EnableAutoBuild)
	d.Set("enable_notification", resp.Branch.EnableNotification)
	d.Set("enable_pull_request_preview", resp.Branch.EnablePullRequestPreview)
	if err := d.Set("environment_variables", aws.StringValueMap(resp.Branch.EnvironmentVariables)); err != nil {
		return fmt.Errorf("error setting environment_variables: %s", err)
	}
	d.Set("framework", resp.Branch.Framework)
	d.Set("pull_request_environment_name", resp.Branch.PullRequestEnvironmentName)
	d.Set("source_branch", resp.Branch.SourceBranch)
	d.Set("stage", resp.Branch.Stage)
	d.Set("ttl", resp.Branch.Ttl)

	// Generate SNS topic name for notification
	d.Set("sns_topic_name", fmt.Sprintf("amplify-%s_%s", app_id, branch_name))

	if err := d.Set("tags", keyvaluetags.AmplifyKeyValueTags(resp.Branch.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsAmplifyBranchUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Updating Amplify Branch: %s", d.Id())

	s := strings.Split(d.Id(), "/")
	app_id := s[0]
	branch_name := s[2]

	params := &amplify.UpdateBranchInput{
		AppId:      aws.String(app_id),
		BranchName: aws.String(branch_name),
	}

	if d.HasChange("backend_environment_arn") {
		params.BackendEnvironmentArn = aws.String(d.Get("backend_environment_arn").(string))
	}

	if d.HasChange("basic_auth_config") {
		enable, credentials := expandAmplifyBasicAuthConfig(d.Get("basic_auth_config").([]interface{}))
		params.EnableBasicAuth = enable
		params.BasicAuthCredentials = credentials
	}

	if d.HasChange("build_spec") {
		params.BuildSpec = aws.String(d.Get("build_spec").(string))
	}

	if d.HasChange("description") {
		params.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("display_name") {
		params.DisplayName = aws.String(d.Get("display_name").(string))
	}

	if d.HasChange("enable_auto_build") {
		params.EnableAutoBuild = aws.Bool(d.Get("enable_auto_build").(bool))
	}

	if d.HasChange("enable_notification") {
		params.EnableNotification = aws.Bool(d.Get("enable_notification").(bool))
	}

	if d.HasChange("enable_pull_request_preview") {
		params.EnablePullRequestPreview = aws.Bool(d.Get("enable_pull_request_preview").(bool))
	}

	if d.HasChange("environment_variables") {
		v := d.Get("environment_variables").(map[string]interface{})
		params.EnvironmentVariables = expandAmplifyEnvironmentVariables(v)
	}

	if d.HasChange("framework") {
		params.Framework = aws.String(d.Get("framework").(string))
	}

	if d.HasChange("pull_request_environment_name") {
		params.PullRequestEnvironmentName = aws.String(d.Get("pull_request_environment_name").(string))
	}

	if d.HasChange("stage") {
		params.Stage = aws.String(d.Get("stage").(string))
	}

	if d.HasChange("ttl") {
		params.Ttl = aws.String(d.Get("ttl").(string))
	}

	_, err := conn.UpdateBranch(params)
	if err != nil {
		return fmt.Errorf("Error updating Amplify Branch: %s", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.AmplifyUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsAmplifyBranchRead(d, meta)
}

func resourceAwsAmplifyBranchDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Deleting Amplify Branch: %s", d.Id())

	s := strings.Split(d.Id(), "/")
	app_id := s[0]
	branch_name := s[2]

	params := &amplify.DeleteBranchInput{
		AppId:      aws.String(app_id),
		BranchName: aws.String(branch_name),
	}

	_, err := conn.DeleteBranch(params)
	if err != nil {
		return fmt.Errorf("Error deleting Amplify Branch: %s", err)
	}

	return nil
}
