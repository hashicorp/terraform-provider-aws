// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_distribution", name="Distribution")
// @Tags(identifierAttribute="arn")
func ResourceDistribution() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceDistributionCreate,
		ReadWithoutTimeout:   resourceDistributionRead,
		UpdateWithoutTimeout: resourceDistributionUpdate,
		DeleteWithoutTimeout: resourceDistributionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				// Set non API attributes to their Default settings in the schema
				d.Set("retain_on_delete", false)
				d.Set("wait_for_deployment", true)
				return []*schema.ResourceData{d}, nil
			},
		},
		MigrateState:  resourceDistributionMigrateState,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aliases": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      AliasesHash,
			},
			"ordered_cache_behavior": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_methods": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"cached_methods": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"cache_policy_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"compress": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"default_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"field_level_encryption_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"forwarded_values": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cookies": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"forward": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(cloudfront.ItemSelection_Values(), false),
												},
												"whitelisted_names": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"headers": {
										Type:     schema.TypeSet,
										Optional: true,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"query_string": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"query_string_cache_keys": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"lambda_function_association": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 4,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.EventType_Values(), false),
									},
									"lambda_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"include_body": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
							Set: LambdaFunctionAssociationHash,
						},
						"function_association": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 2,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.EventType_Values(), false),
									},
									"function_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"max_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"min_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"origin_request_policy_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"path_pattern": {
							Type:     schema.TypeString,
							Required: true,
						},
						"realtime_log_config_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"response_headers_policy_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"smooth_streaming": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"target_origin_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"trusted_key_groups": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"trusted_signers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"viewer_protocol_policy": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(cloudfront.ViewerProtocolPolicy_Values(), false),
						},
					},
				},
			},
			"continuous_deployment_policy_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"comment": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"custom_error_response": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      CustomErrorResponseHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"error_caching_min_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"error_code": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"response_code": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"response_page_path": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"default_cache_behavior": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_methods": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"cached_methods": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"cache_policy_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"compress": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"default_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"field_level_encryption_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"forwarded_values": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cookies": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"forward": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(cloudfront.ItemSelection_Values(), false),
												},
												"whitelisted_names": {
													Type:     schema.TypeSet,
													Optional: true,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"headers": {
										Type:     schema.TypeSet,
										Optional: true,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"query_string": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"query_string_cache_keys": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"lambda_function_association": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 4,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.EventType_Values(), false),
									},
									"lambda_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"include_body": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
							Set: LambdaFunctionAssociationHash,
						},
						"function_association": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 2,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.EventType_Values(), false),
									},
									"function_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"max_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"min_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"origin_request_policy_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"realtime_log_config_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"response_headers_policy_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"smooth_streaming": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"target_origin_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"trusted_key_groups": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"trusted_signers": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"viewer_protocol_policy": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(cloudfront.ViewerProtocolPolicy_Values(), false),
						},
					},
				},
			},
			"default_root_object": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"http_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      cloudfront.HttpVersionHttp2,
				ValidateFunc: validation.StringInSlice(cloudfront.HttpVersion_Values(), false),
			},
			"logging_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"include_cookies": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
					},
				},
			},
			"origin_group": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      OriginGroupHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"origin_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"failover_criteria": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status_codes": {
										Type:     schema.TypeSet,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeInt},
									},
								},
							},
						},
						"member": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 2,
							MaxItems: 2,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"origin_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"origin": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      OriginHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_attempts": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      3,
							ValidateFunc: validation.IntBetween(1, 3),
						},
						"connection_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntBetween(1, 10),
						},
						"custom_origin_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"http_port": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"https_port": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"origin_keepalive_timeout": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      5,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"origin_read_timeout": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      30,
										ValidateFunc: validation.IntBetween(1, 180),
									},
									"origin_protocol_policy": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.OriginProtocolPolicy_Values(), false),
									},
									"origin_ssl_protocols": {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringInSlice(cloudfront.SslProtocol_Values(), false),
										},
									},
								},
							},
						},
						"domain_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"custom_header": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      OriginCustomHeaderHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
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
						"origin_access_control_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"origin_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"origin_path": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"origin_shield": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"origin_shield_region": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringMatch(regionRegexp, "must be a valid AWS Region Code"),
									},
								},
							},
						},
						"s3_origin_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"origin_access_identity": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"price_class": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      cloudfront.PriceClassPriceClassAll,
				ValidateFunc: validation.StringInSlice(cloudfront.PriceClass_Values(), false),
			},
			"restrictions": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"geo_restriction": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"locations": {
										Type:     schema.TypeSet,
										Optional: true,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"restriction_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.GeoRestrictionType_Values(), false),
									},
								},
							},
						},
					},
				},
			},
			"viewer_certificate": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"acm_certificate_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"cloudfront_default_certificate": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"iam_certificate_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"minimum_protocol_version": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      cloudfront.MinimumProtocolVersionTlsv1,
							ValidateFunc: validation.StringInSlice(cloudfront.MinimumProtocolVersion_Values(), false),
						},
						"ssl_support_method": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(cloudfront.SSLSupportMethod_Values(), false),
						},
					},
				},
			},
			"web_acl_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"trusted_key_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"items": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key_group_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"key_pair_ids": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			// Terraform AWS Provider 3.0 name change:
			// enables TF Plugin SDK to ignore pre-existing attribute state
			// associated with previous naming i.e. active_trusted_signers
			"trusted_signers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"items": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"aws_account_number": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"key_pair_ids": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"in_progress_validation_batches": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// retain_on_delete is a non-API attribute that may help facilitate speedy
			// deletion of a resoruce. It's mainly here for testing purposes, so
			// enable at your own risk.
			"retain_on_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"wait_for_deployment": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"is_ipv6_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"staging": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDistributionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)

	input := &cloudfront.CreateDistributionWithTagsInput{
		DistributionConfigWithTags: &cloudfront.DistributionConfigWithTags{
			DistributionConfig: expandDistributionConfig(d),
			Tags:               &cloudfront.Tags{Items: []*cloudfront.Tag{}},
		},
	}

	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.DistributionConfigWithTags.Tags.Items = tags
	}

	// ACM and IAM certificate eventual consistency.
	// InvalidViewerCertificate: The specified SSL certificate doesn't exist, isn't in us-east-1 region, isn't valid, or doesn't include a valid certificate chain.
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 1*time.Minute, func() (interface{}, error) {
		return conn.CreateDistributionWithTagsWithContext(ctx, input)
	}, cloudfront.ErrCodeInvalidViewerCertificate)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Distribution: %s", err)
	}

	d.SetId(aws.StringValue(outputRaw.(*cloudfront.CreateDistributionWithTagsOutput).Distribution.Id))

	if d.Get("wait_for_deployment").(bool) {
		log.Printf("[DEBUG] Waiting until CloudFront Distribution (%s) is deployed", d.Id())
		if err := WaitDistributionDeployed(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting until CloudFront Distribution (%s) is deployed: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDistributionRead(ctx, d, meta)...)
}

func resourceDistributionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)

	output, err := FindDistributionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.CloudFront, create.ErrActionReading, ResNameDistribution, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CloudFront, create.ErrActionReading, ResNameDistribution, d.Id(), err)
	}

	// Update attributes from DistributionConfig
	err = flattenDistributionConfig(d, output.Distribution.DistributionConfig)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Distribution (%s): %s", d.Id(), err)
	}

	// Update other attributes outside of DistributionConfig
	if err := d.Set("trusted_key_groups", flattenActiveTrustedKeyGroups(output.Distribution.ActiveTrustedKeyGroups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting trusted_key_groups: %s", err)
	}
	if err := d.Set("trusted_signers", flattenActiveTrustedSigners(output.Distribution.ActiveTrustedSigners)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting trusted_signers: %s", err)
	}
	d.Set("status", output.Distribution.Status)
	d.Set("domain_name", output.Distribution.DomainName)
	d.Set("last_modified_time", aws.String(output.Distribution.LastModifiedTime.String()))
	d.Set("in_progress_validation_batches", output.Distribution.InProgressInvalidationBatches)
	d.Set("etag", output.ETag)
	d.Set("arn", output.Distribution.ARN)
	d.Set("hosted_zone_id", meta.(*conns.AWSClient).CloudFrontDistributionHostedZoneID())

	return diags
}

func resourceDistributionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &cloudfront.UpdateDistributionInput{
			Id:                 aws.String(d.Id()),
			DistributionConfig: expandDistributionConfig(d),
			IfMatch:            aws.String(d.Get("etag").(string)),
		}

		// ACM and IAM certificate eventual consistency.
		// InvalidViewerCertificate: The specified SSL certificate doesn't exist, isn't in us-east-1 region, isn't valid, or doesn't include a valid certificate chain.
		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 1*time.Minute, func() (interface{}, error) {
			return conn.UpdateDistributionWithContext(ctx, input)
		}, cloudfront.ErrCodeInvalidViewerCertificate)

		// Refresh our ETag if it is out of date and attempt update again.
		if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodePreconditionFailed) {
			getDistributionInput := &cloudfront.GetDistributionInput{
				Id: aws.String(d.Id()),
			}
			var getDistributionOutput *cloudfront.GetDistributionOutput

			log.Printf("[DEBUG] Refreshing CloudFront Distribution (%s) ETag", d.Id())
			getDistributionOutput, err = conn.GetDistributionWithContext(ctx, getDistributionInput)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "refreshing CloudFront Distribution (%s) ETag: %s", d.Id(), err)
			}

			if getDistributionOutput == nil {
				return sdkdiag.AppendErrorf(diags, "refreshing CloudFront Distribution (%s) ETag: empty response", d.Id())
			}

			input.IfMatch = getDistributionOutput.ETag

			_, err = conn.UpdateDistributionWithContext(ctx, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudFront Distribution (%s): %s", d.Id(), err)
		}

		if d.Get("wait_for_deployment").(bool) {
			log.Printf("[DEBUG] Waiting until CloudFront Distribution (%s) is deployed", d.Id())
			if err := WaitDistributionDeployed(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting until CloudFront Distribution (%s) is deployed: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceDistributionRead(ctx, d, meta)...)
}

func resourceDistributionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)

	if d.Get("arn").(string) == "" {
		diags = append(diags, resourceDistributionRead(ctx, d, meta)...)
	}

	if v := d.Get("continuous_deployment_policy_id").(string); v != "" {
		if err := disableContinuousDeploymentPolicy(ctx, conn, v); err != nil {
			return create.AppendDiagError(diags, names.CloudFront, create.ErrActionDeleting, ResNameDistribution, d.Id(), err)
		}

		if err := WaitDistributionDeployed(ctx, conn, d.Id()); err != nil && !tfresource.NotFound(err) {
			return sdkdiag.AppendErrorf(diags, "waiting until CloudFront Distribution (%s) is deployed: %s", d.Id(), err)
		}
	}

	if err := disableDistribution(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling CloudFront Distribution (%s): %s", d.Id(), err)
	}

	if d.Get("retain_on_delete").(bool) {
		log.Printf("[WARN] Removing CloudFront Distribution ID %q with `retain_on_delete` set. Please delete this distribution manually.", d.Id())
		return diags
	}

	err := deleteDistribution(ctx, conn, d.Id())
	if err == nil || tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchDistribution) {
		return diags
	}

	// Disable distribution if it is not yet disabled and attempt deletion again.
	// Here we update via the deployed configuration to ensure we are not submitting an out of date
	// configuration from the Terraform configuration, should other changes have occurred manually.
	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeDistributionNotDisabled) {
		if err = disableDistribution(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling CloudFront Distribution (%s): %s", d.Id(), err)
		}

		_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, 3*time.Minute, func() (interface{}, error) {
			return nil, deleteDistribution(ctx, conn, d.Id())
		}, cloudfront.ErrCodeDistributionNotDisabled)
	}

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodePreconditionFailed, cloudfront.ErrCodeInvalidIfMatchVersion) {
		_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, 1*time.Minute, func() (interface{}, error) {
			return nil, deleteDistribution(ctx, conn, d.Id())
		}, cloudfront.ErrCodePreconditionFailed, cloudfront.ErrCodeInvalidIfMatchVersion)
	}

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchDistribution) {
		return diags
	}

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeDistributionNotDisabled) {
		if err = disableDistribution(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling CloudFront Distribution (%s): %s", d.Id(), err)
		}

		err = deleteDistribution(ctx, conn, d.Id())
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Distribution (%s): %s", d.Id(), err)
	}

	return diags
}

