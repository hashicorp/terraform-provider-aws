// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_glue_connection")
func DataSourceConnection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectionRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_properties": {
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
				Elem:      &schema.Schema{Type: schema.TypeString},
			},
			"connection_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"match_criteria": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"physical_connection_requirements": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZone: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"security_group_id_list": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).GlueClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id := d.Get(names.AttrID).(string)
	catalogID, connectionName, err := DecodeConnectionID(id)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding Glue Connection %s: %s", id, err)
	}

	connection, err := FindConnectionByName(ctx, conn, connectionName, catalogID)
	if err != nil {
		if tfresource.NotFound(err) {
			return sdkdiag.AppendErrorf(diags, "Glue Connection (%s) not found", id)
		}
		return sdkdiag.AppendErrorf(diags, "reading Glue Connection (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set(names.AttrCatalogID, catalogID)
	d.Set("connection_type", connection.ConnectionType)
	d.Set(names.AttrName, connection.Name)
	d.Set(names.AttrDescription, connection.Description)

	connectionArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("connection/%s", connectionName),
	}.String()
	d.Set(names.AttrARN, connectionArn)

	if err := d.Set("connection_properties", connection.ConnectionProperties); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting connection_properties: %s", err)
	}

	if err := d.Set("physical_connection_requirements", flattenPhysicalConnectionRequirements(connection.PhysicalConnectionRequirements)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting physical_connection_requirements: %s", err)
	}

	if err := d.Set("match_criteria", flex.FlattenStringValueList(connection.MatchCriteria)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting match_criteria: %s", err)
	}

	tags, err := listTags(ctx, conn, connectionArn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Glue Connection (%s): %s", connectionArn, err)
	}

	//lintignore:AWSR002
	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
