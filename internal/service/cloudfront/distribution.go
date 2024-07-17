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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
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
			names.AttrARN: {
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
			names.AttrComment: {
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
									names.AttrFunctionARN: {
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
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Required: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrHostedZoneID: {
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
						names.AttrBucket: {
							Type:     schema.TypeString,
							Required: true,
						},
						"include_cookies": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						names.AttrPrefix: {
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
									names.AttrFunctionARN: {
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
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrValue: {
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
						names.AttrDomainName: {
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
									names.AttrEnabled: {
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
			names.AttrStatus: {
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
						names.AttrEnabled: {
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
						names.AttrEnabled: {
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
		if err := d.Set("aliases", flattenAliases(distributionConfig.Aliases)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting aliases: %s", err)
		}
	}
	d.Set(names.AttrARN, output.Distribution.ARN)
	d.Set("caller_reference", distributionConfig.CallerReference)
	if aws.ToString(distributionConfig.Comment) != "" {
		d.Set(names.AttrComment, distributionConfig.Comment)
	}
	// Not having this set for staging distributions causes IllegalUpdate errors when making updates of any kind.
	// If this absolutely must not be optional/computed, the policy ID will need to be retrieved and set for each
	// API call for staging distributions.
	d.Set("continuous_deployment_policy_id", distributionConfig.ContinuousDeploymentPolicyId)
	if distributionConfig.CustomErrorResponses != nil {
		if err := d.Set("custom_error_response", flattenCustomErrorResponses(distributionConfig.CustomErrorResponses)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting custom_error_response: %s", err)
		}
	}
	if err := d.Set("default_cache_behavior", []interface{}{flattenDefaultCacheBehavior(distributionConfig.DefaultCacheBehavior)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_cache_behavior: %s", err)
	}
	d.Set("default_root_object", distributionConfig.DefaultRootObject)
	d.Set(names.AttrDomainName, output.Distribution.DomainName)
	d.Set(names.AttrEnabled, distributionConfig.Enabled)
	d.Set("etag", output.ETag)
	d.Set("http_version", distributionConfig.HttpVersion)
	d.Set(names.AttrHostedZoneID, meta.(*conns.AWSClient).CloudFrontDistributionHostedZoneID(ctx))
	d.Set("in_progress_validation_batches", output.Distribution.InProgressInvalidationBatches)
	d.Set("is_ipv6_enabled", distributionConfig.IsIPV6Enabled)
	d.Set("last_modified_time", aws.String(output.Distribution.LastModifiedTime.String()))
	if distributionConfig.Logging != nil && aws.ToBool(distributionConfig.Logging.Enabled) {
		if err := d.Set("logging_config", flattenLoggingConfig(distributionConfig.Logging)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting logging_config: %s", err)
		}
	} else {
		d.Set("logging_config", []interface{}{})
	}
	if distributionConfig.CacheBehaviors != nil {
		if err := d.Set("ordered_cache_behavior", flattenCacheBehaviors(distributionConfig.CacheBehaviors)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ordered_cache_behavior: %s", err)
		}
	}
	if aws.ToInt32(distributionConfig.Origins.Quantity) > 0 {
		if err := d.Set("origin", flattenOrigins(distributionConfig.Origins)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting origin: %s", err)
		}
	}
	if aws.ToInt32(distributionConfig.OriginGroups.Quantity) > 0 {
		if err := d.Set("origin_group", flattenOriginGroups(distributionConfig.OriginGroups)); err != nil {
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
	d.Set(names.AttrStatus, output.Distribution.Status)
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

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
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
			var etag string
			etag, err = distroETag(ctx, conn, d.Id())

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

	if d.Get(names.AttrARN).(string) == "" {
		diags = append(diags, resourceDistributionRead(ctx, d, meta)...)
	}

	if v := d.Get("continuous_deployment_policy_id").(string); v != "" {
		err := disableContinuousDeploymentPolicy(ctx, conn, v)

		switch {
		case tfresource.NotFound(err):
		case err != nil:
			return sdkdiag.AppendFromErr(diags, err)
		default:
			if _, err := waitDistributionDeployed(ctx, conn, d.Id()); err != nil && !tfresource.NotFound(err) {
				return sdkdiag.AppendErrorf(diags, "waiting for CloudFront Distribution (%s) deploy: %s", d.Id(), err)
			}
		}
	}

	if err := disableDistribution(ctx, conn, d.Id()); err != nil {
		if tfresource.NotFound(err) {
			return diags
		}

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
		if err := disableDistribution(ctx, conn, d.Id()); err != nil {
			if tfresource.NotFound(err) {
				return diags
			}

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
		_, err = tfresource.RetryWhenIsOneOf2[*awstypes.PreconditionFailed, *awstypes.InvalidIfMatchVersion](ctx, timeout, func() (interface{}, error) {
			return nil, deleteDistribution(ctx, conn, d.Id())
		})
	}

	if errs.IsA[*awstypes.DistributionNotDisabled](err) {
		if err := disableDistribution(ctx, conn, d.Id()); err != nil {
			if tfresource.NotFound(err) {
				return diags
			}

			return sdkdiag.AppendFromErr(diags, err)
		}

		err = deleteDistribution(ctx, conn, d.Id())
	}

	if errs.IsA[*awstypes.NoSuchDistribution](err) { // nosemgrep:dgryski.semgrep-go.oddifsequence.odd-sequence-ifs
		return diags
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

		return output, aws.ToString(output.Distribution.Status), nil
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

func expandDistributionConfig(d *schema.ResourceData) *awstypes.DistributionConfig {
	apiObject := &awstypes.DistributionConfig{
		CacheBehaviors:               expandCacheBehaviors(d.Get("ordered_cache_behavior").([]interface{})),
		CallerReference:              aws.String(id.UniqueId()),
		Comment:                      aws.String(d.Get(names.AttrComment).(string)),
		ContinuousDeploymentPolicyId: aws.String(d.Get("continuous_deployment_policy_id").(string)),
		CustomErrorResponses:         expandCustomErrorResponses(d.Get("custom_error_response").(*schema.Set).List()),
		DefaultCacheBehavior:         expandDefaultCacheBehavior(d.Get("default_cache_behavior").([]interface{})[0].(map[string]interface{})),
		DefaultRootObject:            aws.String(d.Get("default_root_object").(string)),
		Enabled:                      aws.Bool(d.Get(names.AttrEnabled).(bool)),
		IsIPV6Enabled:                aws.Bool(d.Get("is_ipv6_enabled").(bool)),
		HttpVersion:                  awstypes.HttpVersion(d.Get("http_version").(string)),
		Origins:                      expandOrigins(d.Get("origin").(*schema.Set).List()),
		PriceClass:                   awstypes.PriceClass(d.Get("price_class").(string)),
		Staging:                      aws.Bool(d.Get("staging").(bool)),
		WebACLId:                     aws.String(d.Get("web_acl_id").(string)),
	}

	if v, ok := d.GetOk("aliases"); ok {
		apiObject.Aliases = expandAliases(v.(*schema.Set).List())
	} else {
		apiObject.Aliases = expandAliases([]interface{}{})
	}

	if v, ok := d.GetOk("caller_reference"); ok {
		apiObject.CallerReference = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.Logging = expandLoggingConfig(v.([]interface{})[0].(map[string]interface{}))
	} else {
		apiObject.Logging = expandLoggingConfig(nil)
	}

	if v, ok := d.GetOk("origin_group"); ok {
		apiObject.OriginGroups = expandOriginGroups(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("restrictions"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.Restrictions = expandRestrictions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("viewer_certificate"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.ViewerCertificate = expandViewerCertificate(v.([]interface{})[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCacheBehavior(tfMap map[string]interface{}) *awstypes.CacheBehavior {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CacheBehavior{
		CachePolicyId:           aws.String(tfMap["cache_policy_id"].(string)),
		Compress:                aws.Bool(tfMap["compress"].(bool)),
		FieldLevelEncryptionId:  aws.String(tfMap["field_level_encryption_id"].(string)),
		OriginRequestPolicyId:   aws.String(tfMap["origin_request_policy_id"].(string)),
		ResponseHeadersPolicyId: aws.String(tfMap["response_headers_policy_id"].(string)),
		TargetOriginId:          aws.String(tfMap["target_origin_id"].(string)),
		ViewerProtocolPolicy:    awstypes.ViewerProtocolPolicy(tfMap["viewer_protocol_policy"].(string)),
	}

	if v, ok := tfMap["allowed_methods"]; ok {
		apiObject.AllowedMethods = expandAllowedMethods(v.(*schema.Set).List())
	}

	if tfMap["cache_policy_id"].(string) == "" {
		apiObject.DefaultTTL = aws.Int64(int64(tfMap["default_ttl"].(int)))
		apiObject.MaxTTL = aws.Int64(int64(tfMap["max_ttl"].(int)))
		apiObject.MinTTL = aws.Int64(int64(tfMap["min_ttl"].(int)))
	}

	if v, ok := tfMap["cached_methods"]; ok {
		apiObject.AllowedMethods.CachedMethods = expandCachedMethods(v.(*schema.Set).List())
	}

	if v, ok := tfMap["forwarded_values"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ForwardedValues = expandForwardedValues(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["function_association"]; ok {
		apiObject.FunctionAssociations = expandFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := tfMap["lambda_function_association"]; ok {
		apiObject.LambdaFunctionAssociations = expandLambdaFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := tfMap["path_pattern"]; ok {
		apiObject.PathPattern = aws.String(v.(string))
	}

	if v, ok := tfMap["realtime_log_config_arn"]; ok && v.(string) != "" {
		apiObject.RealtimeLogConfigArn = aws.String(v.(string))
	}

	if v, ok := tfMap["smooth_streaming"]; ok {
		apiObject.SmoothStreaming = aws.Bool(v.(bool))
	}

	if v, ok := tfMap["trusted_key_groups"]; ok {
		apiObject.TrustedKeyGroups = expandTrustedKeyGroups(v.([]interface{}))
	} else {
		apiObject.TrustedKeyGroups = expandTrustedKeyGroups([]interface{}{})
	}

	if v, ok := tfMap["trusted_signers"]; ok {
		apiObject.TrustedSigners = expandTrustedSigners(v.([]interface{}))
	} else {
		apiObject.TrustedSigners = expandTrustedSigners([]interface{}{})
	}

	return apiObject
}

func expandCacheBehaviors(tfList []interface{}) *awstypes.CacheBehaviors {
	var items []awstypes.CacheBehavior

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandCacheBehavior(tfMap)

		if apiObject == nil {
			continue
		}

		items = append(items, *apiObject)
	}

	return &awstypes.CacheBehaviors{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenCacheBehavior(apiObject *awstypes.CacheBehavior) map[string]interface{} {
	tfMap := make(map[string]interface{})

	tfMap["cache_policy_id"] = aws.ToString(apiObject.CachePolicyId)
	tfMap["compress"] = aws.ToBool(apiObject.Compress)
	tfMap["field_level_encryption_id"] = aws.ToString(apiObject.FieldLevelEncryptionId)
	tfMap["viewer_protocol_policy"] = apiObject.ViewerProtocolPolicy
	tfMap["target_origin_id"] = aws.ToString(apiObject.TargetOriginId)
	tfMap["min_ttl"] = aws.ToInt64(apiObject.MinTTL)
	tfMap["origin_request_policy_id"] = aws.ToString(apiObject.OriginRequestPolicyId)
	tfMap["realtime_log_config_arn"] = aws.ToString(apiObject.RealtimeLogConfigArn)
	tfMap["response_headers_policy_id"] = aws.ToString(apiObject.ResponseHeadersPolicyId)

	if apiObject.AllowedMethods != nil {
		tfMap["allowed_methods"] = flattenAllowedMethods(apiObject.AllowedMethods)
	}

	if apiObject.AllowedMethods.CachedMethods != nil {
		tfMap["cached_methods"] = flattenCachedMethods(apiObject.AllowedMethods.CachedMethods)
	}

	if apiObject.DefaultTTL != nil {
		tfMap["default_ttl"] = aws.ToInt64(apiObject.DefaultTTL)
	}

	if apiObject.ForwardedValues != nil {
		tfMap["forwarded_values"] = []interface{}{flattenForwardedValues(apiObject.ForwardedValues)}
	}

	if len(apiObject.FunctionAssociations.Items) > 0 {
		tfMap["function_association"] = flattenFunctionAssociations(apiObject.FunctionAssociations)
	}

	if len(apiObject.LambdaFunctionAssociations.Items) > 0 {
		tfMap["lambda_function_association"] = flattenLambdaFunctionAssociations(apiObject.LambdaFunctionAssociations)
	}

	if apiObject.MaxTTL != nil {
		tfMap["max_ttl"] = aws.ToInt64(apiObject.MaxTTL)
	}

	if apiObject.PathPattern != nil {
		tfMap["path_pattern"] = aws.ToString(apiObject.PathPattern)
	}

	if apiObject.SmoothStreaming != nil {
		tfMap["smooth_streaming"] = aws.ToBool(apiObject.SmoothStreaming)
	}

	if len(apiObject.TrustedKeyGroups.Items) > 0 {
		tfMap["trusted_key_groups"] = flattenTrustedKeyGroups(apiObject.TrustedKeyGroups)
	}

	if len(apiObject.TrustedSigners.Items) > 0 {
		tfMap["trusted_signers"] = flattenTrustedSigners(apiObject.TrustedSigners)
	}

	return tfMap
}

func flattenCacheBehaviors(apiObject *awstypes.CacheBehaviors) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfList := []interface{}{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenCacheBehavior(&v))
	}

	return tfList
}

func expandDefaultCacheBehavior(tfMap map[string]interface{}) *awstypes.DefaultCacheBehavior {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DefaultCacheBehavior{
		CachePolicyId:           aws.String(tfMap["cache_policy_id"].(string)),
		Compress:                aws.Bool(tfMap["compress"].(bool)),
		FieldLevelEncryptionId:  aws.String(tfMap["field_level_encryption_id"].(string)),
		OriginRequestPolicyId:   aws.String(tfMap["origin_request_policy_id"].(string)),
		ResponseHeadersPolicyId: aws.String(tfMap["response_headers_policy_id"].(string)),
		TargetOriginId:          aws.String(tfMap["target_origin_id"].(string)),
		ViewerProtocolPolicy:    awstypes.ViewerProtocolPolicy(tfMap["viewer_protocol_policy"].(string)),
	}

	if v, ok := tfMap["allowed_methods"]; ok {
		apiObject.AllowedMethods = expandAllowedMethods(v.(*schema.Set).List())
	}

	if tfMap["cache_policy_id"].(string) == "" {
		apiObject.MinTTL = aws.Int64(int64(tfMap["min_ttl"].(int)))
		apiObject.MaxTTL = aws.Int64(int64(tfMap["max_ttl"].(int)))
		apiObject.DefaultTTL = aws.Int64(int64(tfMap["default_ttl"].(int)))
	}

	if v, ok := tfMap["cached_methods"]; ok {
		apiObject.AllowedMethods.CachedMethods = expandCachedMethods(v.(*schema.Set).List())
	}

	if forwardedValuesFlat, ok := tfMap["forwarded_values"].([]interface{}); ok && len(forwardedValuesFlat) == 1 {
		apiObject.ForwardedValues = expandForwardedValues(tfMap["forwarded_values"].([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["function_association"]; ok {
		apiObject.FunctionAssociations = expandFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := tfMap["lambda_function_association"]; ok {
		apiObject.LambdaFunctionAssociations = expandLambdaFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := tfMap["realtime_log_config_arn"]; ok && v.(string) != "" {
		apiObject.RealtimeLogConfigArn = aws.String(v.(string))
	}

	if v, ok := tfMap["smooth_streaming"]; ok {
		apiObject.SmoothStreaming = aws.Bool(v.(bool))
	}

	if v, ok := tfMap["trusted_key_groups"]; ok {
		apiObject.TrustedKeyGroups = expandTrustedKeyGroups(v.([]interface{}))
	} else {
		apiObject.TrustedKeyGroups = expandTrustedKeyGroups([]interface{}{})
	}

	if v, ok := tfMap["trusted_signers"]; ok {
		apiObject.TrustedSigners = expandTrustedSigners(v.([]interface{}))
	} else {
		apiObject.TrustedSigners = expandTrustedSigners([]interface{}{})
	}

	return apiObject
}

func flattenDefaultCacheBehavior(apiObject *awstypes.DefaultCacheBehavior) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"cache_policy_id":            aws.ToString(apiObject.CachePolicyId),
		"compress":                   aws.ToBool(apiObject.Compress),
		"field_level_encryption_id":  aws.ToString(apiObject.FieldLevelEncryptionId),
		"viewer_protocol_policy":     apiObject.ViewerProtocolPolicy,
		"target_origin_id":           aws.ToString(apiObject.TargetOriginId),
		"min_ttl":                    aws.ToInt64(apiObject.MinTTL),
		"origin_request_policy_id":   aws.ToString(apiObject.OriginRequestPolicyId),
		"realtime_log_config_arn":    aws.ToString(apiObject.RealtimeLogConfigArn),
		"response_headers_policy_id": aws.ToString(apiObject.ResponseHeadersPolicyId),
	}

	if apiObject.AllowedMethods != nil {
		tfMap["allowed_methods"] = flattenAllowedMethods(apiObject.AllowedMethods)
	}

	if apiObject.AllowedMethods.CachedMethods != nil {
		tfMap["cached_methods"] = flattenCachedMethods(apiObject.AllowedMethods.CachedMethods)
	}

	if apiObject.DefaultTTL != nil {
		tfMap["default_ttl"] = aws.ToInt64(apiObject.DefaultTTL)
	}

	if apiObject.ForwardedValues != nil {
		tfMap["forwarded_values"] = []interface{}{flattenForwardedValues(apiObject.ForwardedValues)}
	}

	if len(apiObject.FunctionAssociations.Items) > 0 {
		tfMap["function_association"] = flattenFunctionAssociations(apiObject.FunctionAssociations)
	}

	if len(apiObject.LambdaFunctionAssociations.Items) > 0 {
		tfMap["lambda_function_association"] = flattenLambdaFunctionAssociations(apiObject.LambdaFunctionAssociations)
	}

	if apiObject.MaxTTL != nil {
		tfMap["max_ttl"] = aws.ToInt64(apiObject.MaxTTL)
	}

	if apiObject.SmoothStreaming != nil {
		tfMap["smooth_streaming"] = aws.ToBool(apiObject.SmoothStreaming)
	}

	if len(apiObject.TrustedKeyGroups.Items) > 0 {
		tfMap["trusted_key_groups"] = flattenTrustedKeyGroups(apiObject.TrustedKeyGroups)
	}

	if len(apiObject.TrustedSigners.Items) > 0 {
		tfMap["trusted_signers"] = flattenTrustedSigners(apiObject.TrustedSigners)
	}

	return tfMap
}

func expandTrustedKeyGroups(tfList []interface{}) *awstypes.TrustedKeyGroups {
	apiObject := &awstypes.TrustedKeyGroups{}

	if len(tfList) > 0 {
		apiObject.Enabled = aws.Bool(true)
		apiObject.Items = flex.ExpandStringValueList(tfList)
		apiObject.Quantity = aws.Int32(int32(len(tfList)))
	} else {
		apiObject.Enabled = aws.Bool(false)
		apiObject.Quantity = aws.Int32(0)
	}

	return apiObject
}

func flattenTrustedKeyGroups(apiObject *awstypes.TrustedKeyGroups) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringValueList(apiObject.Items)
	}

	return []interface{}{}
}

func expandTrustedSigners(tfList []interface{}) *awstypes.TrustedSigners {
	apiObject := &awstypes.TrustedSigners{}

	if len(tfList) > 0 {
		apiObject.Enabled = aws.Bool(true)
		apiObject.Items = flex.ExpandStringValueList(tfList)
		apiObject.Quantity = aws.Int32(int32(len(tfList)))
	} else {
		apiObject.Enabled = aws.Bool(false)
		apiObject.Quantity = aws.Int32(0)
	}

	return apiObject
}

func flattenTrustedSigners(apiObject *awstypes.TrustedSigners) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringValueList(apiObject.Items)
	}

	return []interface{}{}
}

func expandLambdaFunctionAssociation(tfMap map[string]interface{}) *awstypes.LambdaFunctionAssociation {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LambdaFunctionAssociation{}

	if v, ok := tfMap["event_type"]; ok {
		apiObject.EventType = awstypes.EventType(v.(string))
	}

	if v, ok := tfMap["include_body"]; ok {
		apiObject.IncludeBody = aws.Bool(v.(bool))
	}

	if v, ok := tfMap["lambda_arn"]; ok {
		apiObject.LambdaFunctionARN = aws.String(v.(string))
	}

	return apiObject
}

func expandLambdaFunctionAssociations(v interface{}) *awstypes.LambdaFunctionAssociations {
	if v == nil {
		return &awstypes.LambdaFunctionAssociations{
			Quantity: aws.Int32(0),
		}
	}

	tfList := v.([]interface{})

	var items []awstypes.LambdaFunctionAssociation

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := expandLambdaFunctionAssociation(tfMap)

		if item == nil {
			continue
		}

		items = append(items, *item)
	}

	return &awstypes.LambdaFunctionAssociations{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func expandFunctionAssociation(tfMap map[string]interface{}) *awstypes.FunctionAssociation {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FunctionAssociation{}

	if v, ok := tfMap["event_type"]; ok {
		apiObject.EventType = awstypes.EventType(v.(string))
	}

	if v, ok := tfMap[names.AttrFunctionARN]; ok {
		apiObject.FunctionARN = aws.String(v.(string))
	}

	return apiObject
}

func expandFunctionAssociations(v interface{}) *awstypes.FunctionAssociations {
	if v == nil {
		return &awstypes.FunctionAssociations{
			Quantity: aws.Int32(0),
		}
	}

	tfList := v.([]interface{})

	var items []awstypes.FunctionAssociation

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := expandFunctionAssociation(tfMap)

		if item == nil {
			continue
		}

		items = append(items, *item)
	}

	return &awstypes.FunctionAssociations{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenLambdaFunctionAssociation(apiObject *awstypes.LambdaFunctionAssociation) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if apiObject != nil {
		tfMap["event_type"] = apiObject.EventType
		tfMap["include_body"] = aws.ToBool(apiObject.IncludeBody)
		tfMap["lambda_arn"] = aws.ToString(apiObject.LambdaFunctionARN)
	}

	return tfMap
}

func flattenLambdaFunctionAssociations(apiObject *awstypes.LambdaFunctionAssociations) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenLambdaFunctionAssociation(&v))
	}

	return tfList
}

func flattenFunctionAssociation(apiObject *awstypes.FunctionAssociation) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if apiObject != nil {
		tfMap["event_type"] = apiObject.EventType
		tfMap[names.AttrFunctionARN] = aws.ToString(apiObject.FunctionARN)
	}

	return tfMap
}

func flattenFunctionAssociations(apiObject *awstypes.FunctionAssociations) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenFunctionAssociation(&v))
	}

	return tfList
}

func expandForwardedValues(tfMap map[string]interface{}) *awstypes.ForwardedValues {
	if len(tfMap) < 1 {
		return nil
	}

	apiObject := &awstypes.ForwardedValues{
		QueryString: aws.Bool(tfMap["query_string"].(bool)),
	}

	if v, ok := tfMap["cookies"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.Cookies = expandCookiePreference(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["headers"]; ok {
		apiObject.Headers = expandForwardedValuesHeaders(v.(*schema.Set).List())
	}

	if v, ok := tfMap["query_string_cache_keys"]; ok {
		apiObject.QueryStringCacheKeys = expandQueryStringCacheKeys(v.([]interface{}))
	}

	return apiObject
}

func flattenForwardedValues(apiObject *awstypes.ForwardedValues) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	tfMap["query_string"] = aws.ToBool(apiObject.QueryString)

	if apiObject.Cookies != nil {
		tfMap["cookies"] = []interface{}{flattenCookiePreference(apiObject.Cookies)}
	}

	if apiObject.Headers != nil {
		tfMap["headers"] = flattenForwardedValuesHeaders(apiObject.Headers)
	}

	if apiObject.QueryStringCacheKeys != nil {
		tfMap["query_string_cache_keys"] = flattenQueryStringCacheKeys(apiObject.QueryStringCacheKeys)
	}

	return tfMap
}

func expandForwardedValuesHeaders(tfList []interface{}) *awstypes.Headers {
	return &awstypes.Headers{
		Items:    flex.ExpandStringValueList(tfList),
		Quantity: aws.Int32(int32(len(tfList))),
	}
}

func flattenForwardedValuesHeaders(apiObject *awstypes.Headers) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringValueList(apiObject.Items)
	}

	return []interface{}{}
}

func expandQueryStringCacheKeys(tfList []interface{}) *awstypes.QueryStringCacheKeys {
	return &awstypes.QueryStringCacheKeys{
		Items:    flex.ExpandStringValueList(tfList),
		Quantity: aws.Int32(int32(len(tfList))),
	}
}

func flattenQueryStringCacheKeys(apiObject *awstypes.QueryStringCacheKeys) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringValueList(apiObject.Items)
	}

	return []interface{}{}
}

func expandCookiePreference(tfMap map[string]interface{}) *awstypes.CookiePreference {
	apiObject := &awstypes.CookiePreference{
		Forward: awstypes.ItemSelection(tfMap["forward"].(string)),
	}

	if v, ok := tfMap["whitelisted_names"]; ok {
		apiObject.WhitelistedNames = expandCookiePreferenceCookieNames(v.(*schema.Set).List())
	}

	return apiObject
}

func flattenCookiePreference(apiObject *awstypes.CookiePreference) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	tfMap["forward"] = apiObject.Forward

	if apiObject.WhitelistedNames != nil {
		tfMap["whitelisted_names"] = flattenCookiePreferenceCookieNames(apiObject.WhitelistedNames)
	}

	return tfMap
}

func expandCookiePreferenceCookieNames(tfList []interface{}) *awstypes.CookieNames {
	return &awstypes.CookieNames{
		Items:    flex.ExpandStringValueList(tfList),
		Quantity: aws.Int32(int32(len(tfList))),
	}
}

func flattenCookiePreferenceCookieNames(apiObject *awstypes.CookieNames) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringValueList(apiObject.Items)
	}

	return []interface{}{}
}

func expandAllowedMethods(tfList []interface{}) *awstypes.AllowedMethods {
	return &awstypes.AllowedMethods{
		Items:    flex.ExpandStringyValueList[awstypes.Method](tfList),
		Quantity: aws.Int32(int32(len(tfList))),
	}
}

func flattenAllowedMethods(apiObject *awstypes.AllowedMethods) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringyValueList(apiObject.Items)
	}

	return nil
}

func expandCachedMethods(tfList []interface{}) *awstypes.CachedMethods {
	return &awstypes.CachedMethods{
		Items:    flex.ExpandStringyValueList[awstypes.Method](tfList),
		Quantity: aws.Int32(int32(len(tfList))),
	}
}

func flattenCachedMethods(apiObject *awstypes.CachedMethods) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringyValueList(apiObject.Items)
	}

	return nil
}

func expandOrigins(tfList []interface{}) *awstypes.Origins {
	var items []awstypes.Origin

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := expandOrigin(tfMap)

		if item == nil {
			continue
		}

		items = append(items, *item)
	}

	return &awstypes.Origins{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenOrigins(apiObject *awstypes.Origins) []interface{} {
	if apiObject.Items == nil {
		return nil
	}

	tfList := []interface{}{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenOrigin(&v))
	}

	return tfList
}

func expandOrigin(tfMap map[string]interface{}) *awstypes.Origin {
	apiObject := &awstypes.Origin{
		DomainName: aws.String(tfMap[names.AttrDomainName].(string)),
		Id:         aws.String(tfMap["origin_id"].(string)),
	}

	if v, ok := tfMap["connection_attempts"]; ok {
		apiObject.ConnectionAttempts = aws.Int32(int32(v.(int)))
	}

	if v, ok := tfMap["connection_timeout"]; ok {
		apiObject.ConnectionTimeout = aws.Int32(int32(v.(int)))
	}

	if v, ok := tfMap["custom_header"]; ok {
		apiObject.CustomHeaders = expandCustomHeaders(v.(*schema.Set).List())
	}

	if v, ok := tfMap["custom_origin_config"]; ok {
		if v := v.([]interface{}); len(v) > 0 {
			apiObject.CustomOriginConfig = expandCustomOriginConfig(v[0].(map[string]interface{}))
		}
	}

	if v, ok := tfMap["origin_access_control_id"]; ok {
		apiObject.OriginAccessControlId = aws.String(v.(string))
	}

	if v, ok := tfMap["origin_path"]; ok {
		apiObject.OriginPath = aws.String(v.(string))
	}

	if v, ok := tfMap["origin_shield"]; ok {
		if v := v.([]interface{}); len(v) > 0 {
			apiObject.OriginShield = expandOriginShield(v[0].(map[string]interface{}))
		}
	}

	if v, ok := tfMap["s3_origin_config"]; ok {
		if v := v.([]interface{}); len(v) > 0 {
			apiObject.S3OriginConfig = expandS3OriginConfig(v[0].(map[string]interface{}))
		}
	}

	// if both custom and s3 origin are missing, add an empty s3 origin
	// One or the other must be specified, but the S3 origin can be "empty"
	if apiObject.S3OriginConfig == nil && apiObject.CustomOriginConfig == nil {
		apiObject.S3OriginConfig = &awstypes.S3OriginConfig{
			OriginAccessIdentity: aws.String(""),
		}
	}

	return apiObject
}

func flattenOrigin(apiObject *awstypes.Origin) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})
	tfMap[names.AttrDomainName] = aws.ToString(apiObject.DomainName)
	tfMap["origin_id"] = aws.ToString(apiObject.Id)

	if apiObject.ConnectionAttempts != nil {
		tfMap["connection_attempts"] = aws.ToInt32(apiObject.ConnectionAttempts)
	}

	if apiObject.ConnectionTimeout != nil {
		tfMap["connection_timeout"] = aws.ToInt32(apiObject.ConnectionTimeout)
	}

	if apiObject.CustomHeaders != nil {
		tfMap["custom_header"] = flattenCustomHeaders(apiObject.CustomHeaders)
	}

	if apiObject.CustomOriginConfig != nil {
		tfMap["custom_origin_config"] = []interface{}{flattenCustomOriginConfig(apiObject.CustomOriginConfig)}
	}

	if apiObject.OriginAccessControlId != nil {
		tfMap["origin_access_control_id"] = aws.ToString(apiObject.OriginAccessControlId)
	}

	if apiObject.OriginPath != nil {
		tfMap["origin_path"] = aws.ToString(apiObject.OriginPath)
	}

	if apiObject.OriginShield != nil && aws.ToBool(apiObject.OriginShield.Enabled) {
		tfMap["origin_shield"] = []interface{}{flattenOriginShield(apiObject.OriginShield)}
	}

	if apiObject.S3OriginConfig != nil && aws.ToString(apiObject.S3OriginConfig.OriginAccessIdentity) != "" {
		tfMap["s3_origin_config"] = []interface{}{flattenS3OriginConfig(apiObject.S3OriginConfig)}
	}

	return tfMap
}

func expandOriginGroups(tfList []interface{}) *awstypes.OriginGroups {
	var items []awstypes.OriginGroup

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := expandOriginGroup(tfMap)

		if item == nil {
			continue
		}

		items = append(items, *item)
	}

	return &awstypes.OriginGroups{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenOriginGroups(apiObject *awstypes.OriginGroups) []interface{} {
	if apiObject.Items == nil {
		return nil
	}

	var tfList []interface{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenOriginGroup(&v))
	}

	return tfList
}

func expandOriginGroup(tfMap map[string]interface{}) *awstypes.OriginGroup {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OriginGroup{
		FailoverCriteria: expandOriginGroupFailoverCriteria(tfMap["failover_criteria"].([]interface{})[0].(map[string]interface{})),
		Id:               aws.String(tfMap["origin_id"].(string)),
		Members:          expandMembers(tfMap["member"].([]interface{})),
	}

	return apiObject
}

func flattenOriginGroup(apiObject *awstypes.OriginGroup) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})
	tfMap["origin_id"] = aws.ToString(apiObject.Id)

	if apiObject.FailoverCriteria != nil {
		tfMap["failover_criteria"] = flattenOriginGroupFailoverCriteria(apiObject.FailoverCriteria)
	}

	if apiObject.Members != nil {
		tfMap["member"] = flattenOriginGroupMembers(apiObject.Members)
	}

	return tfMap
}

func expandOriginGroupFailoverCriteria(tfMap map[string]interface{}) *awstypes.OriginGroupFailoverCriteria {
	apiObject := &awstypes.OriginGroupFailoverCriteria{}

	if v, ok := tfMap["status_codes"]; ok {
		codes := flex.ExpandInt32ValueList(v.(*schema.Set).List())

		apiObject.StatusCodes = &awstypes.StatusCodes{
			Items:    codes,
			Quantity: aws.Int32(int32(len(codes))),
		}
	}

	return apiObject
}

func flattenOriginGroupFailoverCriteria(apiObject *awstypes.OriginGroupFailoverCriteria) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.StatusCodes.Items; v != nil {
		tfMap["status_codes"] = flex.FlattenInt32ValueList(apiObject.StatusCodes.Items)
	}

	return []interface{}{tfMap}
}

func expandMembers(tfList []interface{}) *awstypes.OriginGroupMembers {
	var items []awstypes.OriginGroupMember

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := awstypes.OriginGroupMember{
			OriginId: aws.String(tfMap["origin_id"].(string)),
		}

		items = append(items, item)
	}

	return &awstypes.OriginGroupMembers{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenOriginGroupMembers(apiObject *awstypes.OriginGroupMembers) []interface{} {
	if apiObject.Items == nil {
		return nil
	}

	tfList := []interface{}{}

	for _, apiObject := range apiObject.Items {
		tfMap := map[string]interface{}{
			"origin_id": aws.ToString(apiObject.OriginId),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandCustomHeaders(tfList []interface{}) *awstypes.CustomHeaders {
	var items []awstypes.OriginCustomHeader

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := expandOriginCustomHeader(tfMap)

		if item == nil {
			continue
		}

		items = append(items, *item)
	}

	return &awstypes.CustomHeaders{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenCustomHeaders(apiObject *awstypes.CustomHeaders) []interface{} {
	if apiObject.Items == nil {
		return nil
	}

	tfList := []interface{}{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenOriginCustomHeader(&v))
	}

	return tfList
}

func expandOriginCustomHeader(tfMap map[string]interface{}) *awstypes.OriginCustomHeader {
	if tfMap == nil {
		return nil
	}

	return &awstypes.OriginCustomHeader{
		HeaderName:  aws.String(tfMap[names.AttrName].(string)),
		HeaderValue: aws.String(tfMap[names.AttrValue].(string)),
	}
}

func flattenOriginCustomHeader(apiObject *awstypes.OriginCustomHeader) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	return map[string]interface{}{
		names.AttrName:  aws.ToString(apiObject.HeaderName),
		names.AttrValue: aws.ToString(apiObject.HeaderValue),
	}
}

func expandCustomOriginConfig(tfMap map[string]interface{}) *awstypes.CustomOriginConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CustomOriginConfig{
		HTTPPort:               aws.Int32(int32(tfMap["http_port"].(int))),
		HTTPSPort:              aws.Int32(int32(tfMap["https_port"].(int))),
		OriginKeepaliveTimeout: aws.Int32(int32(tfMap["origin_keepalive_timeout"].(int))),
		OriginProtocolPolicy:   awstypes.OriginProtocolPolicy(tfMap["origin_protocol_policy"].(string)),
		OriginReadTimeout:      aws.Int32(int32(tfMap["origin_read_timeout"].(int))),
		OriginSslProtocols:     expandCustomOriginConfigSSL(tfMap["origin_ssl_protocols"].(*schema.Set).List()),
	}

	return apiObject
}

func flattenCustomOriginConfig(apiObject *awstypes.CustomOriginConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"http_port":                aws.ToInt32(apiObject.HTTPPort),
		"https_port":               aws.ToInt32(apiObject.HTTPSPort),
		"origin_keepalive_timeout": aws.ToInt32(apiObject.OriginKeepaliveTimeout),
		"origin_protocol_policy":   apiObject.OriginProtocolPolicy,
		"origin_read_timeout":      aws.ToInt32(apiObject.OriginReadTimeout),
		"origin_ssl_protocols":     flattenCustomOriginConfigSSL(apiObject.OriginSslProtocols),
	}

	return tfMap
}

func expandCustomOriginConfigSSL(tfList []interface{}) *awstypes.OriginSslProtocols {
	return &awstypes.OriginSslProtocols{
		Items:    flex.ExpandStringyValueList[awstypes.SslProtocol](tfList),
		Quantity: aws.Int32(int32(len(tfList))),
	}
}

func flattenCustomOriginConfigSSL(apiObject *awstypes.OriginSslProtocols) []interface{} {
	if apiObject == nil {
		return nil
	}

	return flex.FlattenStringyValueList(apiObject.Items)
}

func expandS3OriginConfig(tfMap map[string]interface{}) *awstypes.S3OriginConfig {
	if tfMap == nil {
		return nil
	}

	return &awstypes.S3OriginConfig{
		OriginAccessIdentity: aws.String(tfMap["origin_access_identity"].(string)),
	}
}

func expandOriginShield(tfMap map[string]interface{}) *awstypes.OriginShield {
	if tfMap == nil {
		return nil
	}

	return &awstypes.OriginShield{
		Enabled:            aws.Bool(tfMap[names.AttrEnabled].(bool)),
		OriginShieldRegion: aws.String(tfMap["origin_shield_region"].(string)),
	}
}

func flattenS3OriginConfig(apiObject *awstypes.S3OriginConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	return map[string]interface{}{
		"origin_access_identity": aws.ToString(apiObject.OriginAccessIdentity),
	}
}

func flattenOriginShield(apiObject *awstypes.OriginShield) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	return map[string]interface{}{
		names.AttrEnabled:      aws.ToBool(apiObject.Enabled),
		"origin_shield_region": aws.ToString(apiObject.OriginShieldRegion),
	}
}

func expandCustomErrorResponses(tfList []interface{}) *awstypes.CustomErrorResponses {
	var items []awstypes.CustomErrorResponse

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := expandCustomErrorResponse(tfMap)

		if item == nil {
			continue
		}

		items = append(items, *item)
	}

	return &awstypes.CustomErrorResponses{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenCustomErrorResponses(apiObject *awstypes.CustomErrorResponses) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfList := []interface{}{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenCustomErrorResponse(&v))
	}

	return tfList
}

func expandCustomErrorResponse(tfMap map[string]interface{}) *awstypes.CustomErrorResponse {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CustomErrorResponse{
		ErrorCode: aws.Int32(int32(tfMap["error_code"].(int))),
	}

	if v, ok := tfMap["error_caching_min_ttl"]; ok {
		apiObject.ErrorCachingMinTTL = aws.Int64(int64(v.(int)))
	}

	if v, ok := tfMap["response_code"]; ok && v.(int) != 0 {
		apiObject.ResponseCode = flex.IntValueToString(v.(int))
	} else {
		apiObject.ResponseCode = aws.String("")
	}

	if v, ok := tfMap["response_page_path"]; ok {
		apiObject.ResponsePagePath = aws.String(v.(string))
	}

	return apiObject
}

func flattenCustomErrorResponse(apiObject *awstypes.CustomErrorResponse) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})
	tfMap["error_code"] = aws.ToInt32(apiObject.ErrorCode)

	if apiObject.ErrorCachingMinTTL != nil {
		tfMap["error_caching_min_ttl"] = aws.ToInt64(apiObject.ErrorCachingMinTTL)
	}

	if apiObject.ResponseCode != nil {
		tfMap["response_code"] = flex.StringToIntValue(apiObject.ResponseCode)
	}

	if apiObject.ResponsePagePath != nil {
		tfMap["response_page_path"] = aws.ToString(apiObject.ResponsePagePath)
	}

	return tfMap
}

