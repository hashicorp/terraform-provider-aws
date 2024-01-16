// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_docdb_orderable_db_instance")
func DataSourceOrderableDBInstance() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrderableDBInstanceRead,
		Schema: map[string]*schema.Schema{
			"availability_zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"engine": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  engineDocDB,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"instance_class": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"preferred_instance_classes"},
			},
			"license_model": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "na",
			},
			"preferred_instance_classes": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"instance_class"},
			},
			"vpc": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceOrderableDBInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	input := &docdb.DescribeOrderableDBInstanceOptionsInput{}

	if v, ok := d.GetOk("instance_class"); ok {
		input.DBInstanceClass = aws.String(v.(string))
	}

	if v, ok := d.GetOk("engine"); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk("engine_version"); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("license_model"); ok {
		input.LicenseModel = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc"); ok {
		input.Vpc = aws.Bool(v.(bool))
	}

	var orderableDBInstance *docdb.OrderableDBInstanceOption
	var err error
	if preferredInstanceClasses := flex.ExpandStringValueList(d.Get("preferred_instance_classes").([]interface{})); len(preferredInstanceClasses) > 0 {
		var orderableDBInstances []*docdb.OrderableDBInstanceOption

		orderableDBInstances, err = findOrderableDBInstances(ctx, conn, input)
		if err == nil {
		PreferredInstanceClassLoop:
			for _, preferredInstanceClass := range preferredInstanceClasses {
				for _, v := range orderableDBInstances {
					if preferredInstanceClass == aws.StringValue(v.DBInstanceClass) {
						orderableDBInstance = v
						break PreferredInstanceClassLoop
					}
				}
			}

			if orderableDBInstance == nil {
				err = tfresource.NewEmptyResultError(input)
			}
		}
	} else {
		orderableDBInstance, err = findOrderableDBInstance(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("DocumentDB Orderable DB Instance", err))
	}

	d.SetId(aws.StringValue(orderableDBInstance.DBInstanceClass))
	d.Set("availability_zones", tfslices.ApplyToAll(orderableDBInstance.AvailabilityZones, func(v *docdb.AvailabilityZone) string {
		return aws.StringValue(v.Name)
	}))
	d.Set("engine", orderableDBInstance.Engine)
	d.Set("engine_version", orderableDBInstance.EngineVersion)
	d.Set("instance_class", orderableDBInstance.DBInstanceClass)
	d.Set("license_model", orderableDBInstance.LicenseModel)
	d.Set("vpc", orderableDBInstance.Vpc)

	return diags
}

func findOrderableDBInstance(ctx context.Context, conn *docdb.DocDB, input *docdb.DescribeOrderableDBInstanceOptionsInput) (*docdb.OrderableDBInstanceOption, error) {
	output, err := findOrderableDBInstances(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findOrderableDBInstances(ctx context.Context, conn *docdb.DocDB, input *docdb.DescribeOrderableDBInstanceOptionsInput) ([]*docdb.OrderableDBInstanceOption, error) {
	var output []*docdb.OrderableDBInstanceOption

	err := conn.DescribeOrderableDBInstanceOptionsPagesWithContext(ctx, input, func(page *docdb.DescribeOrderableDBInstanceOptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.OrderableDBInstanceOptions {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
