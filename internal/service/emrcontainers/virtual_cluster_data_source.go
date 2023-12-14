// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emrcontainers

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_emrcontainers_virtual_cluster")
func DataSourceVirtualCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVirtualClusterRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container_provider": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"info": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"eks_info": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"namespace": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"virtual_cluster_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceVirtualClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EMRContainersConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id := d.Get("virtual_cluster_id").(string)
	vc, err := FindVirtualClusterByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Containers Virtual Cluster (%s): %s", id, err)
	}

	d.SetId(aws.StringValue(vc.Id))
	d.Set("arn", vc.Arn)
	if vc.ContainerProvider != nil {
		if err := d.Set("container_provider", []interface{}{flattenContainerProvider(vc.ContainerProvider)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting container_provider: %s", err)
		}
	} else {
		d.Set("container_provider", nil)
	}
	d.Set("created_at", aws.TimeValue(vc.CreatedAt).String())
	d.Set("name", vc.Name)
	d.Set("state", vc.State)
	d.Set("virtual_cluster_id", vc.Id)

	if err := d.Set("tags", KeyValueTags(ctx, vc.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
