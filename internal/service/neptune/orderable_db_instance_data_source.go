// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_neptune_orderable_db_instance")
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
				Default:  engineNeptune,
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
				Default:  "amazon-license",
			},
			"max_iops_per_db_instance": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"max_iops_per_gib": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"max_storage_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"min_iops_per_db_instance": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"min_iops_per_gib": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"min_storage_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"multi_az_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"preferred_instance_classes": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"instance_class"},
			},
			"read_replica_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrStorageType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"supports_enhanced_monitoring": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supports_iam_database_authentication": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supports_iops": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supports_performance_insights": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supports_storage_encryption": {
				Type:     schema.TypeBool,
				Computed: true,
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
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	input := &neptune.DescribeOrderableDBInstanceOptionsInput{}

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

	var orderableDBInstance *neptune.OrderableDBInstanceOption
	var err error
	if preferredInstanceClasses := flex.ExpandStringValueList(d.Get("preferred_instance_classes").([]interface{})); len(preferredInstanceClasses) > 0 {
		var orderableDBInstances []*neptune.OrderableDBInstanceOption

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
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Neptune Orderable DB Instance", err))
	}

	d.SetId(aws.StringValue(orderableDBInstance.DBInstanceClass))
	d.Set(names.AttrAvailabilityZones, tfslices.ApplyToAll(orderableDBInstance.AvailabilityZones, func(v *neptune.AvailabilityZone) string {
		return aws.StringValue(v.Name)
	}))
	d.Set(names.AttrEngine, orderableDBInstance.Engine)
	d.Set(names.AttrEngineVersion, orderableDBInstance.EngineVersion)
	d.Set("license_model", orderableDBInstance.LicenseModel)
	d.Set("max_iops_per_db_instance", orderableDBInstance.MaxIopsPerDbInstance)
	d.Set("max_iops_per_gib", orderableDBInstance.MaxIopsPerGib)
	d.Set("max_storage_size", orderableDBInstance.MaxStorageSize)
	d.Set("min_iops_per_db_instance", orderableDBInstance.MinIopsPerDbInstance)
	d.Set("min_iops_per_gib", orderableDBInstance.MinIopsPerGib)
	d.Set("min_storage_size", orderableDBInstance.MinStorageSize)
	d.Set("multi_az_capable", orderableDBInstance.MultiAZCapable)
	d.Set("instance_class", orderableDBInstance.DBInstanceClass)
	d.Set("read_replica_capable", orderableDBInstance.ReadReplicaCapable)
	d.Set(names.AttrStorageType, orderableDBInstance.StorageType)
	d.Set("supports_enhanced_monitoring", orderableDBInstance.SupportsEnhancedMonitoring)
	d.Set("supports_iam_database_authentication", orderableDBInstance.SupportsIAMDatabaseAuthentication)
	d.Set("supports_iops", orderableDBInstance.SupportsIops)
	d.Set("supports_performance_insights", orderableDBInstance.SupportsPerformanceInsights)
	d.Set("supports_storage_encryption", orderableDBInstance.SupportsStorageEncryption)
	d.Set("vpc", orderableDBInstance.Vpc)

	return diags
}

func findOrderableDBInstance(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeOrderableDBInstanceOptionsInput) (*neptune.OrderableDBInstanceOption, error) {
	output, err := findOrderableDBInstances(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findOrderableDBInstances(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeOrderableDBInstanceOptionsInput) ([]*neptune.OrderableDBInstanceOption, error) {
	var output []*neptune.OrderableDBInstanceOption

	err := conn.DescribeOrderableDBInstanceOptionsPagesWithContext(ctx, input, func(page *neptune.DescribeOrderableDBInstanceOptionsOutput, lastPage bool) bool {
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
