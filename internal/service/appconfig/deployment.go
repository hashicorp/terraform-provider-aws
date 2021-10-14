package appconfig

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeploymentCreate,
		Read:   resourceDeploymentRead,
		Update: resourceDeploymentUpdate,
		Delete: resourceDeploymentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z0-9]{4,7}`), ""),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_profile_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z0-9]{4,7}`), ""),
			},
			"configuration_version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"deployment_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"deployment_strategy_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`(^[a-z0-9]{4,7}$|^AppConfig\.[A-Za-z0-9]{9,40}$)`), ""),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"environment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z0-9]{4,7}`), ""),
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &appconfig.StartDeploymentInput{
		ApplicationId:          aws.String(d.Get("application_id").(string)),
		EnvironmentId:          aws.String(d.Get("environment_id").(string)),
		ConfigurationProfileId: aws.String(d.Get("configuration_profile_id").(string)),
		ConfigurationVersion:   aws.String(d.Get("configuration_version").(string)),
		DeploymentStrategyId:   aws.String(d.Get("deployment_strategy_id").(string)),
		Description:            aws.String(d.Get("description").(string)),
		Tags:                   tags.IgnoreAws().AppconfigTags(),
	}

	output, err := conn.StartDeployment(input)

	if err != nil {
		return fmt.Errorf("error starting AppConfig Deployment: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error starting AppConfig Deployment: empty response")
	}

	appID := aws.StringValue(output.ApplicationId)
	envID := aws.StringValue(output.EnvironmentId)
	deployNum := aws.Int64Value(output.DeploymentNumber)

	d.SetId(fmt.Sprintf("%s/%s/%d", appID, envID, deployNum))

	return resourceDeploymentRead(d, meta)
}

func resourceDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	appID, envID, deploymentNum, err := resourceAwsAppconfigDeploymentParseID(d.Id())

	if err != nil {
		return err
	}

	input := &appconfig.GetDeploymentInput{
		ApplicationId:    aws.String(appID),
		DeploymentNumber: aws.Int64(int64(deploymentNum)),
		EnvironmentId:    aws.String(envID),
	}

	output, err := conn.GetDeployment(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Appconfig Deployment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppConfig Deployment (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting AppConfig Deployment (%s): empty response", d.Id())
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/%s/environment/%s/deployment/%d", aws.StringValue(output.ApplicationId), aws.StringValue(output.EnvironmentId), aws.Int64Value(output.DeploymentNumber)),
		Service:   "appconfig",
	}.String()

	d.Set("application_id", output.ApplicationId)
	d.Set("arn", arn)
	d.Set("configuration_profile_id", output.ConfigurationProfileId)
	d.Set("configuration_version", output.ConfigurationVersion)
	d.Set("deployment_number", output.DeploymentNumber)
	d.Set("deployment_strategy_id", output.DeploymentStrategyId)
	d.Set("description", output.Description)
	d.Set("environment_id", output.EnvironmentId)
	d.Set("state", output.State)

	tags, err := tftags.AppconfigListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for AppConfig Deployment (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := tftags.AppconfigUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating AppConfig Deployment (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceDeploymentRead(d, meta)
}

func resourceDeploymentDelete(_ *schema.ResourceData, _ interface{}) error {
	log.Printf("[WARN] Cannot destroy AppConfig Deployment. Terraform will remove this resource from the state file, however this resource remains.")
	return nil
}

func resourceAwsAppconfigDeploymentParseID(id string) (string, string, int, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", 0, fmt.Errorf("unexpected format of ID (%q), expected ApplicationID:EnvironmentID:DeploymentNumber", id)
	}

	num, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", "", 0, fmt.Errorf("error parsing AppConfig Deployment resource ID deployment_number: %w", err)
	}

	return parts[0], parts[1], num, nil
}
