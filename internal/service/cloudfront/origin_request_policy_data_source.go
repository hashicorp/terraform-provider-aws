package cloudfront

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceOriginRequestPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOriginRequestPolicyRead,

		Schema: map[string]*schema.Schema{
			"comment": {
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
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "name"},
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

func dataSourceOriginRequestPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	var originRequestPolicyID string

	if v, ok := d.GetOk("id"); ok {
		originRequestPolicyID = v.(string)
	} else {
		name := d.Get("name").(string)
		input := &cloudfront.ListOriginRequestPoliciesInput{}

		err := ListOriginRequestPoliciesPages(conn, input, func(page *cloudfront.ListOriginRequestPoliciesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, policySummary := range page.OriginRequestPolicyList.Items {
				if originRequestPolicy := policySummary.OriginRequestPolicy; aws.StringValue(originRequestPolicy.OriginRequestPolicyConfig.Name) == name {
					originRequestPolicyID = aws.StringValue(originRequestPolicy.Id)

					return false
				}
			}

			return !lastPage
		})

		if err != nil {
			return fmt.Errorf("error listing CloudFront Origin Request Policies: %w", err)
		}

		if originRequestPolicyID == "" {
			return fmt.Errorf("no matching CloudFront Origin Request Policy (%s)", name)
		}
	}

	output, err := FindOriginRequestPolicyByID(conn, originRequestPolicyID)

	if err != nil {
		return fmt.Errorf("error reading CloudFront Origin Request Policy (%s): %w", originRequestPolicyID, err)
	}

	d.SetId(originRequestPolicyID)

	apiObject := output.OriginRequestPolicy.OriginRequestPolicyConfig
	d.Set("comment", apiObject.Comment)
	if apiObject.CookiesConfig != nil {
		if err := d.Set("cookies_config", []interface{}{flattenOriginRequestPolicyCookiesConfig(apiObject.CookiesConfig)}); err != nil {
			return fmt.Errorf("error setting cookies_config: %w", err)
		}
	} else {
		d.Set("cookies_config", nil)
	}
	d.Set("etag", output.ETag)
	if apiObject.HeadersConfig != nil {
		if err := d.Set("headers_config", []interface{}{flattenOriginRequestPolicyHeadersConfig(apiObject.HeadersConfig)}); err != nil {
			return fmt.Errorf("error setting headers_config: %w", err)
		}
	} else {
		d.Set("headers_config", nil)
	}
	d.Set("name", apiObject.Name)
	if apiObject.QueryStringsConfig != nil {
		if err := d.Set("query_strings_config", []interface{}{flattenOriginRequestPolicyQueryStringsConfig(apiObject.QueryStringsConfig)}); err != nil {
			return fmt.Errorf("error setting query_strings_config: %w", err)
		}
	} else {
		d.Set("query_strings_config", nil)
	}

	return nil
}
