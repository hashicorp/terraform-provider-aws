package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBucket() *schema.Resource {
	return &schema.Resource{
		Create:               resourceBucketCreate,
		Read:                 resourceBucketRead,
		Update:               resourceBucketUpdate,
		DeleteWithoutTimeout: resourceBucketDelete,

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
			"bucket_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"bucket"},
				ValidateFunc:  validation.StringLenBetween(0, 63-resource.UniqueIDSuffixLength),
			},

			"bucket_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"bucket_regional_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"acl": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"grant"},
				ValidateFunc:  validation.StringInSlice(BucketCannedACL_Values(), false),
				Deprecated:    "Use the aws_s3_bucket_acl resource instead",
			},

			"grant": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"acl"},
				Deprecated:    "Use the aws_s3_bucket_acl resource instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							// TypeAmazonCustomerByEmail is not currently supported
							ValidateFunc: validation.StringInSlice([]string{
								s3.TypeCanonicalUser,
								s3.TypeGroup,
							}, false),
						},
						"uri": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"permissions": {
							Type:     schema.TypeSet,
							Required: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(s3.Permission_Values(), false),
							},
						},
					},
				},
			},

			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Deprecated:       "Use the aws_s3_bucket_policy resource instead",
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
			},

			"cors_rule": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				Deprecated: "Use the aws_s3_bucket_cors_configuration resource instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_methods": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_origins": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"expose_headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"max_age_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},

			"website": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				MaxItems:   1,
				Deprecated: "Use the aws_s3_bucket_website_configuration resource instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_document": {
							Type:     schema.TypeString,
							Optional: true,
							ExactlyOneOf: []string{
								"website.0.index_document",
								"website.0.redirect_all_requests_to",
							},
						},

						"error_document": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"redirect_all_requests_to": {
							Type: schema.TypeString,
							ExactlyOneOf: []string{
								"website.0.index_document",
								"website.0.redirect_all_requests_to",
							},
							ConflictsWith: []string{
								"website.0.error_document",
								"website.0.routing_rules",
							},
							Optional: true,
						},

						"routing_rules": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsJSON,
							StateFunc: func(v interface{}) string {
								json, _ := structure.NormalizeJsonString(v)
								return json
							},
						},
					},
				},
			},

			"hosted_zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website_endpoint": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "Use the aws_s3_bucket_website_configuration resource",
			},
			"website_domain": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "Use the aws_s3_bucket_website_configuration resource",
			},

			"versioning": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				MaxItems:   1,
				Deprecated: "Use the aws_s3_bucket_versioning resource instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"mfa_delete": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"logging": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				MaxItems:   1,
				Deprecated: "Use the aws_s3_bucket_logging resource instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"lifecycle_rule": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				Deprecated: "Use the aws_s3_bucket_lifecycle_configuration resource instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tags": tftags.TagsSchema(),
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"abort_incomplete_multipart_upload_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"expiration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validBucketLifecycleTimestamp,
									},
									"days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"expired_object_delete_marker": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"noncurrent_version_expiration": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"transition": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validBucketLifecycleTimestamp,
									},
									"days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"storage_class": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(s3.TransitionStorageClass_Values(), false),
									},
								},
							},
						},
						"noncurrent_version_transition": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"storage_class": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(s3.TransitionStorageClass_Values(), false),
									},
								},
							},
						},
					},
				},
			},

			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"acceleration_status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Deprecated:   "Use the aws_s3_bucket_accelerate_configuration resource instead",
				ValidateFunc: validation.StringInSlice(s3.BucketAccelerateStatus_Values(), false),
			},

			"request_payer": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Deprecated:   "Use the aws_s3_bucket_request_payment_configuration resource instead",
				ValidateFunc: validation.StringInSlice(s3.Payer_Values(), false),
			},

			"replication_configuration": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				MaxItems:   1,
				Deprecated: "Use the aws_s3_bucket_replication_configuration resource instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role": {
							Type:     schema.TypeString,
							Required: true,
						},
						"rules": {
							Type:     schema.TypeSet,
							Required: true,
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
													ValidateFunc: verify.ValidAccountID,
												},
												"bucket": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidARN,
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
												"replication_time": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"minutes": {
																Type:         schema.TypeInt,
																Optional:     true,
																Default:      15,
																ValidateFunc: validation.IntBetween(15, 15),
															},
															"status": {
																Type:         schema.TypeString,
																Optional:     true,
																Default:      s3.ReplicationTimeStatusEnabled,
																ValidateFunc: validation.StringInSlice(s3.ReplicationTimeStatus_Values(), false),
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
															"minutes": {
																Type:         schema.TypeInt,
																Optional:     true,
																Default:      15,
																ValidateFunc: validation.IntBetween(10, 15),
															},
															"status": {
																Type:         schema.TypeString,
																Optional:     true,
																Default:      s3.MetricsStatusEnabled,
																ValidateFunc: validation.StringInSlice(s3.MetricsStatus_Values(), false),
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
												"tags": tftags.TagsSchema(),
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
					},
				},
			},

			"server_side_encryption_configuration": {
				Type:       schema.TypeList,
				MaxItems:   1,
				Optional:   true,
				Computed:   true,
				Deprecated: "Use the aws_s3_bucket_server_side_encryption_configuration resource instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"apply_server_side_encryption_by_default": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"kms_master_key_id": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"sse_algorithm": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(s3.ServerSideEncryption_Values(), false),
												},
											},
										},
									},
									"bucket_key_enabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},

			"object_lock_enabled": {
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true, // Can be removed when object_lock_configuration.0.object_lock_enabled is removed
				ForceNew:      true,
				ConflictsWith: []string{"object_lock_configuration"},
			},

			"object_lock_configuration": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				MaxItems:   1,
				Deprecated: "Use the top-level parameter object_lock_enabled and the aws_s3_bucket_object_lock_configuration resource instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object_lock_enabled": {
							Type:          schema.TypeString,
							Optional:      true,
							ForceNew:      true,
							ConflictsWith: []string{"object_lock_enabled"},
							ValidateFunc:  validation.StringInSlice(s3.ObjectLockEnabled_Values(), false),
							Deprecated:    "Use the top-level parameter object_lock_enabled instead",
						},

						"rule": {
							Type:       schema.TypeList,
							Optional:   true,
							Deprecated: "Use the aws_s3_bucket_object_lock_configuration resource instead",
							MaxItems:   1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_retention": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mode": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(s3.ObjectLockRetentionMode_Values(), false),
												},

												"days": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},

												"years": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBucketCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	// Get the bucket and acl
	var bucket string
	if v, ok := d.GetOk("bucket"); ok {
		bucket = v.(string)
	} else if v, ok := d.GetOk("bucket_prefix"); ok {
		bucket = resource.PrefixedUniqueId(v.(string))
	} else {
		bucket = resource.UniqueId()
	}

	log.Printf("[DEBUG] S3 bucket create: %s", bucket)

	req := &s3.CreateBucketInput{
		Bucket:                     aws.String(bucket),
		ObjectLockEnabledForBucket: aws.Bool(d.Get("object_lock_enabled").(bool)),
	}

	if acl, ok := d.GetOk("acl"); ok {
		acl := acl.(string)
		req.ACL = aws.String(acl)
		log.Printf("[DEBUG] S3 bucket %s has canned ACL %s", bucket, acl)
	} else {
		// Use default value previously available in v3.x of the provider
		req.ACL = aws.String(s3.BucketCannedACLPrivate)
		log.Printf("[DEBUG] S3 bucket %s has default canned ACL %s", bucket, s3.BucketCannedACLPrivate)
	}

	awsRegion := meta.(*conns.AWSClient).Region
	log.Printf("[DEBUG] S3 bucket create: %s, using region: %s", bucket, awsRegion)

	// Special case us-east-1 region and do not set the LocationConstraint.
	// See "Request Elements: http://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketPUT.html
	if awsRegion != endpoints.UsEast1RegionID {
		req.CreateBucketConfiguration = &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(awsRegion),
		}
	}

	if err := ValidBucketName(bucket, awsRegion); err != nil {
		return fmt.Errorf("error validating S3 Bucket (%s) name: %w", bucket, err)
	}

	// S3 Object Lock can only be enabled on bucket creation.
	objectLockConfiguration := expandObjectLockConfiguration(d.Get("object_lock_configuration").([]interface{}))
	if objectLockConfiguration != nil && aws.StringValue(objectLockConfiguration.ObjectLockEnabled) == s3.ObjectLockEnabledEnabled {
		req.ObjectLockEnabledForBucket = aws.Bool(true)
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateBucket(req)

		if tfawserr.ErrCodeEquals(err, ErrCodeOperationAborted) {
			return resource.RetryableError(fmt.Errorf("error creating S3 Bucket (%s), retrying: %w", bucket, err))
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.CreateBucket(req)
	}
	if err != nil {
		return fmt.Errorf("error creating S3 Bucket (%s): %w", bucket, err)
	}

	// Assign the bucket name as the resource ID
	d.SetId(bucket)
	return resourceBucketUpdate(d, meta)
}

func resourceBucketUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		// Retry due to S3 eventual consistency
		_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
			terr := BucketUpdateTags(conn, d.Id(), o, n)
			return nil, terr
		}, s3.ErrCodeNoSuchBucket)
		if err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) tags: %s", d.Id(), err)
		}
	}

	// Note: Order of argument updates below is important

	if d.HasChange("policy") {
		if err := resourceBucketInternalPolicyUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) Policy: %w", d.Id(), err)
		}
	}

	if d.HasChange("cors_rule") {
		if err := resourceBucketInternalCorsUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) CORS Rules: %w", d.Id(), err)
		}
	}

	if d.HasChange("website") {
		if err := resourceBucketInternalWebsiteUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) Website: %w", d.Id(), err)
		}
	}

	if d.HasChange("versioning") {
		v := d.Get("versioning").([]interface{})

		if d.IsNewResource() {
			if versioning := expandVersioningWhenIsNewResource(v); versioning != nil {
				err := resourceBucketInternalVersioningUpdate(conn, d.Id(), versioning)
				if err != nil {
					return fmt.Errorf("error updating (new) S3 Bucket (%s) Versioning: %w", d.Id(), err)
				}
			}
		} else {
			if err := resourceBucketInternalVersioningUpdate(conn, d.Id(), expandVersioning(v)); err != nil {
				return fmt.Errorf("error updating S3 Bucket (%s) Versioning: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("acl") && !d.IsNewResource() {
		if err := resourceBucketInternalACLUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) ACL: %w", d.Id(), err)
		}
	}

	if d.HasChange("grant") {
		if err := resourceBucketInternalGrantsUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) Grants: %w", d.Id(), err)
		}
	}

	if d.HasChange("logging") {
		if err := resourceBucketInternalLoggingUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) Logging: %w", d.Id(), err)
		}
	}

	if d.HasChange("lifecycle_rule") {
		if err := resourceBucketInternalLifecycleUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) Lifecycle Rules: %w", d.Id(), err)
		}
	}

	if d.HasChange("acceleration_status") {
		if err := resourceBucketInternalAccelerationUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) Acceleration Status: %w", d.Id(), err)
		}
	}

	if d.HasChange("request_payer") {
		if err := resourceBucketInternalRequestPayerUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) Request Payer: %w", d.Id(), err)
		}
	}

	if d.HasChange("replication_configuration") {
		if err := resourceBucketInternalReplicationConfigurationUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) Replication configuration: %w", d.Id(), err)
		}
	}

	if d.HasChange("server_side_encryption_configuration") {
		if err := resourceBucketInternalServerSideEncryptionConfigurationUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) Server-side Encryption configuration: %w", d.Id(), err)
		}
	}

	if d.HasChange("object_lock_configuration") {
		if err := resourceBucketInternalObjectLockConfigurationUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) Object Lock configuration: %w", d.Id(), err)
		}
	}

	return resourceBucketRead(d, meta)
}

func resourceBucketRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &s3.HeadBucketInput{
		Bucket: aws.String(d.Id()),
	}

	err := resource.Retry(bucketCreatedTimeout, func() *resource.RetryError {
		_, err := conn.HeadBucket(input)

		if d.IsNewResource() && tfawserr.ErrStatusCodeEquals(err, http.StatusNotFound) {
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

	if tfresource.TimedOut(err) {
		_, err = conn.HeadBucket(input)
	}

	if !d.IsNewResource() && tfawserr.ErrStatusCodeEquals(err, http.StatusNotFound) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Bucket (%s): %w", d.Id(), err)
	}

	d.Set("bucket", d.Id())

	d.Set("bucket_domain_name", meta.(*conns.AWSClient).PartitionHostname(fmt.Sprintf("%s.s3", d.Get("bucket").(string))))

	// Read the policy if configured outside this resource e.g. with aws_s3_bucket_policy resource
	pol, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetBucketPolicy(&s3.GetBucketPolicyInput{
			Bucket: aws.String(d.Id()),
		})
	}, s3.ErrCodeNoSuchBucket)

	// The call to HeadBucket above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketPolicy, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil && !tfawserr.ErrCodeEquals(err, ErrCodeNoSuchBucketPolicy, ErrCodeNotImplemented) {
		return fmt.Errorf("error getting S3 bucket (%s) policy: %w", d.Id(), err)
	}

	if output, ok := pol.(*s3.GetBucketPolicyOutput); ok {
		d.Set("policy", output.Policy)
	} else {
		d.Set("policy", nil)
	}

	// Read the Grant ACL.
	// In the event grants are not configured on the bucket, the API returns an empty array
	apResponse, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetBucketAcl(&s3.GetBucketAclInput{
			Bucket: aws.String(d.Id()),
		})
	}, s3.ErrCodeNoSuchBucket)

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketAcl, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting S3 Bucket (%s) ACL: %w", d.Id(), err)
	}

	if aclOutput, ok := apResponse.(*s3.GetBucketAclOutput); ok {
		if err := d.Set("grant", flattenGrants(aclOutput)); err != nil {
			return fmt.Errorf("error setting grant %s", err)
		}
	} else {
		d.Set("grant", nil)
	}

	// Read the CORS
	corsResponse, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetBucketCors(&s3.GetBucketCorsInput{
			Bucket: aws.String(d.Id()),
		})
	}, s3.ErrCodeNoSuchBucket)

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketCors, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil && !tfawserr.ErrCodeEquals(err, ErrCodeNoSuchCORSConfiguration) {
		return fmt.Errorf("error getting S3 Bucket CORS configuration: %s", err)
	}

	if output, ok := corsResponse.(*s3.GetBucketCorsOutput); ok {
		if err := d.Set("cors_rule", flattenBucketCorsRules(output.CORSRules)); err != nil {
			return fmt.Errorf("error setting cors_rule: %w", err)
		}
	} else {
		d.Set("cors_rule", nil)
	}

	// Read the website configuration
	wsResponse, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetBucketWebsite(&s3.GetBucketWebsiteInput{
			Bucket: aws.String(d.Id()),
		})
	}, s3.ErrCodeNoSuchBucket)

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketWebsite, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil && !tfawserr.ErrCodeEquals(err,
		ErrCodeMethodNotAllowed,
		ErrCodeNotImplemented,
		ErrCodeNoSuchWebsiteConfiguration,
		ErrCodeXNotImplemented,
	) {
		return fmt.Errorf("error getting S3 Bucket website configuration: %w", err)
	}

	if ws, ok := wsResponse.(*s3.GetBucketWebsiteOutput); ok {
		website, err := flattenBucketWebsite(ws)
		if err != nil {
			return err
		}
		if err := d.Set("website", website); err != nil {
			return fmt.Errorf("error setting website: %w", err)
		}
	} else {
		d.Set("website", nil)
	}

	// Read the versioning configuration

	versioningResponse, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetBucketVersioning(&s3.GetBucketVersioningInput{
			Bucket: aws.String(d.Id()),
		})
	}, s3.ErrCodeNoSuchBucket)

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketVersioning, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting S3 Bucket versioning (%s): %w", d.Id(), err)
	}

	if versioning, ok := versioningResponse.(*s3.GetBucketVersioningOutput); ok {
		if err := d.Set("versioning", flattenVersioning(versioning)); err != nil {
			return fmt.Errorf("error setting versioning: %w", err)
		}
	}

	// Read the acceleration status

	accelerateResponse, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetBucketAccelerateConfiguration(&s3.GetBucketAccelerateConfigurationInput{
			Bucket: aws.String(d.Id()),
		})
	}, s3.ErrCodeNoSuchBucket)

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketAccelerateConfiguration, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	// Amazon S3 Transfer Acceleration might not be supported in the region
	if err != nil && !tfawserr.ErrCodeEquals(err, ErrCodeMethodNotAllowed, ErrCodeUnsupportedArgument, ErrCodeNotImplemented) {
		return fmt.Errorf("error getting S3 Bucket (%s) accelerate configuration: %w", d.Id(), err)
	}

	if accelerate, ok := accelerateResponse.(*s3.GetBucketAccelerateConfigurationOutput); ok {
		d.Set("acceleration_status", accelerate.Status)
	}

	// Read the request payer configuration.

	payerResponse, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetBucketRequestPayment(&s3.GetBucketRequestPaymentInput{
			Bucket: aws.String(d.Id()),
		})
	}, s3.ErrCodeNoSuchBucket)

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketRequestPayment, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil && !tfawserr.ErrCodeEquals(err, ErrCodeNotImplemented) {
		return fmt.Errorf("error getting S3 Bucket request payment: %s", err)
	}

	if payer, ok := payerResponse.(*s3.GetBucketRequestPaymentOutput); ok {
		d.Set("request_payer", payer.Payer)
	}

	// Read the logging configuration if configured outside this resource
	loggingResponse, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetBucketLogging(&s3.GetBucketLoggingInput{
			Bucket: aws.String(d.Id()),
		})
	}, s3.ErrCodeNoSuchBucket)

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketLogging, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil && !tfawserr.ErrCodeEquals(err, ErrCodeNotImplemented) {
		return fmt.Errorf("error getting S3 Bucket logging: %s", err)
	}

	if logging, ok := loggingResponse.(*s3.GetBucketLoggingOutput); ok {
		if err := d.Set("logging", flattenBucketLoggingEnabled(logging.LoggingEnabled)); err != nil {
			return fmt.Errorf("error setting logging: %s", err)
		}
	} else {
		d.Set("logging", nil)
	}

	// Read the lifecycle configuration

	lifecycleResponse, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetBucketLifecycleConfiguration(&s3.GetBucketLifecycleConfigurationInput{
			Bucket: aws.String(d.Id()),
		})
	}, s3.ErrCodeNoSuchBucket)

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketLifecycleConfiguration, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil && !tfawserr.ErrCodeEquals(err, ErrCodeNoSuchLifecycleConfiguration) {
		return fmt.Errorf("error getting S3 Bucket (%s) Lifecycle Configuration: %w", d.Id(), err)
	}

	if lifecycle, ok := lifecycleResponse.(*s3.GetBucketLifecycleConfigurationOutput); ok {
		if err := d.Set("lifecycle_rule", flattenBucketLifecycleRules(lifecycle.Rules)); err != nil {
			return fmt.Errorf("error setting lifecycle_rule: %s", err)
		}
	} else {
		d.Set("lifecycle_rule", nil)
	}

	// Read the bucket replication configuration if configured outside this resource

	replicationResponse, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetBucketReplication(&s3.GetBucketReplicationInput{
			Bucket: aws.String(d.Id()),
		})
	}, s3.ErrCodeNoSuchBucket)

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketReplication, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil && !tfawserr.ErrCodeEquals(err, ErrCodeReplicationConfigurationNotFound) {
		return fmt.Errorf("error getting S3 Bucket replication: %w", err)
	}

	if replication, ok := replicationResponse.(*s3.GetBucketReplicationOutput); ok {
		if err := d.Set("replication_configuration", flattenBucketReplicationConfiguration(replication.ReplicationConfiguration)); err != nil {
			return fmt.Errorf("error setting replication_configuration: %w", err)
		}
	} else {
		// Still need to set for the non-existent case
		d.Set("replication_configuration", nil)
	}

	// Read the bucket server side encryption configuration

	encryptionResponse, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetBucketEncryption(&s3.GetBucketEncryptionInput{
			Bucket: aws.String(d.Id()),
		})
	}, s3.ErrCodeNoSuchBucket)

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketEncryption, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil && !tfawserr.ErrMessageContains(err, ErrCodeServerSideEncryptionConfigurationNotFound, "encryption configuration was not found") {
		return fmt.Errorf("error getting S3 Bucket encryption: %w", err)
	}

	if encryption, ok := encryptionResponse.(*s3.GetBucketEncryptionOutput); ok {
		if err := d.Set("server_side_encryption_configuration", flattenServerSideEncryptionConfiguration(encryption.ServerSideEncryptionConfiguration)); err != nil {
			return fmt.Errorf("error setting server_side_encryption_configuration: %w", err)
		}
	} else {
		d.Set("server_side_encryption_configuration", nil)
	}

	// Object Lock configuration.
	resp, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetObjectLockConfiguration(&s3.GetObjectLockConfigurationInput{
			Bucket: aws.String(d.Id()),
		})
	}, s3.ErrCodeNoSuchBucket)

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetObjectLockConfiguration, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	// Object lock not supported in all partitions (extra guard, also guards in read func)
	if err != nil && !tfawserr.ErrCodeEquals(err, ErrCodeMethodNotAllowed, ErrCodeNotImplemented, ErrCodeObjectLockConfigurationNotFound) {
		if meta.(*conns.AWSClient).Partition == endpoints.AwsPartitionID || meta.(*conns.AWSClient).Partition == endpoints.AwsUsGovPartitionID {
			return fmt.Errorf("error getting S3 Bucket (%s) Object Lock configuration: %w", d.Id(), err)
		}
	}

	if err != nil {
		log.Printf("[WARN] Unable to read S3 bucket (%s) Object Lock Configuration: %s", d.Id(), err)
	}

	if output, ok := resp.(*s3.GetObjectLockConfigurationOutput); ok && output.ObjectLockConfiguration != nil {
		d.Set("object_lock_enabled", aws.StringValue(output.ObjectLockConfiguration.ObjectLockEnabled) == s3.ObjectLockEnabledEnabled)
		if err := d.Set("object_lock_configuration", flattenObjectLockConfiguration(output.ObjectLockConfiguration)); err != nil {
			return fmt.Errorf("error setting object_lock_configuration: %w", err)
		}
	} else {
		d.Set("object_lock_enabled", false)
		d.Set("object_lock_configuration", nil)
	}

	// Add the region as an attribute
	discoveredRegion, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return s3manager.GetBucketRegionWithClient(context.Background(), conn, d.Id(), func(r *request.Request) {
			// By default, GetBucketRegion forces virtual host addressing, which
			// is not compatible with many non-AWS implementations. Instead, pass
			// the provider s3_force_path_style configuration, which defaults to
			// false, but allows override.
			r.Config.S3ForcePathStyle = conn.Config.S3ForcePathStyle

			// By default, GetBucketRegion uses anonymous credentials when doing
			// a HEAD request to get the bucket region. This breaks in aws-cn regions
			// when the account doesn't have an ICP license to host public content.
			// Use the current credentials when getting the bucket region.
			r.Config.Credentials = conn.Config.Credentials
		})
	}, "NotFound")

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as s3manager.GetBucketRegionWithClient, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting S3 Bucket location: %s", err)
	}

	region := discoveredRegion.(string)
	if err := d.Set("region", region); err != nil {
		return err
	}

	// Add the bucket_regional_domain_name as an attribute
	regionalEndpoint, err := BucketRegionalDomainName(d.Get("bucket").(string), region)
	if err != nil {
		return err
	}
	d.Set("bucket_regional_domain_name", regionalEndpoint)

	// Add the hosted zone ID for this bucket's region as an attribute
	hostedZoneID, err := HostedZoneIDForRegion(region)
	if err != nil {
		log.Printf("[WARN] %s", err)
	} else {
		d.Set("hosted_zone_id", hostedZoneID)
	}

	// Add website_endpoint as an attribute
	websiteEndpoint, err := websiteEndpoint(meta.(*conns.AWSClient), d)

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketLocation, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}
	if websiteEndpoint != nil {
		if err := d.Set("website_endpoint", websiteEndpoint.Endpoint); err != nil {
			return err
		}
		if err := d.Set("website_domain", websiteEndpoint.Domain); err != nil {
			return err
		}
	}

	// Retry due to S3 eventual consistency
	tagsRaw, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return BucketListTags(conn, d.Id())
	}, s3.ErrCodeNoSuchBucket)

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketTagging, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing tags for S3 Bucket (%s): %s", d.Id(), err)
	}

	tags, ok := tagsRaw.(tftags.KeyValueTags)

	if !ok {
		return fmt.Errorf("error listing tags for S3 Bucket (%s): unable to convert tags", d.Id())
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "s3",
		Resource:  d.Id(),
	}.String()
	d.Set("arn", arn)

	return nil
}

func resourceBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	log.Printf("[DEBUG] Deleting S3 Bucket: %s", d.Id())
	_, err := conn.DeleteBucketWithContext(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, ErrCodeBucketNotEmpty) {
		if d.Get("force_destroy").(bool) {
			// Use a S3 service client that can handle multiple slashes in URIs.
			// While aws_s3_object resources cannot create these object
			// keys, other AWS services and applications using the S3 Bucket can.
			conn = meta.(*conns.AWSClient).S3ConnURICleaningDisabled

			// bucket may have things delete them
			log.Printf("[DEBUG] S3 Bucket attempting to forceDestroy %s", err)

			// Delete everything including locked objects.
			// Don't ignore any object errors or we could recurse infinitely.
			var objectLockEnabled bool
			objectLockConfiguration := expandObjectLockConfiguration(d.Get("object_lock_configuration").([]interface{}))
			if objectLockConfiguration != nil {
				objectLockEnabled = aws.StringValue(objectLockConfiguration.ObjectLockEnabled) == s3.ObjectLockEnabledEnabled
			}

			if n, err := EmptyBucket(ctx, conn, d.Id(), objectLockEnabled); err != nil {
				return diag.Errorf("emptying S3 Bucket (%s): %s", d.Id(), err)
			} else {
				log.Printf("[DEBUG] Deleted %d S3 objects", n)
			}

			// this line recurses until all objects are deleted or an error is returned
			return resourceBucketDelete(ctx, d, meta)
		}
	}

	if err != nil {
		return diag.Errorf("deleting S3 Bucket (%s): %s", d.Id(), err)
	}

	return nil
}

// https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region
func BucketRegionalDomainName(bucket string, region string) (string, error) {
	// Return a default AWS Commercial domain name if no region is provided
	// Otherwise EndpointFor() will return BUCKET.s3..amazonaws.com
	if region == "" {
		return fmt.Sprintf("%s.s3.amazonaws.com", bucket), nil //lintignore:AWSR001
	}
	endpoint, err := endpoints.DefaultResolver().EndpointFor(endpoints.S3ServiceID, region)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s", bucket, strings.TrimPrefix(endpoint.URL, "https://")), nil
}

type S3Website struct {
	Endpoint, Domain string
}

func WebsiteEndpoint(client *conns.AWSClient, bucket string, region string) *S3Website {
	domain := WebsiteDomainURL(client, region)
	return &S3Website{Endpoint: fmt.Sprintf("%s.%s", bucket, domain), Domain: domain}
}

func WebsiteDomainURL(client *conns.AWSClient, region string) string {
	region = normalizeRegion(region)

	// Different regions have different syntax for website endpoints
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteEndpoints.html
	// https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_website_region_endpoints
	if isOldRegion(region) {
		return fmt.Sprintf("s3-website-%s.amazonaws.com", region) //lintignore:AWSR001
	}
	return client.RegionalHostname("s3-website")
}

func websiteEndpoint(client *conns.AWSClient, d *schema.ResourceData) (*S3Website, error) {
	// If the bucket doesn't have a website configuration, return an empty
	// endpoint
	if _, ok := d.GetOk("website"); !ok {
		return nil, nil
	}

	bucket := d.Get("bucket").(string)

	// Lookup the region for this bucket

	locationResponse, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return client.S3Conn.GetBucketLocation(
			&s3.GetBucketLocationInput{
				Bucket: aws.String(bucket),
			},
		)
	}, s3.ErrCodeNoSuchBucket)
	if err != nil {
		return nil, err
	}
	location := locationResponse.(*s3.GetBucketLocationOutput)
	var region string
	if location.LocationConstraint != nil {
		region = aws.StringValue(location.LocationConstraint)
	}

	return WebsiteEndpoint(client, bucket, region), nil
}

