// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emrcontainers

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_emrcontainers_virtual_cluster", name="Virtual Cluster")
// @Tags
func dataSourceVirtualCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVirtualClusterRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container_provider": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
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
												names.AttrNamespace: {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"virtual_cluster_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceVirtualClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRContainersClient(ctx)

	id := d.Get("virtual_cluster_id").(string)
	vc, err := findVirtualClusterByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Containers Virtual Cluster (%s): %s", id, err)
	}

	d.SetId(aws.ToString(vc.Id))
	d.Set(names.AttrARN, vc.Arn)
	if vc.ContainerProvider != nil {
		if err := d.Set("container_provider", []interface{}{flattenContainerProvider(vc.ContainerProvider)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting container_provider: %s", err)
		}
	} else {
		d.Set("container_provider", nil)
	}
	d.Set(names.AttrCreatedAt, aws.ToTime(vc.CreatedAt).String())
	d.Set(names.AttrName, vc.Name)
	d.Set(names.AttrState, vc.State)
	d.Set("virtual_cluster_id", vc.Id)

	setTagsOut(ctx, vc.Tags)

	return diags
}
