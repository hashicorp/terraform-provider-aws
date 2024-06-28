// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudfront_origin_request_policy", name="Origin Request Policy")
func dataSourceOriginRequestPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOriginRequestPolicyRead,

		Schema: map[string]*schema.Schema{
			names.AttrComment: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cookies_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cookie_behavior": {
							Computed: true,
							Type:     schema.TypeString,
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
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"headers_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"header_behavior": {
							Computed: true,
							Type:     schema.TypeString,
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
			names.AttrID: {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{names.AttrID, names.AttrName},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{names.AttrID, names.AttrName},
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
	}
}

func dataSourceOriginRequestPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	var originRequestPolicyID string

	if v, ok := d.GetOk(names.AttrID); ok {
		originRequestPolicyID = v.(string)
	} else {
		name := d.Get(names.AttrName).(string)
		input := &cloudfront.ListOriginRequestPoliciesInput{}

		err := listOriginRequestPoliciesPages(ctx, conn, input, func(page *cloudfront.ListOriginRequestPoliciesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, policySummary := range page.OriginRequestPolicyList.Items {
				if originRequestPolicy := policySummary.OriginRequestPolicy; aws.ToString(originRequestPolicy.OriginRequestPolicyConfig.Name) == name {
					originRequestPolicyID = aws.ToString(originRequestPolicy.Id)

					return false
				}
			}

			return !lastPage
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing CloudFront Origin Request Policies: %s", err)
		}

		if originRequestPolicyID == "" {
			return sdkdiag.AppendErrorf(diags, "no matching CloudFront Origin Request Policy (%s)", name)
		}
	}

	output, err := findOriginRequestPolicyByID(ctx, conn, originRequestPolicyID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Origin Request Policy (%s): %s", originRequestPolicyID, err)
	}

	d.SetId(originRequestPolicyID)

	apiObject := output.OriginRequestPolicy.OriginRequestPolicyConfig
	d.Set(names.AttrComment, apiObject.Comment)
	if apiObject.CookiesConfig != nil {
		if err := d.Set("cookies_config", []interface{}{flattenOriginRequestPolicyCookiesConfig(apiObject.CookiesConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting cookies_config: %s", err)
		}
	} else {
		d.Set("cookies_config", nil)
	}
	d.Set("etag", output.ETag)
	if apiObject.HeadersConfig != nil {
		if err := d.Set("headers_config", []interface{}{flattenOriginRequestPolicyHeadersConfig(apiObject.HeadersConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting headers_config: %s", err)
		}
	} else {
		d.Set("headers_config", nil)
	}
	d.Set(names.AttrName, apiObject.Name)
	if apiObject.QueryStringsConfig != nil {
		if err := d.Set("query_strings_config", []interface{}{flattenOriginRequestPolicyQueryStringsConfig(apiObject.QueryStringsConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting query_strings_config: %s", err)
		}
	} else {
		d.Set("query_strings_config", nil)
	}

	return diags
}
