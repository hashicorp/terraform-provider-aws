package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"log"
	"regexp"
)

func resourceAwsImageBuilderImagePipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsImageBuilderImagePipelineCreate,
		Read:   resourceAwsImageBuilderImagePipelineRead,
		Update: resourceAwsImageBuilderImagePipelineUpdate,
		Delete: resourceAwsImageBuilderImagePipelineDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"distribution_configuration_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):distribution-configuration/[a-z0-9-_]+$`), "valid distribution configuration ARN must be provided"),
			},
			"enhanced_image_metadata_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"image_recipe_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true, // this should be updatable, but recreating the recipe causes TF to error saying it's depended on by other resources, the pipeline
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):image-recipe/[a-z0-9-_]+/\d+\.\d+\.\d+$`), "valid image recipe ARN must be provided"),
			},
			"image_tests_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image_tests_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"timeout_minutes": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      720,
							ValidateFunc: validation.IntBetween(60, 1440),
						},
					},
				},
			},
			"infrastructure_configuration_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):infrastructure-configuration/[a-z0-9-_]+$`), "valid infrastructure configuration ARN must be provided"),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^[-_A-Za-z-0-9][-_A-Za-z0-9 ]{1,126}[-_A-Za-z-0-9]$"), "valid name must be provided"),
			},
			"schedule": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pipeline_execution_start_condition": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "EXPRESSION_MATCH_ONLY",
							ValidateFunc: validation.StringInSlice([]string{"EXPRESSION_MATCH_ONLY", "EXPRESSION_MATCH_AND_DEPENDENCY_UPDATES_AVAILABLE"}, false),
						},
						"schedule_expression": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
					},
				},
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"ENABLED", "DISABLED"}, false),
				Default:      "ENABLED",
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsImageBuilderImagePipelineCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.CreateImagePipelineInput{
		EnhancedImageMetadataEnabled:   aws.Bool(d.Get("enhanced_image_metadata_enabled").(bool)),
		ImageRecipeArn:                 aws.String(d.Get("image_recipe_arn").(string)),
		InfrastructureConfigurationArn: aws.String(d.Get("infrastructure_configuration_arn").(string)),
		Name:                           aws.String(d.Get("name").(string)),
		Status:                         aws.String(d.Get("status").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.SetDescription(v.(string))
	}

	if v, ok := d.GetOk("distribution_configuration_arn"); ok {
		input.SetDistributionConfigurationArn(v.(string))
	}

	if v, ok := d.GetOk("image_tests_configuration"); ok {
		input.ImageTestsConfiguration = expandAwsImageBuilderTestConfiguration(v)
	}

	if v, ok := d.GetOk("schedule"); ok {
		input.Schedule = expandAwsImageBuilderPipelineSchedule(v)
	}

	if v, ok := d.GetOk("tags"); ok {
		input.SetTags(keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().ImagebuilderTags())
	}

	log.Printf("[DEBUG] Creating Image Pipeline: %#v", input)

	resp, err := conn.CreateImagePipeline(input)
	if err != nil {
		return fmt.Errorf("error creating Image Pipeline: %s", err)
	}

	d.SetId(aws.StringValue(resp.ImagePipelineArn))

	return resourceAwsImageBuilderImagePipelineRead(d, meta)
}

func resourceAwsImageBuilderImagePipelineRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	resp, err := conn.GetImagePipeline(&imagebuilder.GetImagePipelineInput{
		ImagePipelineArn: aws.String(d.Id()),
	})

	if isAWSErr(err, imagebuilder.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Image Pipeline (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Image Pipeline (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.ImagePipeline.Arn)
	d.Set("description", resp.ImagePipeline.Description)
	d.Set("distribution_configuration_arn", resp.ImagePipeline.DistributionConfigurationArn)
	d.Set("enhanced_image_metadata_enabled", resp.ImagePipeline.EnhancedImageMetadataEnabled)
	d.Set("image_recipe_arn", resp.ImagePipeline.ImageRecipeArn)
	d.Set("image_tests_configuration", flattenAwsImageBuilderTestConfiguration(resp.ImagePipeline.ImageTestsConfiguration))
	d.Set("infrastructure_configuration_arn", resp.ImagePipeline.InfrastructureConfigurationArn)
	d.Set("name", resp.ImagePipeline.Name)
	// TODO: platform?? this isn't in create so why in read? auto field?

	if resp.ImagePipeline.Schedule != nil {
		d.Set("schedule", flattenAwsImageBuilderPipelineSchedule(resp.ImagePipeline.Schedule))
	}
	d.Set("status", resp.ImagePipeline.Status)

	tags, err := keyvaluetags.ImagebuilderListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for Image Pipeline (%s): %s", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(meta.(*AWSClient).IgnoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsImageBuilderImagePipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.UpdateImagePipelineInput{
		ImagePipelineArn:               aws.String(d.Id()),
		ImageRecipeArn:                 aws.String(d.Get("image_recipe_arn").(string)),
		InfrastructureConfigurationArn: aws.String(d.Get("infrastructure_configuration_arn").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("distribution_configuration_arn"); ok {
		input.DistributionConfigurationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enhanced_image_metadata_enabled"); ok {
		input.EnhancedImageMetadataEnabled = aws.Bool(v.(bool))
	}

	if d.HasChange("image_tests_configuration") {
		input.ImageTestsConfiguration = expandAwsImageBuilderTestConfiguration(d.Get("image_tests_configuration"))
	}

	if d.HasChange("schedule") {
		input.Schedule = expandAwsImageBuilderPipelineSchedule(d.Get("schedule"))
	}

	if v, ok := d.GetOk("status"); ok {
		input.Status = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Updating Image Pipeline: %#v", input)
	_, err := conn.UpdateImagePipeline(input)
	if err != nil {
		return fmt.Errorf("error updating Image Pipeline  (%s): %s", d.Id(), err)
	}

	return resourceAwsImageBuilderImagePipelineRead(d, meta)
}

func resourceAwsImageBuilderImagePipelineDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	_, err := conn.DeleteImagePipeline(&imagebuilder.DeleteImagePipelineInput{
		ImagePipelineArn: aws.String(d.Id()),
	})

	if isAWSErr(err, imagebuilder.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Image Pipeline (%s): %s", d.Id(), err)
	}

	return nil
}

func expandAwsImageBuilderTestConfiguration(d interface{}) *imagebuilder.ImageTestsConfiguration {
	testconfig := d.([]interface{})[0].(map[string]interface{})

	itc := &imagebuilder.ImageTestsConfiguration{
		ImageTestsEnabled: testconfig["image_tests_enabled"].(*bool),
		TimeoutMinutes:    testconfig["timeout_minutes"].(*int64),
	}

	return itc
}

func expandAwsImageBuilderPipelineSchedule(d interface{}) *imagebuilder.Schedule {
	schedconfig := d.([]interface{})[0].(map[string]interface{})

	ibsched := &imagebuilder.Schedule{
		PipelineExecutionStartCondition: schedconfig["pipeline_execution_start_condition"].(*string),
		ScheduleExpression:              schedconfig["schedule_expression"].(*string),
	}

	return ibsched
}

func flattenAwsImageBuilderTestConfiguration(configuration *imagebuilder.ImageTestsConfiguration) map[string]interface{} {
	testconfig := make(map[string]interface{})

	if configuration != nil {
		testconfig["timeout_minutes"] = configuration.TimeoutMinutes
		testconfig["image_tests_enabled"] = configuration.ImageTestsEnabled
	}

	return testconfig
}

func flattenAwsImageBuilderPipelineSchedule(schedule *imagebuilder.Schedule) map[string]interface{} {
	schedconfig := make(map[string]interface{}, 2)

	schedconfig["schedule_expression"] = schedule.ScheduleExpression
	schedconfig["pipeline_execution_start_condition"] = schedule.PipelineExecutionStartCondition

	return schedconfig
}
