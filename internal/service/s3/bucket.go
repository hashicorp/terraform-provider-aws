// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"golang.org/x/exp/slices"
)

const (
	// General timeout for S3 bucket changes to propagate.
	// See https://docs.aws.amazon.com/AmazonS3/latest/userguide/Welcome.html#ConsistencyModel.
	s3BucketPropagationTimeout = 2 * time.Minute // nosemgrep:ci.s3-in-const-name, ci.s3-in-var-name
)

// @SDKResource("aws_s3_bucket", name="Bucket")
// @Tags
func ResourceBucket() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketCreate,
		ReadWithoutTimeout:   resourceBucketRead,
		UpdateWithoutTimeout: resourceBucketUpdate,
		DeleteWithoutTimeout: resourceBucketDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Read:   schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"acceleration_status": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Deprecated:       "Use the aws_s3_bucket_accelerate_configuration resource instead",
				ValidateDiagFunc: enum.Validate[types.BucketAccelerateStatus](),
			},
			"acl": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"grant"},
				ValidateFunc:  validation.StringInSlice(bucketCannedACL_Values(), false),
				Deprecated:    "Use the aws_s3_bucket_acl resource instead",
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bucket": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"bucket_prefix"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 63),
					validation.StringDoesNotMatch(directoryBucketNameRegex, `must not be in the format [bucket_name]--[azid]--x-s3. Use the aws_s3_directory_bucket resource to manage S3 Express buckets`),
				),
			},
			"bucket_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bucket_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"bucket"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 63-id.UniqueIDSuffixLength),
				),
			},
			"bucket_regional_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
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
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
						"permissions": {
							Type:     schema.TypeSet,
							Required: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[types.Permission](),
							},
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							// TypeAmazonCustomerByEmail is not currently supported
							ValidateFunc: validation.StringInSlice(enum.Slice(
								types.TypeCanonicalUser,
								types.TypeGroup,
							), false),
						},
						"uri": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lifecycle_rule": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				Deprecated: "Use the aws_s3_bucket_lifecycle_configuration resource instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"abort_incomplete_multipart_upload_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
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
						"id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
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
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.TransitionStorageClass](),
									},
								},
							},
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tags": tftags.TagsSchema(),
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
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.TransitionStorageClass](),
									},
								},
							},
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
			"object_lock_configuration": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				MaxItems:   1,
				Deprecated: "Use the top-level parameter object_lock_enabled and the aws_s3_bucket_object_lock_configuration resource instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object_lock_enabled": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ConflictsWith:    []string{"object_lock_enabled"},
							ValidateDiagFunc: enum.Validate[types.ObjectLockEnabled](),
							Deprecated:       "Use the top-level parameter object_lock_enabled instead",
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
												"days": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"mode": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.ObjectLockRetentionMode](),
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
			"object_lock_enabled": {
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true, // Can be removed when object_lock_configuration.0.object_lock_enabled is removed
				ForceNew:      true,
				ConflictsWith: []string{"object_lock_configuration"},
			},
			"policy": {
				Type:                  schema.TypeString,
				Optional:              true,
				Computed:              true,
				Deprecated:            "Use the aws_s3_bucket_policy resource instead",
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
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
									"delete_marker_replication_status": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(enum.Slice(types.DeleteMarkerReplicationStatusEnabled), false),
									},
									"destination": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_control_translation": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 1,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"owner": {
																Type:             schema.TypeString,
																Required:         true,
																ValidateDiagFunc: enum.Validate[types.OwnerOverride](),
															},
														},
													},
												},
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
																Type:             schema.TypeString,
																Optional:         true,
																Default:          types.MetricsStatusEnabled,
																ValidateDiagFunc: enum.Validate[types.MetricsStatus](),
															},
														},
													},
												},
												"replica_kms_key_id": {
													Type:     schema.TypeString,
													Optional: true,
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
																Type:             schema.TypeString,
																Optional:         true,
																Default:          types.ReplicationTimeStatusEnabled,
																ValidateDiagFunc: enum.Validate[types.ReplicationTimeStatus](),
															},
														},
													},
												},
												"storage_class": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.StorageClass](),
												},
											},
										},
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
									"id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 255),
									},
									"prefix": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
									},
									"priority": {
										Type:     schema.TypeInt,
										Optional: true,
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
									"status": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.ReplicationRuleStatus](),
									},
								},
							},
						},
					},
				},
			},
			"request_payer": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Deprecated:       "Use the aws_s3_bucket_request_payment_configuration resource instead",
				ValidateDiagFunc: enum.Validate[types.Payer](),
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
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.ServerSideEncryption](),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
			"website": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				MaxItems:   1,
				Deprecated: "Use the aws_s3_bucket_website_configuration resource instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"error_document": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"index_document": {
							Type:     schema.TypeString,
							Optional: true,
							ExactlyOneOf: []string{
								"website.0.index_document",
								"website.0.redirect_all_requests_to",
							},
						},
						"redirect_all_requests_to": {
							Type:     schema.TypeString,
							Optional: true,
							ExactlyOneOf: []string{
								"website.0.index_document",
								"website.0.redirect_all_requests_to",
							},
							ConflictsWith: []string{
								"website.0.error_document",
								"website.0.routing_rules",
							},
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
			"website_domain": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "Use the aws_s3_bucket_website_configuration resource",
			},
			"website_endpoint": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "Use the aws_s3_bucket_website_configuration resource",
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := create.Name(d.Get("bucket").(string), d.Get("bucket_prefix").(string))
	region := meta.(*conns.AWSClient).Region

	// Special case: us-east-1 does not return error if the bucket already exists and is owned by
	// current account. It also resets the Bucket ACLs.
	if region == names.USEast1RegionID {
		if err := findBucket(ctx, conn, bucket); err == nil {
			return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s): %s", bucket, errors.New(errCodeBucketAlreadyExists))
		}
	}

	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
		// NOTE: Please, do not add any other fields here unless the field is
		// supported in *all* AWS partitions (including ISO partitions) and by
		// third-party S3 API implementations.
	}

	if v, ok := d.GetOk("acl"); ok {
		input.ACL = types.BucketCannedACL(v.(string))
	} else {
		// Use default value previously available in v3.x of the provider.
		input.ACL = types.BucketCannedACLPrivate
	}

	// Special case us-east-1 region and do not set the LocationConstraint.
	// See "Request Elements: http://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketPUT.html
	if region != names.USEast1RegionID {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}

	if err := validBucketName(bucket, region); err != nil {
		return sdkdiag.AppendErrorf(diags, "validating S3 Bucket (%s) name: %s", bucket, err)
	}

	// S3 Object Lock is not supported on all partitions.
	if v, ok := d.GetOk("object_lock_enabled"); ok {
		input.ObjectLockEnabledForBucket = aws.Bool(v.(bool))
	}

	// S3 Object Lock can only be enabled on bucket creation.
	if v := expandBucketObjectLockConfiguration(d.Get("object_lock_configuration").([]interface{})); v != nil && v.ObjectLockEnabled == types.ObjectLockEnabledEnabled {
		input.ObjectLockEnabledForBucket = aws.Bool(true)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.CreateBucket(ctx, input)
	}, errCodeOperationAborted)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s): %s", bucket, err)
	}

	d.SetId(bucket)

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return nil, findBucket(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBucketUpdate(ctx, d, meta)...)
}

func resourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	err := findBucket(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "s3",
		Resource:  d.Id(),
	}.String()
	d.Set("arn", arn)
	d.Set("bucket", d.Id())
	d.Set("bucket_domain_name", meta.(*conns.AWSClient).PartitionHostname(d.Id()+".s3"))
	d.Set("bucket_prefix", create.NamePrefixFromName(d.Id()))

	//
	// Bucket Policy.
	//
	// Read the policy if configured outside this resource e.g. with aws_s3_bucket_policy resource.
	policy, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() (string, error) {
		return findBucketPolicy(ctx, conn, d.Id())
	})

	// The call to HeadBucket above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as GetBucketPolicy, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	switch {
	case err == nil:
		policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), policy)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set("policy", policyToSet)
	case tfawserr.ErrCodeEquals(err, errCodeNoSuchBucketPolicy, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("policy", nil)
	default:
		return diag.Errorf("reading S3 Bucket (%s) policy: %s", d.Id(), err)
	}

	//
	// Bucket ACL.
	//
	bucketACL, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() (*s3.GetBucketAclOutput, error) {
		return findBucketACL(ctx, conn, d.Id(), "")
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	switch {
	case err == nil:
		if err := d.Set("grant", flattenGrants(bucketACL)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting grant: %s", err)
		}
	case tfawserr.ErrCodeEquals(err, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("grant", nil)
	default:
		return diag.Errorf("reading S3 Bucket (%s) ACL: %s", d.Id(), err)
	}

	//
	// Bucket CORS Configuration.
	//
	corsRules, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() ([]types.CORSRule, error) {
		return findCORSRules(ctx, conn, d.Id(), "")
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	switch {
	case err == nil:
		if err := d.Set("cors_rule", flattenBucketCORSRules(corsRules)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting cors_rule: %s", err)
		}
	case tfawserr.ErrCodeEquals(err, errCodeNoSuchCORSConfiguration, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("cors_rule", nil)
	default:
		return diag.Errorf("reading S3 Bucket (%s) CORS configuration: %s", d.Id(), err)
	}

	//
	// Bucket Website Configuration.
	//
	bucketWebsite, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() (*s3.GetBucketWebsiteOutput, error) {
		return findBucketWebsite(ctx, conn, d.Id(), "")
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	switch {
	case err == nil:
		website, err := flattenBucketWebsite(bucketWebsite)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		if err := d.Set("website", website); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting website: %s", err)
		}
	case tfawserr.ErrCodeEquals(err, errCodeNoSuchWebsiteConfiguration, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("website", nil)
	default:
		return diag.Errorf("reading S3 Bucket (%s) website configuration: %s", d.Id(), err)
	}

	//
	// Bucket Versioning.
	//
	bucketVersioning, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() (*s3.GetBucketVersioningOutput, error) {
		return findBucketVersioning(ctx, conn, d.Id(), "")
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	switch {
	case err == nil:
		if err := d.Set("versioning", flattenVersioning(bucketVersioning)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting versioning: %s", err)
		}
	case tfawserr.ErrCodeEquals(err, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("versioning", nil)
	default:
		return diag.Errorf("reading S3 Bucket (%s) versioning: %s", d.Id(), err)
	}

	//
	// Bucket Accelerate Configuration.
	//
	bucketAccelerate, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() (*s3.GetBucketAccelerateConfigurationOutput, error) {
		return findBucketAccelerateConfiguration(ctx, conn, d.Id(), "")
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	switch {
	case err == nil:
		d.Set("acceleration_status", bucketAccelerate.Status)
	case tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed, errCodeUnsupportedArgument, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("acceleration_status", nil)
	default:
		return diag.Errorf("reading S3 Bucket (%s) accelerate configuration: %s", d.Id(), err)
	}

	//
	// Bucket Request Payment Configuration.
	//
	bucketRequestPayment, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() (*s3.GetBucketRequestPaymentOutput, error) {
		return findBucketRequestPayment(ctx, conn, d.Id(), "")
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	switch {
	case err == nil:
		d.Set("request_payer", bucketRequestPayment.Payer)
	case tfawserr.ErrCodeEquals(err, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("request_payer", nil)
	default:
		return diag.Errorf("reading S3 Bucket (%s) request payment configuration: %s", d.Id(), err)
	}

	//
	// Bucket Logging.
	//
	loggingEnabled, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() (*types.LoggingEnabled, error) {
		return findLoggingEnabled(ctx, conn, d.Id(), "")
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	switch {
	case err == nil:
		if err := d.Set("logging", flattenBucketLoggingEnabled(loggingEnabled)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting logging: %s", err)
		}
	case tfawserr.ErrCodeEquals(err, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("logging", nil)
	default:
		return diag.Errorf("reading S3 Bucket (%s) logging: %s", d.Id(), err)
	}

	//
	// Bucket Lifecycle Configuration.
	//
	lifecycleRules, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() ([]types.LifecycleRule, error) {
		return findLifecycleRules(ctx, conn, d.Id(), "")
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	switch {
	case err == nil:
		if err := d.Set("lifecycle_rule", flattenBucketLifecycleRules(ctx, lifecycleRules)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting lifecycle_rule: %s", err)
		}
	case tfawserr.ErrCodeEquals(err, errCodeNoSuchLifecycleConfiguration, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("lifecycle_rule", nil)
	default:
		return diag.Errorf("reading S3 Bucket (%s) lifecycle configuration: %s", d.Id(), err)
	}

	//
	// Bucket Replication Configuration.
	//
	replicationConfiguration, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() (*types.ReplicationConfiguration, error) {
		return findReplicationConfiguration(ctx, conn, d.Id())
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	switch {
	case err == nil:
		if err := d.Set("replication_configuration", flattenBucketReplicationConfiguration(ctx, replicationConfiguration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting replication_configuration: %s", err)
		}
	case tfawserr.ErrCodeEquals(err, errCodeReplicationConfigurationNotFound, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("replication_configuration", nil)
	default:
		return diag.Errorf("reading S3 Bucket (%s) replication configuration: %s", d.Id(), err)
	}

	//
	// Bucket Server-side Encryption Configuration.
	//
	encryptionConfiguration, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() (*types.ServerSideEncryptionConfiguration, error) {
		return findServerSideEncryptionConfiguration(ctx, conn, d.Id(), "")
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	switch {
	case err == nil:
		if err := d.Set("server_side_encryption_configuration", flattenServerSideEncryptionConfiguration(encryptionConfiguration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting server_side_encryption_configuration: %s", err)
		}
	case tfawserr.ErrCodeEquals(err, errCodeReplicationConfigurationNotFound, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("server_side_encryption_configuration", nil)
	default:
		return diag.Errorf("reading S3 Bucket (%s) server-side encryption configuration: %s", d.Id(), err)
	}

	//
	// Bucket Object Lock Configuration.
	//
	objLockConfig, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() (*types.ObjectLockConfiguration, error) {
		return findObjectLockConfiguration(ctx, conn, d.Id(), "")
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	switch {
	case err == nil:
		if err := d.Set("object_lock_configuration", flattenObjectLockConfiguration(objLockConfig)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting object_lock_configuration: %s", err)
		}
		d.Set("object_lock_enabled", objLockConfig.ObjectLockEnabled == types.ObjectLockEnabledEnabled)
	case tfawserr.ErrCodeEquals(err, errCodeObjectLockConfigurationNotFoundError, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("object_lock_configuration", nil)
		d.Set("object_lock_enabled", nil)
	default:
		if partition := meta.(*conns.AWSClient).Partition; partition == names.StandardPartitionID || partition == names.USGovCloudPartitionID {
			return diag.Errorf("reading S3 Bucket (%s) object lock configuration: %s", d.Id(), err)
		}
		log.Printf("[WARN] Unable to read S3 Bucket (%s) Object Lock Configuration: %s", d.Id(), err)
		d.Set("object_lock_configuration", nil)
		d.Set("object_lock_enabled", nil)
	}

	//
	// Bucket Region etc.
	//
	region, err := manager.GetBucketRegion(ctx, conn, d.Id(), func(o *s3.Options) {
		o.UsePathStyle = meta.(*conns.AWSClient).S3UsePathStyle()
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) location: %s", d.Id(), err)
	}

	d.Set("region", region)
	d.Set("bucket_regional_domain_name", bucketRegionalDomainName(d.Id(), region))

	hostedZoneID, err := hostedZoneIDForRegion(region)
	if err != nil {
		log.Printf("[WARN] %s", err)
	} else {
		d.Set("hosted_zone_id", hostedZoneID)
	}

	if _, ok := d.GetOk("website"); ok {
		endpoint, domain := bucketWebsiteEndpointAndDomain(d.Id(), region)
		d.Set("website_domain", domain)
		d.Set("website_endpoint", endpoint)
	} else {
		d.Set("website_domain", nil)
		d.Set("website_endpoint", nil)
	}

	//
	// Bucket Tags.
	//
	tags, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() (tftags.KeyValueTags, error) {
		return BucketListTags(ctx, conn, d.Id())
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	switch {
	case err == nil:
		setTagsOut(ctx, Tags(tags))
	case tfawserr.ErrCodeEquals(err, errCodeNotImplemented, errCodeXNotImplemented):
	default:
		return sdkdiag.AppendErrorf(diags, "listing tags for S3 Bucket (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceBucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	// Note: Order of argument updates below is important.

	//
	// Bucket Policy.
	//
	if d.HasChange("policy") {
		policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if policy == "" {
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
				return conn.DeleteBucketPolicy(ctx, &s3.DeleteBucketPolicyInput{
					Bucket: aws.String(d.Id()),
				})
			}, errCodeNoSuchBucket)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket (%s) policy: %s", d.Id(), err)
			}
		} else {
			input := &s3.PutBucketPolicyInput{
				Bucket: aws.String(d.Id()),
				Policy: aws.String(policy),
			}

			_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
				return conn.PutBucketPolicy(ctx, input)
			}, errCodeMalformedPolicy, errCodeNoSuchBucket)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) policy: %s", d.Id(), err)
			}
		}
	}

	//
	// Bucket CORS Configuration.
	//
	if d.HasChange("cors_rule") {
		if v, ok := d.GetOk("cors_rule"); !ok || len(v.([]interface{})) == 0 {
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
				return conn.DeleteBucketCors(ctx, &s3.DeleteBucketCorsInput{
					Bucket: aws.String(d.Id()),
				})
			}, errCodeNoSuchBucket)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket (%s) CORS configuration: %s", d.Id(), err)
			}
		} else {
			input := &s3.PutBucketCorsInput{
				Bucket: aws.String(d.Id()),
				CORSConfiguration: &types.CORSConfiguration{
					CORSRules: expandCORSRules(d.Get("cors_rule").(*schema.Set).List()),
				},
			}

			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
				return conn.PutBucketCors(ctx, input)
			}, errCodeNoSuchBucket)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) CORS configuration: %s", d.Id(), err)
			}
		}
	}

	//
	// Bucket Website Configuration.
	//
	if d.HasChange("website") {
		if v, ok := d.GetOk("website"); !ok || len(v.([]interface{})) == 0 || v.([]interface{})[0] == nil {
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
				return conn.DeleteBucketWebsite(ctx, &s3.DeleteBucketWebsiteInput{
					Bucket: aws.String(d.Id()),
				})
			}, errCodeNoSuchBucket)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket (%s) website configuration: %s", d.Id(), err)
			}
		} else {
			websiteConfig, err := expandBucketWebsiteConfiguration(v.([]interface{}))
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			input := &s3.PutBucketWebsiteInput{
				Bucket:               aws.String(d.Id()),
				WebsiteConfiguration: websiteConfig,
			}

			_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
				return conn.PutBucketWebsite(ctx, input)
			}, errCodeNoSuchBucket)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) website configuration: %s", d.Id(), err)
			}
		}
	}

	//
	// Bucket Versioning.
	//
	if d.HasChange("versioning") {
		v := d.Get("versioning").([]interface{})
		var versioningConfig *types.VersioningConfiguration

		if d.IsNewResource() {
			versioningConfig = expandBucketVersioningConfigurationCreate(v)
		} else {
			versioningConfig = expandBucketVersioningConfigurationUpdate(v)
		}

		input := &s3.PutBucketVersioningInput{
			Bucket:                  aws.String(d.Id()),
			VersioningConfiguration: versioningConfig,
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
			return conn.PutBucketVersioning(ctx, input)
		}, errCodeNoSuchBucket)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) versioning: %s", d.Id(), err)
		}
	}

	//
	// Bucket ACL.
	//
	if (d.HasChange("acl") && !d.IsNewResource()) || (d.HasChange("grant") && d.Get("grant").(*schema.Set).Len() == 0) {
		acl := types.BucketCannedACL(d.Get("acl").(string))
		if acl == "" {
			// Use default value previously available in v3.x of the provider.
			acl = types.BucketCannedACLPrivate
		}
		input := &s3.PutBucketAclInput{
			ACL:    acl,
			Bucket: aws.String(d.Id()),
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
			return conn.PutBucketAcl(ctx, input)
		}, errCodeNoSuchBucket)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) ACL: %s", d.Id(), err)
		}
	}

	if d.HasChange("grant") && d.Get("grant").(*schema.Set).Len() > 0 {
		bucketACL, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutUpdate), func() (*s3.GetBucketAclOutput, error) {
			return findBucketACL(ctx, conn, d.Id(), "")
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) ACL: %s", d.Id(), err)
		}

		input := &s3.PutBucketAclInput{
			AccessControlPolicy: &types.AccessControlPolicy{
				Grants: expandGrants(d.Get("grant").(*schema.Set).List()),
				Owner:  bucketACL.Owner,
			},
			Bucket: aws.String(d.Id()),
		}

		_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
			return conn.PutBucketAcl(ctx, input)
		}, errCodeNoSuchBucket)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) ACL: %s", d.Id(), err)
		}
	}

	//
	// Bucket Logging.
	//
	if d.HasChange("logging") {
		input := &s3.PutBucketLoggingInput{
			Bucket:              aws.String(d.Id()),
			BucketLoggingStatus: &types.BucketLoggingStatus{},
		}

		if v, ok := d.GetOk("logging"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			input.BucketLoggingStatus.LoggingEnabled = &types.LoggingEnabled{}

			if v, ok := tfMap["target_bucket"].(string); ok && v != "" {
				input.BucketLoggingStatus.LoggingEnabled.TargetBucket = aws.String(v)
			}

			if v, ok := tfMap["target_prefix"].(string); ok && v != "" {
				input.BucketLoggingStatus.LoggingEnabled.TargetPrefix = aws.String(v)
			}
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
			return conn.PutBucketLogging(ctx, input)
		}, errCodeNoSuchBucket)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) logging: %s", d.Id(), err)
		}
	}

	//
	// Bucket Lifecycle Configuration.
	//
	if d.HasChange("lifecycle_rule") {
		if v, ok := d.GetOk("lifecycle_rule"); !ok || len(v.([]interface{})) == 0 {
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
				return conn.DeleteBucketLifecycle(ctx, &s3.DeleteBucketLifecycleInput{
					Bucket: aws.String(d.Id()),
				})
			}, errCodeNoSuchBucket)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket (%s) lifecycle configuration: %s", d.Id(), err)
			}
		} else {
			input := &s3.PutBucketLifecycleConfigurationInput{
				Bucket: aws.String(d.Id()),
				LifecycleConfiguration: &types.BucketLifecycleConfiguration{
					Rules: expandBucketLifecycleRules(ctx, v.([]interface{})),
				},
			}

			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
				return conn.PutBucketLifecycleConfiguration(ctx, input)
			}, errCodeNoSuchBucket)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) lifecycle configuration: %s", d.Id(), err)
			}
		}
	}

	//
	// Bucket Accelerate Configuration.
	//
	if d.HasChange("acceleration_status") {
		input := &s3.PutBucketAccelerateConfigurationInput{
			AccelerateConfiguration: &types.AccelerateConfiguration{
				Status: types.BucketAccelerateStatus(d.Get("acceleration_status").(string)),
			},
			Bucket: aws.String(d.Id()),
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
			return conn.PutBucketAccelerateConfiguration(ctx, input)
		}, errCodeNoSuchBucket)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) accelerate configuration: %s", d.Id(), err)
		}
	}

	//
	// Bucket Request Payment Configuration.
	//
	if d.HasChange("request_payer") {
		input := &s3.PutBucketRequestPaymentInput{
			Bucket: aws.String(d.Id()),
			RequestPaymentConfiguration: &types.RequestPaymentConfiguration{
				Payer: types.Payer(d.Get("request_payer").(string)),
			},
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
			return conn.PutBucketRequestPayment(ctx, input)
		}, errCodeNoSuchBucket)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) request payment configuration: %s", d.Id(), err)
		}
	}

	//
	// Bucket Replication Configuration.
	//
	if d.HasChange("replication_configuration") {
		if v, ok := d.GetOk("replication_configuration"); !ok || len(v.([]interface{})) == 0 || v.([]interface{})[0] == nil {
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
				return conn.DeleteBucketReplication(ctx, &s3.DeleteBucketReplicationInput{
					Bucket: aws.String(d.Id()),
				})
			}, errCodeNoSuchBucket)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket (%s) replication configuration: %s", d.Id(), err)
			}
		} else {
			hasVersioning := false

			// Validate that bucket versioning is enabled.
			if v, ok := d.GetOk("versioning"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				tfMap := v.([]interface{})[0].(map[string]interface{})

				if tfMap["enabled"].(bool) {
					hasVersioning = true
				}
			}

			if !hasVersioning {
				return sdkdiag.AppendErrorf(diags, "versioning must be enabled on S3 Bucket (%s) to allow replication", d.Id())
			}

			input := &s3.PutBucketReplicationInput{
				Bucket:                   aws.String(d.Id()),
				ReplicationConfiguration: expandBucketReplicationConfiguration(ctx, v.([]interface{})),
			}

			_, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutUpdate),
				func() (interface{}, error) {
					return conn.PutBucketReplication(ctx, input)
				},
				func(err error) (bool, error) {
					if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) || tfawserr.ErrMessageContains(err, errCodeInvalidRequest, "Versioning must be 'Enabled' on the bucket") {
						return true, err
					}

					return false, err
				},
			)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) replication configuration: %s", d.Id(), err)
			}
		}
	}

	//
	// Bucket Server-side Encryption Configuration.
	//
	if d.HasChange("server_side_encryption_configuration") {
		if v, ok := d.GetOk("replication_configuration"); !ok || len(v.([]interface{})) == 0 {
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
				return conn.DeleteBucketEncryption(ctx, &s3.DeleteBucketEncryptionInput{
					Bucket: aws.String(d.Id()),
				})
			}, errCodeNoSuchBucket)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket (%s) server-side encryption configuration: %s", d.Id(), err)
			}
		} else {
			input := &s3.PutBucketEncryptionInput{
				Bucket: aws.String(d.Id()),
				ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
					Rules: expandBucketServerSideEncryptionRules(v.([]interface{})),
				},
			}

			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
				return conn.PutBucketEncryption(ctx, input)
			}, errCodeNoSuchBucket, errCodeOperationAborted)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) server-side encryption configuration: %s", d.Id(), err)
			}
		}
	}

	//
	// Bucket Object Lock Configuration.
	//
	if d.HasChange("object_lock_configuration") {
		// S3 Object Lock configuration cannot be deleted, only updated.
		input := &s3.PutObjectLockConfigurationInput{
			Bucket:                  aws.String(d.Id()),
			ObjectLockConfiguration: expandBucketObjectLockConfiguration(d.Get("object_lock_configuration").([]interface{})),
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
			return conn.PutObjectLockConfiguration(ctx, input)
		}, errCodeNoSuchBucket)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) object lock configuration: %s", d.Id(), err)
		}
	}

	//
	// Bucket Tags.
	//
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		// Retry due to S3 eventual consistency.
		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
			terr := BucketUpdateTags(ctx, conn, d.Id(), o, n)
			return nil, terr
		}, errCodeNoSuchBucket)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating S3 Bucket (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceBucketRead(ctx, d, meta)...)
}

func resourceBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	log.Printf("[INFO] Deleting S3 Bucket: %s", d.Id())
	_, err := conn.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return diags
	}

	if tfawserr.ErrCodeEquals(err, errCodeBucketNotEmpty) {
		if d.Get("force_destroy").(bool) {
			// Delete everything including locked objects.
			// Don't ignore any object errors or we could recurse infinitely.
			var objectLockEnabled bool
			if v := expandBucketObjectLockConfiguration(d.Get("object_lock_configuration").([]interface{})); v != nil {
				objectLockEnabled = v.ObjectLockEnabled == types.ObjectLockEnabledEnabled
			}

			if n, err := emptyBucket(ctx, conn, d.Id(), objectLockEnabled); err != nil {
				return diag.Errorf("emptying S3 Bucket (%s): %s", d.Id(), err)
			} else {
				log.Printf("[DEBUG] Deleted %d S3 objects", n)
			}

			// Recurse until all objects are deleted or an error is returned.
			return resourceBucketDelete(ctx, d, meta)
		}
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return nil, findBucket(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findBucket(ctx context.Context, conn *s3.Client, bucket string, optFns ...func(*s3.Options)) error {
	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	_, err := conn.HeadBucket(ctx, input, optFns...)

	// For directory buckets that no longer exist it's the CreateSession call invoked by HeadBucket that returns "NoSuchBucket",
	// and that error code is flattend into HeadBucket's error message -- hence the 'errs.Contains' call.
	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotFound) || tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) || errs.Contains(err, errCodeNoSuchBucket) {
		return &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	return err
}

func retryWhenNoSuchBucketError[T any](ctx context.Context, timeout time.Duration, f func() (T, error)) (T, error) {
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (interface{}, error) {
		return f()
	}, errCodeNoSuchBucket)

	if err != nil {
		var zero T
		return zero, err
	}

	return outputRaw.(T), nil
}

