package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAppconfigDeploymentStrategy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppconfigDeploymentStrategyCreate,
		Read:   resourceAwsAppconfigDeploymentStrategyRead,
		Update: resourceAwsAppconfigDeploymentStrategyUpdate,
		Delete: resourceAwsAppconfigDeploymentStrategyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
				),
			},
			"deployment_duration_in_minutes": {
				Type:     schema.TypeInt,
				Required: true,
				ValidateFunc: validation.All(
					validation.IntBetween(0, 1440),
				),
			},
			"growth_factor": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"replicate_to": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"NONE", "SSM_DOCUMENT",
				}, false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1024),
				),
			},
			"final_bake_time_in_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
				ValidateFunc: validation.All(
					validation.IntBetween(0, 1440),
				),
			},
			"growth_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"EXPONENTIAL", "LINEAR",
				}, false),
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsAppconfigDeploymentStrategyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	input := &appconfig.CreateDeploymentStrategyInput{
		Name:                        aws.String(d.Get("name").(string)),
		Description:                 aws.String(d.Get("description").(string)),
		DeploymentDurationInMinutes: aws.Int64(int64(d.Get("deployment_duration_in_minutes").(int))),
		FinalBakeTimeInMinutes:      aws.Int64(int64(d.Get("final_bake_time_in_minutes").(int))),
		ReplicateTo:                 aws.String(d.Get("replicate_to").(string)),
		GrowthType:                  aws.String(d.Get("growth_type").(string)),
		GrowthFactor:                aws.Float64(d.Get("growth_factor").(float64)),
		Tags:                        keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AppconfigTags(),
	}

	strategy, err := conn.CreateDeploymentStrategy(input)
	if err != nil {
		return fmt.Errorf("Error creating AppConfig DeploymentStrategy: %s", err)
	}

	d.SetId(aws.StringValue(strategy.Id))

	return resourceAwsAppconfigDeploymentStrategyRead(d, meta)
}

func resourceAwsAppconfigDeploymentStrategyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &appconfig.GetDeploymentStrategyInput{
		DeploymentStrategyId: aws.String(d.Id()),
	}

	output, err := conn.GetDeploymentStrategy(input)

	if !d.IsNewResource() && isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Appconfig DeploymentStrategy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppConfig DeploymentStrategy (%s): %s", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting AppConfig DeploymentStrategy (%s): empty response", d.Id())
	}

	d.Set("name", output.Name)
	d.Set("description", output.Description)
	d.Set("deployment_duration_in_minutes", output.DeploymentDurationInMinutes)
	d.Set("final_bake_time_in_minutes", output.FinalBakeTimeInMinutes)
	d.Set("growth_factor", output.GrowthFactor)
	d.Set("replicate_to", output.ReplicateTo)
	d.Set("growth_type", output.GrowthType)

	strategyARN := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("deploymentstrategy/%s", d.Id()),
		Service:   "appconfig",
	}.String()
	d.Set("arn", strategyARN)

	tags, err := keyvaluetags.AppconfigListTags(conn, strategyARN)
	if err != nil {
		return fmt.Errorf("error getting tags for AppConfig DeploymentStrategy (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsAppconfigDeploymentStrategyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	updateInput := &appconfig.UpdateDeploymentStrategyInput{
		DeploymentStrategyId: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		updateInput.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("growth_type") {
		updateInput.GrowthType = aws.String(d.Get("growth_type").(string))
	}

	if d.HasChange("deployment_duration_in_minutes") {
		updateInput.DeploymentDurationInMinutes = aws.Int64(int64(d.Get("deployment_duration_in_minutes").(int)))
	}

	if d.HasChange("growth_factor") {
		updateInput.GrowthFactor = aws.Float64(d.Get("growth_factor").(float64))
	}

	if d.HasChange("final_bake_time_in_minutes") {
		updateInput.FinalBakeTimeInMinutes = aws.Int64(int64(d.Get("final_bake_time_in_minutes").(int)))
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.AppconfigUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating AppConfig DeploymentStrategy (%s) tags: %s", d.Id(), err)
		}
	}

	_, err := conn.UpdateDeploymentStrategy(updateInput)
	if err != nil {
		return fmt.Errorf("error updating AppConfig DeploymentStrategy (%s): %s", d.Id(), err)
	}

	return resourceAwsAppconfigDeploymentStrategyRead(d, meta)
}

func resourceAwsAppconfigDeploymentStrategyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	input := &appconfig.DeleteDeploymentStrategyInput{
		DeploymentStrategyId: aws.String(d.Id()),
	}

	_, err := conn.DeleteDeploymentStrategy(input)

	if isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Appconfig DeploymentStrategy (%s): %s", d.Id(), err)
	}

	return nil
}
