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
)

// @SDKDataSource("aws_ssoadmin_instances")
func DataSourceInstances() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstancesRead,

		Schema: map[string]*schema.Schema{
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"identity_store_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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

	var instances []map[string]interface{}

	for _, v := range output {
		instances = append(instances, map[string]interface{}{
			"arn":               aws.ToString(v.InstanceArn),
			"identity_store_id": aws.ToString(v.IdentityStoreId),
		})
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("instances", instances)

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
