package ecr

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceRepositoryCreate,
		Read:   resourceRepositoryRead,
		Update: resourceRepositoryUpdate,
		Delete: resourceRepositoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								ecr.EncryptionTypeAes256,
								ecr.EncryptionTypeKms,
							}, false),
							Default:  ecr.EncryptionTypeAes256,
							ForceNew: true,
						},
						"kms_key": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				ForceNew:         true,
			},
			"image_scanning_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scan_on_push": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"image_tag_mutability": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ecr.ImageTagMutabilityMutable,
				ValidateFunc: validation.StringInSlice([]string{
					ecr.ImageTagMutabilityMutable,
					ecr.ImageTagMutabilityImmutable,
				}, false),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := ecr.CreateRepositoryInput{
		ImageTagMutability:      aws.String(d.Get("image_tag_mutability").(string)),
		RepositoryName:          aws.String(d.Get("name").(string)),
		EncryptionConfiguration: expandRepositoryEncryptionConfiguration(d.Get("encryption_configuration").([]interface{})),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	imageScanningConfigs := d.Get("image_scanning_configuration").([]interface{})
	if len(imageScanningConfigs) > 0 {
		imageScanningConfig := imageScanningConfigs[0]
		if imageScanningConfig != nil {
			configMap := imageScanningConfig.(map[string]interface{})
			input.ImageScanningConfiguration = &ecr.ImageScanningConfiguration{
				ScanOnPush: aws.Bool(configMap["scan_on_push"].(bool)),
			}
		}
	}

	log.Printf("[DEBUG] Creating ECR repository: %#v", input)
	out, err := conn.CreateRepository(&input)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if input.Tags != nil && meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating ECR Repository (%s) with tags: %s. Trying create without tags.", d.Get("name").(string), err)
		input.Tags = nil

		out, err = conn.CreateRepository(&input)
	}

	if err != nil {
		return fmt.Errorf("failed creating ECR Repository (%s): %w", d.Get("name").(string), err)
	}

	repository := *out.Repository // nosemgrep: prefer-aws-go-sdk-pointer-conversion-assignment // false positive

	log.Printf("[DEBUG] ECR repository created: %q", *repository.RepositoryArn)

	d.SetId(aws.StringValue(repository.RepositoryName))

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if input.Tags == nil && len(tags) > 0 && meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID {
		err := UpdateTags(conn, aws.StringValue(repository.RepositoryArn), nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed adding tags after create for ECR Repository (%s): %s", d.Id(), err)
			return resourceRepositoryRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("failed adding tags after create for ECR Repository (%s): %w", d.Id(), err)
		}
	}

	return resourceRepositoryRead(d, meta)
}

func resourceRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading ECR repository %s", d.Id())
	var out *ecr.DescribeRepositoriesOutput
	input := &ecr.DescribeRepositoriesInput{
		RepositoryNames: aws.StringSlice([]string{d.Id()}),
	}

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		out, err = conn.DescribeRepositories(input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		out, err = conn.DescribeRepositories(input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
		log.Printf("[WARN] ECR Repository (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ECR Repository (%s): %w", d.Id(), err)
	}

	if out == nil || len(out.Repositories) == 0 || out.Repositories[0] == nil {
		return fmt.Errorf("error reading ECR Repository (%s): empty response", d.Id())
	}

	repository := out.Repositories[0]
	arn := aws.StringValue(repository.RepositoryArn)

	d.Set("arn", arn)
	d.Set("name", repository.RepositoryName)
	d.Set("registry_id", repository.RegistryId)
	d.Set("repository_url", repository.RepositoryUri)
	d.Set("image_tag_mutability", repository.ImageTagMutability)

	if err := d.Set("image_scanning_configuration", flattenImageScanningConfiguration(repository.ImageScanningConfiguration)); err != nil {
		return fmt.Errorf("error setting image_scanning_configuration for ECR Repository (%s): %w", arn, err)
	}

	if err := d.Set("encryption_configuration", flattenRepositoryEncryptionConfiguration(repository.EncryptionConfiguration)); err != nil {
		return fmt.Errorf("error setting encryption_configuration for ECR Repository (%s): %w", arn, err)
	}

	tags, err := ListTags(conn, arn)

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed listing tags for ECR Repository (%s): %s", d.Id(), err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed listing tags for ECR Repository (%s): %w", d.Id(), err)
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

