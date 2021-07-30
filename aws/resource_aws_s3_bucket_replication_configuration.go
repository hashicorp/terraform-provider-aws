package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsS3BucketReplicationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketReplicationConfigurationCreate,
		Read:   resourceAwsS3BucketReplicationConfigurationRead,
		Update: resourceAwsS3BucketReplicationConfigurationUpdate,
		Delete: resourceAwsS3BucketReplicationConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"bucket_prefix"},
				ValidateFunc:  validation.StringLenBetween(0, 63),
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rules": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      rulesHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"destination": {
							Type:     schema.TypeList,
							MaxItems: 1,
							MinItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateAwsAccountId,
									},
									"bucket": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
									"storage_class": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(s3.StorageClass_Values(), false),
									},
									"replica_kms_key_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"access_control_translation": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"owner": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(s3.OwnerOverride_Values(), false),
												},
											},
										},
									},
								},
							},
						},
						"source_selection_criteria": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sse_kms_encrypted_objects": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
						"status": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(s3.ReplicationRuleStatus_Values(), false),
						},
						"priority": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"filter": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"prefix": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
									},
									"tags": tagsSchema(),
								},
							},
						},
						"delete_marker_replication_status": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{s3.DeleteMarkerReplicationStatusEnabled}, false),
						},
					},
				},
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsS3BucketReplicationConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsS3BucketReplicationConfigurationUpdate(d, meta)
}

func resourceAwsS3BucketReplicationConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsS3BucketReplicationConfigurationRead(d, meta)
}

func resourceAwsS3BucketReplicationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	input := &s3.HeadBucketInput{
		Bucket: aws.String(d.Get("bucket").(string)),
	}

	err := resource.Retry(s3BucketCreationTimeout, func() *resource.RetryError {
		_, err := s3conn.HeadBucket(input)

		if d.IsNewResource() && isAWSErrRequestFailureStatusCode(err, 404) {
			return resource.RetryableError(err)
		}

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	// Read the bucket replication configuration
	replicationResponse, err := retryOnAwsCode(s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
		return s3conn.GetBucketReplication(&s3.GetBucketReplicationInput{
			Bucket: aws.String(d.Get("bucket").(string)),
		})
	})
	if err != nil && !isAWSErr(err, "ReplicationConfigurationNotFoundError", "") {
		return fmt.Errorf("error getting S3 Bucket replication: %s", err)
	}
	replication, ok := replicationResponse.(*s3.GetBucketReplicationOutput)
	if !ok || replication == nil {
		return fmt.Errorf("error reading replication_configuration")
	}

	return nil
}

func resourceAwsS3BucketReplicationConfigurationDelete(d *schema.ResourceData, meta interface{}) error {

	return nil
}
