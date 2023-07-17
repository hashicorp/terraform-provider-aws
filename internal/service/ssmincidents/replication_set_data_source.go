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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"kms_key_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_message": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": tftags.TagsSchemaComputed(),
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
			"status": {
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
	client := meta.(*conns.AWSClient).SSMIncidentsClient(ctx)

	arn, err := getReplicationSetARN(ctx, client)

	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionReading, ResNameReplicationSet, d.Id(), err)
	}

	d.SetId(arn)

	replicationSet, err := FindReplicationSetByID(ctx, client, d.Id())

	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionReading, ResNameReplicationSet, d.Id(), err)
	}

	d.Set("arn", replicationSet.Arn)
	d.Set("created_by", replicationSet.CreatedBy)
	d.Set("deletion_protected", replicationSet.DeletionProtected)
	d.Set("last_modified_by", replicationSet.LastModifiedBy)
	d.Set("status", replicationSet.Status)

	if err := d.Set("region", flattenRegions(replicationSet.RegionMap)); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionSetting, ResNameReplicationSet, d.Id(), err)
	}

	tags, err := listTags(ctx, client, d.Id())

	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionReading, DSNameReplicationSet, d.Id(), err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	//lintignore:AWSR002
	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionSetting, DSNameReplicationSet, d.Id(), err)
	}

	return nil
}
