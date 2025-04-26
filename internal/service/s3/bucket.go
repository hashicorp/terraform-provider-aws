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
	"slices"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// General timeout for S3 bucket changes to propagate.
	// See https://docs.aws.amazon.com/AmazonS3/latest/userguide/Welcome.html#ConsistencyModel.
	bucketPropagationTimeout = 2 * time.Minute
)

// @SDKResource("aws_s3_bucket", name="Bucket")
// @Tags(identifierAttribute="bucket", resourceType="Bucket")
// @Testing(importIgnore="force_destroy")
func resourceBucket() *schema.Resource {
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
				Deprecated:       "acceleration_status is deprecated. Use the aws_s3_bucket_accelerate_configuration resource instead.",
				ValidateDiagFunc: enum.Validate[types.BucketAccelerateStatus](),
			},
			"acl": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"grant"},
				ValidateFunc:  validation.StringInSlice(bucketCannedACL_Values(), false),
				Deprecated:    "acl is deprecated. Use the aws_s3_bucket_acl resource instead.",
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrBucket: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrBucketPrefix},
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 63),
					validation.StringDoesNotMatch(directoryBucketNameRegex, `must not be in the format [bucket_name]--[azid]--x-s3. Use the aws_s3_directory_bucket resource to manage S3 Express buckets`),
				),
			},
			"bucket_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrBucketPrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrBucket},
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
				Deprecated: "cors_rule is deprecated. Use the aws_s3_bucket_cors_configuration resource instead.",
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
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"grant": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"acl"},
				Deprecated:    "grant is deprecated. Use the aws_s3_bucket_acl resource instead.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrPermissions: {
							Type:     schema.TypeSet,
							Required: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[types.Permission](),
							},
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Required: true,
							// TypeAmazonCustomerByEmail is not currently supported
							ValidateFunc: validation.StringInSlice(enum.Slice(
								types.TypeCanonicalUser,
								types.TypeGroup,
							), false),
						},
						names.AttrURI: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lifecycle_rule": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				Deprecated: "lifecycle_rule is deprecated. Use the aws_s3_bucket_lifecycle_configuration resource instead.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"abort_incomplete_multipart_upload_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						names.AttrEnabled: {
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
						names.AttrID: {
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
									names.AttrStorageClass: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.TransitionStorageClass](),
									},
								},
							},
						},
						names.AttrPrefix: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrTags: tftags.TagsSchema(),
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
									names.AttrStorageClass: {
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
				Deprecated: "logging is deprecated. Use the aws_s3_bucket_logging resource instead.",
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
				Deprecated: "object_lock_configuration is deprecated. Use the top-level parameter object_lock_enabled and the aws_s3_bucket_object_lock_configuration resource instead.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object_lock_enabled": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ConflictsWith:    []string{"object_lock_enabled"},
							ValidateDiagFunc: enum.Validate[types.ObjectLockEnabled](),
							Deprecated:       "object_lock_enabled is deprecated. Use the top-level parameter object_lock_enabled instead.",
						},
						names.AttrRule: {
							Type:       schema.TypeList,
							Optional:   true,
							Deprecated: "rule is deprecated. Use the aws_s3_bucket_object_lock_configuration resource instead.",
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
												names.AttrMode: {
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
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Optional:              true,
				Computed:              true,
				Deprecated:            "policy is deprecated. Use the aws_s3_bucket_policy resource instead.",
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			names.AttrRegion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_configuration": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				MaxItems:   1,
				Deprecated: "replication_configuration is deprecated. Use the aws_s3_bucket_replication_configuration resource instead.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrRole: {
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
									names.AttrDestination: {
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
															names.AttrOwner: {
																Type:             schema.TypeString,
																Required:         true,
																ValidateDiagFunc: enum.Validate[types.OwnerOverride](),
															},
														},
													},
												},
												names.AttrAccountID: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidAccountID,
												},
												names.AttrBucket: {
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
															names.AttrStatus: {
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
															names.AttrStatus: {
																Type:             schema.TypeString,
																Optional:         true,
																Default:          types.ReplicationTimeStatusEnabled,
																ValidateDiagFunc: enum.Validate[types.ReplicationTimeStatus](),
															},
														},
													},
												},
												names.AttrStorageClass: {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.StorageClass](),
												},
											},
										},
									},
									names.AttrFilter: {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrPrefix: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 1024),
												},
												names.AttrTags: tftags.TagsSchema(),
											},
										},
									},
									names.AttrID: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 255),
									},
									names.AttrPrefix: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
									},
									names.AttrPriority: {
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
															names.AttrEnabled: {
																Type:     schema.TypeBool,
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									names.AttrStatus: {
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
				Deprecated:       "request_payer is deprecated. Use the aws_s3_bucket_request_payment_configuration resource instead.",
				ValidateDiagFunc: enum.Validate[types.Payer](),
			},
			"server_side_encryption_configuration": {
				Type:       schema.TypeList,
				MaxItems:   1,
				Optional:   true,
				Computed:   true,
				Deprecated: "server_side_encryption_configuration is deprecated. Use the aws_s3_bucket_server_side_encryption_configuration resource instead.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrRule: {
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
				Deprecated: "versioning is deprecated. Use the aws_s3_bucket_versioning resource instead.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
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
				Deprecated: "website is deprecated. Use the aws_s3_bucket_website_configuration resource instead.",
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
							StateFunc: func(v any) string {
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
				Deprecated: "website_domain is deprecated. Use the aws_s3_bucket_website_configuration resource instead.",
			},
			"website_endpoint": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "website_endpoint is deprecated. Use the aws_s3_bucket_website_configuration resource instead.",
			},
		},
	}
}

func resourceBucketCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := create.Name(d.Get(names.AttrBucket).(string), d.Get(names.AttrBucketPrefix).(string))
	region := meta.(*conns.AWSClient).Region(ctx)

	if err := validBucketName(bucket, region); err != nil {
		return sdkdiag.AppendErrorf(diags, "validating S3 Bucket (%s) name: %s", bucket, err)
	}

	// Special case: us-east-1 does not return error if the bucket already exists and is owned by
	// current account. It also resets the Bucket ACLs.
	if region == endpoints.UsEast1RegionID {
		if _, err := findBucket(ctx, conn, bucket); err == nil {
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

	// See https://docs.aws.amazon.com/AmazonS3/latest/API/API_CreateBucket.html#AmazonS3-CreateBucket-request-LocationConstraint.
	if region != endpoints.UsEast1RegionID {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}

	// S3 Object Lock is not supported on all partitions.
	if v, ok := d.GetOk("object_lock_enabled"); ok {
		input.ObjectLockEnabledForBucket = aws.Bool(v.(bool))
	}

	// S3 Object Lock can only be enabled on bucket creation.
	if v := expandBucketObjectLockConfiguration(d.Get("object_lock_configuration").([]any)); v != nil && v.ObjectLockEnabled == types.ObjectLockEnabledEnabled {
		input.ObjectLockEnabledForBucket = aws.Bool(true)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (any, error) {
		return conn.CreateBucket(ctx, input)
	}, errCodeOperationAborted)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s): %s", bucket, err)
	}

	d.SetId(bucket)

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (any, error) {
		return findBucket(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket (%s) create: %s", d.Id(), err)
	}

	if err := bucketCreateTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting S3 Bucket (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceBucketUpdate(ctx, d, meta)...)
}

func resourceBucketRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	_, err := findBucket(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "s3",
		Resource:  d.Id(),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrBucket, d.Id())
	d.Set("bucket_domain_name", meta.(*conns.AWSClient).PartitionHostname(ctx, d.Id()+".s3"))
	d.Set(names.AttrBucketPrefix, create.NamePrefixFromName(d.Id()))

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
		policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), policy)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set(names.AttrPolicy, policyToSet)
	case tfresource.NotFound(err), tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set(names.AttrPolicy, nil)
	default:
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) policy: %s", d.Id(), err)
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
		if err := d.Set("grant", flattenBucketGrants(bucketACL)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting grant: %s", err)
		}
	case tfresource.NotFound(err), tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("grant", nil)
	default:
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) ACL: %s", d.Id(), err)
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
	case tfresource.NotFound(err), tfawserr.ErrCodeEquals(err, errCodeNoSuchCORSConfiguration, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("cors_rule", nil)
	default:
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) CORS configuration: %s", d.Id(), err)
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
	case tfresource.NotFound(err), tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("website", nil)
	default:
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) website configuration: %s", d.Id(), err)
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
		if err := d.Set("versioning", flattenBucketVersioning(bucketVersioning)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting versioning: %s", err)
		}
	case tfresource.NotFound(err), tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("versioning", nil)
	default:
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) versioning: %s", d.Id(), err)
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
	case tfresource.NotFound(err), tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented, errCodeUnsupportedArgument, errCodeUnsupportedOperation):
		d.Set("acceleration_status", nil)
	default:
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) accelerate configuration: %s", d.Id(), err)
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
	case tfresource.NotFound(err), tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("request_payer", nil)
	default:
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) request payment configuration: %s", d.Id(), err)
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
	case tfresource.NotFound(err), tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("logging", nil)
	default:
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) logging: %s", d.Id(), err)
	}

	//
	// Bucket Lifecycle Configuration.
	//
	lifecycleRules, err := retryWhenNoSuchBucketError(ctx, d.Timeout(schema.TimeoutRead), func() ([]types.LifecycleRule, error) {
		output, err := findBucketLifecycleConfiguration(ctx, conn, d.Id(), "")

		if err != nil {
			return nil, err
		}

		return output.Rules, nil
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
	case tfresource.NotFound(err), tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("lifecycle_rule", nil)
	default:
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) lifecycle configuration: %s", d.Id(), err)
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
	case tfresource.NotFound(err), tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("replication_configuration", nil)
	default:
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) replication configuration: %s", d.Id(), err)
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
		if err := d.Set("server_side_encryption_configuration", flattenBucketServerSideEncryptionConfiguration(encryptionConfiguration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting server_side_encryption_configuration: %s", err)
		}
	case tfresource.NotFound(err), tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented, errCodeUnsupportedOperation):
		d.Set("server_side_encryption_configuration", nil)
	default:
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) server-side encryption configuration: %s", d.Id(), err)
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
	case tfresource.NotFound(err), tfawserr.ErrCodeEquals(err, errCodeMethodNotAllowed, errCodeNotImplemented, errCodeXNotImplemented):
		d.Set("object_lock_configuration", nil)
		d.Set("object_lock_enabled", nil)
	default:
		if partition := meta.(*conns.AWSClient).Partition(ctx); partition == endpoints.AwsPartitionID || partition == endpoints.AwsUsGovPartitionID {
			return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) object lock configuration: %s", d.Id(), err)
		}
		log.Printf("[WARN] Unable to read S3 Bucket (%s) Object Lock Configuration: %s", d.Id(), err)
		d.Set("object_lock_configuration", nil)
		d.Set("object_lock_enabled", nil)
	}

	//
	// Bucket Region etc.
	//
	region, err := manager.GetBucketRegion(ctx, conn, d.Id(), func(o *s3.Options) {
		o.UsePathStyle = meta.(*conns.AWSClient).S3UsePathStyle(ctx)
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) location: %s", d.Id(), err)
	}

	d.Set(names.AttrRegion, region)
	d.Set("bucket_regional_domain_name", bucketRegionalDomainName(d.Id(), region))

	hostedZoneID, err := hostedZoneIDForRegion(region)
	if err != nil {
		log.Printf("[WARN] %s", err)
	} else {
		d.Set(names.AttrHostedZoneID, hostedZoneID)
	}

	if _, ok := d.GetOk("website"); ok {
		endpoint, domain := bucketWebsiteEndpointAndDomain(d.Id(), region)
		d.Set("website_domain", domain)
		d.Set("website_endpoint", endpoint)
	}

	return diags
}

func resourceBucketUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	// Note: Order of argument updates below is important.

	//
	// Bucket Policy.
	//
	if d.HasChange(names.AttrPolicy) {
		policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if policy == "" {
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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

			_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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
		if v, ok := d.GetOk("cors_rule"); !ok || len(v.([]any)) == 0 {
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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
					CORSRules: expandBucketCORSRules(d.Get("cors_rule").([]any)),
				},
			}

			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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
		if v, ok := d.GetOk("website"); !ok || len(v.([]any)) == 0 || v.([]any)[0] == nil {
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
				return conn.DeleteBucketWebsite(ctx, &s3.DeleteBucketWebsiteInput{
					Bucket: aws.String(d.Id()),
				})
			}, errCodeNoSuchBucket)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket (%s) website configuration: %s", d.Id(), err)
			}
		} else {
			websiteConfig, err := expandBucketWebsiteConfiguration(v.([]any))
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			input := &s3.PutBucketWebsiteInput{
				Bucket:               aws.String(d.Id()),
				WebsiteConfiguration: websiteConfig,
			}

			_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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
		v := d.Get("versioning").([]any)
		var versioningConfig *types.VersioningConfiguration

		if d.IsNewResource() {
			versioningConfig = expandBucketVersioningConfigurationCreate(v)
		} else {
			versioningConfig = expandBucketVersioningConfigurationUpdate(v)
		}

		if versioningConfig != nil {
			input := &s3.PutBucketVersioningInput{
				Bucket:                  aws.String(d.Id()),
				VersioningConfiguration: versioningConfig,
			}

			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
				return conn.PutBucketVersioning(ctx, input)
			}, errCodeNoSuchBucket)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) versioning: %s", d.Id(), err)
			}
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

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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
				Grants: expandBucketGrants(d.Get("grant").(*schema.Set).List()),
				Owner:  bucketACL.Owner,
			},
			Bucket: aws.String(d.Id()),
		}

		_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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

		if v, ok := d.GetOk("logging"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			tfMap := v.([]any)[0].(map[string]any)

			input.BucketLoggingStatus.LoggingEnabled = &types.LoggingEnabled{}

			if v, ok := tfMap["target_bucket"].(string); ok && v != "" {
				input.BucketLoggingStatus.LoggingEnabled.TargetBucket = aws.String(v)
			}

			if v, ok := tfMap["target_prefix"].(string); ok && v != "" {
				input.BucketLoggingStatus.LoggingEnabled.TargetPrefix = aws.String(v)
			}
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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
		if v, ok := d.GetOk("lifecycle_rule"); !ok || len(v.([]any)) == 0 {
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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
					Rules: expandBucketLifecycleRules(ctx, v.([]any)),
				},
			}

			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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
		if v, ok := d.GetOk("replication_configuration"); !ok || len(v.([]any)) == 0 || v.([]any)[0] == nil {
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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
			if v, ok := d.GetOk("versioning"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				tfMap := v.([]any)[0].(map[string]any)

				if tfMap[names.AttrEnabled].(bool) {
					hasVersioning = true
				}
			}

			if !hasVersioning {
				return sdkdiag.AppendErrorf(diags, "versioning must be enabled on S3 Bucket (%s) to allow replication", d.Id())
			}

			input := &s3.PutBucketReplicationInput{
				Bucket:                   aws.String(d.Id()),
				ReplicationConfiguration: expandBucketReplicationConfiguration(ctx, v.([]any)),
			}

			_, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutUpdate),
				func() (any, error) {
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
		if v, ok := d.GetOk("server_side_encryption_configuration"); !ok || len(v.([]any)) == 0 {
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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
					Rules: expandBucketServerSideEncryptionRules(v.([]any)),
				},
			}

			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
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
			ObjectLockConfiguration: expandBucketObjectLockConfiguration(d.Get("object_lock_configuration").([]any)),
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
			return conn.PutObjectLockConfiguration(ctx, input)
		}, errCodeNoSuchBucket)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) object lock configuration: %s", d.Id(), err)
		}
	}

	return append(diags, resourceBucketRead(ctx, d, meta)...)
}

func resourceBucketDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
		if d.Get(names.AttrForceDestroy).(bool) {
			// Delete everything including locked objects.
			// Don't ignore any object errors or we could recurse infinitely.
			var objectLockEnabled bool
			if v := expandBucketObjectLockConfiguration(d.Get("object_lock_configuration").([]any)); v != nil {
				objectLockEnabled = v.ObjectLockEnabled == types.ObjectLockEnabledEnabled
			}

			if n, err := emptyBucket(ctx, conn, d.Id(), objectLockEnabled); err != nil {
				return sdkdiag.AppendErrorf(diags, "emptying S3 Bucket (%s): %s", d.Id(), err)
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

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete), func() (any, error) {
		return findBucket(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findBucket(ctx context.Context, conn *s3.Client, bucket string, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	input := s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	output, err := conn.HeadBucket(ctx, &input, optFns...)

	// For directory buckets that no longer exist it's the CreateSession call invoked by HeadBucket that returns "NoSuchBucket",
	// and that error code is flattend into HeadBucket's error message -- hence the 'errs.Contains' call.
	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotFound) || tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) || errs.Contains(err, errCodeNoSuchBucket) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findBucketRegion(ctx context.Context, awsClient *conns.AWSClient, bucket string, optFns ...func(*s3.Options)) (string, error) {
	optFns = append(slices.Clone(optFns),
		func(o *s3.Options) {
			// By default, GetBucketRegion forces virtual host addressing, which
			// is not compatible with many non-AWS implementations. Instead, pass
			// the provider s3_force_path_style configuration, which defaults to
			// false, but allows override.
			o.UsePathStyle = awsClient.S3UsePathStyle(ctx)
		},
		func(o *s3.Options) {
			// By default, GetBucketRegion uses anonymous credentials when doing
			// a HEAD request to get the bucket region. This breaks in aws-cn regions
			// when the account doesn't have an ICP license to host public content.
			// Use the current credentials when getting the bucket region.
			o.Credentials = awsClient.CredentialsProvider(ctx)
		})

	region, err := manager.GetBucketRegion(ctx, awsClient.S3Client(ctx), bucket, optFns...)

	if errs.IsA[manager.BucketNotFound](err) {
		return "", &retry.NotFoundError{
			LastError:   err,
			LastRequest: bucket,
		}
	}

	if err != nil {
		return "", err
	}

	return region, nil
}

func retryWhenNoSuchBucketError[T any](ctx context.Context, timeout time.Duration, f func() (T, error)) (T, error) {
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (any, error) {
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
	return fmt.Sprintf("%s.s3.%s.%s", bucket, region, names.PartitionForRegion(region).DNSSuffix())
}

func bucketWebsiteEndpointAndDomain(bucket, region string) (string, string) {
	var domain string

	// Default to us-east-1 if the bucket doesn't have a region:
	// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketGETlocation.html
	if region == "" {
		region = endpoints.UsEast1RegionID
	}

	// Different regions have different syntax for website endpoints:
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteEndpoints.html
	// https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_website_region_endpoints
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
	if slices.Contains(oldRegions, region) {
		domain = fmt.Sprintf("s3-website-%s.amazonaws.com", region) //lintignore:AWSR001
	} else {
		domain = fmt.Sprintf("s3-website.%s.%s", region, names.PartitionForRegion(region).DNSSuffix())
	}

	return fmt.Sprintf("%s.%s", bucket, domain), domain
}

//
// Bucket CORS Configuration.
//

func expandBucketCORSRules(l []any) []types.CORSRule {
	if len(l) == 0 {
		return nil
	}

	var rules []types.CORSRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		rule := types.CORSRule{}

		if v, ok := tfMap["allowed_headers"].([]any); ok && len(v) > 0 {
			rule.AllowedHeaders = flex.ExpandStringValueListEmpty(v)
		}

		if v, ok := tfMap["allowed_methods"].([]any); ok && len(v) > 0 {
			rule.AllowedMethods = flex.ExpandStringValueListEmpty(v)
		}

		if v, ok := tfMap["allowed_origins"].([]any); ok && len(v) > 0 {
			rule.AllowedOrigins = flex.ExpandStringValueListEmpty(v)
		}

		if v, ok := tfMap["expose_headers"].([]any); ok && len(v) > 0 {
			rule.ExposeHeaders = flex.ExpandStringValueListEmpty(v)
		}

		if v, ok := tfMap["max_age_seconds"].(int); ok {
			rule.MaxAgeSeconds = aws.Int32(int32(v))
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenBucketCORSRules(rules []types.CORSRule) []any {
	var results []any

	for _, rule := range rules {
		m := map[string]any{
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

//
// Bucket Website Configuration.
//

func expandBucketWebsiteConfiguration(l []any) (*types.WebsiteConfiguration, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	tfMap, ok := l[0].(map[string]any)
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

func flattenBucketWebsite(apiObject *s3.GetBucketWebsiteOutput) ([]any, error) {
	if apiObject == nil {
		return []any{}, nil
	}

	m := make(map[string]any)

	if v := apiObject.IndexDocument; v != nil {
		m["index_document"] = aws.ToString(v.Suffix)
	}

	if v := apiObject.ErrorDocument; v != nil {
		m["error_document"] = aws.ToString(v.Key)
	}

	if apiObject := apiObject.RedirectAllRequestsTo; apiObject != nil {
		hostName := aws.ToString(apiObject.HostName)

		if apiObject.Protocol == "" {
			m["redirect_all_requests_to"] = hostName
		} else {
			var host string
			var path string
			var query string

			parsedHostName, err := url.Parse(hostName)
			if err == nil {
				host = parsedHostName.Host
				path = parsedHostName.Path
				query = parsedHostName.RawQuery
			} else {
				host = hostName
			}

			m["redirect_all_requests_to"] = (&url.URL{
				Scheme:   string(apiObject.Protocol),
				Host:     host,
				Path:     path,
				RawQuery: query,
			}).String()
		}
	}

	if apiObject := apiObject.RoutingRules; apiObject != nil {
		rr, err := normalizeRoutingRules(apiObject)
		if err != nil {
			return nil, err
		}
		m["routing_rules"] = rr
	}

	// We have special handling for the website configuration,
	// so only return the configuration if there is one.
	if len(m) == 0 {
		return []any{}, nil
	}

	return []any{m}, nil
}

//
// Bucket Versioning.
//

func expandBucketVersioningConfigurationCreate(l []any) *types.VersioningConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.VersioningConfiguration{}

	// Only set and return a non-nil VersioningConfiguration with at least one of
	// MFADelete or Status enabled as the PutBucketVersioning API request
	// does not need to be made for new buckets that don't require versioning.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/4494.

	if v, ok := tfMap[names.AttrEnabled].(bool); ok && v {
		apiObject.Status = types.BucketVersioningStatusEnabled
	}

	if v, ok := tfMap["mfa_delete"].(bool); ok && v {
		apiObject.MFADelete = types.MFADeleteEnabled
	}

	if apiObject.MFADelete == "" && apiObject.Status == "" {
		return nil
	}

	return apiObject
}

func expandBucketVersioningConfigurationUpdate(l []any) *types.VersioningConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.VersioningConfiguration{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
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

func flattenBucketVersioning(config *s3.GetBucketVersioningOutput) []any {
	if config == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrEnabled: config.Status == types.BucketVersioningStatusEnabled,
		"mfa_delete":      config.MFADelete == types.MFADeleteStatusEnabled,
	}

	return []any{m}
}

//
// Bucket ACL.
//

func expandBucketGrants(l []any) []types.Grant {
	var grants []types.Grant

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		if v, ok := tfMap[names.AttrPermissions].(*schema.Set); ok {
			for _, rawPermission := range v.List() {
				permission, ok := rawPermission.(string)
				if !ok {
					continue
				}

				grantee := &types.Grantee{}

				if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
					grantee.ID = aws.String(v)
				}

				if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
					grantee.Type = types.Type(v)
				}

				if v, ok := tfMap[names.AttrURI].(string); ok && v != "" {
					grantee.URI = aws.String(v)
				}

				grant := types.Grant{
					Grantee:    grantee,
					Permission: types.Permission(permission),
				}

				grants = append(grants, grant)
			}
		}
	}

	return grants
}

func flattenBucketGrants(apiObject *s3.GetBucketAclOutput) []any {
	if len(apiObject.Grants) == 0 {
		return []any{}
	}

	getGrant := func(grants []any, grantee map[string]any) (any, bool) {
		for _, grant := range grants {
			tfMap := grant.(map[string]any)
			if tfMap[names.AttrType] == grantee[names.AttrType] && tfMap[names.AttrID] == grantee[names.AttrID] && tfMap[names.AttrURI] == grantee[names.AttrURI] && tfMap[names.AttrPermissions].(*schema.Set).Len() > 0 {
				return grant, true
			}
		}
		return nil, false
	}

	results := make([]any, 0, len(apiObject.Grants))

	for _, apiObject := range apiObject.Grants {
		grantee := apiObject.Grantee

		m := map[string]any{
			names.AttrType: grantee.Type,
		}

		if grantee.ID != nil {
			m[names.AttrID] = aws.ToString(grantee.ID)
		}

		if grantee.URI != nil {
			m[names.AttrURI] = aws.ToString(grantee.URI)
		}

		if v, ok := getGrant(results, m); ok {
			v.(map[string]any)[names.AttrPermissions].(*schema.Set).Add(string(apiObject.Permission))
		} else {
			m[names.AttrPermissions] = schema.NewSet(schema.HashString, []any{string(apiObject.Permission)})
			results = append(results, m)
		}
	}

	return results
}

//
// Bucket Logging.
//

func flattenBucketLoggingEnabled(apiObject *types.LoggingEnabled) []any {
	if apiObject == nil {
		return []any{}
	}

	m := make(map[string]any)

	if apiObject.TargetBucket != nil {
		m["target_bucket"] = aws.ToString(apiObject.TargetBucket)
	}

	if apiObject.TargetPrefix != nil {
		m["target_prefix"] = aws.ToString(apiObject.TargetPrefix)
	}

	return []any{m}
}

//
// Bucket Lifecycle Configuration.
//

func expandBucketLifecycleRules(ctx context.Context, tfList []any) []types.LifecycleRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []types.LifecycleRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := types.LifecycleRule{}

		if v, ok := tfMap["abort_incomplete_multipart_upload_days"].(int); ok && v > 0 {
			apiObject.AbortIncompleteMultipartUpload = &types.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: aws.Int32(int32(v)),
			}
		}

		if v, ok := tfMap["expiration"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.Expiration = expandBucketLifecycleExpiration(v)
		}

		var filter *types.LifecycleRuleFilter
		prefix := tfMap[names.AttrPrefix].(string)
		if tags := Tags(tftags.New(ctx, tfMap[names.AttrTags]).IgnoreAWS()); len(tags) > 0 {
			filter = &types.LifecycleRuleFilter{
				And: &types.LifecycleRuleAndOperator{
					Prefix: aws.String(prefix),
					Tags:   tags,
				},
			}
		} else {
			filter = &types.LifecycleRuleFilter{
				Prefix: aws.String(prefix),
			}
		}
		apiObject.Filter = filter

		if v, ok := tfMap[names.AttrID].(string); ok {
			apiObject.ID = aws.String(v)
		} else {
			apiObject.ID = aws.String(id.PrefixedUniqueId("tf-s3-lifecycle-"))
		}

		if v, ok := tfMap["noncurrent_version_expiration"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.NoncurrentVersionExpiration = expandBucketNoncurrentVersionExpiration(v[0].(map[string]any))
		}

		if v, ok := tfMap["noncurrent_version_transition"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.NoncurrentVersionTransitions = expandBucketNoncurrentVersionTransition(v.List())
		}

		if v, ok := tfMap[names.AttrEnabled].(bool); ok && v {
			apiObject.Status = types.ExpirationStatusEnabled
		} else {
			apiObject.Status = types.ExpirationStatusDisabled
		}

		if v, ok := tfMap["transition"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.Transitions = expandBucketTransitions(v.List())
		}

		// As a lifecycle rule requires 1 or more transition/expiration actions,
		// we explicitly pass a default ExpiredObjectDeleteMarker value to be able to create
		// the rule while keeping the policy unaffected if the conditions are not met.
		if apiObject.AbortIncompleteMultipartUpload == nil && apiObject.Expiration == nil && apiObject.NoncurrentVersionExpiration == nil && apiObject.NoncurrentVersionTransitions == nil && apiObject.Transitions == nil {
			apiObject.Expiration = &types.LifecycleExpiration{ExpiredObjectDeleteMarker: aws.Bool(false)}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandBucketLifecycleExpiration(tfList []any) *types.LifecycleExpiration {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &types.LifecycleExpiration{}

	if tfList[0] == nil {
		return apiObject
	}

	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap["date"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v+"T00:00:00Z")
		apiObject.Date = aws.Time(t)
	} else if v, ok := tfMap["days"].(int); ok && v > 0 {
		apiObject.Days = aws.Int32(int32(v))
	} else if v, ok := tfMap["expired_object_delete_marker"].(bool); ok {
		apiObject.ExpiredObjectDeleteMarker = aws.Bool(v)
	}

	return apiObject
}

func expandBucketNoncurrentVersionExpiration(tfMap map[string]any) *types.NoncurrentVersionExpiration {
	if len(tfMap) == 0 {
		return nil
	}

	var apiObject *types.NoncurrentVersionExpiration

	if v, ok := tfMap["days"].(int); ok {
		apiObject = &types.NoncurrentVersionExpiration{
			NoncurrentDays: aws.Int32(int32(v)),
		}
	}

	return apiObject
}

func expandBucketNoncurrentVersionTransition(tfList []any) []types.NoncurrentVersionTransition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []types.NoncurrentVersionTransition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := types.NoncurrentVersionTransition{}

		if v, ok := tfMap["days"].(int); ok {
			apiObject.NoncurrentDays = aws.Int32(int32(v))
		}

		if v, ok := tfMap[names.AttrStorageClass].(string); ok && v != "" {
			apiObject.StorageClass = types.TransitionStorageClass(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandBucketTransitions(tfList []any) []types.Transition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []types.Transition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := types.Transition{}

		if v, ok := tfMap["date"].(string); ok && v != "" {
			t, _ := time.Parse(time.RFC3339, v+"T00:00:00Z")
			apiObject.Date = aws.Time(t)
		} else if v, ok := tfMap["days"].(int); ok && v >= 0 {
			apiObject.Days = aws.Int32(int32(v))
		}

		if v, ok := tfMap[names.AttrStorageClass].(string); ok && v != "" {
			apiObject.StorageClass = types.TransitionStorageClass(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenBucketLifecycleRules(ctx context.Context, apiObjects []types.LifecycleRule) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)

		if apiObject := apiObject.AbortIncompleteMultipartUpload; apiObject != nil {
			if v := apiObject.DaysAfterInitiation; v != nil {
				tfMap["abort_incomplete_multipart_upload_days"] = aws.ToInt32(v)
			}
		}

		if v := apiObject.Expiration; v != nil {
			tfMap["expiration"] = flattenBucketLifecycleExpiration(v)
		}

		if apiObject := apiObject.Filter; apiObject != nil {
			if v := apiObject.And; v != nil {
				if v := v.Prefix; v != nil {
					tfMap[names.AttrPrefix] = aws.ToString(v)
				}

				if v := v.Tags; v != nil {
					tfMap[names.AttrTags] = keyValueTags(ctx, v).IgnoreAWS().Map()
				}
			}

			if v := apiObject.Prefix; v != nil {
				tfMap[names.AttrPrefix] = aws.ToString(v)
			}

			if v := apiObject.Tag; v != nil {
				tfMap[names.AttrTags] = keyValueTags(ctx, []types.Tag{*v}).IgnoreAWS().Map()
			}
		}

		if v := apiObject.ID; v != nil {
			tfMap[names.AttrID] = aws.ToString(v)
		}

		if v := apiObject.Prefix; v != nil {
			tfMap[names.AttrPrefix] = aws.ToString(v)
		}

		tfMap[names.AttrEnabled] = apiObject.Status == types.ExpirationStatusEnabled

		if apiObject := apiObject.NoncurrentVersionExpiration; apiObject != nil {
			m := make(map[string]any)

			if v := apiObject.NoncurrentDays; v != nil {
				m["days"] = aws.ToInt32(v)
			}

			tfMap["noncurrent_version_expiration"] = []any{m}
		}

		if v := apiObject.NoncurrentVersionTransitions; v != nil {
			tfMap["noncurrent_version_transition"] = flattenBucketNoncurrentVersionTransitions(v)
		}

		if v := apiObject.Transitions; v != nil {
			tfMap["transition"] = flattenBucketTransitions(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenBucketLifecycleExpiration(apiObject *types.LifecycleExpiration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)

	if v := apiObject.Date; v != nil {
		tfMap["date"] = v.Format("2006-01-02")
	}

	if v := apiObject.Days; v != nil {
		tfMap["days"] = aws.ToInt32(v)
	}

	if v := apiObject.ExpiredObjectDeleteMarker; v != nil {
		tfMap["expired_object_delete_marker"] = aws.ToBool(v)
	}

	return []any{tfMap}
}

func flattenBucketNoncurrentVersionTransitions(apiObjects []types.NoncurrentVersionTransition) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrStorageClass: apiObject.StorageClass,
		}

		if v := apiObject.NoncurrentDays; v != nil {
			tfMap["days"] = aws.ToInt32(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenBucketTransitions(apiObjects []types.Transition) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrStorageClass: apiObject.StorageClass,
		}

		if v := apiObject.Date; v != nil {
			tfMap["date"] = v.Format("2006-01-02")
		}

		if v := apiObject.Days; v != nil {
			tfMap["days"] = aws.ToInt32(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

//
// Bucket Replication Configuration.
//

func expandBucketReplicationConfiguration(ctx context.Context, tfList []any) *types.ReplicationConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.ReplicationConfiguration{}

	if v, ok := tfMap[names.AttrRole].(string); ok {
		apiObject.Role = aws.String(v)
	}

	if v, ok := tfMap["rules"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Rules = expandBucketReplicationRules(ctx, v.List())
	}

	return apiObject
}

func expandBucketReplicationRules(ctx context.Context, tfList []any) []types.ReplicationRule {
	var rules []types.ReplicationRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := types.ReplicationRule{}

		if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
			apiObject.Status = types.ReplicationRuleStatus(v)
		} else {
			continue
		}

		if v, ok := tfMap[names.AttrDestination].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.Destination = expandBucketDestination(v)
		} else {
			apiObject.Destination = &types.Destination{}
		}

		if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
			apiObject.ID = aws.String(v)
		}

		if v, ok := tfMap["source_selection_criteria"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.SourceSelectionCriteria = expandBucketSourceSelectionCriteria(v)
		}

		if v, ok := tfMap[names.AttrFilter].([]any); ok && len(v) > 0 && v[0] != nil {
			// XML schema V2.
			tfFilterMap := v[0].(map[string]any)
			var filter *types.ReplicationRuleFilter

			if tags := Tags(tftags.New(ctx, tfFilterMap[names.AttrTags]).IgnoreAWS()); len(tags) > 0 {
				filter = &types.ReplicationRuleFilter{
					And: &types.ReplicationRuleAndOperator{
						Prefix: aws.String(tfFilterMap[names.AttrPrefix].(string)),
						Tags:   tags,
					},
				}
			} else {
				filter = &types.ReplicationRuleFilter{
					Prefix: aws.String(tfFilterMap[names.AttrPrefix].(string)),
				}
			}

			apiObject.Filter = filter
			apiObject.Priority = aws.Int32(int32(tfMap[names.AttrPriority].(int)))

			if v, ok := tfMap["delete_marker_replication_status"].(string); ok && v != "" {
				apiObject.DeleteMarkerReplication = &types.DeleteMarkerReplication{
					Status: types.DeleteMarkerReplicationStatus(v),
				}
			} else {
				apiObject.DeleteMarkerReplication = &types.DeleteMarkerReplication{
					Status: types.DeleteMarkerReplicationStatusDisabled,
				}
			}
		} else {
			// XML schema V1.
			apiObject.Prefix = aws.String(tfMap[names.AttrPrefix].(string))
		}

		rules = append(rules, apiObject)
	}

	return rules
}

func expandBucketDestination(tfList []any) *types.Destination {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.Destination{}

	if v, ok := tfMap[names.AttrBucket].(string); ok {
		apiObject.Bucket = aws.String(v)
	}

	if v, ok := tfMap[names.AttrStorageClass].(string); ok && v != "" {
		apiObject.StorageClass = types.StorageClass(v)
	}

	if v, ok := tfMap["replica_kms_key_id"].(string); ok && v != "" {
		apiObject.EncryptionConfiguration = &types.EncryptionConfiguration{
			ReplicaKmsKeyID: aws.String(v),
		}
	}

	if v, ok := tfMap[names.AttrAccountID].(string); ok && v != "" {
		apiObject.Account = aws.String(v)
	}

	if v, ok := tfMap["access_control_translation"].([]any); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]any)

		apiObject.AccessControlTranslation = &types.AccessControlTranslation{
			Owner: types.OwnerOverride(tfMap[names.AttrOwner].(string)),
		}
	}

	if v, ok := tfMap["metrics"].([]any); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]any)

		apiObject.Metrics = &types.Metrics{
			EventThreshold: &types.ReplicationTimeValue{
				Minutes: aws.Int32(int32(tfMap["minutes"].(int))),
			},
			Status: types.MetricsStatus(tfMap[names.AttrStatus].(string)),
		}
	}

	if v, ok := tfMap["replication_time"].([]any); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]any)

		apiObject.ReplicationTime = &types.ReplicationTime{
			Status: types.ReplicationTimeStatus(tfMap[names.AttrStatus].(string)),
			Time: &types.ReplicationTimeValue{
				Minutes: aws.Int32(int32(tfMap["minutes"].(int))),
			},
		}
	}

	return apiObject
}

