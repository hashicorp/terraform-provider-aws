// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_glue_catalog_database")
func DataSourceCatalogDatabase() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCatalogDatabaseRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"catalog_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"create_table_default_permission": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"target_database": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"catalog_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"database_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCatalogDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	catalogID := createCatalogID(d, meta.(*conns.AWSClient).AccountID)
	name := d.Get("name").(string)

	d.SetId(fmt.Sprintf("%s:%s", catalogID, name))

	input := &glue.GetDatabaseInput{
		CatalogId: aws.String(catalogID),
		Name:      aws.String(name),
	}

	out, err := conn.GetDatabaseWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return diag.Errorf("No Glue Database %s found for catalog_id: %s", name, catalogID)
		}

		return diag.Errorf("reading Glue Catalog Database: %s", err)
	}

	db := out.Database
	dbArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("database/%s", aws.StringValue(db.Name)),
	}.String()
	d.Set("arn", dbArn)
	d.Set("name", db.Name)
	d.Set("catalog_id", catalogID)
	d.Set("description", db.Description)
	d.Set("location_uri", db.LocationUri)
	d.Set("create_table_default_permission", db.CreateTableDefaultPermissions)

	if err := d.Set("parameters", aws.StringValueMap(db.Parameters)); err != nil {
		return diag.Errorf("setting parameters: %s", err)
	}

	if db.TargetDatabase != nil {
		if err := d.Set("target_database", []interface{}{flattenDatabaseTargetDatabase(db.TargetDatabase)}); err != nil {
			return diag.Errorf("setting target_database: %s", err)
		}
	} else {
		d.Set("target_database", nil)
	}

	return nil
}
