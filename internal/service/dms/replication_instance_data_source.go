// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dms_replication_instance", name="Replication Instance")
func dataSourceReplicationInstance() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReplicationInstanceRead,

		Schema: map[string]*schema.Schema{
			names.AttrAllocatedStorage: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrAutoMinorVersionUpgrade: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"multi_az": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"network_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPreferredMaintenanceWindow: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPubliclyAccessible: {
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
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
		},
	}
}

func dataSourceReplicationInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	rID := d.Get("replication_instance_id").(string)
	instance, err := findReplicationInstanceByID(ctx, conn, rID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Replication Instance (%s): %s", rID, err)
	}

	d.SetId(aws.ToString(instance.ReplicationInstanceIdentifier))
	d.Set(names.AttrAllocatedStorage, instance.AllocatedStorage)
	d.Set(names.AttrAutoMinorVersionUpgrade, instance.AutoMinorVersionUpgrade)
	d.Set(names.AttrAvailabilityZone, instance.AvailabilityZone)
	d.Set(names.AttrEngineVersion, instance.EngineVersion)
	d.Set(names.AttrKMSKeyARN, instance.KmsKeyId)
	d.Set("multi_az", instance.MultiAZ)
	d.Set("network_type", instance.NetworkType)
	d.Set(names.AttrPreferredMaintenanceWindow, instance.PreferredMaintenanceWindow)
	d.Set(names.AttrPubliclyAccessible, instance.PubliclyAccessible)
	arn := aws.ToString(instance.ReplicationInstanceArn)
	d.Set("replication_instance_arn", arn)
	d.Set("replication_instance_class", instance.ReplicationInstanceClass)
	d.Set("replication_instance_id", instance.ReplicationInstanceIdentifier)
	d.Set("replication_instance_private_ips", instance.ReplicationInstancePrivateIpAddresses)
	d.Set("replication_instance_public_ips", instance.ReplicationInstancePublicIpAddresses)
	d.Set("replication_subnet_group_id", instance.ReplicationSubnetGroup.ReplicationSubnetGroupIdentifier)
	vpcSecurityGroupIDs := tfslices.ApplyToAll(instance.VpcSecurityGroups, func(sg awstypes.VpcSecurityGroupMembership) string {
		return aws.ToString(sg.VpcSecurityGroupId)
	})
	d.Set(names.AttrVPCSecurityGroupIDs, vpcSecurityGroupIDs)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for DMS Replication Instance (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set(names.AttrTags, tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