func expandBucketSourceSelectionCriteria(tfList []any) *types.SourceSelectionCriteria {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.SourceSelectionCriteria{}

	if v, ok := tfMap["sse_kms_encrypted_objects"].([]any); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]any)
		kmsEncryptedObjects := &types.SseKmsEncryptedObjects{}

		if tfMap[names.AttrEnabled].(bool) {
			kmsEncryptedObjects.Status = types.SseKmsEncryptedObjectsStatusEnabled
		} else {
			kmsEncryptedObjects.Status = types.SseKmsEncryptedObjectsStatusDisabled
		}

		apiObject.SseKmsEncryptedObjects = kmsEncryptedObjects
	}

	return apiObject
}

func flattenBucketReplicationConfiguration(ctx context.Context, apiObject *types.ReplicationConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)

	if v := apiObject.Role; v != nil {
		tfMap[names.AttrRole] = aws.ToString(v)
	}

	if v := apiObject.Rules; len(v) > 0 {
		tfMap["rules"] = flattenBucketReplicationRules(ctx, v)
	}

	return []any{tfMap}
}

func flattenBucketReplicationRules(ctx context.Context, apiObjects []types.ReplicationRule) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrStatus: apiObject.Status,
		}

		if apiObject := apiObject.DeleteMarkerReplication; apiObject != nil {
			if apiObject.Status == types.DeleteMarkerReplicationStatusEnabled {
				tfMap["delete_marker_replication_status"] = apiObject.Status
			}
		}

		if apiObject.Destination != nil {
			tfMap[names.AttrDestination] = flattenBucketDestination(apiObject.Destination)
		}

		if apiObject.Filter != nil {
			tfMap[names.AttrFilter] = flattenBucketReplicationRuleFilter(ctx, apiObject.Filter)
		}

		if apiObject.ID != nil {
			tfMap[names.AttrID] = aws.ToString(apiObject.ID)
		}

		if apiObject.Prefix != nil {
			tfMap[names.AttrPrefix] = aws.ToString(apiObject.Prefix)
		}

		if apiObject.Priority != nil {
			tfMap[names.AttrPriority] = aws.ToInt32(apiObject.Priority)
		}

		if apiObject.SourceSelectionCriteria != nil {
			tfMap["source_selection_criteria"] = flattenBucketSourceSelectionCriteria(apiObject.SourceSelectionCriteria)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenBucketDestination(apiObject *types.Destination) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrStorageClass: apiObject.StorageClass,
	}

	if apiObject := apiObject.AccessControlTranslation; apiObject != nil {
		m := map[string]any{
			names.AttrOwner: apiObject.Owner,
		}

		tfMap["access_control_translation"] = []any{m}
	}

	if apiObject.Account != nil {
		tfMap[names.AttrAccountID] = aws.ToString(apiObject.Account)
	}

	if apiObject.Bucket != nil {
		tfMap[names.AttrBucket] = aws.ToString(apiObject.Bucket)
	}

	if apiObject := apiObject.EncryptionConfiguration; apiObject != nil {
		if apiObject.ReplicaKmsKeyID != nil {
			tfMap["replica_kms_key_id"] = aws.ToString(apiObject.ReplicaKmsKeyID)
		}
	}

	if apiObject := apiObject.Metrics; apiObject != nil {
		m := map[string]any{
			names.AttrStatus: apiObject.Status,
		}

		if apiObject.EventThreshold != nil {
			m["minutes"] = aws.ToInt32(apiObject.EventThreshold.Minutes)
		}

		tfMap["metrics"] = []any{m}
	}

	if apiObject := apiObject.ReplicationTime; apiObject != nil {
		m := map[string]any{
			"minutes":        aws.ToInt32(apiObject.Time.Minutes),
			names.AttrStatus: apiObject.Status,
		}

		tfMap["replication_time"] = []any{m}
	}

	return []any{tfMap}
}

