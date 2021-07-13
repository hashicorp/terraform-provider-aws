package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsRdsGlobalCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRdsGlobalClusterRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"global_cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"global_cluster_members": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"db_cluster_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_writer": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"global_cluster_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_db_cluster_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsRdsGlobalClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	globalClusterIdentifier := d.Get("global_cluster_identifier").(string)

	params := &rds.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: aws.String(globalClusterIdentifier),
	}
	log.Printf("[DEBUG] Reading Global RDS Cluster: %s", params)
	resp, err := conn.DescribeGlobalClusters(params)

	if err != nil {
		return fmt.Errorf("Error retrieving Global RDS cluster: %w", err)
	}

	if resp == nil {
		return fmt.Errorf("Error retrieving Global RDS cluster: empty response for: %s", params)
	}

	var globalCluster *rds.GlobalCluster
	for _, c := range resp.GlobalClusters {
		if aws.StringValue(c.GlobalClusterIdentifier) == globalClusterIdentifier {
			globalCluster = c
			break
		}
	}

	if globalCluster == nil {
		return fmt.Errorf("Error retrieving Global RDS cluster: cluster not found in response for: %s", params)
	}

	d.SetId(aws.StringValue(globalCluster.GlobalClusterIdentifier))

	d.Set("arn", globalCluster.GlobalClusterArn)
	d.Set("database_name", globalCluster.DatabaseName)
	d.Set("deletion_protection", globalCluster.DeletionProtection)
	d.Set("engine", globalCluster.Engine)
	d.Set("engine_version", globalCluster.EngineVersion)
	d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)

	var gcmList []interface{}
	for _, gcm := range globalCluster.GlobalClusterMembers {
		gcmMap := map[string]interface{}{
			"db_cluster_arn": aws.StringValue(gcm.DBClusterArn),
			"is_writer":      aws.BoolValue(gcm.IsWriter),
		}

		gcmList = append(gcmList, gcmMap)
	}
	if err := d.Set("global_cluster_members", gcmList); err != nil {
		return fmt.Errorf("error setting global_cluster_members: %w", err)
	}

	if err := d.Set("global_cluster_members", flattenRdsGlobalClusterMembers(globalCluster.GlobalClusterMembers)); err != nil {
		return fmt.Errorf("error setting global_cluster_members: %w", err)
	}

	d.Set("global_cluster_resource_id", globalCluster.GlobalClusterResourceId)
	d.Set("storage_encrypted", globalCluster.StorageEncrypted)

	return nil
}
