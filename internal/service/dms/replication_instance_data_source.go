// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dms_replication_instance")
func DataSourceReplicationInstance() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReplicationInstanceRead,

		Schema: map[string]*schema.Schema{
			"allocated_storage": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"multi_az": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"preferred_maintenance_window": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"replication_instance_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_instance_class": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"replication_instance_private_ips": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"replication_instance_public_ips": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"replication_subnet_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameReplicationInstance = "Replication Instance Data Source"
)

func dataSourceReplicationInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DMSConn(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	rID := d.Get("replication_instance_id").(string)

	instance, err := FindReplicationInstanceByID(ctx, conn, rID)
	if err != nil {
		return create.DiagError(names.DMS, create.ErrActionReading, DSNameReplicationTask, d.Id(), err)
	}

	d.SetId(aws.StringValue(instance.ReplicationInstanceIdentifier))

	d.Set("allocated_storage", instance.AllocatedStorage)
	d.Set("auto_minor_version_upgrade", instance.AutoMinorVersionUpgrade)
	d.Set("availability_zone", instance.AvailabilityZone)
	d.Set("engine_version", instance.EngineVersion)
	d.Set("kms_key_arn", instance.KmsKeyId)
	d.Set("multi_az", instance.MultiAZ)
	d.Set("preferred_maintenance_window", instance.PreferredMaintenanceWindow)
	d.Set("publicly_accessible", instance.PubliclyAccessible)
	d.Set("replication_instance_arn", instance.ReplicationInstanceArn)
	d.Set("replication_instance_class", instance.ReplicationInstanceClass)
	d.Set("replication_instance_id", instance.ReplicationInstanceIdentifier)

	if err := d.Set("replication_instance_private_ips", aws.StringValueSlice(instance.ReplicationInstancePrivateIpAddresses)); err != nil {
		return create.DiagError(names.DMS, create.ErrActionReading, DSNameReplicationTask, d.Id(), err)
	}

	if err := d.Set("replication_instance_public_ips", aws.StringValueSlice(instance.ReplicationInstancePublicIpAddresses)); err != nil {
		return create.DiagError(names.DMS, create.ErrActionReading, DSNameReplicationTask, d.Id(), err)
	}

	d.Set("replication_subnet_group_id", instance.ReplicationSubnetGroup.ReplicationSubnetGroupIdentifier)

	vpc_security_group_ids := []string{}
	for _, sg := range instance.VpcSecurityGroups {
		vpc_security_group_ids = append(vpc_security_group_ids, aws.StringValue(sg.VpcSecurityGroupId))
	}

	if err := d.Set("vpc_security_group_ids", vpc_security_group_ids); err != nil {
		return create.DiagError(names.DMS, create.ErrActionReading, DSNameReplicationTask, d.Id(), err)
	}

	tags, err := listTags(ctx, conn, aws.StringValue(instance.ReplicationInstanceArn))

	if err != nil {
		return create.DiagError(names.DMS, create.ErrActionReading, DSNameReplicationInstance, d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.DMS, create.ErrActionSetting, DSNameReplicationInstance, d.Id(), err)
	}

	return nil
}

func FindReplicationInstanceByID(ctx context.Context, conn *dms.DatabaseMigrationService, id string) (*dms.ReplicationInstance, error) {
	input := &dms.DescribeReplicationInstancesInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("replication-instance-id"),
				Values: []*string{aws.String(id)},
			},
		},
	}
	response, err := conn.DescribeReplicationInstancesWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if response == nil || len(response.ReplicationInstances) == 0 || response.ReplicationInstances[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return response.ReplicationInstances[0], nil
}
