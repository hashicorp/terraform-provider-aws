// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_kms_alias", name="Alias")
func dataSourceAlias() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAliasRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validNameForDataSource,
			},
			"target_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	target := d.Get(names.AttrName).(string)
	alias, err := findAliasByName(ctx, conn, target)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Alias (%s): %s", target, err)
	}

	d.SetId(aws.StringValue(alias.AliasArn))
	d.Set(names.AttrARN, alias.AliasArn)

	// ListAliases can return an alias for an AWS service key (e.g.
	// alias/aws/rds) without a TargetKeyId if the alias has not yet been
	// used for the first time. In that situation, calling DescribeKey will
	// associate an actual key with the alias, and the next call to
	// ListAliases will have a TargetKeyId for the alias.
	//
	// For a simpler codepath, we always call DescribeKey with the alias
	// name to get the target key's ARN and Id direct from AWS.
	//
	// https://docs.aws.amazon.com/kms/latest/APIReference/API_ListAliases.html

	keyMetadata, err := findKeyByID(ctx, conn, target)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Key (%s): %s", target, err)
	}

	d.Set("target_key_arn", keyMetadata.Arn)
	d.Set("target_key_id", keyMetadata.KeyId)

	return diags
}
