// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_docdb_orderable_db_instance")
func DataSourceOrderableDBInstance() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrderableDBInstanceRead,
		Schema: map[string]*schema.Schema{
			names.AttrAvailabilityZones: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  engineDocDB,
			},
			names.AttrEngineVersion: {
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
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	input := &docdb.DescribeOrderableDBInstanceOptionsInput{}

	if v, ok := d.GetOk("instance_class"); ok {
		input.DBInstanceClass = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEngine); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("license_model"); ok {
		input.LicenseModel = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc"); ok {
		input.Vpc = aws.Bool(v.(bool))
	}

	var orderableDBInstance *awstypes.OrderableDBInstanceOption
	var err error
	if preferredInstanceClasses := flex.ExpandStringValueList(d.Get("preferred_instance_classes").([]interface{})); len(preferredInstanceClasses) > 0 {
		var orderableDBInstances []awstypes.OrderableDBInstanceOption

		orderableDBInstances, err = findOrderableDBInstances(ctx, conn, input)
		if err == nil {
		PreferredInstanceClassLoop:
			for _, preferredInstanceClass := range preferredInstanceClasses {
				for _, v := range orderableDBInstances {
					if preferredInstanceClass == aws.ToString(v.DBInstanceClass) {
						oi := &v
						orderableDBInstance = oi
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

	d.SetId(aws.ToString(orderableDBInstance.DBInstanceClass))
	d.Set(names.AttrAvailabilityZones, tfslices.ApplyToAll(orderableDBInstance.AvailabilityZones, func(v awstypes.AvailabilityZone) string {
		return aws.ToString(v.Name)
	}))
	d.Set(names.AttrEngine, orderableDBInstance.Engine)
	d.Set(names.AttrEngineVersion, orderableDBInstance.EngineVersion)
	d.Set("instance_class", orderableDBInstance.DBInstanceClass)
	d.Set("license_model", orderableDBInstance.LicenseModel)
	d.Set("vpc", orderableDBInstance.Vpc)

	return diags
}

func findOrderableDBInstance(ctx context.Context, conn *docdb.Client, input *docdb.DescribeOrderableDBInstanceOptionsInput) (*awstypes.OrderableDBInstanceOption, error) {
	output, err := findOrderableDBInstances(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findOrderableDBInstances(ctx context.Context, conn *docdb.Client, input *docdb.DescribeOrderableDBInstanceOptionsInput) ([]awstypes.OrderableDBInstanceOption, error) {
	var output []awstypes.OrderableDBInstanceOption

	pages := docdb.NewDescribeOrderableDBInstanceOptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.OrderableDBInstanceOptions {
			if !reflect.ValueOf(v).IsZero() {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