func expandLoggingConfig(tfMap map[string]interface{}) *awstypes.LoggingConfig {
	apiObject := &awstypes.LoggingConfig{}

	if tfMap != nil {
		apiObject.Bucket = aws.String(tfMap[names.AttrBucket].(string))
		apiObject.Enabled = aws.Bool(true)
		apiObject.IncludeCookies = aws.Bool(tfMap["include_cookies"].(bool))
		apiObject.Prefix = aws.String(tfMap[names.AttrPrefix].(string))
	} else {
		apiObject.Bucket = aws.String("")
		apiObject.Enabled = aws.Bool(false)
		apiObject.IncludeCookies = aws.Bool(false)
		apiObject.Prefix = aws.String("")
	}

	return apiObject
}

func flattenLoggingConfig(apiObject *awstypes.LoggingConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrBucket:  aws.ToString(apiObject.Bucket),
		"include_cookies": aws.ToBool(apiObject.IncludeCookies),
		names.AttrPrefix:  aws.ToString(apiObject.Prefix),
	}

	return []interface{}{tfMap}
}

func expandAliases(tfList []interface{}) *awstypes.Aliases {
	apiObject := &awstypes.Aliases{
		Quantity: aws.Int32(int32(len(tfList))),
	}

	if len(tfList) > 0 {
		apiObject.Items = flex.ExpandStringValueList(tfList)
	}

	return apiObject
}

