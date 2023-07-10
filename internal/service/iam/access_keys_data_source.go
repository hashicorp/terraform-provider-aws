// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_access_keys")
func DataSourceAccessKeys() *schema.Resource {
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
						"status": {
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

const (
	DSNameAccessKeys = "Access Keys Data Source"
)

func dataSourceAccessKeysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	username := d.Get("user").(string)
	out, err := FindAccessKeys(ctx, conn, username)

	if err != nil {
		return create.DiagError(names.IAM, create.ErrActionReading, DSNameAccessKeys, username, err)
	}

	d.SetId(username)

	if err := d.Set("access_keys", flattenAccessKeys(out)); err != nil {
		return create.DiagError(names.IAM, create.ErrActionSetting, DSNameAccessKeys, d.Id(), err)
	}

	return nil
}

func flattenAccessKeys(apiObjects []*iam.AccessKeyMetadata) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}
		tfList = append(tfList, flattenAccessKey(apiObject))
	}

	return tfList
}

func flattenAccessKey(apiObject *iam.AccessKeyMetadata) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AccessKeyId; v != nil {
		m["access_key_id"] = aws.ToString(v)
	}
	if v := apiObject.CreateDate; v != nil {
		m["create_date"] = aws.ToTime(v).Format(time.RFC3339)
	}
	if v := apiObject.Status; v != nil {
		m["status"] = aws.ToString(v)
	}

	return m
}
