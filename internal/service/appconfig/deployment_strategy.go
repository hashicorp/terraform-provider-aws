package appconfig

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDeploymentStrategy() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeploymentStrategyCreate,
		Read:   resourceDeploymentStrategyRead,
		Update: resourceDeploymentStrategyUpdate,
		Delete: resourceDeploymentStrategyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_duration_in_minutes": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 1440),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"final_bake_time_in_minutes": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 1440),
			},
			"growth_factor": {
				Type:         schema.TypeFloat,
				Required:     true,
				ValidateFunc: validation.FloatBetween(1.0, 100.0),
			},
			"growth_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      appconfig.GrowthTypeLinear,
				ValidateFunc: validation.StringInSlice(appconfig.GrowthType_Values(), false),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"replicate_to": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(appconfig.ReplicateTo_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDeploymentStrategyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &appconfig.CreateDeploymentStrategyInput{
		DeploymentDurationInMinutes: aws.Int64(int64(d.Get("deployment_duration_in_minutes").(int))),
		GrowthFactor:                aws.Float64(d.Get("growth_factor").(float64)),
		GrowthType:                  aws.String(d.Get("growth_type").(string)),
		Name:                        aws.String(name),
		ReplicateTo:                 aws.String(d.Get("replicate_to").(string)),
		Tags:                        tags.IgnoreAws().AppconfigTags(),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("final_bake_time_in_minutes"); ok {
		input.FinalBakeTimeInMinutes = aws.Int64(int64(v.(int)))
	}

	strategy, err := conn.CreateDeploymentStrategy(input)

	if err != nil {
		return fmt.Errorf("error creating AppConfig Deployment Strategy (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(strategy.Id))

	return resourceDeploymentStrategyRead(d, meta)
}

func resourceDeploymentStrategyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &appconfig.GetDeploymentStrategyInput{
		DeploymentStrategyId: aws.String(d.Id()),
	}

	output, err := conn.GetDeploymentStrategy(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Appconfig Deployment Strategy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppConfig Deployment Strategy (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting AppConfig Deployment Strategy (%s): empty response", d.Id())
	}

	d.Set("description", output.Description)
	d.Set("deployment_duration_in_minutes", output.DeploymentDurationInMinutes)
	d.Set("final_bake_time_in_minutes", output.FinalBakeTimeInMinutes)
	d.Set("growth_factor", output.GrowthFactor)
	d.Set("growth_type", output.GrowthType)
	d.Set("name", output.Name)
	d.Set("replicate_to", output.ReplicateTo)

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("deploymentstrategy/%s", d.Id()),
		Service:   "appconfig",
	}.String()
	d.Set("arn", arn)

	tags, err := tftags.AppconfigListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for AppConfig Deployment Strategy (%s): %w", d.Id(), err)
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

func resourceDeploymentStrategyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn

	if d.HasChangesExcept("tags", "tags_all") {
		updateInput := &appconfig.UpdateDeploymentStrategyInput{
			DeploymentStrategyId: aws.String(d.Id()),
		}

		if d.HasChange("deployment_duration_in_minutes") {
			updateInput.DeploymentDurationInMinutes = aws.Int64(int64(d.Get("deployment_duration_in_minutes").(int)))
		}

		if d.HasChange("description") {
			updateInput.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("final_bake_time_in_minutes") {
			updateInput.FinalBakeTimeInMinutes = aws.Int64(int64(d.Get("final_bake_time_in_minutes").(int)))
		}

		if d.HasChange("growth_factor") {
			updateInput.GrowthFactor = aws.Float64(d.Get("growth_factor").(float64))
		}

		if d.HasChange("growth_type") {
			updateInput.GrowthType = aws.String(d.Get("growth_type").(string))
		}

		_, err := conn.UpdateDeploymentStrategy(updateInput)

		if err != nil {
			return fmt.Errorf("error updating AppConfig Deployment Strategy (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.AppconfigUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating AppConfig Deployment Strategy (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceDeploymentStrategyRead(d, meta)
}

func resourceDeploymentStrategyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn

	input := &appconfig.DeleteDeploymentStrategyInput{
		DeploymentStrategyId: aws.String(d.Id()),
	}

	_, err := conn.DeleteDeploymentStrategy(input)

	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Appconfig Deployment Strategy (%s): %w", d.Id(), err)
	}

	return nil
}
