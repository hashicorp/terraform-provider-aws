// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_ssoadmin_instances")
func DataSourceInstances() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstancesRead,

		Schema: map[string]*schema.Schema{
			"arns": {
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
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	output, err := findInstanceMetadatas(ctx, conn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Instances: %s", err)
	}

	var identityStoreIDs, arns []string

	for _, v := range output {
		identityStoreIDs = append(identityStoreIDs, aws.StringValue(v.IdentityStoreId))
		arns = append(arns, aws.StringValue(v.InstanceArn))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("arns", arns)
	d.Set("identity_store_ids", identityStoreIDs)

	return diags
}

func findInstanceMetadatas(ctx context.Context, conn *ssoadmin.SSOAdmin) ([]*ssoadmin.InstanceMetadata, error) {
	input := &ssoadmin.ListInstancesInput{}
	var output []*ssoadmin.InstanceMetadata

	err := conn.ListInstancesPagesWithContext(ctx, input, func(page *ssoadmin.ListInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Instances {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
