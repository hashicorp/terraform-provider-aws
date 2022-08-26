package cloudfront

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceOriginAccessIdentities() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOriginAccessIdentitiesRead,

		Schema: map[string]*schema.Schema{
			"comments": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"iam_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ids": {
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

	var comments []interface{}

	if v, ok := d.GetOk("comments"); ok && v.(*schema.Set).Len() > 0 {
		comments = v.(*schema.Set).List()
	}

	var output []*cloudfront.OriginAccessIdentitySummary

	err := conn.ListCloudFrontOriginAccessIdentitiesPages(&cloudfront.ListCloudFrontOriginAccessIdentitiesInput{}, func(page *cloudfront.ListCloudFrontOriginAccessIdentitiesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CloudFrontOriginAccessIdentityList.Items {
			if v == nil {
				continue
			}

			if len(comments) > 0 {
				if _, ok := verify.SliceContainsString(comments, aws.StringValue(v.Comment)); !ok {
					continue
				}
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("listing CloudFront origin access identities: %w", err)
	}

	var iamARNs, ids, s3CanonicalUserIDs []string

	for _, v := range output {
		// See https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-restricting-access-to-s3.html#private-content-updating-s3-bucket-policies-principal.
		iamARN := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   "iam",
			AccountID: "cloudfront",
			Resource:  fmt.Sprintf("user/CloudFront Origin Access Identity %s", *v.Id),
		}.String()
		iamARNs = append(iamARNs, iamARN)
		ids = append(ids, aws.StringValue(v.Id))
		s3CanonicalUserIDs = append(s3CanonicalUserIDs, aws.StringValue(v.S3CanonicalUserId))
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)
	d.Set("iam_arns", iamARNs)
	d.Set("ids", ids)
	d.Set("s3_canonical_user_ids", s3CanonicalUserIDs)

	return nil
}