func isOldRegion(region string) bool {
	oldRegions := []string{
		endpoints.ApNortheast1RegionID,
		endpoints.ApSoutheast1RegionID,
		endpoints.ApSoutheast2RegionID,
		endpoints.EuWest1RegionID,
		endpoints.SaEast1RegionID,
		endpoints.UsEast1RegionID,
		endpoints.UsGovWest1RegionID,
		endpoints.UsWest1RegionID,
		endpoints.UsWest2RegionID,
	}
	for _, r := range oldRegions {
		if region == r {
			return true
		}
	}
	return false
}

func normalizeRegion(region string) string {
	// Default to us-east-1 if the bucket doesn't have a region:
	// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketGETlocation.html
	if region == "" {
		region = endpoints.UsEast1RegionID
	}

	return region
}

////////////////////////////////////////// Argument-Specific Update Functions //////////////////////////////////////////

func resourceBucketInternalAccelerationUpdate(conn *s3.S3, d *schema.ResourceData) error {
	input := &s3.PutBucketAccelerateConfigurationInput{
		Bucket: aws.String(d.Id()),
		AccelerateConfiguration: &s3.AccelerateConfiguration{
			Status: aws.String(d.Get("acceleration_status").(string)),
		},
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketAccelerateConfiguration(input)
	}, s3.ErrCodeNoSuchBucket)

	return err
}

func resourceBucketInternalACLUpdate(conn *s3.S3, d *schema.ResourceData) error {
	acl := d.Get("acl").(string)
	if acl == "" {
		// Use default value previously available in v3.x of the provider
		acl = s3.BucketCannedACLPrivate
	}

	input := &s3.PutBucketAclInput{
		Bucket: aws.String(d.Id()),
		ACL:    aws.String(acl),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketAcl(input)
	}, s3.ErrCodeNoSuchBucket)

	return err
}

func resourceBucketInternalCorsUpdate(conn *s3.S3, d *schema.ResourceData) error {
	rawCors := d.Get("cors_rule").([]interface{})

	if len(rawCors) == 0 {
		// Delete CORS
		_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
			return conn.DeleteBucketCors(&s3.DeleteBucketCorsInput{
				Bucket: aws.String(d.Id()),
			})
		}, s3.ErrCodeNoSuchBucket)

		if err != nil {
			return fmt.Errorf("error deleting S3 Bucket (%s) CORS: %w", d.Id(), err)
		}

		return nil
	}
	// Put CORS
	rules := make([]*s3.CORSRule, 0, len(rawCors))
	for _, cors := range rawCors {
		// Prevent panic
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7546
		corsMap, ok := cors.(map[string]interface{})
		if !ok {
			continue
		}
		r := &s3.CORSRule{}
		for k, v := range corsMap {
			if k == "max_age_seconds" {
				r.MaxAgeSeconds = aws.Int64(int64(v.(int)))
			} else {
				vMap := make([]*string, len(v.([]interface{})))
				for i, vv := range v.([]interface{}) {
					if str, ok := vv.(string); ok {
						vMap[i] = aws.String(str)
					}
				}
				switch k {
				case "allowed_headers":
					r.AllowedHeaders = vMap
				case "allowed_methods":
					r.AllowedMethods = vMap
				case "allowed_origins":
					r.AllowedOrigins = vMap
				case "expose_headers":
					r.ExposeHeaders = vMap
				}
			}
		}
		rules = append(rules, r)
	}

	input := &s3.PutBucketCorsInput{
		Bucket: aws.String(d.Id()),
		CORSConfiguration: &s3.CORSConfiguration{
			CORSRules: rules,
		},
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketCors(input)
	}, s3.ErrCodeNoSuchBucket)

	return err
}

func resourceBucketInternalGrantsUpdate(conn *s3.S3, d *schema.ResourceData) error {
	grants := d.Get("grant").(*schema.Set)

	if grants.Len() == 0 {
		log.Printf("[DEBUG] S3 bucket: %s, Grants fallback to canned ACL", d.Id())

		if err := resourceBucketInternalACLUpdate(conn, d); err != nil {
			return fmt.Errorf("error fallback to canned ACL, %s", err)
		}

		return nil
	}

	resp, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetBucketAcl(&s3.GetBucketAclInput{
			Bucket: aws.String(d.Id()),
		})
	}, s3.ErrCodeNoSuchBucket)

	if err != nil {
		return fmt.Errorf("error getting S3 Bucket (%s) ACL: %s", d.Id(), err)
	}

	output := resp.(*s3.GetBucketAclOutput)

	if output == nil {
		return fmt.Errorf("error getting S3 Bucket (%s) ACL: empty output", d.Id())
	}

	input := &s3.PutBucketAclInput{
		Bucket: aws.String(d.Id()),
		AccessControlPolicy: &s3.AccessControlPolicy{
			Grants: expandGrants(grants.List()),
			Owner:  output.Owner,
		},
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketAcl(input)
	}, s3.ErrCodeNoSuchBucket)

	return err
}

func resourceBucketInternalLifecycleUpdate(conn *s3.S3, d *schema.ResourceData) error {
	lifecycleRules := d.Get("lifecycle_rule").([]interface{})

	if len(lifecycleRules) == 0 || lifecycleRules[0] == nil {
		input := &s3.DeleteBucketLifecycleInput{
			Bucket: aws.String(d.Id()),
		}

		_, err := conn.DeleteBucketLifecycle(input)

		if err != nil {
			return fmt.Errorf("error removing S3 Bucket (%s) lifecycle: %w", d.Id(), err)
		}

		return nil
	}

	rules := make([]*s3.LifecycleRule, 0, len(lifecycleRules))

	for i, lifecycleRule := range lifecycleRules {
		r := lifecycleRule.(map[string]interface{})

		rule := &s3.LifecycleRule{}

		// Filter
		tags := Tags(tftags.New(r["tags"]).IgnoreAWS())
		filter := &s3.LifecycleRuleFilter{}
		if len(tags) > 0 {
			lifecycleRuleAndOp := &s3.LifecycleRuleAndOperator{}
			lifecycleRuleAndOp.SetPrefix(r["prefix"].(string))
			lifecycleRuleAndOp.SetTags(tags)
			filter.SetAnd(lifecycleRuleAndOp)
		} else {
			filter.SetPrefix(r["prefix"].(string))
		}
		rule.SetFilter(filter)

		// ID
		if val, ok := r["id"].(string); ok && val != "" {
			rule.ID = aws.String(val)
		} else {
			rule.ID = aws.String(resource.PrefixedUniqueId("tf-s3-lifecycle-"))
		}

		// Enabled
		if val, ok := r["enabled"].(bool); ok && val {
			rule.Status = aws.String(s3.ExpirationStatusEnabled)
		} else {
			rule.Status = aws.String(s3.ExpirationStatusDisabled)
		}

		// AbortIncompleteMultipartUpload
		if val, ok := r["abort_incomplete_multipart_upload_days"].(int); ok && val > 0 {
			rule.AbortIncompleteMultipartUpload = &s3.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: aws.Int64(int64(val)),
			}
		}

		// Expiration
		expiration := d.Get(fmt.Sprintf("lifecycle_rule.%d.expiration", i)).([]interface{})
		if len(expiration) > 0 && expiration[0] != nil {
			e := expiration[0].(map[string]interface{})
			i := &s3.LifecycleExpiration{}
			if val, ok := e["date"].(string); ok && val != "" {
				t, err := time.Parse(time.RFC3339, fmt.Sprintf("%sT00:00:00Z", val))
				if err != nil {
					return fmt.Errorf("Error Parsing AWS S3 Bucket Lifecycle Expiration Date: %s", err.Error())
				}
				i.Date = aws.Time(t)
			} else if val, ok := e["days"].(int); ok && val > 0 {
				i.Days = aws.Int64(int64(val))
			} else if val, ok := e["expired_object_delete_marker"].(bool); ok {
				i.ExpiredObjectDeleteMarker = aws.Bool(val)
			}
			rule.Expiration = i
		}

		// NoncurrentVersionExpiration
		nc_expiration := d.Get(fmt.Sprintf("lifecycle_rule.%d.noncurrent_version_expiration", i)).([]interface{})
		if len(nc_expiration) > 0 && nc_expiration[0] != nil {
			e := nc_expiration[0].(map[string]interface{})

			if val, ok := e["days"].(int); ok && val > 0 {
				rule.NoncurrentVersionExpiration = &s3.NoncurrentVersionExpiration{
					NoncurrentDays: aws.Int64(int64(val)),
				}
			}
		}

		// Transitions
		transitions := d.Get(fmt.Sprintf("lifecycle_rule.%d.transition", i)).(*schema.Set).List()
		if len(transitions) > 0 {
			rule.Transitions = make([]*s3.Transition, 0, len(transitions))
			for _, transition := range transitions {
				transition := transition.(map[string]interface{})
				i := &s3.Transition{}
				if val, ok := transition["date"].(string); ok && val != "" {
					t, err := time.Parse(time.RFC3339, fmt.Sprintf("%sT00:00:00Z", val))
					if err != nil {
						return fmt.Errorf("Error Parsing AWS S3 Bucket Lifecycle Expiration Date: %s", err.Error())
					}
					i.Date = aws.Time(t)
				} else if val, ok := transition["days"].(int); ok && val >= 0 {
					i.Days = aws.Int64(int64(val))
				}
				if val, ok := transition["storage_class"].(string); ok && val != "" {
					i.StorageClass = aws.String(val)
				}

				rule.Transitions = append(rule.Transitions, i)
			}
		}
		// NoncurrentVersionTransitions
		nc_transitions := d.Get(fmt.Sprintf("lifecycle_rule.%d.noncurrent_version_transition", i)).(*schema.Set).List()
		if len(nc_transitions) > 0 {
			rule.NoncurrentVersionTransitions = make([]*s3.NoncurrentVersionTransition, 0, len(nc_transitions))
			for _, transition := range nc_transitions {
				transition := transition.(map[string]interface{})
				i := &s3.NoncurrentVersionTransition{}
				if val, ok := transition["days"].(int); ok && val >= 0 {
					i.NoncurrentDays = aws.Int64(int64(val))
				}
				if val, ok := transition["storage_class"].(string); ok && val != "" {
					i.StorageClass = aws.String(val)
				}

				rule.NoncurrentVersionTransitions = append(rule.NoncurrentVersionTransitions, i)
			}
		}

		// As a lifecycle rule requires 1 or more transition/expiration actions,
		// we explicitly pass a default ExpiredObjectDeleteMarker value to be able to create
		// the rule while keeping the policy unaffected if the conditions are not met.
		if rule.Expiration == nil && rule.NoncurrentVersionExpiration == nil &&
			rule.Transitions == nil && rule.NoncurrentVersionTransitions == nil &&
			rule.AbortIncompleteMultipartUpload == nil {
			rule.Expiration = &s3.LifecycleExpiration{ExpiredObjectDeleteMarker: aws.Bool(false)}
		}

		rules = append(rules, rule)
	}

	input := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(d.Id()),
		LifecycleConfiguration: &s3.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketLifecycleConfiguration(input)
	}, s3.ErrCodeNoSuchBucket)

	return err
}

