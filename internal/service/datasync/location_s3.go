package datasync

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLocationS3() *schema.Resource {
	return &schema.Resource{
		Create: resourceLocationS3Create,
		Read:   resourceLocationS3Read,
		Update: resourceLocationS3Update,
		Delete: resourceLocationS3Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"agent_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"s3_bucket_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"s3_config": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_access_role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"s3_storage_class": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(datasync.S3StorageClass_Values(), false),
			},
			"subdirectory": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// Ignore missing trailing slash
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "/" {
						return false
					}
					if strings.TrimSuffix(old, "/") == strings.TrimSuffix(new, "/") {
						return true
					}
					return false
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationS3Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &datasync.CreateLocationS3Input{
		S3BucketArn:  aws.String(d.Get("s3_bucket_arn").(string)),
		S3Config:     expandDataSyncS3Config(d.Get("s3_config").([]interface{})),
		Subdirectory: aws.String(d.Get("subdirectory").(string)),
		Tags:         Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("agent_arns"); ok {
		input.AgentArns = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("s3_storage_class"); ok {
		input.S3StorageClass = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating DataSync Location S3: %s", input)

	var output *datasync.CreateLocationS3Output
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.CreateLocationS3(input)

		// Retry for IAM eventual consistency on error:
		// InvalidRequestException: Unable to assume role. Reason: Access denied when calling sts:AssumeRole
		if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "Unable to assume role") {
			return resource.RetryableError(err)
		}

		// Retry for IAM eventual consistency on error:
		// InvalidRequestException: DataSync location access test failed: could not perform s3:ListObjectsV2 on bucket
		if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "access test failed") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateLocationS3(input)
	}

	if err != nil {
		return fmt.Errorf("error creating DataSync Location S3: %s", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return resourceLocationS3Read(d, meta)
}

func resourceLocationS3Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &datasync.DescribeLocationS3Input{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DataSync Location S3: %s", input)
	output, err := conn.DescribeLocationS3(input)

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
		log.Printf("[WARN] DataSync Location S3 %q not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Location S3 (%s): %s", d.Id(), err)
	}

	subdirectory, err := SubdirectoryFromLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return err
	}

	d.Set("agent_arns", flex.FlattenStringSet(output.AgentArns))
	d.Set("arn", output.LocationArn)
	if err := d.Set("s3_config", flattenDataSyncS3Config(output.S3Config)); err != nil {
		return fmt.Errorf("error setting s3_config: %s", err)
	}
	d.Set("s3_storage_class", output.S3StorageClass)
	d.Set("subdirectory", subdirectory)
	d.Set("uri", output.LocationUri)

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Location S3 (%s): %s", d.Id(), err)
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

func resourceLocationS3Update(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating DataSync Location S3 (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceLocationS3Read(d, meta)
}

func resourceLocationS3Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location S3: %s", input)
	_, err := conn.DeleteLocation(input)

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Location S3 (%s): %s", d.Id(), err)
	}

	return nil
}