func deleteDistribution(ctx context.Context, conn *cloudfront.CloudFront, id string) error {
	etag, err := distroETag(ctx, conn, id)
	if err != nil {
		return err
	}

	in := &cloudfront.DeleteDistributionInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}

	if _, err := conn.DeleteDistributionWithContext(ctx, in); err != nil {
		return err
	}

	if err := WaitDistributionDeleted(ctx, conn, id); err != nil {
		return err
	}

	return nil
}

func distroETag(ctx context.Context, conn *cloudfront.CloudFront, id string) (string, error) {
	output, err := FindDistributionByID(ctx, conn, id)
	if err != nil {
		return "", err
	}

	return aws.StringValue(output.ETag), nil
}

func disableDistribution(ctx context.Context, conn *cloudfront.CloudFront, id string) error {
	out, err := FindDistributionByID(ctx, conn, id)
	if err != nil {
		return err
	}

	if aws.StringValue(out.Distribution.Status) == "InProgress" {
		if err := WaitDistributionDeployed(ctx, conn, id); err != nil {
			return err
		}

		out, err = FindDistributionByID(ctx, conn, id)
		if err != nil {
			return err
		}
	}

	if !aws.BoolValue(out.Distribution.DistributionConfig.Enabled) {
		return nil
	}

	in := &cloudfront.UpdateDistributionInput{
		DistributionConfig: out.Distribution.DistributionConfig,
		Id:                 aws.String(id),
		IfMatch:            out.ETag,
	}
	in.DistributionConfig.Enabled = aws.Bool(false)

	_, err = conn.UpdateDistributionWithContext(ctx, in)
	if err != nil {
		return err
	}

	if err := WaitDistributionDeployed(ctx, conn, id); err != nil {
		return err
	}

	return nil
}

func FindDistributionByID(ctx context.Context, conn *cloudfront.CloudFront, id string) (*cloudfront.GetDistributionOutput, error) {
	input := &cloudfront.GetDistributionInput{
		Id: aws.String(id),
	}

	output, err := conn.GetDistributionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchDistribution) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Distribution == nil || output.Distribution.DistributionConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// resourceAwsCloudFrontWebDistributionWaitUntilDeployed blocks until the
// distribution is deployed. It currently takes exactly 15 minutes to deploy
// but that might change in the future.
func WaitDistributionDeployed(ctx context.Context, conn *cloudfront.CloudFront, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"InProgress"},
		Target:     []string{"Deployed"},
		Refresh:    distributionDeployRefreshFunc(ctx, conn, id),
		Timeout:    90 * time.Minute,
		MinTimeout: 15 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func WaitDistributionDeleted(ctx context.Context, conn *cloudfront.CloudFront, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"InProgress", "Deployed"},
		Target:     []string{},
		Refresh:    distributionDeleteRefreshFunc(ctx, conn, id),
		Timeout:    90 * time.Minute,
		MinTimeout: 15 * time.Second,
		Delay:      15 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func distributionDeleteRefreshFunc(ctx context.Context, conn *cloudfront.CloudFront, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindDistributionByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if out == nil {
			return nil, "", nil
		}

		return out.Distribution, aws.StringValue(out.Distribution.Status), nil
	}
}

func distributionDeployRefreshFunc(ctx context.Context, conn *cloudfront.CloudFront, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindDistributionByID(ctx, conn, id)
		if err != nil {
			return nil, "", err
		}

		if out == nil {
			return nil, "", nil
		}

		return out.Distribution, aws.StringValue(out.Distribution.Status), nil
	}
}
