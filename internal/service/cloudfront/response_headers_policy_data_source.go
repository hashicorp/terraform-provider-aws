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
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:          schema.TypeString,
				ConflictsWith: []string{"name"},
				Optional:      true,
			},
			"name": {
				Type:          schema.TypeString,
				ConflictsWith: []string{"id"},
				Optional:      true,
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
		},
	}
}
func dataSourceResponseHeadersPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	if d.Id() == "" {
		if err := dataSourceResponseHeadersPolicyFindByName(d, conn); err != nil {
			return fmt.Errorf("unable to locate response headers policy by name: %s", err.Error())
		}
	}

	if d.Id() != "" {
		d.Set("id", d.Id())
		request := &cloudfront.GetResponseHeadersPolicyInput{
			Id: aws.String(d.Id()),
		}

		resp, err := conn.GetResponseHeadersPolicy(request)
		if err != nil {
			return fmt.Errorf("unable to retrieve response headers policy with ID %s: %s", d.Id(), err.Error())
		}
		d.Set("etag", resp.ETag)

		if err := setCloudFrontResponseHeadersPolicy(d, resp.ResponseHeadersPolicy.ResponseHeadersPolicyConfig); err != nil {
			return fmt.Errorf("unable to store response headers policy in config: %s", err.Error())
		}
	}

	return nil
}

func dataSourceResponseHeadersPolicyFindByName(d *schema.ResourceData, conn *cloudfront.CloudFront) error {
	var responseHeadersPolicy *cloudfront.ResponseHeadersPolicy
	request := &cloudfront.ListResponseHeadersPoliciesInput{}
	resp, err := conn.ListResponseHeadersPolicies(request)
	if err != nil {
		return err
	}

	for _, policySummary := range resp.ResponseHeadersPolicyList.Items {
		if aws.StringValue(policySummary.ResponseHeadersPolicy.ResponseHeadersPolicyConfig.Name) == d.Get("name").(string) {
			responseHeadersPolicy = policySummary.ResponseHeadersPolicy
			break
		}
	}

	if responseHeadersPolicy != nil {
		d.SetId(aws.StringValue(responseHeadersPolicy.Id))
	}
	return nil
}
