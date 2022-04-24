package cloudfront

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DataSourceOriginAccessIdentities() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOriginAccessIdentitiesRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"s3_canonical_user_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceOriginAccessIdentitiesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	comments := extractComments(d)

	input := &cloudfront.ListCloudFrontOriginAccessIdentitiesInput{}

	var results []*cloudfront.OriginAccessIdentitySummary

	err := conn.ListCloudFrontOriginAccessIdentitiesPages(input, func(page *cloudfront.ListCloudFrontOriginAccessIdentitiesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		for _, originAccessIdentity := range page.CloudFrontOriginAccessIdentityList.Items {

			if originAccessIdentity == nil {
				continue
			}
			if len(comments) > 0 {

				if ok := SliceContainsString(comments, *originAccessIdentity.Comment); !ok {
					continue
				}

			}

			results = append(results, originAccessIdentity)
		}
		return !lastPage

	})
	if err != nil {
		return fmt.Errorf("error reading Cloudfront OriginAccessIdentities: %w", err)
	}

	var arns, ids, s3UserIds []string

	for _, r := range results {
		iamArn := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   "iam",
			AccountID: "cloudfront",
			Resource:  fmt.Sprintf("user/CloudFront Origin Access Identity %s", *r.Id),
		}.String()
		arns = append(arns, iamArn)
		ids = append(ids, aws.StringValue(r.Id))
		s3UserIds = append(s3UserIds, *r.S3CanonicalUserId)
	}

	accountId := meta.(*conns.AWSClient).AccountID
	d.SetId(fmt.Sprintf("originaccessidentities-%s", accountId))

	if err := d.Set("arns", arns); err != nil {
		return fmt.Errorf("error setting arns: %w", err)
	}

	if err := d.Set("ids", ids); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	if err := d.Set("s3_canonical_user_ids", s3UserIds); err != nil {
		return fmt.Errorf("error setting s3_canonical_user_ids: %w", err)
	}

	return nil
}

func extractComments(d *schema.ResourceData) []*string {
	if v, ok := d.GetOk("filter"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		if u, ok := tfMap["name"].(string); ok && u == "comment" {
			if w, ok := tfMap["values"].(*schema.Set); ok && w.Len() > 0 {
				return flex.ExpandStringSet(w)
			}
		}
	}
	return nil
}

func SliceContainsString(slice []*string, s string) bool {
	for _, value := range slice {
		if strings.EqualFold(*value, s) {
			return true
		}
	}
	return false
}