func flattenBucketReplicationRuleFilter(ctx context.Context, apiObject *types.ReplicationRuleFilter) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)

	if v := apiObject.And; v != nil {
		tfMap[names.AttrPrefix] = aws.ToString(v.Prefix)
		tfMap[names.AttrTags] = keyValueTags(ctx, v.Tags).IgnoreAWS().Map()
	}

	if v := apiObject.Prefix; v != nil {
		tfMap[names.AttrPrefix] = aws.ToString(v)
	}

	if v := apiObject.Tag; v != nil {
		tfMap[names.AttrTags] = keyValueTags(ctx, []types.Tag{*v}).IgnoreAWS().Map()
	}

	return []any{tfMap}
}

func flattenBucketSourceSelectionCriteria(apiObject *types.SourceSelectionCriteria) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)

	if v := apiObject.SseKmsEncryptedObjects; v != nil {
		tfMap["sse_kms_encrypted_objects"] = flattenBucketSSEKMSEncryptedObjects(v)
	}

	return []any{tfMap}
}

func flattenBucketSSEKMSEncryptedObjects(apiObject *types.SseKmsEncryptedObjects) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)

	if apiObject.Status == types.SseKmsEncryptedObjectsStatusEnabled {
		tfMap[names.AttrEnabled] = true
	} else if apiObject.Status == types.SseKmsEncryptedObjectsStatusDisabled {
		tfMap[names.AttrEnabled] = false
	}

	return []any{tfMap}
}

