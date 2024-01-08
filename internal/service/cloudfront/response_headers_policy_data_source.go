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

// @SDKDataSource("aws_cloudfront_response_headers_policy")
func DataSourceResponseHeadersPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResponseHeadersPolicyRead,

		Schema: map[string]*schema.Schema{
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cors_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_control_allow_credentials": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"access_control_allow_headers": {
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
						"access_control_allow_methods": {
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
						"access_control_allow_origins": {
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
						"access_control_expose_headers": {
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
						"access_control_max_age_sec": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"origin_override": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"custom_headers_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"items": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"header": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"value": {
										Type:     schema.TypeString,
										Computed: true,
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
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"remove_headers_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"items": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"header": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"security_headers_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content_security_policy": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"content_security_policy": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"content_type_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"override": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"frame_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"frame_option": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"referrer_policy": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"referrer_policy": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"strict_transport_security": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_control_max_age_sec": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"include_subdomains": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"preload": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"xss_protection": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode_block": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"protection": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"report_uri": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"server_timing_headers_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"sampling_rate": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
					}},
			},
		},
	}
}

func dataSourceResponseHeadersPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn(ctx)

	var responseHeadersPolicyID string

	if v, ok := d.GetOk("id"); ok {
		responseHeadersPolicyID = v.(string)
	} else {
		name := d.Get("name").(string)
		input := &cloudfront.ListResponseHeadersPoliciesInput{}

		err := ListResponseHeadersPoliciesPages(ctx, conn, input, func(page *cloudfront.ListResponseHeadersPoliciesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, policySummary := range page.ResponseHeadersPolicyList.Items {
				if responseHeadersPolicy := policySummary.ResponseHeadersPolicy; aws.StringValue(responseHeadersPolicy.ResponseHeadersPolicyConfig.Name) == name {
					responseHeadersPolicyID = aws.StringValue(responseHeadersPolicy.Id)

					return false
				}
			}

			return !lastPage
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing CloudFront Response Headers Policies: %s", err)
		}

		if responseHeadersPolicyID == "" {
			return sdkdiag.AppendErrorf(diags, "no matching CloudFront Response Headers Policy (%s)", name)
		}
	}

	output, err := FindResponseHeadersPolicyByID(ctx, conn, responseHeadersPolicyID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Response Headers Policy (%s): %s", responseHeadersPolicyID, err)
	}

	d.SetId(responseHeadersPolicyID)

	apiObject := output.ResponseHeadersPolicy.ResponseHeadersPolicyConfig
	d.Set("comment", apiObject.Comment)
	if apiObject.CorsConfig != nil {
		if err := d.Set("cors_config", []interface{}{flattenResponseHeadersPolicyCorsConfig(apiObject.CorsConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting cors_config: %s", err)
		}
	} else {
		d.Set("cors_config", nil)
	}
	if apiObject.CustomHeadersConfig != nil {
		if err := d.Set("custom_headers_config", []interface{}{flattenResponseHeadersPolicyCustomHeadersConfig(apiObject.CustomHeadersConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting custom_headers_config: %s", err)
		}
	} else {
		d.Set("custom_headers_config", nil)
	}
	d.Set("etag", output.ETag)
	d.Set("name", apiObject.Name)
	if apiObject.RemoveHeadersConfig != nil {
		if err := d.Set("remove_headers_config", []interface{}{flattenResponseHeadersPolicyRemoveHeadersConfig(apiObject.RemoveHeadersConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting remove_headers_config: %s", err)
		}
	} else {
		d.Set("remove_headers_config", nil)
	}
	if apiObject.SecurityHeadersConfig != nil {
		if err := d.Set("security_headers_config", []interface{}{flattenResponseHeadersPolicySecurityHeadersConfig(apiObject.SecurityHeadersConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting security_headers_config: %s", err)
		}
	} else {
		d.Set("security_headers_config", nil)
	}

	if apiObject.ServerTimingHeadersConfig != nil {
		if err := d.Set("server_timing_headers_config", []interface{}{flattenResponseHeadersPolicyServerTimingHeadersConfig(apiObject.ServerTimingHeadersConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting server_timing_headers_config: %s", err)
		}
	} else {
		d.Set("server_timing_headers_config", nil)
	}

	return diags
}
