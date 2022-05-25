package redshift

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"allow_version_upgrade": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"aqua_configuration_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automated_snapshot_retention_period": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_relocation_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"bucket_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"node_role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"cluster_parameter_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_revision_number": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_security_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"cluster_subnet_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_iam_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"elastic_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_logging": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enhanced_vpc_routing": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"iam_roles": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maintenance_track_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"manual_snapshot_retention_period": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"number_of_nodes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
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
			"s3_key_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"log_destination_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"log_exports": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchema(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_security_group_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	clusterID := d.Get("cluster_identifier").(string)
	rsc, err := FindClusterByID(conn, clusterID)

	if err != nil {
		return fmt.Errorf("reading Redshift Cluster (%s): %w", clusterID, err)
	}

	d.SetId(clusterID)
	d.Set("allow_version_upgrade", rsc.AllowVersionUpgrade)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   redshift.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("cluster:%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("automated_snapshot_retention_period", rsc.AutomatedSnapshotRetentionPeriod)
	if rsc.AquaConfiguration != nil {
		d.Set("aqua_configuration_status", rsc.AquaConfiguration.AquaConfigurationStatus)
	}
	d.Set("availability_zone", rsc.AvailabilityZone)
	azr, err := clusterAvailabilityZoneRelocationStatus(rsc)
	if err != nil {
		return err
	}
	d.Set("availability_zone_relocation_enabled", azr)
	d.Set("cluster_identifier", rsc.ClusterIdentifier)
	if err := d.Set("cluster_nodes", flattenClusterNodes(rsc.ClusterNodes)); err != nil {
		return fmt.Errorf("setting cluster_nodes: %w", err)
	}

	if len(rsc.ClusterParameterGroups) > 0 {
		d.Set("cluster_parameter_group_name", rsc.ClusterParameterGroups[0].ParameterGroupName)
	}

	d.Set("cluster_public_key", rsc.ClusterPublicKey)
	d.Set("cluster_revision_number", rsc.ClusterRevisionNumber)

	var csg []string
	for _, g := range rsc.ClusterSecurityGroups {
		csg = append(csg, aws.StringValue(g.ClusterSecurityGroupName))
	}
	d.Set("cluster_security_groups", csg)

	d.Set("cluster_subnet_group_name", rsc.ClusterSubnetGroupName)

	if len(rsc.ClusterNodes) > 1 {
		d.Set("cluster_type", clusterTypeMultiNode)
	} else {
		d.Set("cluster_type", clusterTypeSingleNode)
	}

	d.Set("cluster_version", rsc.ClusterVersion)
	d.Set("database_name", rsc.DBName)

	if rsc.ElasticIpStatus != nil {
		d.Set("elastic_ip", rsc.ElasticIpStatus.ElasticIp)
	}

	d.Set("encrypted", rsc.Encrypted)

	if rsc.Endpoint != nil {
		d.Set("endpoint", rsc.Endpoint.Address)
	}

	d.Set("enhanced_vpc_routing", rsc.EnhancedVpcRouting)

	var iamRoles []string
	for _, i := range rsc.IamRoles {
		iamRoles = append(iamRoles, aws.StringValue(i.IamRoleArn))
	}
	d.Set("iam_roles", iamRoles)

	d.Set("kms_key_id", rsc.KmsKeyId)
	d.Set("master_username", rsc.MasterUsername)
	d.Set("node_type", rsc.NodeType)
	d.Set("number_of_nodes", rsc.NumberOfNodes)
	d.Set("port", rsc.Endpoint.Port)
	d.Set("preferred_maintenance_window", rsc.PreferredMaintenanceWindow)
	d.Set("publicly_accessible", rsc.PubliclyAccessible)
	d.Set("default_iam_role_arn", rsc.DefaultIamRoleArn)
	d.Set("maintenance_track_name", rsc.MaintenanceTrackName)
	d.Set("manual_snapshot_retention_period", rsc.ManualSnapshotRetentionPeriod)

	if err := d.Set("tags", KeyValueTags(rsc.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	d.Set("vpc_id", rsc.VpcId)

	var vpcg []string
	for _, g := range rsc.VpcSecurityGroups {
		vpcg = append(vpcg, aws.StringValue(g.VpcSecurityGroupId))
	}
	d.Set("vpc_security_group_ids", vpcg)

	loggingStatus, err := conn.DescribeLoggingStatus(&redshift.DescribeLoggingStatusInput{
		ClusterIdentifier: aws.String(clusterID),
	})

	if err != nil {
		return fmt.Errorf("reading Redshift Cluster (%s) logging status: %w", d.Id(), err)
	}

	if loggingStatus != nil && aws.BoolValue(loggingStatus.LoggingEnabled) {
		d.Set("enable_logging", loggingStatus.LoggingEnabled)
		d.Set("bucket_name", loggingStatus.BucketName)
		d.Set("s3_key_prefix", loggingStatus.S3KeyPrefix)
		d.Set("log_exports", flex.FlattenStringSet(loggingStatus.LogExports))
		d.Set("log_destination_type", loggingStatus.LogDestinationType)
	}

	return nil
}