//
// Bucket Server-side Encryption Configuration.
//

func expandBucketServerSideEncryptionRules(l []any) []types.ServerSideEncryptionRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]any)
	if !ok {
		return nil
	}

	var rules []types.ServerSideEncryptionRule

	if l, ok := tfMap[names.AttrRule].([]any); ok && len(l) > 0 {
		for _, tfMapRaw := range l {
			tfMap, ok := tfMapRaw.(map[string]any)
			if !ok {
				continue
			}

			rule := types.ServerSideEncryptionRule{}

			if v, ok := tfMap["apply_server_side_encryption_by_default"].([]any); ok && len(v) > 0 && v[0] != nil {
				rule.ApplyServerSideEncryptionByDefault = expandBucketServerSideEncryptionByDefault(v)
			}

			if v, ok := tfMap["bucket_key_enabled"].(bool); ok {
				rule.BucketKeyEnabled = aws.Bool(v)
			}

			rules = append(rules, rule)
		}
	}

	return rules
}

func expandBucketServerSideEncryptionByDefault(l []any) *types.ServerSideEncryptionByDefault {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]any)
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

func flattenBucketServerSideEncryptionConfiguration(apiObject *types.ServerSideEncryptionConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrRule: flattenBucketServerSideEncryptionRules(apiObject.Rules),
	}

	return []any{m}
}

