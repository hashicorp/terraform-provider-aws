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
				Default:  "docdb",
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

	var instanceClassResults []*docdb.OrderableDBInstanceOption

	err := conn.DescribeOrderableDBInstanceOptionsPagesWithContext(ctx, input, func(resp *docdb.DescribeOrderableDBInstanceOptionsOutput, lastPage bool) bool {
		for _, instanceOption := range resp.OrderableDBInstanceOptions {
			if instanceOption == nil {
				continue
			}

			instanceClassResults = append(instanceClassResults, instanceOption)
		}
		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB orderable DB instance options: %s", err)
	}

	if len(instanceClassResults) == 0 {
		return sdkdiag.AppendErrorf(diags, "no DocumentDB Orderable DB Instance options found matching criteria; try different search")
	}

	// preferred classes
	var found *docdb.OrderableDBInstanceOption
	if l := d.Get("preferred_instance_classes").([]interface{}); len(l) > 0 {
		for _, elem := range l {
			preferredInstanceClass, ok := elem.(string)

			if !ok {
				continue
			}

			for _, instanceClassResult := range instanceClassResults {
				if preferredInstanceClass == aws.StringValue(instanceClassResult.DBInstanceClass) {
					found = instanceClassResult
					break
				}
			}

			if found != nil {
				break
			}
		}
	}

	if found == nil && len(instanceClassResults) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple DocumentDB DB Instance Classes (%v) match the criteria; try a different search", instanceClassResults)
	}

	if found == nil && len(instanceClassResults) == 1 {
		found = instanceClassResults[0]
	}

	if found == nil {
		return sdkdiag.AppendErrorf(diags, "no DocumentDB DB Instance Classes match the criteria; try a different search")
	}

	d.SetId(aws.StringValue(found.DBInstanceClass))

	d.Set("instance_class", found.DBInstanceClass)

	var availabilityZones []string
	for _, az := range found.AvailabilityZones {
		availabilityZones = append(availabilityZones, aws.StringValue(az.Name))
	}
	d.Set("availability_zones", availabilityZones)

	d.Set("engine", found.Engine)
	d.Set("engine_version", found.EngineVersion)
	d.Set("license_model", found.LicenseModel)
	d.Set("vpc", found.Vpc)

	return diags
}
