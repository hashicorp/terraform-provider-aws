// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @SDKDataSource("aws_timestreamwrite_database", name="Database")
func DataSourceDatabase() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDatabaseRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"table_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameDatabase = "Database Data Source"
)

func dataSourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).TimestreamWriteClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("database_name").(string)

	out, err := findDatabaseByName(ctx, conn, name)
	if err != nil {
		return create.AppendDiagError(diags, names.TimestreamWrite, create.ErrActionReading, DSNameDatabase, name, err)
	}

	d.SetId(aws.ToString(out.DatabaseName))

	d.Set("arn", out.Arn)
	d.Set("database_name", out.DatabaseName)
	d.Set("kms_key_id", out.KmsKeyId)
	d.Set("table_count", out.TableCount)

	tags, err := listTags(ctx, conn, d.Get("arn").(string))
	if err != nil {
		return diag.Errorf("listing tags for timestream table (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return diags
}