func flattenBucketServerSideEncryptionRules(rules []types.ServerSideEncryptionRule) []any {
	var results []any

	for _, rule := range rules {
		m := make(map[string]any)

		if apiObject := rule.ApplyServerSideEncryptionByDefault; apiObject != nil {
			m["apply_server_side_encryption_by_default"] = []any{
				map[string]any{
					"kms_master_key_id": aws.ToString(apiObject.KMSMasterKeyID),
					"sse_algorithm":     apiObject.SSEAlgorithm,
				},
			}
		}

		if rule.BucketKeyEnabled != nil {
			m["bucket_key_enabled"] = aws.ToBool(rule.BucketKeyEnabled)
		}

		results = append(results, m)
	}

	return results
}

//
// Bucket Object Lock Configuration.
//

func expandBucketObjectLockConfiguration(l []any) *types.ObjectLockConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.ObjectLockConfiguration{}

	if v, ok := tfMap["object_lock_enabled"].(string); ok && v != "" {
		apiObject.ObjectLockEnabled = types.ObjectLockEnabled(v)
	}

	if v, ok := tfMap[names.AttrRule].([]any); ok && len(v) > 0 {
		tfMap := v[0].(map[string]any)

		if v, ok := tfMap["default_retention"].([]any); ok && len(v) > 0 && v[0] != nil {
			tfMap := v[0].(map[string]any)

			apiObject.Rule = &types.ObjectLockRule{
				DefaultRetention: &types.DefaultRetention{},
			}

			if v, ok := tfMap["days"].(int); ok && v > 0 {
				apiObject.Rule.DefaultRetention.Days = aws.Int32(int32(v))
			}
			if v, ok := tfMap[names.AttrMode].(string); ok && v != "" {
				apiObject.Rule.DefaultRetention.Mode = types.ObjectLockRetentionMode(v)
			}
			if v, ok := tfMap["years"].(int); ok && v > 0 {
				apiObject.Rule.DefaultRetention.Years = aws.Int32(int32(v))
			}
		}
	}

	return apiObject
}

func flattenObjectLockConfiguration(apiObject *types.ObjectLockConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	m := map[string]any{
		"object_lock_enabled": apiObject.ObjectLockEnabled,
	}

	if apiObject.Rule != nil && apiObject.Rule.DefaultRetention != nil {
		apiObject := apiObject.Rule.DefaultRetention
		tfMap := map[string]any{
			"default_retention": []any{
				map[string]any{
					"days":         aws.ToInt32(apiObject.Days),
					names.AttrMode: apiObject.Mode,
					"years":        aws.ToInt32(apiObject.Years),
				},
			},
		}

		m[names.AttrRule] = []any{tfMap}
	}

	return []any{m}
}

// validBucketName validates any S3 bucket name that is not inside the us-east-1 region.
// Buckets outside of this region have to be DNS-compliant. After the same restrictions are
// applied to buckets in the us-east-1 region, this function can be refactored as a SchemaValidateFunc
func validBucketName(value string, region string) error {
	if region != endpoints.UsEast1RegionID {
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

func validBucketLifecycleTimestamp(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := time.Parse(time.RFC3339, fmt.Sprintf("%sT00:00:00Z", value))
	if err != nil {
		errors = append(errors, fmt.Errorf(
			"%q cannot be parsed as RFC3339 Timestamp Format", value))
	}

	return
}