func flattenAliases(apiObject *awstypes.Aliases) []interface{} {
	if apiObject == nil {
		return nil
	}

	if apiObject.Items != nil {
		return flex.FlattenStringValueList(apiObject.Items)
	}

	return []interface{}{}
}

func expandRestrictions(tfMap map[string]interface{}) *awstypes.Restrictions {
	if tfMap == nil {
		return nil
	}

	return &awstypes.Restrictions{
		GeoRestriction: expandGeoRestriction(tfMap["geo_restriction"].([]interface{})[0].(map[string]interface{})),
	}
}

func flattenRestrictions(apiObject *awstypes.Restrictions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"geo_restriction": []interface{}{flattenGeoRestriction(apiObject.GeoRestriction)},
	}

	return []interface{}{tfMap}
}

func expandGeoRestriction(tfMap map[string]interface{}) *awstypes.GeoRestriction {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.GeoRestriction{
		Quantity:        aws.Int32(0),
		RestrictionType: awstypes.GeoRestrictionType(tfMap["restriction_type"].(string)),
	}

	if v, ok := tfMap["locations"]; ok {
		v := v.(*schema.Set)
		apiObject.Items = flex.ExpandStringValueSet(v)
		apiObject.Quantity = aws.Int32(int32(v.Len()))
	}

	return apiObject
}

