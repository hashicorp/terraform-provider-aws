// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resourceInstanceResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrUsername: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			names.AttrPassword: {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			names.AttrDeletionProtection: {
				Type:     schema.TypeBool,
				Optional: true,
			},

			names.AttrEngine: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"ca_cert_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"character_set_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			names.AttrAllocatedStorage: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			names.AttrStorageType: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			names.AttrIdentifier: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"identifier_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"instance_class": {
				Type:     schema.TypeString,
				Required: true,
			},

			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"backup_retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"backup_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			names.AttrIOPS: {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"license_model": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"maintenance_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"max_allocated_storage": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"multi_az": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			names.AttrPort: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			names.AttrPubliclyAccessible: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"security_group_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			names.AttrFinalSnapshotIdentifier: {
				Type:     schema.TypeString,
				Optional: true,
			},

			"s3_import": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucketName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrBucketPrefix: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"ingestion_role": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"source_engine": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"source_engine_version": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},

			"skip_final_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"copy_tags_to_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			names.AttrParameterGroupName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			names.AttrAddress: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrApplyImmediately: {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"replicate_source_db": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"replicas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"snapshot_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			names.AttrAutoMinorVersionUpgrade: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			names.AttrAllowMajorVersionUpgrade: {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"monitoring_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"monitoring_interval": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},

			"option_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"timezone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"iam_database_authentication_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			names.AttrResourceID: {
				Type:     schema.TypeString,
				Computed: true,
			},

			"enabled_cloudwatch_logs_exports": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			names.AttrDomain: {
				Type:     schema.TypeString,
				Optional: true,
			},

			"domain_iam_role_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"performance_insights_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"performance_insights_kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"performance_insights_retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			names.AttrTags: tftags.TagsSchema(),
		},
	}
}

func instanceStateUpgradeV0(_ context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState == nil {
		return nil, nil
	}

	rawState["delete_automated_backups"] = true

	return rawState, nil
}

func resourceInstanceResourceV1() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrAddress: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAllocatedStorage: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrAllowMajorVersionUpgrade: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			// apply_immediately is used to determine when the update modifications
			// take place.
			// See http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html
			names.AttrApplyImmediately: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAutoMinorVersionUpgrade: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"backup_retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"backup_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"blue_green_update": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"ca_cert_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"character_set_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"copy_tags_to_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"custom_iam_instance_profile": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"customer_owned_ip_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"db_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"delete_automated_backups": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrDeletionProtection: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrDomain: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain_iam_role_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled_cloudwatch_logs_exports": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"engine_version_actual": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFinalSnapshotIdentifier: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_database_authentication_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrIdentifier: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"identifier_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"instance_class": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrIOPS: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"latest_restorable_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"license_model": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"listener_endpoint": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAddress: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrHostedZoneID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrPort: {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"maintenance_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"manage_master_user_password": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"master_user_secret": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"secret_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"secret_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"master_user_secret_kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"max_allocated_storage": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"monitoring_interval": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"monitoring_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"multi_az": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"nchar_character_set_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"network_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"option_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrParameterGroupName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrPassword: {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"performance_insights_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"performance_insights_kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"performance_insights_retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrPubliclyAccessible: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"replica_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"replicas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"replicate_source_db": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"restore_to_point_in_time": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"restore_time": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source_db_instance_automated_backups_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source_db_instance_identifier": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source_dbi_resource_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"use_latest_restorable_time": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"s3_import": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucketName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrBucketPrefix: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"ingestion_role": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"source_engine": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"source_engine_version": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"skip_final_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"snapshot_identifier": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"storage_throughput": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrStorageType: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"timezone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrUsername: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func instanceStateUpgradeV1(_ context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState == nil {
		return nil, nil
	}

	rawState[names.AttrID] = rawState[names.AttrResourceID]

	return rawState, nil
}
