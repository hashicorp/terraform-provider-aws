// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const keyRequestPageSize = 1000

// @SDKDataSource("aws_s3_objects", name="Objects")
func dataSourceObjects() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceObjectsRead,

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:     schema.TypeString,
				Required: true,
			},
			"common_prefixes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"delimiter": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encoding_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.EncodingType](),
			},
			"fetch_owner": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"max_keys": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1000,
			},
			"owners": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrPrefix: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"request_charged": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"request_payer": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.RequestPayer](),
			},
			"start_after": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceObjectsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)
	var optFns []func(*s3.Options)

	bucket := d.Get(names.AttrBucket).(string)
	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}
	// Via S3 access point: "Invalid configuration: region from ARN `us-east-1` does not match client region `aws-global` and UseArnRegion is `false`".
	if arn.IsARN(bucket) && conn.Options().Region == names.GlobalRegionID {
		optFns = append(optFns, func(o *s3.Options) { o.UseARNRegion = true })
	}
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	if v, ok := d.GetOk("delimiter"); ok {
		input.Delimiter = aws.String(v.(string))
	}

	if v, ok := d.GetOk("encoding_type"); ok {
		input.EncodingType = types.EncodingType(v.(string))
	}

	if v, ok := d.GetOk("fetch_owner"); ok {
		input.FetchOwner = aws.Bool(v.(bool))
	}

	// "input.MaxKeys" refers to max keys returned in a single request
	// (i.e. page size), not the total number of keys returned if you page
	// through the results. "max_keys" does refer to total keys returned.
	maxKeys := int64(d.Get("max_keys").(int))
	if maxKeys <= keyRequestPageSize {
		input.MaxKeys = aws.Int32(int32(maxKeys))
	}

	if v, ok := d.GetOk(names.AttrPrefix); ok {
		input.Prefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("request_payer"); ok {
		input.RequestPayer = types.RequestPayer(v.(string))
	}

	if v, ok := d.GetOk("start_after"); ok {
		input.StartAfter = aws.String(v.(string))
	}

	var nKeys int64
	var commonPrefixes, keys, owners []string
	var requestCharged string

	pages := s3.NewListObjectsV2Paginator(conn, input)
pageLoop:
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx, optFns...)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing S3 Bucket (%s) Objects: %s", bucket, err)
		}

		requestCharged = string(page.RequestCharged)

		for _, v := range page.CommonPrefixes {
			commonPrefixes = append(commonPrefixes, aws.ToString(v.Prefix))
		}

		for _, v := range page.Contents {
			if nKeys >= maxKeys {
				break pageLoop
			}

			keys = append(keys, aws.ToString(v.Key))

			if v := v.Owner; v != nil {
				owners = append(owners, aws.ToString(v.ID))
			}

			nKeys++
		}
	}

	d.SetId(bucket)
	d.Set("common_prefixes", commonPrefixes)
	d.Set("keys", keys)
	d.Set("owners", owners)
	d.Set("request_charged", requestCharged)

	return diags
}
