// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_access_keys", name="Access Keys")
func dataSourceAccessKeys() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAccessKeysRead,
		Schema: map[string]*schema.Schema{
			"access_keys": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"create_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"user": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAccessKeysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	username := d.Get("user").(string)
	output, err := findAccessKeysByUser(ctx, conn, username)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Access Keys (%s): %s", username, err)
	}

	d.SetId(username)
	if err := d.Set("access_keys", flattenAccessKeys(output)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting access_keys: %s", err)
	}

	return diags
}

func flattenAccessKeys(apiObjects []awstypes.AccessKeyMetadata) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == (awstypes.AccessKeyMetadata{}) {
			continue
		}
		tfList = append(tfList, flattenAccessKey(apiObject))
	}

	return tfList
}

func flattenAccessKey(apiObject awstypes.AccessKeyMetadata) map[string]interface{} {
	if apiObject == (awstypes.AccessKeyMetadata{}) {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AccessKeyId; v != nil {
		m["access_key_id"] = aws.ToString(v)
	}
	if v := apiObject.CreateDate; v != nil {
		m["create_date"] = aws.ToTime(v).Format(time.RFC3339)
	}
	if v := apiObject.Status; v != "" {
		m[names.AttrStatus] = v
	}

	return m
}