func resourceBucketInternalLoggingUpdate(conn *s3.S3, d *schema.ResourceData) error {
	logging := d.Get("logging").([]interface{})
	loggingStatus := &s3.BucketLoggingStatus{}

	if len(logging) > 0 {
		c := logging[0].(map[string]interface{})

		loggingEnabled := &s3.LoggingEnabled{}
		if val, ok := c["target_bucket"].(string); ok {
			loggingEnabled.TargetBucket = aws.String(val)
		}
		if val, ok := c["target_prefix"].(string); ok {
			loggingEnabled.TargetPrefix = aws.String(val)
		}

		loggingStatus.LoggingEnabled = loggingEnabled
	}

	input := &s3.PutBucketLoggingInput{
		Bucket:              aws.String(d.Id()),
		BucketLoggingStatus: loggingStatus,
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketLogging(input)
	}, s3.ErrCodeNoSuchBucket)

	return err
}

func resourceBucketInternalObjectLockConfigurationUpdate(conn *s3.S3, d *schema.ResourceData) error {
	// S3 Object Lock configuration cannot be deleted, only updated.
	req := &s3.PutObjectLockConfigurationInput{
		Bucket:                  aws.String(d.Id()),
		ObjectLockConfiguration: expandObjectLockConfiguration(d.Get("object_lock_configuration").([]interface{})),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutObjectLockConfiguration(req)
	}, s3.ErrCodeNoSuchBucket)

	return err
}

func resourceBucketInternalPolicyUpdate(conn *s3.S3, d *schema.ResourceData) error {
	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is an invalid JSON: %w", policy, err)
	}

	if policy == "" {
		_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
			return conn.DeleteBucketPolicy(&s3.DeleteBucketPolicyInput{
				Bucket: aws.String(d.Id()),
			})
		}, s3.ErrCodeNoSuchBucket)

		if err != nil {
			return fmt.Errorf("error deleting S3 Bucket (%s) policy: %w", d.Id(), err)
		}

		return nil
	}

	params := &s3.PutBucketPolicyInput{
		Bucket: aws.String(d.Id()),
		Policy: aws.String(policy),
	}

	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.PutBucketPolicy(params)
		if tfawserr.ErrCodeEquals(err, ErrCodeMalformedPolicy, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutBucketPolicy(params)
	}

	return err
}

func resourceBucketInternalReplicationConfigurationUpdate(conn *s3.S3, d *schema.ResourceData) error {
	replicationConfiguration := d.Get("replication_configuration").([]interface{})

	if len(replicationConfiguration) == 0 {
		input := &s3.DeleteBucketReplicationInput{
			Bucket: aws.String(d.Id()),
		}

		_, err := conn.DeleteBucketReplication(input)

		if err != nil {
			return fmt.Errorf("error removing S3 Bucket (%s) Replication: %w", d.Id(), err)
		}

		return nil
	}

	hasVersioning := false
	// Validate that bucket versioning is enabled
	if versioning, ok := d.GetOk("versioning"); ok {
		v := versioning.([]interface{})

		if v[0].(map[string]interface{})["enabled"].(bool) {
			hasVersioning = true
		}
	}

	if !hasVersioning {
		return fmt.Errorf("versioning must be enabled to allow S3 bucket replication")
	}

	input := &s3.PutBucketReplicationInput{
		Bucket:                   aws.String(d.Id()),
		ReplicationConfiguration: expandBucketReplicationConfiguration(replicationConfiguration),
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.PutBucketReplication(input)
		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) || tfawserr.ErrMessageContains(err, ErrCodeInvalidRequest, "Versioning must be 'Enabled' on the bucket") {
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

	return err
}

func resourceBucketInternalRequestPayerUpdate(conn *s3.S3, d *schema.ResourceData) error {
	payer := d.Get("request_payer").(string)

	input := &s3.PutBucketRequestPaymentInput{
		Bucket: aws.String(d.Id()),
		RequestPaymentConfiguration: &s3.RequestPaymentConfiguration{
			Payer: aws.String(payer),
		},
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketRequestPayment(input)
	}, s3.ErrCodeNoSuchBucket)

	return err
}

func resourceBucketInternalServerSideEncryptionConfigurationUpdate(conn *s3.S3, d *schema.ResourceData) error {
	serverSideEncryptionConfiguration := d.Get("server_side_encryption_configuration").([]interface{})

	if len(serverSideEncryptionConfiguration) == 0 {
		input := &s3.DeleteBucketEncryptionInput{
			Bucket: aws.String(d.Id()),
		}

		_, err := conn.DeleteBucketEncryption(input)

		if err != nil {
			return fmt.Errorf("error removing S3 Bucket (%s) Server-side Encryption: %w", d.Id(), err)
		}

		return nil
	}

	c := serverSideEncryptionConfiguration[0].(map[string]interface{})

	rc := &s3.ServerSideEncryptionConfiguration{}

	rcRules := c["rule"].([]interface{})
	var rules []*s3.ServerSideEncryptionRule
	for _, v := range rcRules {
		rr := v.(map[string]interface{})
		rrDefault := rr["apply_server_side_encryption_by_default"].([]interface{})
		sseAlgorithm := rrDefault[0].(map[string]interface{})["sse_algorithm"].(string)
		kmsMasterKeyId := rrDefault[0].(map[string]interface{})["kms_master_key_id"].(string)
		rcDefaultRule := &s3.ServerSideEncryptionByDefault{
			SSEAlgorithm: aws.String(sseAlgorithm),
		}
		if kmsMasterKeyId != "" {
			rcDefaultRule.KMSMasterKeyID = aws.String(kmsMasterKeyId)
		}
		rcRule := &s3.ServerSideEncryptionRule{
			ApplyServerSideEncryptionByDefault: rcDefaultRule,
		}

		if val, ok := rr["bucket_key_enabled"].(bool); ok {
			rcRule.BucketKeyEnabled = aws.Bool(val)
		}

		rules = append(rules, rcRule)
	}

	rc.Rules = rules

	input := &s3.PutBucketEncryptionInput{
		Bucket:                            aws.String(d.Id()),
		ServerSideEncryptionConfiguration: rc,
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(
		propagationTimeout,
		func() (interface{}, error) {
			return conn.PutBucketEncryption(input)
		},
		s3.ErrCodeNoSuchBucket,
		ErrCodeOperationAborted,
	)

	return err
}

func resourceBucketInternalVersioningUpdate(conn *s3.S3, bucket string, versioningConfig *s3.VersioningConfiguration) error {
	input := &s3.PutBucketVersioningInput{
		Bucket:                  aws.String(bucket),
		VersioningConfiguration: versioningConfig,
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketVersioning(input)
	}, s3.ErrCodeNoSuchBucket)

	return err
}

func resourceBucketInternalWebsiteUpdate(conn *s3.S3, d *schema.ResourceData) error {
	ws := d.Get("website").([]interface{})

	if len(ws) == 0 {
		input := &s3.DeleteBucketWebsiteInput{
			Bucket: aws.String(d.Id()),
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
			return conn.DeleteBucketWebsite(input)
		}, s3.ErrCodeNoSuchBucket)

		if err != nil {
			return fmt.Errorf("error deleting S3 Bucket (%s) Website: %w", d.Id(), err)
		}

		d.Set("website_endpoint", "")
		d.Set("website_domain", "")

		return nil
	}

	websiteConfig, err := expandWebsiteConfiguration(ws)
	if err != nil {
		return fmt.Errorf("error expanding S3 Bucket (%s) website configuration: %w", d.Id(), err)
	}

	input := &s3.PutBucketWebsiteInput{
		Bucket:               aws.String(d.Id()),
		WebsiteConfiguration: websiteConfig,
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketWebsite(input)
	}, s3.ErrCodeNoSuchBucket)

	return err
}

///////////////////////////////////////////// Expand and Flatten functions /////////////////////////////////////////////

// Cors Rule functions

