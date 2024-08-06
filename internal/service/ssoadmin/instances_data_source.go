// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssoadmin_instances")
func DataSourceInstances() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstancesRead,

		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"identity_store_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInstancesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	output, err := findInstanceMetadatas(ctx, conn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Instances: %s", err)
	}

	var identityStoreIDs, arns []string

	for _, v := range output {
		identityStoreIDs = append(identityStoreIDs, aws.ToString(v.IdentityStoreId))
		arns = append(arns, aws.ToString(v.InstanceArn))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrARNs, arns)
	d.Set("identity_store_ids", identityStoreIDs)

	return diags
}

func findInstanceMetadatas(ctx context.Context, conn *ssoadmin.Client) ([]awstypes.InstanceMetadata, error) {
	input := &ssoadmin.ListInstancesInput{}
	var output []awstypes.InstanceMetadata

	paginator := ssoadmin.NewListInstancesPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		if page != nil {
			output = append(output, page.Instances...)
		}
	}

	return output, nil
}