// https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region
func bucketRegionalDomainName(bucket, region string) string {
	// Return a default AWS Commercial domain name if no Region is provided.
	if region == "" {
		return fmt.Sprintf("%s.s3.amazonaws.com", bucket) //lintignore:AWSR001
	}
	return fmt.Sprintf("%s.s3.%s.%s", bucket, region, names.DNSSuffixForPartition(names.PartitionForRegion(region)))
}

func bucketWebsiteEndpointAndDomain(bucket, region string) (string, string) {
	var domain string

	// Default to us-east-1 if the bucket doesn't have a region:
	// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketGETlocation.html
	if region == "" {
		region = names.USEast1RegionID
	}

	// Different regions have different syntax for website endpoints:
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteEndpoints.html
	// https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_website_region_endpoints
	oldRegions := []string{
		names.APNortheast1RegionID,
		names.APSoutheast1RegionID,
		names.APSoutheast2RegionID,
		names.EUWest1RegionID,
		names.SAEast1RegionID,
		names.USEast1RegionID,
		names.USGovWest1RegionID,
		names.USWest1RegionID,
		names.USWest2RegionID,
	}
	if slices.Contains(oldRegions, region) {
		domain = fmt.Sprintf("s3-website-%s.amazonaws.com", region) //lintignore:AWSR001
	} else {
		dnsSuffix := names.DNSSuffixForPartition(names.PartitionForRegion(region))
		domain = fmt.Sprintf("s3-website.%s.%s", region, dnsSuffix)
	}

	return fmt.Sprintf("%s.%s", bucket, domain), domain
}

