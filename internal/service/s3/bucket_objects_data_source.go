// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

// WARNING: This code is DEPRECATED and will be removed in a future release!!
// DO NOT apply fixes or enhancements to the data source in this file.
// INSTEAD, apply fixes and enhancements to the data source in "objects_data_source.go".

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_s3_bucket_objects", name="Bucket Objects")
func dataSourceBucketObjects() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBucketObjectsRead,

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Deprecated: "Use the aws_s3_objects data source instead",
				Type:       schema.TypeString,
				Required:   true,
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
				Type:     schema.TypeString,
				Optional: true,
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
			"start_after": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		DeprecationMessage: `use the aws_s3_objects data source instead`,
	}
}

func dataSourceBucketObjectsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	if s, ok := d.GetOk("delimiter"); ok {
		input.Delimiter = aws.String(s.(string))
	}

	if s, ok := d.GetOk("encoding_type"); ok {
		input.EncodingType = types.EncodingType(s.(string))
	}

	// "listInput.MaxKeys" refers to max keys returned in a single request
	// (i.e., page size), not the total number of keys returned if you page
	// through the results. "maxKeys" does refer to total keys returned.
	maxKeys := int64(d.Get("max_keys").(int))
	if maxKeys <= keyRequestPageSize {
		input.MaxKeys = aws.Int32(int32(maxKeys))
	}

	if s, ok := d.GetOk(names.AttrPrefix); ok {
		input.Prefix = aws.String(s.(string))
	}

	if s, ok := d.GetOk("start_after"); ok {
		input.StartAfter = aws.String(s.(string))
	}

	if b, ok := d.GetOk("fetch_owner"); ok {
		input.FetchOwner = aws.Bool(b.(bool))
	}

	var nKeys int64
	var commonPrefixes, keys, owners []string

	pages := s3.NewListObjectsV2Paginator(conn, input)
pageLoop:
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing S3 Bucket (%s) Objects: %s", bucket, err)
		}

		for _, commonPrefix := range page.CommonPrefixes {
			commonPrefixes = append(commonPrefixes, aws.ToString(commonPrefix.Prefix))
		}

		for _, object := range page.Contents {
			if nKeys >= maxKeys {
				break pageLoop
			}

			keys = append(keys, aws.ToString(object.Key))

			if object.Owner != nil {
				owners = append(owners, aws.ToString(object.Owner.ID))
			}

			nKeys++
		}
	}

	d.SetId(bucket)
	d.Set("common_prefixes", commonPrefixes)
	d.Set("keys", keys)
	d.Set("owners", owners)

	return diags
}