func flattenBucketCorsRules(rules []*s3.CORSRule) []interface{} {
	var results []interface{}

	for _, rule := range rules {
		if rule == nil {
			continue
		}

		m := make(map[string]interface{})

		if len(rule.AllowedHeaders) > 0 {
			m["allowed_headers"] = flex.FlattenStringList(rule.AllowedHeaders)
		}

		if len(rule.AllowedMethods) > 0 {
			m["allowed_methods"] = flex.FlattenStringList(rule.AllowedMethods)
		}

		if len(rule.AllowedOrigins) > 0 {
			m["allowed_origins"] = flex.FlattenStringList(rule.AllowedOrigins)
		}

		if len(rule.ExposeHeaders) > 0 {
			m["expose_headers"] = flex.FlattenStringList(rule.ExposeHeaders)
		}

		if rule.MaxAgeSeconds != nil {
			m["max_age_seconds"] = int(aws.Int64Value(rule.MaxAgeSeconds))
		}

		results = append(results, m)
	}

	return results
}

// Grants functions

func expandGrants(l []interface{}) []*s3.Grant {
	var grants []*s3.Grant

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		if v, ok := tfMap["permissions"].(*schema.Set); ok {
			for _, rawPermission := range v.List() {
				permission, ok := rawPermission.(string)
				if !ok {
					continue
				}

				grantee := &s3.Grantee{}

				if v, ok := tfMap["id"].(string); ok && v != "" {
					grantee.SetID(v)
				}

				if v, ok := tfMap["type"].(string); ok && v != "" {
					grantee.SetType(v)
				}

				if v, ok := tfMap["uri"].(string); ok && v != "" {
					grantee.SetURI(v)
				}

				g := &s3.Grant{
					Grantee:    grantee,
					Permission: aws.String(permission),
				}

				grants = append(grants, g)
			}
		}
	}
	return grants
}

func flattenGrants(ap *s3.GetBucketAclOutput) []interface{} {
	if len(ap.Grants) == 0 {
		return []interface{}{}
	}

	getGrant := func(grants []interface{}, grantee map[string]interface{}) (interface{}, bool) {
		for _, pg := range grants {
			pgt := pg.(map[string]interface{})
			if pgt["type"] == grantee["type"] && pgt["id"] == grantee["id"] && pgt["uri"] == grantee["uri"] &&
				pgt["permissions"].(*schema.Set).Len() > 0 {
				return pg, true
			}
		}
		return nil, false
	}

	grants := make([]interface{}, 0, len(ap.Grants))
	for _, granteeObject := range ap.Grants {
		grantee := make(map[string]interface{})
		grantee["type"] = aws.StringValue(granteeObject.Grantee.Type)

		if granteeObject.Grantee.ID != nil {
			grantee["id"] = aws.StringValue(granteeObject.Grantee.ID)
		}
		if granteeObject.Grantee.URI != nil {
			grantee["uri"] = aws.StringValue(granteeObject.Grantee.URI)
		}
		if pg, ok := getGrant(grants, grantee); ok {
			pg.(map[string]interface{})["permissions"].(*schema.Set).Add(aws.StringValue(granteeObject.Permission))
		} else {
			grantee["permissions"] = schema.NewSet(schema.HashString, []interface{}{aws.StringValue(granteeObject.Permission)})
			grants = append(grants, grantee)
		}
	}

	return grants
}

// Lifecycle Rule functions

func flattenBucketLifecycleRuleExpiration(expiration *s3.LifecycleExpiration) []interface{} {
	if expiration == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if expiration.Date != nil {
		m["date"] = (aws.TimeValue(expiration.Date)).Format("2006-01-02")
	}
	if expiration.Days != nil {
		m["days"] = int(aws.Int64Value(expiration.Days))
	}
	if expiration.ExpiredObjectDeleteMarker != nil {
		m["expired_object_delete_marker"] = aws.BoolValue(expiration.ExpiredObjectDeleteMarker)
	}

	return []interface{}{m}
}

func flattenBucketLifecycleRules(lifecycleRules []*s3.LifecycleRule) []interface{} {
	if len(lifecycleRules) == 0 {
		return []interface{}{}
	}

	var results []interface{}

	for _, lifecycleRule := range lifecycleRules {
		if lifecycleRule == nil {
			continue
		}

		rule := make(map[string]interface{})

		// AbortIncompleteMultipartUploadDays
		if lifecycleRule.AbortIncompleteMultipartUpload != nil {
			if lifecycleRule.AbortIncompleteMultipartUpload.DaysAfterInitiation != nil {
				rule["abort_incomplete_multipart_upload_days"] = int(aws.Int64Value(lifecycleRule.AbortIncompleteMultipartUpload.DaysAfterInitiation))
			}
		}

		// ID
		if lifecycleRule.ID != nil {
			rule["id"] = aws.StringValue(lifecycleRule.ID)
		}

		// Filter
		if filter := lifecycleRule.Filter; filter != nil {
			if filter.And != nil {
				// Prefix
				if filter.And.Prefix != nil {
					rule["prefix"] = aws.StringValue(filter.And.Prefix)
				}
				// Tag
				if len(filter.And.Tags) > 0 {
					rule["tags"] = KeyValueTags(filter.And.Tags).IgnoreAWS().Map()
				}
			} else {
				// Prefix
				if filter.Prefix != nil {
					rule["prefix"] = aws.StringValue(filter.Prefix)
				}
				// Tag
				if filter.Tag != nil {
					rule["tags"] = KeyValueTags([]*s3.Tag{filter.Tag}).IgnoreAWS().Map()
				}
			}
		}

		// Prefix
		if lifecycleRule.Prefix != nil {
			rule["prefix"] = aws.StringValue(lifecycleRule.Prefix)
		}

		// Enabled
		if lifecycleRule.Status != nil {
			if aws.StringValue(lifecycleRule.Status) == s3.ExpirationStatusEnabled {
				rule["enabled"] = true
			} else {
				rule["enabled"] = false
			}
		}

		// Expiration
		if lifecycleRule.Expiration != nil {
			rule["expiration"] = flattenBucketLifecycleRuleExpiration(lifecycleRule.Expiration)
		}

		// NoncurrentVersionExpiration
		if lifecycleRule.NoncurrentVersionExpiration != nil {
			e := make(map[string]interface{})
			if lifecycleRule.NoncurrentVersionExpiration.NoncurrentDays != nil {
				e["days"] = int(aws.Int64Value(lifecycleRule.NoncurrentVersionExpiration.NoncurrentDays))
			}
			rule["noncurrent_version_expiration"] = []interface{}{e}
		}

		// NoncurrentVersionTransition
		if len(lifecycleRule.NoncurrentVersionTransitions) > 0 {
			rule["noncurrent_version_transition"] = flattenBucketLifecycleRuleNoncurrentVersionTransitions(lifecycleRule.NoncurrentVersionTransitions)
		}

		// Transition
		if len(lifecycleRule.Transitions) > 0 {
			rule["transition"] = flattenBucketLifecycleRuleTransitions(lifecycleRule.Transitions)
		}

		results = append(results, rule)
	}

	return results
}

func flattenBucketLifecycleRuleNoncurrentVersionTransitions(transitions []*s3.NoncurrentVersionTransition) []interface{} {
	if len(transitions) == 0 {
		return []interface{}{}
	}

	var results []interface{}

	for _, t := range transitions {
		m := make(map[string]interface{})

		if t.NoncurrentDays != nil {
			m["days"] = int(aws.Int64Value(t.NoncurrentDays))
		}

		if t.StorageClass != nil {
			m["storage_class"] = aws.StringValue(t.StorageClass)
		}

		results = append(results, m)
	}

	return results
}

func flattenBucketLifecycleRuleTransitions(transitions []*s3.Transition) []interface{} {
	if len(transitions) == 0 {
		return []interface{}{}
	}

	var results []interface{}

	for _, t := range transitions {
		m := make(map[string]interface{})

		if t.Date != nil {
			m["date"] = (aws.TimeValue(t.Date)).Format("2006-01-02")
		}
		if t.Days != nil {
			m["days"] = int(aws.Int64Value(t.Days))
		}
		if t.StorageClass != nil {
			m["storage_class"] = aws.StringValue(t.StorageClass)
		}

		results = append(results, m)
	}

	return results
}

// Logging functions

func flattenBucketLoggingEnabled(loggingEnabled *s3.LoggingEnabled) []interface{} {
	if loggingEnabled == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if loggingEnabled.TargetBucket != nil {
		m["target_bucket"] = aws.StringValue(loggingEnabled.TargetBucket)
	}
	if loggingEnabled.TargetPrefix != nil {
		m["target_prefix"] = aws.StringValue(loggingEnabled.TargetPrefix)
	}

	return []interface{}{m}
}

// Object Lock Configuration functions

func expandObjectLockConfiguration(vConf []interface{}) *s3.ObjectLockConfiguration {
	if len(vConf) == 0 || vConf[0] == nil {
		return nil
	}

	mConf := vConf[0].(map[string]interface{})

	conf := &s3.ObjectLockConfiguration{}

	if vObjectLockEnabled, ok := mConf["object_lock_enabled"].(string); ok && vObjectLockEnabled != "" {
		conf.ObjectLockEnabled = aws.String(vObjectLockEnabled)
	}

	if vRule, ok := mConf["rule"].([]interface{}); ok && len(vRule) > 0 {
		mRule := vRule[0].(map[string]interface{})

		if vDefaultRetention, ok := mRule["default_retention"].([]interface{}); ok && len(vDefaultRetention) > 0 && vDefaultRetention[0] != nil {
			mDefaultRetention := vDefaultRetention[0].(map[string]interface{})

			conf.Rule = &s3.ObjectLockRule{
				DefaultRetention: &s3.DefaultRetention{},
			}

			if vMode, ok := mDefaultRetention["mode"].(string); ok && vMode != "" {
				conf.Rule.DefaultRetention.Mode = aws.String(vMode)
			}
			if vDays, ok := mDefaultRetention["days"].(int); ok && vDays > 0 {
				conf.Rule.DefaultRetention.Days = aws.Int64(int64(vDays))
			}
			if vYears, ok := mDefaultRetention["years"].(int); ok && vYears > 0 {
				conf.Rule.DefaultRetention.Years = aws.Int64(int64(vYears))
			}
		}
	}

	return conf
}

