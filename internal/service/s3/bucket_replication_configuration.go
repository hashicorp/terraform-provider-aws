package s3

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBucketReplicationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceBucketReplicationConfigurationCreate,
		Read:   resourceBucketReplicationConfigurationRead,
		Update: resourceBucketReplicationConfigurationUpdate,
		Delete: resourceBucketReplicationConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1000,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_marker_replication": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(s3.DeleteMarkerReplicationStatus_Values(), false),
									},
								},
							},
						},
						"destination": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_control_translation": {
										Type:     schema.TypeList,
										Optional: true,
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
									"account": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidAccountID,
									},
									"bucket": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"encryption_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"replica_kms_key_id": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"metrics": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"event_threshold": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"minutes": {
																Type:     schema.TypeInt,
																Required: true,
																// Currently, the S3 API only supports 15 minutes;
																// however, to account for future changes, validation
																// is left at positive integers.
																ValidateFunc: validation.IntAtLeast(0),
															},
														},
													},
												},
												"status": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(s3.MetricsStatus_Values(), false),
												},
											},
										},
									},
									"replication_time": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(s3.ReplicationTimeStatus_Values(), false),
												},
												"time": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"minutes": {
																Type:     schema.TypeInt,
																Required: true,
																// Currently, the S3 API only supports 15 minutes;
																// however, to account for future changes, validation
																// is left at positive integers.
																ValidateFunc: validation.IntAtLeast(0),
															},
														},
													},
												},
											},
										},
									},
									"storage_class": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(s3.StorageClass_Values(), false),
									},
								},
							},
						},
						"existing_object_replication": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(s3.ExistingObjectReplicationStatus_Values(), false),
									},
								},
							},
						},
						"filter": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"and": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 1024),
												},
												"tags": tftags.TagsSchema(),
											},
										},
									},
									"prefix": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
									},
									"tag": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:     schema.TypeString,
													Required: true,
												},
												"value": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
							Deprecated:   "Use filter instead",
						},
						"priority": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"source_selection_criteria": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"replica_modifications": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(s3.ReplicaModificationsStatus_Values(), false),
												},
											},
										},
									},
									"sse_kms_encrypted_objects": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(s3.SseKmsEncryptedObjectsStatus_Values(), false),
												},
											},
										},
									},
								},
							},
						},
						"status": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(s3.ReplicationRuleStatus_Values(), false),
						},
					},
				},
			},
		},
	}
}

func resourceBucketReplicationConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket := d.Get("bucket").(string)

	rc := &s3.ReplicationConfiguration{
		Role:  aws.String(d.Get("role").(string)),
		Rules: ExpandReplicationRules(d.Get("rule").([]interface{})),
	}

	input := &s3.PutBucketReplicationInput{
		Bucket:                   aws.String(bucket),
		ReplicationConfiguration: rc,
	}

	if v, ok := d.GetOk("token"); ok {
		input.Token = aws.String(v.(string))
	}

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err := conn.PutBucketReplication(input)
		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) || tfawserr.ErrMessageContains(err, "InvalidRequest", "Versioning must be 'Enabled' on the bucket") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutBucketReplication(input)
	}

	if err != nil {
		return fmt.Errorf("error creating S3 replication configuration for bucket (%s): %w", bucket, err)
	}

	d.SetId(bucket)

	return resourceBucketReplicationConfigurationRead(d, meta)
}

func resourceBucketReplicationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	input := &s3.GetBucketReplicationInput{
		Bucket: aws.String(d.Id()),
	}

	// Read the bucket replication configuration
	output, err := retryWhenBucketNotFound(func() (interface{}, error) {
		return conn.GetBucketReplication(input)
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeReplicationConfigurationNotFound, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket Replication Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting S3 Bucket Replication Configuration for bucket (%s): %w", d.Id(), err)
	}

	replication, ok := output.(*s3.GetBucketReplicationOutput)

	if !ok || replication == nil || replication.ReplicationConfiguration == nil {
		return fmt.Errorf("error reading S3 Bucket Replication Configuration for bucket (%s): empty output", d.Id())
	}

	r := replication.ReplicationConfiguration

	d.Set("bucket", d.Id())
	d.Set("role", r.Role)
	if err := d.Set("rule", FlattenReplicationRules(r.Rules)); err != nil {
		return fmt.Errorf("error setting rule: %w", err)
	}

	return nil
}

func resourceBucketReplicationConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	rc := &s3.ReplicationConfiguration{
		Role:  aws.String(d.Get("role").(string)),
		Rules: ExpandReplicationRules(d.Get("rule").([]interface{})),
	}

	input := &s3.PutBucketReplicationInput{
		Bucket:                   aws.String(d.Id()),
		ReplicationConfiguration: rc,
	}

	if v, ok := d.GetOk("token"); ok {
		input.Token = aws.String(v.(string))
	}

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err := conn.PutBucketReplication(input)
		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) || tfawserr.ErrMessageContains(err, "InvalidRequest", "Versioning must be 'Enabled' on the bucket") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutBucketReplication(input)
	}

	if err != nil {
		return fmt.Errorf("error updating S3 replication configuration for bucket (%s): %w", d.Id(), err)
	}

	return resourceBucketReplicationConfigurationRead(d, meta)
}

func resourceBucketReplicationConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	input := &s3.DeleteBucketReplicationInput{
		Bucket: aws.String(d.Id()),
	}

	_, err := conn.DeleteBucketReplication(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeReplicationConfigurationNotFound, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 bucket replication configuration for bucket (%s): %w", d.Id(), err)
	}

	return nil
}
