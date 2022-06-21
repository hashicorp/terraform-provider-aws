package sagemaker

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceModelPackageGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceModelPackageGroupCreate,
		Read:   resourceModelPackageGroupRead,
		Update: resourceModelPackageGroupUpdate,
		Delete: resourceModelPackageGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"model_package_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}$`),
						"Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"model_package_group_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceModelPackageGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("model_package_group_name").(string)
	input := &sagemaker.CreateModelPackageGroupInput{
		ModelPackageGroupName: aws.String(name),
	}

	if v, ok := d.GetOk("model_package_group_description"); ok {
		input.ModelPackageGroupDescription = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	_, err := conn.CreateModelPackageGroup(input)
	if err != nil {
		return fmt.Errorf("error creating SageMaker Model Package Group %s: %w", name, err)
	}

	d.SetId(name)

	if _, err := WaitModelPackageGroupCompleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for SageMaker Model Package Group (%s) to be created: %w", d.Id(), err)
	}

	return resourceModelPackageGroupRead(d, meta)
}

func resourceModelPackageGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	mpg, err := FindModelPackageGroupByName(conn, d.Id())
	if err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Model Package Group (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker Model Package Group (%s): %w", d.Id(), err)

	}

	arn := aws.StringValue(mpg.ModelPackageGroupArn)
	d.Set("model_package_group_name", mpg.ModelPackageGroupName)
	d.Set("arn", arn)
	d.Set("model_package_group_description", mpg.ModelPackageGroupDescription)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker Model Package Group (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceModelPackageGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker Model Package Group (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceModelPackageGroupRead(d, meta)
}

func resourceModelPackageGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	input := &sagemaker.DeleteModelPackageGroupInput{
		ModelPackageGroupName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteModelPackageGroup(input); err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker Model Package Group (%s): %w", d.Id(), err)
	}

	if _, err := WaitModelPackageGroupDeleted(conn, d.Id()); err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
			return nil
		}
		return fmt.Errorf("error waiting for SageMaker Model Package Group (%s) to delete: %w", d.Id(), err)
	}

	return nil
}
