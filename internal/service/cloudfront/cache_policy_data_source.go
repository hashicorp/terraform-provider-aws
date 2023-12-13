// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_cloudfront_cache_policy")
func DataSourceCachePolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCachePolicyRead,

		Schema: map[string]*schema.Schema{
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_ttl": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"max_ttl": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"min_ttl": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"parameters_in_cache_key_and_forwarded_to_origin": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cookies_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cookie_behavior": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"cookies": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
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
						"enable_accept_encoding_brotli": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"enable_accept_encoding_gzip": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"headers_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"header_behavior": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"headers": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
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
						"query_strings_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"query_string_behavior": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"query_strings": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
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
					},
				},
			},
		},
	}
}
func dataSourceCachePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)

	var cachePolicyID string

	if v, ok := d.GetOk("id"); ok {
		cachePolicyID = v.(string)
	} else {
		name := d.Get("name").(string)
		input := &cloudfront.ListCachePoliciesInput{}

		err := ListCachePoliciesPages(ctx, conn, input, func(page *cloudfront.ListCachePoliciesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, policySummary := range page.CachePolicyList.Items {
				if cachePolicy := policySummary.CachePolicy; aws.StringValue(cachePolicy.CachePolicyConfig.Name) == name {
					cachePolicyID = aws.StringValue(cachePolicy.Id)

					return false
				}
			}

			return !lastPage
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing CloudFront Cache Policies: %s", err)
		}

		if cachePolicyID == "" {
			return sdkdiag.AppendErrorf(diags, "no matching CloudFront Cache Policy (%s)", name)
		}
	}

	output, err := FindCachePolicyByID(ctx, conn, cachePolicyID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Cache Policy (%s): %s", cachePolicyID, err)
	}

	d.SetId(cachePolicyID)

	apiObject := output.CachePolicy.CachePolicyConfig
	d.Set("comment", apiObject.Comment)
	d.Set("default_ttl", apiObject.DefaultTTL)
	d.Set("etag", output.ETag)
	d.Set("max_ttl", apiObject.MaxTTL)
	d.Set("min_ttl", apiObject.MinTTL)
	d.Set("name", apiObject.Name)
	if apiObject.ParametersInCacheKeyAndForwardedToOrigin != nil {
		if err := d.Set("parameters_in_cache_key_and_forwarded_to_origin", []interface{}{flattenParametersInCacheKeyAndForwardedToOrigin(apiObject.ParametersInCacheKeyAndForwardedToOrigin)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting parameters_in_cache_key_and_forwarded_to_origin: %s", err)
		}
	} else {
		d.Set("parameters_in_cache_key_and_forwarded_to_origin", nil)
	}

	return diags
}