func expandBucketCORSRules(l []interface{}) []types.CORSRule {
	if len(l) == 0 {
		return nil
	}

	var rules []types.CORSRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rule := types.CORSRule{}

		if v, ok := tfMap["allowed_headers"].([]interface{}); ok && len(v) > 0 {
			rule.AllowedHeaders = flex.ExpandStringValueList(v)
		}

		if v, ok := tfMap["allowed_methods"].([]interface{}); ok && len(v) > 0 {
			rule.AllowedMethods = flex.ExpandStringValueList(v)
		}

		if v, ok := tfMap["allowed_origins"].([]interface{}); ok && len(v) > 0 {
			rule.AllowedOrigins = flex.ExpandStringValueList(v)
		}

		if v, ok := tfMap["expose_headers"].([]interface{}); ok && len(v) > 0 {
			rule.ExposeHeaders = flex.ExpandStringValueList(v)
		}

		if v, ok := tfMap["max_age_seconds"].(int); ok {
			rule.MaxAgeSeconds = aws.Int32(int32(v))
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenBucketCORSRules(rules []types.CORSRule) []interface{} {
	var results []interface{}

	for _, rule := range rules {
		m := map[string]interface{}{
			"max_age_seconds": rule.MaxAgeSeconds,
		}

		if len(rule.AllowedHeaders) > 0 {
			m["allowed_headers"] = rule.AllowedHeaders
		}

		if len(rule.AllowedMethods) > 0 {
			m["allowed_methods"] = rule.AllowedMethods
		}

		if len(rule.AllowedOrigins) > 0 {
			m["allowed_origins"] = rule.AllowedOrigins
		}

		if len(rule.ExposeHeaders) > 0 {
			m["expose_headers"] = rule.ExposeHeaders
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

func expandBucketLifecycleRules(ctx context.Context, l []interface{}) []types.LifecycleRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var results []types.LifecycleRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		result := types.LifecycleRule{}

		if v, ok := tfMap["abort_incomplete_multipart_upload"].(int); ok && v > 0 {
			result.AbortIncompleteMultipartUpload = &types.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: aws.Int32(int32(v)),
			}
		}

		if v, ok := tfMap["expiration"].([]interface{}); ok && len(v) > 0 {
			result.Expiration = expandBucketLifecycleRuleExpiration(v)
		}

		var filter types.LifecycleRuleFilter
		prefix := tfMap["prefix"].(string)
		if tags := Tags(tftags.New(ctx, tfMap["tags"]).IgnoreAWS()); len(tags) > 0 {
			filter = &types.LifecycleRuleFilterMemberAnd{
				Value: types.LifecycleRuleAndOperator{
					Prefix: aws.String(prefix),
					Tags:   tags,
				},
			}
		} else {
			filter = &types.LifecycleRuleFilterMemberPrefix{
				Value: prefix,
			}
		}
		result.Filter = filter

		if v, ok := tfMap["id"].(string); ok {
			result.ID = aws.String(v)
		} else {
			result.ID = aws.String(id.PrefixedUniqueId("tf-s3-lifecycle-"))
		}

		if v, ok := tfMap["noncurrent_version_expiration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			result.NoncurrentVersionExpiration = expandBucketLifecycleRuleNoncurrentVersionExpiration(v[0].(map[string]interface{}))
		}

		if v, ok := tfMap["noncurrent_version_transition"].(*schema.Set); ok && v.Len() > 0 {
			result.NoncurrentVersionTransitions = expandBucketLifecycleRuleNoncurrentVersionTransitions(v.List())
		}

		if v, ok := tfMap["enabled"].(bool); ok && v {
			result.Status = types.ExpirationStatusEnabled
		} else {
			result.Status = types.ExpirationStatusDisabled
		}

		if v, ok := tfMap["transition"].(*schema.Set); ok && v.Len() > 0 {
			result.Transitions = expandBucketLifecycleRuleTransitions(v.List())
		}

		// As a lifecycle rule requires 1 or more transition/expiration actions,
		// we explicitly pass a default ExpiredObjectDeleteMarker value to be able to create
		// the rule while keeping the policy unaffected if the conditions are not met.
		if result.AbortIncompleteMultipartUpload == nil && result.Expiration == nil && result.NoncurrentVersionExpiration == nil && result.NoncurrentVersionTransitions == nil && result.Transitions == nil {
			result.Expiration = &types.LifecycleExpiration{ExpiredObjectDeleteMarker: aws.Bool(false)}
		}

		results = append(results, result)
	}

	return results
}

func expandBucketLifecycleRuleExpiration(l []interface{}) *types.LifecycleExpiration {
	if len(l) == 0 {
		return nil
	}

	result := &types.LifecycleExpiration{}

	if l[0] == nil {
		return result
	}

	m := l[0].(map[string]interface{})

	if v, ok := m["date"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v+"T00:00:00Z")
		result.Date = aws.Time(t)
	} else if v, ok := m["days"].(int); ok && v > 0 {
		result.Days = aws.Int32(int32(v))
	} else if v, ok := m["expired_object_delete_marker"].(bool); ok {
		result.ExpiredObjectDeleteMarker = aws.Bool(v)
	}

	return result
}

func expandBucketLifecycleRuleNoncurrentVersionExpiration(m map[string]interface{}) *types.NoncurrentVersionExpiration {
	if len(m) == 0 {
		return nil
	}

	var result *types.NoncurrentVersionExpiration

	if v, ok := m["days"].(int); ok {
		result = &types.NoncurrentVersionExpiration{
			NoncurrentDays: aws.Int32(int32(v)),
		}
	}

	return result
}

func expandBucketLifecycleRuleNoncurrentVersionTransitions(l []interface{}) []types.NoncurrentVersionTransition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var results []types.NoncurrentVersionTransition

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		transition := types.NoncurrentVersionTransition{}

		if v, ok := tfMap["days"].(int); ok {
			transition.NoncurrentDays = aws.Int32(int32(v))
		}

		if v, ok := tfMap["storage_class"].(string); ok && v != "" {
			transition.StorageClass = types.TransitionStorageClass(v)
		}

		results = append(results, transition)
	}

	return results
}

