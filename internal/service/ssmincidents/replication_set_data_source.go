// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssmincidents_replication_set")
func DataSourceReplicationSet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReplicationSetRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			// all other computed fields in alphabetic order
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
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	DSNameReplicationSet = "Replication Set Data Source"
)

func dataSourceReplicationSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	arn, err := getReplicationSetARN(ctx, client)

	if err != nil {
		return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionReading, ResNameReplicationSet, d.Id(), err)
	}

	d.SetId(arn)

	replicationSet, err := FindReplicationSetByID(ctx, client, d.Id())

	if err != nil {
		return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionReading, ResNameReplicationSet, d.Id(), err)
	}

	d.Set(names.AttrARN, replicationSet.Arn)
	d.Set("created_by", replicationSet.CreatedBy)
	d.Set("deletion_protected", replicationSet.DeletionProtected)
	d.Set("last_modified_by", replicationSet.LastModifiedBy)
	d.Set(names.AttrStatus, replicationSet.Status)

	if err := d.Set(names.AttrRegion, flattenRegions(replicationSet.RegionMap)); err != nil {
		return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionSetting, ResNameReplicationSet, d.Id(), err)
	}

	tags, err := listTags(ctx, client, d.Id())

	if err != nil {
		return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionReading, DSNameReplicationSet, d.Id(), err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	//lintignore:AWSR002
	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.AppendDiagError(diags, names.SSMIncidents, create.ErrActionSetting, DSNameReplicationSet, d.Id(), err)
	}

	return diags
}