func flattenObjectLockConfiguration(conf *s3.ObjectLockConfiguration) []interface{} {
	if conf == nil {
		return []interface{}{}
	}

	mConf := map[string]interface{}{
		"object_lock_enabled": aws.StringValue(conf.ObjectLockEnabled),
	}

	if conf.Rule != nil && conf.Rule.DefaultRetention != nil {
		mRule := map[string]interface{}{
			"default_retention": []interface{}{
				map[string]interface{}{
					"mode":  aws.StringValue(conf.Rule.DefaultRetention.Mode),
					"days":  int(aws.Int64Value(conf.Rule.DefaultRetention.Days)),
					"years": int(aws.Int64Value(conf.Rule.DefaultRetention.Years)),
				},
			},
		}

		mConf["rule"] = []interface{}{mRule}
	}

	return []interface{}{mConf}
}

// Replication Configuration functions

func expandBucketReplicationConfiguration(l []interface{}) *s3.ReplicationConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	rc := &s3.ReplicationConfiguration{}

	if val, ok := tfMap["role"].(string); ok {
		rc.Role = aws.String(val)
	}

	if v, ok := tfMap["rules"].(*schema.Set); ok && v.Len() > 0 {
		rc.Rules = expandBucketReplicationConfigurationRules(v.List())
	}

	return rc
}

func expandBucketReplicationConfigurationRules(l []interface{}) []*s3.ReplicationRule {
	var rules []*s3.ReplicationRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rcRule := &s3.ReplicationRule{}

		if status, ok := tfMap["status"].(string); ok && status != "" {
			rcRule.Status = aws.String(status)
		} else {
			continue
		}

		if v, ok := tfMap["id"].(string); ok && v != "" {
			rcRule.ID = aws.String(v)
		}

		if v, ok := tfMap["destination"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rcRule.Destination = expandBucketReplicationConfigurationRulesDestination(v)
		} else {
			rcRule.Destination = &s3.Destination{}
		}

		if v, ok := tfMap["source_selection_criteria"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rcRule.SourceSelectionCriteria = expandBucketReplicationConfigurationRulesSourceSelectionCriteria(v)
		}

		if v, ok := tfMap["filter"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			// XML schema V2.
			rcRule.Priority = aws.Int64(int64(tfMap["priority"].(int)))

			rcRule.Filter = &s3.ReplicationRuleFilter{}

			filter := v[0].(map[string]interface{})
			tags := Tags(tftags.New(filter["tags"]).IgnoreAWS())

			if len(tags) > 0 {
				rcRule.Filter.And = &s3.ReplicationRuleAndOperator{
					Prefix: aws.String(filter["prefix"].(string)),
					Tags:   tags,
				}
			} else {
				rcRule.Filter.Prefix = aws.String(filter["prefix"].(string))
			}

			if dmr, ok := tfMap["delete_marker_replication_status"].(string); ok && dmr != "" {
				rcRule.DeleteMarkerReplication = &s3.DeleteMarkerReplication{
					Status: aws.String(dmr),
				}
			} else {
				rcRule.DeleteMarkerReplication = &s3.DeleteMarkerReplication{
					Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
				}
			}
		} else {
			// XML schema V1.
			rcRule.Prefix = aws.String(tfMap["prefix"].(string))
		}

		rules = append(rules, rcRule)
	}

	return rules
}

func expandBucketReplicationConfigurationRulesDestination(l []interface{}) *s3.Destination {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	ruleDestination := &s3.Destination{}

	if v, ok := tfMap["bucket"].(string); ok {
		ruleDestination.Bucket = aws.String(v)
	}

	if v, ok := tfMap["storage_class"].(string); ok && v != "" {
		ruleDestination.StorageClass = aws.String(v)
	}

	if v, ok := tfMap["replica_kms_key_id"].(string); ok && v != "" {
		ruleDestination.EncryptionConfiguration = &s3.EncryptionConfiguration{
			ReplicaKmsKeyID: aws.String(v),
		}
	}

	if v, ok := tfMap["account_id"].(string); ok && v != "" {
		ruleDestination.Account = aws.String(v)
	}

	if v, ok := tfMap["access_control_translation"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		aclTranslationValues := v[0].(map[string]interface{})
		ruleAclTranslation := &s3.AccessControlTranslation{}
		ruleAclTranslation.Owner = aws.String(aclTranslationValues["owner"].(string))
		ruleDestination.AccessControlTranslation = ruleAclTranslation
	}

	// replication metrics (required for RTC)
	if v, ok := tfMap["metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		metricsConfig := &s3.Metrics{}
		metricsValues := v[0].(map[string]interface{})
		metricsConfig.EventThreshold = &s3.ReplicationTimeValue{}
		metricsConfig.Status = aws.String(metricsValues["status"].(string))
		metricsConfig.EventThreshold.Minutes = aws.Int64(int64(metricsValues["minutes"].(int)))
		ruleDestination.Metrics = metricsConfig
	}

	// replication time control (RTC)
	if v, ok := tfMap["replication_time"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		rtcValues := v[0].(map[string]interface{})
		rtcConfig := &s3.ReplicationTime{}
		rtcConfig.Status = aws.String(rtcValues["status"].(string))
		rtcConfig.Time = &s3.ReplicationTimeValue{}
		rtcConfig.Time.Minutes = aws.Int64(int64(rtcValues["minutes"].(int)))
		ruleDestination.ReplicationTime = rtcConfig
	}

	return ruleDestination
}

func expandBucketReplicationConfigurationRulesSourceSelectionCriteria(l []interface{}) *s3.SourceSelectionCriteria {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	ruleSsc := &s3.SourceSelectionCriteria{}

	if v, ok := tfMap["sse_kms_encrypted_objects"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		sseKmsValues := v[0].(map[string]interface{})
		sseKmsEncryptedObjects := &s3.SseKmsEncryptedObjects{}

		if sseKmsValues["enabled"].(bool) {
			sseKmsEncryptedObjects.Status = aws.String(s3.SseKmsEncryptedObjectsStatusEnabled)
		} else {
			sseKmsEncryptedObjects.Status = aws.String(s3.SseKmsEncryptedObjectsStatusDisabled)
		}
		ruleSsc.SseKmsEncryptedObjects = sseKmsEncryptedObjects
	}

	return ruleSsc
}

func flattenBucketReplicationConfiguration(r *s3.ReplicationConfiguration) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if r.Role != nil {
		m["role"] = aws.StringValue(r.Role)
	}

	if len(r.Rules) > 0 {
		m["rules"] = flattenBucketReplicationConfigurationReplicationRules(r.Rules)
	}

	return []interface{}{m}
}

func flattenBucketReplicationConfigurationReplicationRuleDestination(d *s3.Destination) []interface{} {
	if d == nil {
		return []interface{}{}

	}

	m := make(map[string]interface{})

	if d.Bucket != nil {
		m["bucket"] = aws.StringValue(d.Bucket)
	}

	if d.StorageClass != nil {
		m["storage_class"] = aws.StringValue(d.StorageClass)
	}

	if d.ReplicationTime != nil {
		rtc := map[string]interface{}{
			"minutes": int(aws.Int64Value(d.ReplicationTime.Time.Minutes)),
			"status":  aws.StringValue(d.ReplicationTime.Status),
		}
		m["replication_time"] = []interface{}{rtc}
	}

	if d.Metrics != nil {
		metrics := map[string]interface{}{
			"status": aws.StringValue(d.Metrics.Status),
		}

		if d.Metrics.EventThreshold != nil {
			metrics["minutes"] = int(aws.Int64Value(d.Metrics.EventThreshold.Minutes))
		}

		m["metrics"] = []interface{}{metrics}
	}
	if d.EncryptionConfiguration != nil {
		if d.EncryptionConfiguration.ReplicaKmsKeyID != nil {
			m["replica_kms_key_id"] = aws.StringValue(d.EncryptionConfiguration.ReplicaKmsKeyID)
		}
	}

	if d.Account != nil {
		m["account_id"] = aws.StringValue(d.Account)
	}

	if d.AccessControlTranslation != nil {
		rdt := map[string]interface{}{
			"owner": aws.StringValue(d.AccessControlTranslation.Owner),
		}
		m["access_control_translation"] = []interface{}{rdt}
	}

	return []interface{}{m}
}

func flattenBucketReplicationConfigurationReplicationRuleFilter(filter *s3.ReplicationRuleFilter) []interface{} {
	if filter == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if filter.Prefix != nil {
		m["prefix"] = aws.StringValue(filter.Prefix)
	}

	if filter.Tag != nil {
		m["tags"] = KeyValueTags([]*s3.Tag{filter.Tag}).IgnoreAWS().Map()
	}

	if filter.And != nil {
		m["prefix"] = aws.StringValue(filter.And.Prefix)
		m["tags"] = KeyValueTags(filter.And.Tags).IgnoreAWS().Map()
	}

	return []interface{}{m}
}