func flattenGeoRestriction(apiObject *awstypes.GeoRestriction) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})
	tfMap["restriction_type"] = apiObject.RestrictionType

	if apiObject.Items != nil {
		tfMap["locations"] = flex.FlattenStringValueSet(apiObject.Items)
	}

	return tfMap
}

func expandViewerCertificate(tfMap map[string]interface{}) *awstypes.ViewerCertificate {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ViewerCertificate{}

	if v, ok := tfMap["iam_certificate_id"]; ok && v != "" {
		apiObject.IAMCertificateId = aws.String(v.(string))
		apiObject.SSLSupportMethod = awstypes.SSLSupportMethod(tfMap["ssl_support_method"].(string))
	} else if v, ok := tfMap["acm_certificate_arn"]; ok && v != "" {
		apiObject.ACMCertificateArn = aws.String(v.(string))
		apiObject.SSLSupportMethod = awstypes.SSLSupportMethod(tfMap["ssl_support_method"].(string))
	} else {
		apiObject.CloudFrontDefaultCertificate = aws.Bool(tfMap["cloudfront_default_certificate"].(bool))
	}

	if v, ok := tfMap["minimum_protocol_version"]; ok && v != "" {
		apiObject.MinimumProtocolVersion = awstypes.MinimumProtocolVersion(v.(string))
	}

	return apiObject
}

