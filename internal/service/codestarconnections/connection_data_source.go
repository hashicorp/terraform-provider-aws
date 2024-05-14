// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codestarconnections

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codestarconnections"
	"github.com/aws/aws-sdk-go-v2/service/codestarconnections/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_codestarconnections_connection")
func dataSourceConnection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectionRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
				ExactlyOneOf: []string{names.AttrARN, names.AttrName},
			},
			"connection_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrARN, names.AttrName},
			},
			"provider_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeStarConnectionsClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var connection *types.Connection

	if v, ok := d.GetOk(names.AttrARN); ok {
		arn := v.(string)
		var err error

		connection, err = findConnectionByARN(ctx, conn, arn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CodeStar Connections Connection (%s): %s", arn, err)
		}
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)

		input := &codestarconnections.ListConnectionsInput{}
		pages := codestarconnections.NewListConnectionsPaginator(conn, input)
		for pages.HasMorePages() && connection == nil {
			page, err := pages.NextPage(ctx)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "listing CodeStar Connections Connections: %s", err)
			}

			for _, v := range page.Connections {
				v := v

				if aws.ToString(v.ConnectionName) == name {
					connection = &v
					break
				}
			}
		}

		if connection == nil {
			return sdkdiag.AppendErrorf(diags, "CodeStar Connections Connection (%s): not found", name)
		}
	}

	arn := aws.ToString(connection.ConnectionArn)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set("connection_status", connection.ConnectionStatus)
	d.Set("host_arn", connection.HostArn)
	d.Set(names.AttrName, connection.ConnectionName)
	d.Set("provider_type", connection.ProviderType)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for CodeStar Connections Connection (%s): %s", arn, err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
