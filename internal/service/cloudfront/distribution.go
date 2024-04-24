// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_distribution", name="Distribution")
// @Tags(identifierAttribute="arn")
func resourceDistribution() *schema.Resource {
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
			},
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"comment": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"continuous_deployment_policy_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"custom_error_response": {
				Type:     schema.TypeSet,
				Optional: true,
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
						"cache_policy_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cached_methods": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
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
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.ItemSelection](),
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
						"function_association": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 2,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_type": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.EventType](),
									},
									"function_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
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
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.EventType](),
									},
									"include_body": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"lambda_arn": {
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ViewerProtocolPolicy](),
						},
					},
				},
			},
			"default_root_object": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"http_version": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.HttpVersionHttp2,
				ValidateDiagFunc: enum.Validate[awstypes.HttpVersion](),
			},
			"in_progress_validation_batches": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"is_ipv6_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
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
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.ItemSelection](),
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
						"function_association": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 2,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_type": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.EventType](),
									},
									"function_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
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
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.EventType](),
									},
									"include_body": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"lambda_arn": {
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ViewerProtocolPolicy](),
						},
					},
				},
			},
			"origin_group": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"origin_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},
			"origin": {
				Type:     schema.TypeSet,
				Required: true,
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
						"custom_header": {
							Type:     schema.TypeSet,
							Optional: true,
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
										ValidateFunc: validation.IntAtLeast(1),
									},
									"origin_protocol_policy": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.OriginProtocolPolicy](),
									},
									"origin_ssl_protocols": {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Schema{
											Type:             schema.TypeString,
											ValidateDiagFunc: enum.Validate[awstypes.SslProtocol](),
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
										ValidateFunc: verify.ValidRegionName,
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.PriceClassPriceClassAll,
				ValidateDiagFunc: enum.Validate[awstypes.PriceClass](),
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
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.GeoRestrictionType](),
									},
								},
							},
						},
					},
				},
			},
			// retain_on_delete is a non-API attribute that may help facilitate speedy
			// deletion of a resoruce. It's mainly here for testing purposes, so
			// enable at your own risk.
			"retain_on_delete": {
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.MinimumProtocolVersionTLSv1,
							ValidateDiagFunc: enum.Validate[awstypes.MinimumProtocolVersion](),
						},
						"ssl_support_method": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SSLSupportMethod](),
						},
					},
				},
			},
			"wait_for_deployment": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"web_acl_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDistributionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	input := &cloudfront.CreateDistributionWithTagsInput{
		DistributionConfigWithTags: &awstypes.DistributionConfigWithTags{
			DistributionConfig: expandDistributionConfig(d),
			Tags:               &awstypes.Tags{Items: []awstypes.Tag{}},
		},
	}

	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.DistributionConfigWithTags.Tags.Items = tags
	}

	// ACM and IAM certificate eventual consistency.
	// InvalidViewerCertificate: The specified SSL certificate doesn't exist, isn't in us-east-1 region, isn't valid, or doesn't include a valid certificate chain.
	const (
		timeout = 1 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.InvalidViewerCertificate](ctx, timeout, func() (interface{}, error) {
		return conn.CreateDistributionWithTags(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Distribution: %s", err)
	}

	d.SetId(aws.ToString(outputRaw.(*cloudfront.CreateDistributionWithTagsOutput).Distribution.Id))

	if d.Get("wait_for_deployment").(bool) {
		if _, err := waitDistributionDeployed(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for CloudFront Distribution (%s) deploy: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDistributionRead(ctx, d, meta)...)
}

func resourceDistributionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	output, err := findDistributionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Distribution (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Distribution (%s): %s", d.Id(), err)
	}

	distributionConfig := output.Distribution.DistributionConfig
	if distributionConfig.Aliases != nil {
		if err := d.Set("aliases", FlattenAliases(distributionConfig.Aliases)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting aliases: %s", err)
		}
	}
	d.Set("arn", output.Distribution.ARN)
	d.Set("caller_reference", distributionConfig.CallerReference)
	if aws.ToString(distributionConfig.Comment) != "" {
		d.Set("comment", distributionConfig.Comment)
	}
	// Not having this set for staging distributions causes IllegalUpdate errors when making updates of any kind.
	// If this absolutely must not be optional/computed, the policy ID will need to be retrieved and set for each
	// API call for staging distributions.
	d.Set("continuous_deployment_policy_id", distributionConfig.ContinuousDeploymentPolicyId)
	if distributionConfig.CustomErrorResponses != nil {
		if err := d.Set("custom_error_response", FlattenCustomErrorResponses(distributionConfig.CustomErrorResponses)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting custom_error_response: %s", err)
		}
	}
	if err := d.Set("default_cache_behavior", []interface{}{flattenDefaultCacheBehavior(distributionConfig.DefaultCacheBehavior)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_cache_behavior: %s", err)
	}
	d.Set("default_root_object", distributionConfig.DefaultRootObject)
	d.Set("domain_name", output.Distribution.DomainName)
	d.Set("enabled", distributionConfig.Enabled)
	d.Set("etag", output.ETag)
	d.Set("http_version", distributionConfig.HttpVersion)
	d.Set("hosted_zone_id", meta.(*conns.AWSClient).CloudFrontDistributionHostedZoneID(ctx))
	d.Set("in_progress_validation_batches", output.Distribution.InProgressInvalidationBatches)
	d.Set("is_ipv6_enabled", distributionConfig.IsIPV6Enabled)
	d.Set("last_modified_time", aws.String(output.Distribution.LastModifiedTime.String()))
	if distributionConfig.Logging != nil && aws.ToBool(distributionConfig.Logging.Enabled) {
		if err := d.Set("logging_config", flattenLoggingConfig(distributionConfig.Logging)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting logging_config: %s", err)
		}
	} else {
		err = d.Set("logging_config", []interface{}{})
	}
	if distributionConfig.CacheBehaviors != nil {
		if err := d.Set("ordered_cache_behavior", flattenCacheBehaviors(distributionConfig.CacheBehaviors)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ordered_cache_behavior: %s", err)
		}
	}
	if aws.ToInt32(distributionConfig.Origins.Quantity) > 0 {
		if err := d.Set("origin", FlattenOrigins(distributionConfig.Origins)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting origin: %s", err)
		}
	}
	if aws.ToInt32(distributionConfig.OriginGroups.Quantity) > 0 {
		if err := d.Set("origin_group", FlattenOriginGroups(distributionConfig.OriginGroups)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting origin_group: %s", err)
		}
	}
	d.Set("price_class", distributionConfig.PriceClass)
	if distributionConfig.Restrictions != nil {
		if err := d.Set("restrictions", flattenRestrictions(distributionConfig.Restrictions)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting restrictions: %s", err)
		}
	}
	d.Set("staging", distributionConfig.Staging)
	d.Set("status", output.Distribution.Status)
	if err := d.Set("trusted_key_groups", flattenActiveTrustedKeyGroups(output.Distribution.ActiveTrustedKeyGroups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting trusted_key_groups: %s", err)
	}
	if err := d.Set("trusted_signers", flattenActiveTrustedSigners(output.Distribution.ActiveTrustedSigners)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting trusted_signers: %s", err)
	}
	if err := d.Set("viewer_certificate", flattenViewerCertificate(distributionConfig.ViewerCertificate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting viewer_certificate: %s", err)
	}
	d.Set("web_acl_id", distributionConfig.WebACLId)

	return diags
}

func resourceDistributionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &cloudfront.UpdateDistributionInput{
			DistributionConfig: expandDistributionConfig(d),
			Id:                 aws.String(d.Id()),
			IfMatch:            aws.String(d.Get("etag").(string)),
		}

		// ACM and IAM certificate eventual consistency.
		// InvalidViewerCertificate: The specified SSL certificate doesn't exist, isn't in us-east-1 region, isn't valid, or doesn't include a valid certificate chain.
		const (
			timeout = 1 * time.Minute
		)
		_, err := tfresource.RetryWhenIsA[*awstypes.InvalidViewerCertificate](ctx, timeout, func() (interface{}, error) {
			return conn.UpdateDistribution(ctx, input)
		})

		// Refresh our ETag if it is out of date and attempt update again.
		if errs.IsA[*awstypes.PreconditionFailed](err) {
			etag, err := distroETag(ctx, conn, d.Id())

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			input.IfMatch = aws.String(etag)

			_, err = conn.UpdateDistribution(ctx, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudFront Distribution (%s): %s", d.Id(), err)
		}

		if d.Get("wait_for_deployment").(bool) {
			if _, err := waitDistributionDeployed(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for CloudFront Distribution (%s) deploy: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceDistributionRead(ctx, d, meta)...)
}

func resourceDistributionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	if d.Get("arn").(string) == "" {
		diags = append(diags, resourceDistributionRead(ctx, d, meta)...)
	}

	if v := d.Get("continuous_deployment_policy_id").(string); v != "" {
		if err := disableContinuousDeploymentPolicy(ctx, conn, v); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if _, err := waitDistributionDeployed(ctx, conn, d.Id()); err != nil && !tfresource.NotFound(err) {
			return sdkdiag.AppendErrorf(diags, "waiting for CloudFront Distribution (%s) deploy: %s", d.Id(), err)
		}
	}

	if err := disableDistribution(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.Get("retain_on_delete").(bool) {
		log.Printf("[WARN] Removing CloudFront Distribution ID %q with `retain_on_delete` set. Please delete this distribution manually.", d.Id())
		return diags
	}

	err := deleteDistribution(ctx, conn, d.Id())

	if err == nil || tfresource.NotFound(err) || errs.IsA[*awstypes.NoSuchDistribution](err) {
		return diags
	}

	// Disable distribution if it is not yet disabled and attempt deletion again.
	// Here we update via the deployed configuration to ensure we are not submitting an out of date
	// configuration from the Terraform configuration, should other changes have occurred manually.
	if errs.IsA[*awstypes.DistributionNotDisabled](err) {
		if err = disableDistribution(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		const (
			timeout = 3 * time.Minute
		)
		_, err = tfresource.RetryWhenIsA[*awstypes.DistributionNotDisabled](ctx, timeout, func() (interface{}, error) {
			return nil, deleteDistribution(ctx, conn, d.Id())
		})
	}

	if errs.IsA[*awstypes.PreconditionFailed](err) || errs.IsA[*awstypes.InvalidIfMatchVersion](err) {
		const (
			timeout = 1 * time.Minute
		)
		_, err = tfresource.RetryWhenIsOneOf[*awstypes.PreconditionFailed, *awstypes.InvalidIfMatchVersion](ctx, timeout, func() (interface{}, error) {
			return nil, deleteDistribution(ctx, conn, d.Id())
		})
	}

	if errs.IsA[*awstypes.NoSuchDistribution](err) {
		return diags
	}

	if errs.IsA[*awstypes.DistributionNotDisabled](err) {
		if err = disableDistribution(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		err = deleteDistribution(ctx, conn, d.Id())
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func deleteDistribution(ctx context.Context, conn *cloudfront.Client, id string) error {
	etag, err := distroETag(ctx, conn, id)

	if err != nil {
		return err
	}

	input := &cloudfront.DeleteDistributionInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}

	_, err = conn.DeleteDistribution(ctx, input)

	if err != nil {
		return fmt.Errorf("deleting CloudFront Distribution (%s): %w", id, err)
	}

	if _, err := waitDistributionDeleted(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for CloudFront Distribution (%s) delete: %w", id, err)
	}

	return nil
}

func distroETag(ctx context.Context, conn *cloudfront.Client, id string) (string, error) {
	output, err := findDistributionByID(ctx, conn, id)

	if err != nil {
		return "", fmt.Errorf("reading CloudFront Distribution (%s): %w", id, err)
	}

	return aws.ToString(output.ETag), nil
}

func disableDistribution(ctx context.Context, conn *cloudfront.Client, id string) error {
	output, err := findDistributionByID(ctx, conn, id)

	if err != nil {
		return fmt.Errorf("reading CloudFront Distribution (%s): %w", id, err)
	}

	if aws.ToString(output.Distribution.Status) == distributionStatusInProgress {
		output, err = waitDistributionDeployed(ctx, conn, id)

		if err != nil {
			return fmt.Errorf("waiting for CloudFront Distribution (%s) deploy: %w", id, err)
		}
	}

	if !aws.ToBool(output.Distribution.DistributionConfig.Enabled) {
		return nil
	}

	input := &cloudfront.UpdateDistributionInput{
		DistributionConfig: output.Distribution.DistributionConfig,
		Id:                 aws.String(id),
		IfMatch:            output.ETag,
	}
	input.DistributionConfig.Enabled = aws.Bool(false)

	_, err = conn.UpdateDistribution(ctx, input)

	if err != nil {
		return fmt.Errorf("updating CloudFront Distribution (%s): %w", id, err)
	}

	if _, err := waitDistributionDeployed(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for CloudFront Distribution (%s) deploy: %w", id, err)
	}

	return nil
}

func findDistributionByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetDistributionOutput, error) {
	input := &cloudfront.GetDistributionInput{
		Id: aws.String(id),
	}

	output, err := conn.GetDistribution(ctx, input)

	if errs.IsA[*awstypes.NoSuchDistribution](err) {
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

func statusDistribution(ctx context.Context, conn *cloudfront.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDistributionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output.Distribution, aws.ToString(output.Distribution.Status), nil
	}
}

func waitDistributionDeployed(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetDistributionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{distributionStatusInProgress},
		Target:     []string{distributionStatusDeployed},
		Refresh:    statusDistribution(ctx, conn, id),
		Timeout:    90 * time.Minute,
		MinTimeout: 15 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.GetDistributionOutput); ok {
		return output, err
	}

	return nil, err
}

func waitDistributionDeleted(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetDistributionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{distributionStatusInProgress, distributionStatusDeployed},
		Target:     []string{},
		Refresh:    statusDistribution(ctx, conn, id),
		Timeout:    90 * time.Minute,
		MinTimeout: 15 * time.Second,
		Delay:      15 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.GetDistributionOutput); ok {
		return output, err
	}

	return nil, err
}
