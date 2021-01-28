package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsCloudFrontOriginRequestPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudFrontOriginRequestPolicyRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"etag": {
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

func dataSourceAwsCloudFrontOriginRequestPolicyFindByName(d *schema.ResourceData, conn *cloudfront.CloudFront) error {
	var originRequestPolicy *cloudfront.OriginRequestPolicy
	request := &cloudfront.ListOriginRequestPoliciesInput{}
	resp, err := conn.ListOriginRequestPolicies(request)
	if err != nil {
		return err
	}

	for _, policySummary := range resp.OriginRequestPolicyList.Items {
		if *policySummary.OriginRequestPolicy.OriginRequestPolicyConfig.Name == d.Get("name").(string) {
			originRequestPolicy = policySummary.OriginRequestPolicy
			break
		}
	}

	if originRequestPolicy != nil {
		d.SetId(aws.StringValue(originRequestPolicy.Id))
	}
	return nil
}

func dataSourceAwsCloudFrontOriginRequestPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	if d.Id() == "" {
		if err := dataSourceAwsCloudFrontOriginRequestPolicyFindByName(d, conn); err != nil {
			return err
		}
	}

	if d.Id() != "" {
		request := &cloudfront.GetOriginRequestPolicyInput{
			Id: aws.String(d.Id()),
		}

		resp, err := conn.GetOriginRequestPolicy(request)
		if err != nil {
			return err
		}
		d.Set("etag", aws.StringValue(resp.ETag))

		flattenCloudFrontOriginRequestPolicy(d, resp.OriginRequestPolicy.OriginRequestPolicyConfig)
	}

	return nil
}
