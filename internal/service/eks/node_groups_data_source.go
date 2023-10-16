// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_eks_node_groups")
func DataSourceNodeGroups() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceNodeGroupsRead,

		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceNodeGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).EKSClient(ctx)
	clusterName := d.Get("cluster_name").(string)

	input := &eks.ListNodegroupsInput{
		ClusterName: aws.String(clusterName),
	}

	var nodegroups []string

	paginator := eks.NewListNodegroupsPaginator(client, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing EKS Node Groups: %s", err)
		}

		nodegroups = append(nodegroups, output.Nodegroups...)
	}

	d.SetId(clusterName)
	d.Set("cluster_name", clusterName)
	d.Set("names", nodegroups)

	return diags
}