func flattenViewerCertificate(apiObject *awstypes.ViewerCertificate) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if apiObject.IAMCertificateId != nil {
		tfMap["iam_certificate_id"] = aws.ToString(apiObject.IAMCertificateId)
		tfMap["ssl_support_method"] = apiObject.SSLSupportMethod
	}

	if apiObject.ACMCertificateArn != nil {
		tfMap["acm_certificate_arn"] = aws.ToString(apiObject.ACMCertificateArn)
		tfMap["ssl_support_method"] = apiObject.SSLSupportMethod
	}

	if apiObject.CloudFrontDefaultCertificate != nil {
		tfMap["cloudfront_default_certificate"] = aws.ToBool(apiObject.CloudFrontDefaultCertificate)
	}

	tfMap["minimum_protocol_version"] = apiObject.MinimumProtocolVersion

	return []interface{}{tfMap}
}

func flattenActiveTrustedKeyGroups(apiObject *awstypes.ActiveTrustedKeyGroups) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
		"items":           flattenKGKeyPairIDs(apiObject.Items),
	}

	return []interface{}{tfMap}
}

func flattenKGKeyPairIDs(apiObjects []awstypes.KGKeyPairIds) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"key_group_id": aws.ToString(apiObject.KeyGroupId),
			"key_pair_ids": apiObject.KeyPairIds.Items,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenActiveTrustedSigners(apiObject *awstypes.ActiveTrustedSigners) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
		"items":           flattenSigners(apiObject.Items),
	}

	return []interface{}{tfMap}
}

func flattenSigners(apiObjects []awstypes.Signer) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"aws_account_number": aws.ToString(apiObject.AwsAccountNumber),
			"key_pair_ids":       apiObject.KeyPairIds.Items,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