func flattenBucketReplicationConfigurationReplicationRuleSourceSelectionCriteria(ssc *s3.SourceSelectionCriteria) []interface{} {
	if ssc == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if ssc.SseKmsEncryptedObjects != nil {
		m["sse_kms_encrypted_objects"] = flattenBucketReplicationConfigurationReplicationRuleSourceSelectionCriteriaSSEKMSEncryptedObjects(ssc.SseKmsEncryptedObjects)
	}

	return []interface{}{m}
}

func flattenBucketReplicationConfigurationReplicationRuleSourceSelectionCriteriaSSEKMSEncryptedObjects(objs *s3.SseKmsEncryptedObjects) []interface{} {
	if objs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if aws.StringValue(objs.Status) == s3.SseKmsEncryptedObjectsStatusEnabled {
		m["enabled"] = true
	} else if aws.StringValue(objs.Status) == s3.SseKmsEncryptedObjectsStatusDisabled {
		m["enabled"] = false
	}

	return []interface{}{m}
}

func flattenBucketReplicationConfigurationReplicationRules(rules []*s3.ReplicationRule) []interface{} {
	var results []interface{}

	for _, rule := range rules {
		if rule == nil {
			continue
		}

		m := make(map[string]interface{})

		if rule.Destination != nil {
			m["destination"] = flattenBucketReplicationConfigurationReplicationRuleDestination(rule.Destination)
		}

		if rule.ID != nil {
			m["id"] = aws.StringValue(rule.ID)
		}

		if rule.Prefix != nil {
			m["prefix"] = aws.StringValue(rule.Prefix)
		}
		if rule.Status != nil {
			m["status"] = aws.StringValue(rule.Status)
		}
		if rule.SourceSelectionCriteria != nil {
			m["source_selection_criteria"] = flattenBucketReplicationConfigurationReplicationRuleSourceSelectionCriteria(rule.SourceSelectionCriteria)
		}

		if rule.Priority != nil {
			m["priority"] = int(aws.Int64Value(rule.Priority))
		}

		if rule.Filter != nil {
			m["filter"] = flattenBucketReplicationConfigurationReplicationRuleFilter(rule.Filter)
		}

		if rule.DeleteMarkerReplication != nil {
			if rule.DeleteMarkerReplication.Status != nil && aws.StringValue(rule.DeleteMarkerReplication.Status) == s3.DeleteMarkerReplicationStatusEnabled {
				m["delete_marker_replication_status"] = aws.StringValue(rule.DeleteMarkerReplication.Status)
			}
		}

		results = append(results, m)
	}

	return results
}

// Server Side Encryption Configuration functions

func flattenServerSideEncryptionConfiguration(c *s3.ServerSideEncryptionConfiguration) []interface{} {
	if c == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"rule": flattenServerSideEncryptionConfigurationRules(c.Rules),
	}

	return []interface{}{m}
}

func flattenServerSideEncryptionConfigurationRules(rules []*s3.ServerSideEncryptionRule) []interface{} {
	var results []interface{}

	for _, rule := range rules {
		m := make(map[string]interface{})

		if rule.BucketKeyEnabled != nil {
			m["bucket_key_enabled"] = aws.BoolValue(rule.BucketKeyEnabled)
		}

		if rule.ApplyServerSideEncryptionByDefault != nil {
			m["apply_server_side_encryption_by_default"] = []interface{}{
				map[string]interface{}{
					"kms_master_key_id": aws.StringValue(rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID),
					"sse_algorithm":     aws.StringValue(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm),
				},
			}
		}

		results = append(results, m)
	}

	return results
}

// Versioning functions

func expandVersioning(l []interface{}) *s3.VersioningConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	output := &s3.VersioningConfiguration{}

	if v, ok := tfMap["enabled"].(bool); ok {
		if v {
			output.Status = aws.String(s3.BucketVersioningStatusEnabled)
		} else {
			output.Status = aws.String(s3.BucketVersioningStatusSuspended)
		}
	}

	if v, ok := tfMap["mfa_delete"].(bool); ok {
		if v {
			output.MFADelete = aws.String(s3.MFADeleteEnabled)
		} else {
			output.MFADelete = aws.String(s3.MFADeleteDisabled)
		}
	}

	return output
}

func expandVersioningWhenIsNewResource(l []interface{}) *s3.VersioningConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	output := &s3.VersioningConfiguration{}

	// Only set and return a non-nil VersioningConfiguration with at least one of
	// MFADelete or Status enabled as the PutBucketVersioning API request
	// does not need to be made for new buckets that don't require versioning.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/4494

	if v, ok := tfMap["enabled"].(bool); ok && v {
		output.Status = aws.String(s3.BucketVersioningStatusEnabled)
	}

	if v, ok := tfMap["mfa_delete"].(bool); ok && v {
		output.MFADelete = aws.String(s3.MFADeleteEnabled)
	}

	if output.MFADelete == nil && output.Status == nil {
		return nil
	}

	return output
}

func flattenVersioning(versioning *s3.GetBucketVersioningOutput) []interface{} {
	if versioning == nil {
		return []interface{}{}
	}

	vc := make(map[string]interface{})

	if aws.StringValue(versioning.Status) == s3.BucketVersioningStatusEnabled {
		vc["enabled"] = true
	} else {
		vc["enabled"] = false
	}

	if aws.StringValue(versioning.MFADelete) == s3.MFADeleteEnabled {
		vc["mfa_delete"] = true
	} else {
		vc["mfa_delete"] = false
	}

	return []interface{}{vc}
}

// Website functions

func expandWebsiteConfiguration(l []interface{}) (*s3.WebsiteConfiguration, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	website, ok := l[0].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	websiteConfiguration := &s3.WebsiteConfiguration{}

	if v, ok := website["index_document"].(string); ok && v != "" {
		websiteConfiguration.IndexDocument = &s3.IndexDocument{
			Suffix: aws.String(v),
		}
	}

	if v, ok := website["error_document"].(string); ok && v != "" {
		websiteConfiguration.ErrorDocument = &s3.ErrorDocument{
			Key: aws.String(v),
		}
	}

	if v, ok := website["redirect_all_requests_to"].(string); ok && v != "" {
		redirect, err := url.Parse(v)
		if err == nil && redirect.Scheme != "" {
			var redirectHostBuf bytes.Buffer
			redirectHostBuf.WriteString(redirect.Host)
			if redirect.Path != "" {
				redirectHostBuf.WriteString(redirect.Path)
			}
			if redirect.RawQuery != "" {
				redirectHostBuf.WriteString("?")
				redirectHostBuf.WriteString(redirect.RawQuery)
			}
			websiteConfiguration.RedirectAllRequestsTo = &s3.RedirectAllRequestsTo{
				HostName: aws.String(redirectHostBuf.String()),
				Protocol: aws.String(redirect.Scheme),
			}
		} else {
			websiteConfiguration.RedirectAllRequestsTo = &s3.RedirectAllRequestsTo{
				HostName: aws.String(v),
			}
		}
	}

	if v, ok := website["routing_rules"].(string); ok && v != "" {
		var unmarshaledRules []*s3.RoutingRule
		if err := json.Unmarshal([]byte(v), &unmarshaledRules); err != nil {
			return nil, err
		}
		websiteConfiguration.RoutingRules = unmarshaledRules
	}

	return websiteConfiguration, nil
}

func flattenBucketWebsite(ws *s3.GetBucketWebsiteOutput) ([]interface{}, error) {
	if ws == nil {
		return []interface{}{}, nil
	}

	m := make(map[string]interface{})

	if v := ws.IndexDocument; v != nil {
		m["index_document"] = aws.StringValue(v.Suffix)
	}

	if v := ws.ErrorDocument; v != nil {
		m["error_document"] = aws.StringValue(v.Key)
	}

	if v := ws.RedirectAllRequestsTo; v != nil {
		if v.Protocol == nil {
			m["redirect_all_requests_to"] = aws.StringValue(v.HostName)
		} else {
			var host string
			var path string
			var query string
			parsedHostName, err := url.Parse(aws.StringValue(v.HostName))
			if err == nil {
				host = parsedHostName.Host
				path = parsedHostName.Path
				query = parsedHostName.RawQuery
			} else {
				host = aws.StringValue(v.HostName)
				path = ""
			}

			m["redirect_all_requests_to"] = (&url.URL{
				Host:     host,
				Path:     path,
				Scheme:   aws.StringValue(v.Protocol),
				RawQuery: query,
			}).String()
		}
	}

	if v := ws.RoutingRules; v != nil {
		rr, err := normalizeRoutingRules(v)
		if err != nil {
			return nil, fmt.Errorf("error while marshaling routing rules: %w", err)
		}
		m["routing_rules"] = rr
	}

	// We have special handling for the website configuration,
	// so only return the configuration if there is any
	if len(m) == 0 {
		return []interface{}{}, nil
	}

	return []interface{}{m}, nil
}

func normalizeRoutingRules(w []*s3.RoutingRule) (string, error) {
	withNulls, err := json.Marshal(w)
	if err != nil {
		return "", err
	}

	var rules []map[string]interface{}
	if err := json.Unmarshal(withNulls, &rules); err != nil {
		return "", err
	}

	var cleanRules []map[string]interface{}
	for _, rule := range rules {
		cleanRules = append(cleanRules, removeNil(rule))
	}

	withoutNulls, err := json.Marshal(cleanRules)
	if err != nil {
		return "", err
	}

	return string(withoutNulls), nil
}

func removeNil(data map[string]interface{}) map[string]interface{} {
	withoutNil := make(map[string]interface{})

	for k, v := range data {
		if v == nil {
			continue
		}

		switch v := v.(type) {
		case map[string]interface{}:
			withoutNil[k] = removeNil(v)
		default:
			withoutNil[k] = v
		}
	}

	return withoutNil
}
