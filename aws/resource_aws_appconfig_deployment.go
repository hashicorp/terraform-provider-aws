package aws

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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/appconfig/waiter"
)

func resourceAwsAppconfigDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppconfigDeploymentCreate,
		Read:   resourceAwsAppconfigDeploymentRead,
		Update: resourceAwsAppconfigDeploymentUpdate,
		Delete: resourceAwsAppconfigDeploymentDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppconfigDeploymentImport,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z0-9]{4,7}`), ""),
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
			"deployment_strategy_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z0-9]{4,7}`), ""),
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
			"deployment_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsAppconfigDeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	appID := d.Get("application_id").(string)
	envID := d.Get("environment_id").(string)

	input := &appconfig.StartDeploymentInput{
		ApplicationId:          aws.String(appID),
		EnvironmentId:          aws.String(envID),
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

	deployNum := aws.Int64Value(output.DeploymentNumber)

	d.Set("deployment_number", output.DeploymentNumber)
	d.SetId(appID + "/" + envID + "/" + strconv.FormatInt(deployNum, 10))

	if err := waiter.DeploymentCreated(conn, appID, envID, deployNum); err != nil {
		return fmt.Errorf("error waiting for AppConfig Deployment (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsAppconfigDeploymentRead(d, meta)
}

func parseAwsAppconfigDeploymentID(id string) (string, string, int64, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 3 {
		return "", "", 0, fmt.Errorf("unexpected format of ID (%q), expected <application-id>/<environment-id>/<deployment-number>", id)
	}

	appID := idParts[0]
	envID := idParts[1]
	depS := idParts[2]
	depNum, err := strconv.ParseInt(depS, 10, 64)
	if err != nil {
		return "", "", 0, fmt.Errorf("deployment number is invalid (%s): %w", depS, err)
	}
	return appID, envID, depNum, nil
}

func resourceAwsAppconfigDeploymentImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	appID, envID, depNum, err := parseAwsAppconfigDeploymentID(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("application_id", appID)
	d.Set("environment_id", envID)
	d.Set("deployment_number", depNum)

	return []*schema.ResourceData{d}, nil
}

func resourceAwsAppconfigDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	appID := d.Get("application_id").(string)
	envID := d.Get("environment_id").(string)
	deployNum := d.Get("deployment_number").(int)

	input := &appconfig.GetDeploymentInput{
		ApplicationId:    aws.String(appID),
		EnvironmentId:    aws.String(envID),
		DeploymentNumber: aws.Int64(int64(deployNum)),
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
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("application/%s/environment/%s/deployment/%d", appID, envID, deployNum),
		Service:   "appconfig",
	}.String()

	d.Set("arn", arn)
	d.Set("description", output.Description)
	d.Set("configuration_profile_id", output.ConfigurationProfileId)
	d.Set("configuration_version", output.ConfigurationVersion)
	d.Set("deployment_strategy_id", output.DeploymentStrategyId)

	tags, err := keyvaluetags.AppconfigListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for AppConfig Application (%s): %w", d.Id(), err)
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

func resourceAwsAppconfigDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.AppconfigUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating EFS file system (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsAppconfigDeploymentRead(d, meta)
}

func resourceAwsAppconfigDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy AppConfig Deployment. Terraform will remove this resource from the state file, however this resource remains.")
	return nil
}
