// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssmincidents_replication_set", name="Replication Set")
// @Region(overrideEnabled=false)
// @Tags(identifierAttribute="id")
func dataSourceReplicationSet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReplicationSetRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protected": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"last_modified_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRegion: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrKMSKeyARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatusMessage: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Deprecated: "region is deprecated. Use regions instead.",
			},
			"regions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKeyARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatusMessage: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceReplicationSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	var input ssmincidents.ListReplicationSetsInput
	arn, err := findReplicationSetARN(ctx, conn, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSMIncidents Replication Set: %s", err)
	}

	d.SetId(aws.ToString(arn))

	replicationSet, err := findReplicationSetByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSMIncidents Replication Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, replicationSet.Arn)
	d.Set("created_by", replicationSet.CreatedBy)
	d.Set("deletion_protected", replicationSet.DeletionProtected)
	d.Set("last_modified_by", replicationSet.LastModifiedBy)
	if err := d.Set(names.AttrRegion, flattenRegionInfos(replicationSet.RegionMap)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting region: %s", err)
	}
	if err := d.Set("regions", flattenRegionInfos(replicationSet.RegionMap)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting regions: %s", err)
	}
	d.Set(names.AttrStatus, replicationSet.Status)

	return diags
}

func findReplicationSetARN(context context.Context, conn *ssmincidents.Client, input *ssmincidents.ListReplicationSetsInput) (*string, error) {
	output, err := findReplicationSetARNs(context, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReplicationSetARNs(context context.Context, conn *ssmincidents.Client, input *ssmincidents.ListReplicationSetsInput) ([]string, error) {
	var output []string

	pages := ssmincidents.NewListReplicationSetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(context)

		if err != nil {
			return nil, err
		}

		output = append(output, page.ReplicationSetArns...)
	}

	return output, nil
}