func flattenImageScanningConfiguration(isc *ecr.ImageScanningConfiguration) []map[string]interface{} {
	if isc == nil {
		return nil
	}

	config := make(map[string]interface{})
	config["scan_on_push"] = aws.BoolValue(isc.ScanOnPush)

	return []map[string]interface{}{
		config,
	}
}

func expandRepositoryEncryptionConfiguration(data []interface{}) *ecr.EncryptionConfiguration {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	ec := data[0].(map[string]interface{})
	config := &ecr.EncryptionConfiguration{
		EncryptionType: aws.String(ec["encryption_type"].(string)),
	}
	if v, ok := ec["kms_key"]; ok {
		if s := v.(string); s != "" {
			config.KmsKey = aws.String(v.(string))
		}
	}
	return config
}

func flattenRepositoryEncryptionConfiguration(ec *ecr.EncryptionConfiguration) []map[string]interface{} {
	if ec == nil {
		return nil
	}

	config := map[string]interface{}{
		"encryption_type": aws.StringValue(ec.EncryptionType),
		"kms_key":         aws.StringValue(ec.KmsKey),
	}

	return []map[string]interface{}{
		config,
	}
}

func resourceRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	arn := d.Get("arn").(string)
	conn := meta.(*conns.AWSClient).ECRConn

	if d.HasChange("image_tag_mutability") {
		if err := resourceRepositoryUpdateImageTagMutability(conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("image_scanning_configuration") {
		if err := resourceRepositoryUpdateImageScanningConfiguration(conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(conn, arn, o, n)

		// Some partitions may not support tagging, giving error
		if meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed updating tags for ECR Repository (%s): %s", d.Id(), err)
			return resourceRepositoryRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("failed updating tags for ECR Repository (%s): %w", d.Id(), err)
		}
	}

	return resourceRepositoryRead(d, meta)
}

func resourceRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	_, err := conn.DeleteRepository(&ecr.DeleteRepositoryInput{
		RepositoryName: aws.String(d.Id()),
		RegistryId:     aws.String(d.Get("registry_id").(string)),
		Force:          aws.Bool(true),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting ECR repository: %s", err)
	}

	log.Printf("[DEBUG] Waiting for ECR Repository %q to be deleted", d.Id())
	input := &ecr.DescribeRepositoriesInput{
		RepositoryNames: aws.StringSlice([]string{d.Id()}),
	}
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = conn.DescribeRepositories(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(fmt.Errorf("%q: Timeout while waiting for the ECR Repository to be deleted", d.Id()))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DescribeRepositories(input)
	}

	if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting ECR repository: %s", err)
	}

	log.Printf("[DEBUG] repository %q deleted.", d.Get("name").(string))

	return nil
}

func resourceRepositoryUpdateImageTagMutability(conn *ecr.ECR, d *schema.ResourceData) error {
	input := &ecr.PutImageTagMutabilityInput{
		ImageTagMutability: aws.String(d.Get("image_tag_mutability").(string)),
		RepositoryName:     aws.String(d.Id()),
		RegistryId:         aws.String(d.Get("registry_id").(string)),
	}

	_, err := conn.PutImageTagMutability(input)
	if err != nil {
		return fmt.Errorf("Error setting image tag mutability: %s", err.Error())
	}

	return nil
}
func resourceRepositoryUpdateImageScanningConfiguration(conn *ecr.ECR, d *schema.ResourceData) error {

	var ecrImageScanningConfig ecr.ImageScanningConfiguration
	imageScanningConfigs := d.Get("image_scanning_configuration").([]interface{})
	if len(imageScanningConfigs) > 0 {
		imageScanningConfig := imageScanningConfigs[0]
		if imageScanningConfig != nil {
			configMap := imageScanningConfig.(map[string]interface{})
			ecrImageScanningConfig.ScanOnPush = aws.Bool(configMap["scan_on_push"].(bool))
		}
	}

	input := &ecr.PutImageScanningConfigurationInput{
		ImageScanningConfiguration: &ecrImageScanningConfig,
		RepositoryName:             aws.String(d.Id()),
		RegistryId:                 aws.String(d.Get("registry_id").(string)),
	}

	_, err := conn.PutImageScanningConfiguration(input)
	if err != nil {
		return fmt.Errorf("Error setting image scanning configuration: %s", err.Error())
	}

	return nil
}
