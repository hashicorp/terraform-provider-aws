package cloudfront

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceResponseHeadersPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceResponseHeadersPolicyRead,

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

func dataSourceResponseHeadersPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	var responseHeadersPolicyID string

	if v, ok := d.GetOk("id"); ok {
		responseHeadersPolicyID = v.(string)
	} else {
		name := d.Get("name").(string)
		input := &cloudfront.ListResponseHeadersPoliciesInput{}

		err := ListResponseHeadersPoliciesPages(conn, input, func(page *cloudfront.ListResponseHeadersPoliciesOutput, lastPage bool) bool {
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
			return fmt.Errorf("error listing CloudFront Response Headers Policies: %w", err)
		}

		if responseHeadersPolicyID == "" {
			return fmt.Errorf("no matching CloudFront Response Headers Policy (%s)", name)
		}
	}

	output, err := FindResponseHeadersPolicyByID(conn, responseHeadersPolicyID)

	if err != nil {
		return fmt.Errorf("error reading CloudFront Response Headers Policy (%s): %w", responseHeadersPolicyID, err)
	}

	d.SetId(responseHeadersPolicyID)

	apiObject := output.ResponseHeadersPolicy.ResponseHeadersPolicyConfig
	d.Set("comment", apiObject.Comment)
	if apiObject.CorsConfig != nil {
		if err := d.Set("cors_config", []interface{}{flattenResponseHeadersPolicyCorsConfig(apiObject.CorsConfig)}); err != nil {
			return fmt.Errorf("error setting cors_config: %w", err)
		}
	} else {
		d.Set("cors_config", nil)
	}
	if apiObject.CustomHeadersConfig != nil {
		if err := d.Set("custom_headers_config", []interface{}{flattenResponseHeadersPolicyCustomHeadersConfig(apiObject.CustomHeadersConfig)}); err != nil {
			return fmt.Errorf("error setting custom_headers_config: %w", err)
		}
	} else {
		d.Set("custom_headers_config", nil)
	}
	d.Set("etag", output.ETag)
	d.Set("name", apiObject.Name)
	if apiObject.SecurityHeadersConfig != nil {
		if err := d.Set("security_headers_config", []interface{}{flattenResponseHeadersPolicySecurityHeadersConfig(apiObject.SecurityHeadersConfig)}); err != nil {
			return fmt.Errorf("error setting security_headers_config: %w", err)
		}
	} else {
		d.Set("security_headers_config", nil)
	}

	if apiObject.ServerTimingHeadersConfig != nil {
		if err := d.Set("server_timing_headers_config", []interface{}{flattenResponseHeadersPolicyServerTimingHeadersConfig(apiObject.ServerTimingHeadersConfig)}); err != nil {
			return fmt.Errorf("error setting server_timing_headers_config: %w", err)
		}
	} else {
		d.Set("server_timing_headers_config", nil)
	}

	return nil
}
