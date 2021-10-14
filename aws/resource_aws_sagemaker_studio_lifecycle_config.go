package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceStudioLifecycleConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceStudioLifecycleConfigCreate,
		Read:   resourceStudioLifecycleConfigRead,
		Update: resourceStudioLifecycleConfigUpdate,
		Delete: resourceStudioLifecycleConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"studio_lifecycle_config_app_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.StudioLifecycleConfigAppType_Values(), false),
			},
			"studio_lifecycle_config_content": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 16384),
			},
			"studio_lifecycle_config_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStudioLifecycleConfigCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("studio_lifecycle_config_name").(string)
	input := &sagemaker.CreateStudioLifecycleConfigInput{
		StudioLifecycleConfigName:    aws.String(name),
		StudioLifecycleConfigAppType: aws.String(d.Get("studio_lifecycle_config_app_type").(string)),
		StudioLifecycleConfigContent: aws.String(d.Get("studio_lifecycle_config_content").(string)),
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().SagemakerTags()
	}

	log.Printf("[DEBUG] Creating SageMaker Studio Lifecycle Config : %s", input)
	_, err := conn.CreateStudioLifecycleConfig(input)

	if err != nil {
		return fmt.Errorf("error creating SageMaker Studio Lifecycle Config (%s): %w", name, err)
	}

	d.SetId(name)

	return resourceStudioLifecycleConfigRead(d, meta)
}

func resourceStudioLifecycleConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	image, err := finder.StudioLifecycleConfigByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Studio Lifecycle Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SageMaker Studio Lifecycle Config (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(image.StudioLifecycleConfigArn)
	d.Set("studio_lifecycle_config_name", image.StudioLifecycleConfigName)
	d.Set("studio_lifecycle_config_app_type", image.StudioLifecycleConfigAppType)
	d.Set("studio_lifecycle_config_content", image.StudioLifecycleConfigContent)
	d.Set("arn", arn)

	tags, err := tftags.SagemakerListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker Studio Lifecycle Config (%s): %w", d.Id(), err)
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

func resourceStudioLifecycleConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := tftags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Studio Lifecycle Config (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceStudioLifecycleConfigRead(d, meta)
}

func resourceStudioLifecycleConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	input := &sagemaker.DeleteStudioLifecycleConfigInput{
		StudioLifecycleConfigName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting SageMaker Studio Lifecycle Config: (%s)", d.Id())
	if _, err := conn.DeleteStudioLifecycleConfig(input); err != nil {
		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "does not exist") {
			return nil
		}

		return fmt.Errorf("error deleting SageMaker Studio Lifecycle Config (%s): %w", d.Id(), err)
	}

	return nil
}
