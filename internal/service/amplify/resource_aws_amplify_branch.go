package amplify

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBranch() *schema.Resource {
	return &schema.Resource{
		Create: resourceBranchCreate,
		Read:   resourceBranchRead,
		Update: resourceBranchUpdate,
		Delete: resourceBranchDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"associated_resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"backend_environment_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},

			"basic_auth_credentials": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 2000),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// These credentials are ignored if basic auth is not enabled.
					if d.Get("enable_basic_auth").(bool) {
						return old == new
					}

					return true
				},
			},

			"branch_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z/_.-]{1,255}$`), "should be not be more than 255 letters, numbers, and the symbols /_.-"),
			},

			"custom_domains": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},

			"destination_branch": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[0-9a-z-]{1,255}$`), "should be not be more than 255 lowercase alphanumeric or hyphen characters"),
			},

			"enable_auto_build": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"enable_basic_auth": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"enable_notification": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"enable_performance_mode": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"pull_request_environment_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 20),
			},

			"source_branch": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"stage": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(amplify.Stage_Values(), false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// API returns "NONE" by default.
					if old == StageNone && new == "" {
						return true
					}

					return old == new
				},
			},

			"ttl": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// API returns "5" by default.
					if old == "5" && new == "" {
						return true
					}

					return old == new
				},
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceBranchCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AmplifyConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	appID := d.Get("app_id").(string)
	branchName := d.Get("branch_name").(string)
	id := BranchCreateResourceID(appID, branchName)

	input := &amplify.CreateBranchInput{
		AppId:           aws.String(appID),
		BranchName:      aws.String(branchName),
		EnableAutoBuild: aws.Bool(d.Get("enable_auto_build").(bool)),
	}

	if v, ok := d.GetOk("backend_environment_arn"); ok {
		input.BackendEnvironmentArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("basic_auth_credentials"); ok {
		input.BasicAuthCredentials = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("display_name"); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enable_basic_auth"); ok {
		input.EnableBasicAuth = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("enable_notification"); ok {
		input.EnableNotification = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("enable_performance_mode"); ok {
		input.EnablePerformanceMode = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("enable_pull_request_preview"); ok {
		input.EnablePullRequestPreview = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("environment_variables"); ok && len(v.(map[string]interface{})) > 0 {
		input.EnvironmentVariables = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("framework"); ok {
		input.Framework = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pull_request_environment_name"); ok {
		input.PullRequestEnvironmentName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("stage"); ok {
		input.Stage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ttl"); ok {
		input.Ttl = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().AmplifyTags()
	}

	log.Printf("[DEBUG] Creating Amplify Branch: %s", input)
	_, err := conn.CreateBranch(input)

	if err != nil {
		return fmt.Errorf("error creating Amplify Branch (%s): %w", id, err)
	}

	d.SetId(id)

	return resourceBranchRead(d, meta)
}

func resourceBranchRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AmplifyConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	appID, branchName, err := BranchParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Amplify Branch ID: %w", err)
	}

	branch, err := FindBranchByAppIDAndBranchName(conn, appID, branchName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Amplify Branch (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Amplify Branch (%s): %w", d.Id(), err)
	}

	d.Set("app_id", appID)
	d.Set("arn", branch.BranchArn)
	d.Set("associated_resources", aws.StringValueSlice(branch.AssociatedResources))
	d.Set("backend_environment_arn", branch.BackendEnvironmentArn)
	d.Set("basic_auth_credentials", branch.BasicAuthCredentials)
	d.Set("branch_name", branch.BranchName)
	d.Set("custom_domains", aws.StringValueSlice(branch.CustomDomains))
	d.Set("description", branch.Description)
	d.Set("destination_branch", branch.DestinationBranch)
	d.Set("display_name", branch.DisplayName)
	d.Set("enable_auto_build", branch.EnableAutoBuild)
	d.Set("enable_basic_auth", branch.EnableBasicAuth)
	d.Set("enable_notification", branch.EnableNotification)
	d.Set("enable_performance_mode", branch.EnablePerformanceMode)
	d.Set("enable_pull_request_preview", branch.EnablePullRequestPreview)
	d.Set("environment_variables", aws.StringValueMap(branch.EnvironmentVariables))
	d.Set("framework", branch.Framework)
	d.Set("pull_request_environment_name", branch.PullRequestEnvironmentName)
	d.Set("source_branch", branch.SourceBranch)
	d.Set("stage", branch.Stage)
	d.Set("ttl", branch.Ttl)

	tags := tftags.AmplifyKeyValueTags(branch.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceBranchUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AmplifyConn

	if d.HasChangesExcept("tags", "tags_all") {
		appID, branchName, err := BranchParseResourceID(d.Id())

		if err != nil {
			return fmt.Errorf("error parsing Amplify Branch ID: %w", err)
		}

		input := &amplify.UpdateBranchInput{
			AppId:      aws.String(appID),
			BranchName: aws.String(branchName),
		}

		if d.HasChange("backend_environment_arn") {
			input.BackendEnvironmentArn = aws.String(d.Get("backend_environment_arn").(string))
		}

		if d.HasChange("basic_auth_credentials") {
			input.BasicAuthCredentials = aws.String(d.Get("basic_auth_credentials").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("display_name") {
			input.DisplayName = aws.String(d.Get("display_name").(string))
		}

		if d.HasChange("enable_auto_build") {
			input.EnableAutoBuild = aws.Bool(d.Get("enable_auto_build").(bool))
		}

		if d.HasChange("enable_basic_auth") {
			input.EnableBasicAuth = aws.Bool(d.Get("enable_basic_auth").(bool))
		}

		if d.HasChange("enable_notification") {
			input.EnableNotification = aws.Bool(d.Get("enable_notification").(bool))
		}

		if d.HasChange("enable_performance_mode") {
			input.EnablePullRequestPreview = aws.Bool(d.Get("enable_performance_mode").(bool))
		}

		if d.HasChange("enable_pull_request_preview") {
			input.EnablePullRequestPreview = aws.Bool(d.Get("enable_pull_request_preview").(bool))
		}

		if d.HasChange("environment_variables") {
			if v := d.Get("environment_variables").(map[string]interface{}); len(v) > 0 {
				input.EnvironmentVariables = flex.ExpandStringMap(v)
			} else {
				input.EnvironmentVariables = aws.StringMap(map[string]string{"": ""})
			}
		}

		if d.HasChange("framework") {
			input.Framework = aws.String(d.Get("framework").(string))
		}

		if d.HasChange("pull_request_environment_name") {
			input.PullRequestEnvironmentName = aws.String(d.Get("pull_request_environment_name").(string))
		}

		if d.HasChange("stage") {
			input.Stage = aws.String(d.Get("stage").(string))
		}

		if d.HasChange("ttl") {
			input.Ttl = aws.String(d.Get("ttl").(string))
		}

		_, err = conn.UpdateBranch(input)

		if err != nil {
			return fmt.Errorf("error updating Amplify Branch (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.AmplifyUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceBranchRead(d, meta)
}

func resourceBranchDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AmplifyConn

	appID, branchName, err := BranchParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Amplify Branch ID: %w", err)
	}

	log.Printf("[DEBUG] Deleting Amplify Branch: %s", d.Id())
	_, err = conn.DeleteBranch(&amplify.DeleteBranchInput{
		AppId:      aws.String(appID),
		BranchName: aws.String(branchName),
	})

	if tfawserr.ErrCodeEquals(err, amplify.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Amplify Branch (%s): %w", d.Id(), err)
	}

	return nil
}