func expandBucketLifecycleRuleTransitions(l []interface{}) []types.Transition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var results []types.Transition

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		transition := types.Transition{}

		if v, ok := tfMap["date"].(string); ok && v != "" {
			t, _ := time.Parse(time.RFC3339, v+"T00:00:00Z")
			transition.Date = aws.Time(t)
		} else if v, ok := tfMap["days"].(int); ok && v >= 0 {
			transition.Days = aws.Int32(int32(v))
		}

		if v, ok := tfMap["storage_class"].(string); ok && v != "" {
			transition.StorageClass = types.TransitionStorageClass(v)
		}

		results = append(results, transition)
	}

	return results
}

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

func flattenBucketLifecycleRules(ctx context.Context, lifecycleRules []*s3.LifecycleRule) []interface{} {
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
					rule["tags"] = KeyValueTags(ctx, filter.And.Tags).IgnoreAWS().Map()
				}
			} else {
				// Prefix
				if filter.Prefix != nil {
					rule["prefix"] = aws.StringValue(filter.Prefix)
				}
				// Tag
				if filter.Tag != nil {
					rule["tags"] = KeyValueTags(ctx, []*s3.Tag{filter.Tag}).IgnoreAWS().Map()
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

func expandBucketServerSideEncryptionRules(l []interface{}) []types.ServerSideEncryptionRule {
	var rules []types.ServerSideEncryptionRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rule := types.ServerSideEncryptionRule{}

		if v, ok := tfMap["apply_server_side_encryption_by_default"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.ApplyServerSideEncryptionByDefault = expandBucketServerSideEncryptionByDefault(v)
		}

		if v, ok := tfMap["bucket_key_enabled"].(bool); ok {
			rule.BucketKeyEnabled = aws.Bool(v)
		}

		rules = append(rules, rule)
	}

	return rules
}

func expandBucketServerSideEncryptionByDefault(l []interface{}) *types.ServerSideEncryptionByDefault {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	sse := &types.ServerSideEncryptionByDefault{}

	if v, ok := tfMap["kms_master_key_id"].(string); ok && v != "" {
		sse.KMSMasterKeyID = aws.String(v)
	}

	if v, ok := tfMap["sse_algorithm"].(string); ok && v != "" {
		sse.SSEAlgorithm = types.ServerSideEncryption(v)
	}

	return sse
}

// Object Lock Configuration functions

func expandBucketObjectLockConfiguration(l []interface{}) *types.ObjectLockConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.ObjectLockConfiguration{}

	if v, ok := tfMap["object_lock_enabled"].(string); ok && v != "" {
		apiObject.ObjectLockEnabled = types.ObjectLockEnabled(v)
	}

	if v, ok := tfMap["rule"].([]interface{}); ok && len(v) > 0 {
		tfMap := v[0].(map[string]interface{})

		if v, ok := tfMap["default_retention"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			tfMap := v[0].(map[string]interface{})

			apiObject.Rule = &types.ObjectLockRule{
				DefaultRetention: &types.DefaultRetention{},
			}

			if v, ok := tfMap["days"].(int); ok && v > 0 {
				apiObject.Rule.DefaultRetention.Days = aws.Int32(int32(v))
			}
			if v, ok := tfMap["mode"].(string); ok && v != "" {
				apiObject.Rule.DefaultRetention.Mode = types.ObjectLockRetentionMode(v)
			}
			if v, ok := tfMap["years"].(int); ok && v > 0 {
				apiObject.Rule.DefaultRetention.Years = aws.Int32(int32(v))
			}
		}
	}

	return apiObject
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

func expandBucketReplicationConfiguration(ctx context.Context, l []interface{}) *types.ReplicationConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.ReplicationConfiguration{}

	if v, ok := tfMap["role"].(string); ok {
		apiObject.Role = aws.String(v)
	}

	if v, ok := tfMap["rules"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Rules = expandBucketReplicationConfigurationRules(ctx, v.List())
	}

	return apiObject
}

