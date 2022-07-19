package iam

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceBuckets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceBucketsRead,

		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceBucketsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	input := &s3.ListBucketsInput{}

	var results []*s3.Bucket

	output, err := conn.ListBuckets(input)
	for _, bucket := range output.Buckets {
		if bucket == nil {
			continue
		}

		if v, ok := d.GetOk("name_regex"); ok && !regexp.MustCompile(v.(string)).MatchString(aws.StringValue(bucket.Name)) {
			continue
		}

		results = append(results, bucket)
	}

	if err != nil {
		return fmt.Errorf("error reading S3 buckets: %w", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	var arns, ids []string

	for _, r := range results {
		ids = append(ids, aws.StringValue(r.Name))
	}

	if err := d.Set("ids", ids); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}
