// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudhsmv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_cloudhsm_v2_cluster")
func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"cluster_certificates": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws_hardware_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_csr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hsm_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"manufacturer_hardware_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Conn(ctx)

	clusterID := d.Get("cluster_id").(string)
	input := &cloudhsmv2.DescribeClustersInput{
		Filters: map[string][]*string{
			"clusterIds": aws.StringSlice([]string{clusterID}),
		},
		MaxResults: aws.Int64(1),
	}
	if v, ok := d.GetOk("cluster_state"); ok {
		input.Filters["states"] = aws.StringSlice([]string{v.(string)})
	}

	cluster, err := findCluster(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudHSM v2 Cluster (%s): %s", clusterID, err)
	}

	d.SetId(clusterID)
	if err := d.Set("cluster_certificates", flattenCertificates(cluster)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cluster_certificates: %s", err)
	}
	d.Set("cluster_state", cluster.State)
	d.Set("security_group_id", cluster.SecurityGroup)
	var subnetIDs []string
	for _, v := range cluster.SubnetMapping {
		subnetIDs = append(subnetIDs, aws.StringValue(v))
	}
	d.Set("subnet_ids", subnetIDs)
	d.Set("vpc_id", cluster.VpcId)

	return diags
}