func expandBucketReplicationConfigurationRules(ctx context.Context, l []interface{}) []types.ReplicationRule {
	var rules []types.ReplicationRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rule := types.ReplicationRule{}

		if v, ok := tfMap["status"].(string); ok && v != "" {
			rule.Status = types.ReplicationRuleStatus(v)
		} else {
			continue
		}

		if v, ok := tfMap["destination"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.Destination = expandBucketReplicationConfigurationRulesDestination(v)
		} else {
			rule.Destination = &types.Destination{}
		}

		if v, ok := tfMap["id"].(string); ok && v != "" {
			rule.ID = aws.String(v)
		}

		if v, ok := tfMap["source_selection_criteria"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.SourceSelectionCriteria = expandBucketReplicationConfigurationRulesSourceSelectionCriteria(v)
		}

		if v, ok := tfMap["source_selection_criteria"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.SourceSelectionCriteria = expandBucketReplicationConfigurationRulesSourceSelectionCriteria(v)
		}

		if v, ok := tfMap["filter"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			// XML schema V2.
			tfMap := v[0].(map[string]interface{})
			var filter types.ReplicationRuleFilter

			if tags := Tags(tftags.New(ctx, tfMap["tags"]).IgnoreAWS()); len(tags) > 0 {
				filter = &types.ReplicationRuleFilterMemberAnd{
					Value: types.ReplicationRuleAndOperator{
						Prefix: aws.String(tfMap["prefix"].(string)),
						Tags:   tags,
					},
				}
			} else {
				filter = &types.ReplicationRuleFilterMemberPrefix{
					Value: tfMap["prefix"].(string),
				}
			}

			rule.Filter = filter
			rule.Priority = aws.Int32(int32(tfMap["priority"].(int)))

			if v, ok := tfMap["delete_marker_replication_status"].(string); ok && v != "" {
				rule.DeleteMarkerReplication = &types.DeleteMarkerReplication{
					Status: types.DeleteMarkerReplicationStatus(v),
				}
			} else {
				rule.DeleteMarkerReplication = &types.DeleteMarkerReplication{
					Status: types.DeleteMarkerReplicationStatusDisabled,
				}
			}
		} else {
			// XML schema V1.
			rule.Prefix = aws.String(tfMap["prefix"].(string))
		}

		rules = append(rules, rule)
	}

	return rules
}

func expandBucketReplicationConfigurationRulesDestination(l []interface{}) *types.Destination {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.Destination{}

	if v, ok := tfMap["bucket"].(string); ok {
		apiObject.Bucket = aws.String(v)
	}

	if v, ok := tfMap["storage_class"].(string); ok && v != "" {
		apiObject.StorageClass = types.StorageClass(v)
	}

	if v, ok := tfMap["replica_kms_key_id"].(string); ok && v != "" {
		apiObject.EncryptionConfiguration = &types.EncryptionConfiguration{
			ReplicaKmsKeyID: aws.String(v),
		}
	}

	if v, ok := tfMap["account_id"].(string); ok && v != "" {
		apiObject.Account = aws.String(v)
	}

	if v, ok := tfMap["access_control_translation"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		aclTranslationValues := v[0].(map[string]interface{})
		ruleAclTranslation := &s3.AccessControlTranslation{}
		ruleAclTranslation.Owner = aws.String(aclTranslationValues["owner"].(string))
		apiObject.AccessControlTranslation = ruleAclTranslation
	}

	// replication metrics (required for RTC)
	if v, ok := tfMap["metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		metricsConfig := &s3.Metrics{}
		metricsValues := v[0].(map[string]interface{})
		metricsConfig.EventThreshold = &s3.ReplicationTimeValue{}
		metricsConfig.Status = aws.String(metricsValues["status"].(string))
		metricsConfig.EventThreshold.Minutes = aws.Int64(int64(metricsValues["minutes"].(int)))
		apiObject.Metrics = metricsConfig
	}

	// replication time control (RTC)
	if v, ok := tfMap["replication_time"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		rtcValues := v[0].(map[string]interface{})
		rtcConfig := &s3.ReplicationTime{}
		rtcConfig.Status = aws.String(rtcValues["status"].(string))
		rtcConfig.Time = &s3.ReplicationTimeValue{}
		rtcConfig.Time.Minutes = aws.Int64(int64(rtcValues["minutes"].(int)))
		apiObject.ReplicationTime = rtcConfig
	}

	return apiObject
}

func expandBucketReplicationConfigurationRulesSourceSelectionCriteria(l []interface{}) *types.SourceSelectionCriteria {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.SourceSelectionCriteria{}

	if v, ok := tfMap["sse_kms_encrypted_objects"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]interface{})
		apiObject := &types.SseKmsEncryptedObjects{}

		if tfMap["enabled"].(bool) {
			apiObject.Status = types.SseKmsEncryptedObjectsStatusEnabled
		} else {
			apiObject.Status = types.SseKmsEncryptedObjectsStatusDisabled
		}

		result.SseKmsEncryptedObjects = apiObject
	}

	return result
}

func flattenBucketReplicationConfiguration(ctx context.Context, r *s3.ReplicationConfiguration) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if r.Role != nil {
		m["role"] = aws.StringValue(r.Role)
	}

	if len(r.Rules) > 0 {
		m["rules"] = flattenBucketReplicationConfigurationReplicationRules(ctx, r.Rules)
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

func flattenBucketReplicationConfigurationReplicationRuleFilter(ctx context.Context, filter *s3.ReplicationRuleFilter) []interface{} {
	if filter == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if filter.Prefix != nil {
		m["prefix"] = aws.StringValue(filter.Prefix)
	}

	if filter.Tag != nil {
		m["tags"] = KeyValueTags(ctx, []*s3.Tag{filter.Tag}).IgnoreAWS().Map()
	}

	if filter.And != nil {
		m["prefix"] = aws.StringValue(filter.And.Prefix)
		m["tags"] = KeyValueTags(ctx, filter.And.Tags).IgnoreAWS().Map()
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

func flattenBucketReplicationConfigurationReplicationRules(ctx context.Context, rules []*s3.ReplicationRule) []interface{} {
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
			m["filter"] = flattenBucketReplicationConfigurationReplicationRuleFilter(ctx, rule.Filter)
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

func expandBucketVersioningConfigurationCreate(l []interface{}) *types.VersioningConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.VersioningConfiguration{}

	// Only set and return a non-nil VersioningConfiguration with at least one of
	// MFADelete or Status enabled as the PutBucketVersioning API request
	// does not need to be made for new buckets that don't require versioning.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/4494.

	if v, ok := tfMap["enabled"].(bool); ok && v {
		apiObject.Status = types.BucketVersioningStatusEnabled
	}

	if v, ok := tfMap["mfa_delete"].(bool); ok && v {
		apiObject.MFADelete = types.MFADeleteEnabled
	}

	if itypes.IsZero(&apiObject) {
		return nil
	}

	return apiObject
}

func expandBucketVersioningConfigurationUpdate(l []interface{}) *types.VersioningConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.VersioningConfiguration{}

	if v, ok := tfMap["enabled"].(bool); ok {
		if v {
			apiObject.Status = types.BucketVersioningStatusEnabled
		} else {
			apiObject.Status = types.BucketVersioningStatusSuspended
		}
	}

	if v, ok := tfMap["mfa_delete"].(bool); ok {
		if v {
			apiObject.MFADelete = types.MFADeleteEnabled
		} else {
			apiObject.MFADelete = types.MFADeleteDisabled
		}
	}

	return apiObject
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

func expandBucketWebsiteConfiguration(l []interface{}) (*types.WebsiteConfiguration, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	websiteConfig := &types.WebsiteConfiguration{}

	if v, ok := tfMap["index_document"].(string); ok && v != "" {
		websiteConfig.IndexDocument = &types.IndexDocument{
			Suffix: aws.String(v),
		}
	}

	if v, ok := tfMap["error_document"].(string); ok && v != "" {
		websiteConfig.ErrorDocument = &types.ErrorDocument{
			Key: aws.String(v),
		}
	}

	if v, ok := tfMap["redirect_all_requests_to"].(string); ok && v != "" {
		redirect, err := url.Parse(v)
		if err == nil && redirect.Scheme != "" {
			var buf bytes.Buffer

			buf.WriteString(redirect.Host)
			if redirect.Path != "" {
				buf.WriteString(redirect.Path)
			}
			if redirect.RawQuery != "" {
				buf.WriteString("?")
				buf.WriteString(redirect.RawQuery)
			}
			websiteConfig.RedirectAllRequestsTo = &types.RedirectAllRequestsTo{
				HostName: aws.String(buf.String()),
				Protocol: types.Protocol(redirect.Scheme),
			}
		} else {
			websiteConfig.RedirectAllRequestsTo = &types.RedirectAllRequestsTo{
				HostName: aws.String(v),
			}
		}
	}

	if v, ok := tfMap["routing_rules"].(string); ok && v != "" {
		var routingRules []types.RoutingRule
		if err := json.Unmarshal([]byte(v), &routingRules); err != nil {
			return nil, err
		}
		websiteConfig.RoutingRules = routingRules
	}

	return websiteConfig, nil
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
			return nil, fmt.Errorf("while marshaling routing rules: %w", err)
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

// validBucketName validates any S3 bucket name that is not inside the us-east-1 region.
// Buckets outside of this region have to be DNS-compliant. After the same restrictions are
// applied to buckets in the us-east-1 region, this function can be refactored as a SchemaValidateFunc
func validBucketName(value string, region string) error {
	if region != names.USEast1RegionID {
		if (len(value) < 3) || (len(value) > 63) {
			return fmt.Errorf("%q must contain from 3 to 63 characters", value)
		}
		if !regexache.MustCompile(`^[0-9a-z-.]+$`).MatchString(value) {
			return fmt.Errorf("only lowercase alphanumeric characters and hyphens allowed in %q", value)
		}
		if regexache.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`).MatchString(value) {
			return fmt.Errorf("%q must not be formatted as an IP address", value)
		}
		if strings.HasPrefix(value, `.`) {
			return fmt.Errorf("%q cannot start with a period", value)
		}
		if strings.HasSuffix(value, `.`) {
			return fmt.Errorf("%q cannot end with a period", value)
		}
		if strings.Contains(value, `..`) {
			return fmt.Errorf("%q can be only one period between labels", value)
		}
	} else {
		if len(value) > 255 {
			return fmt.Errorf("%q must contain less than 256 characters", value)
		}
		if !regexache.MustCompile(`^[0-9A-Za-z_.-]+$`).MatchString(value) {
			return fmt.Errorf("only alphanumeric characters, hyphens, periods, and underscores allowed in %q", value)
		}
	}
	return nil
}

func validBucketLifecycleTimestamp(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := time.Parse(time.RFC3339, fmt.Sprintf("%sT00:00:00Z", value))
	if err != nil {
		errors = append(errors, fmt.Errorf(
			"%q cannot be parsed as RFC3339 Timestamp Format", value))
	}

	return
}
