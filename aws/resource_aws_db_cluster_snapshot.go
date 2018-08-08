package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDbClusterSnapshot() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDbClusterSnapshotCreate,
		Read:   resourceAwsDbClusterSnapshotRead,
		Delete: resourceAwsDbClusterSnapshotDelete,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"db_cluster_snapshot_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"allocated_storage": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"availability_zones": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"db_cluster_snapshot_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_encrypted": {
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
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"license_model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"source_db_cluster_snapshot_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsDbClusterSnapshotCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	params := &rds.CreateDBClusterSnapshotInput{
		DBClusterIdentifier:         aws.String(d.Get("db_cluster_identifier").(string)),
		DBClusterSnapshotIdentifier: aws.String(d.Get("db_cluster_snapshot_identifier").(string)),
	}

	_, err := conn.CreateDBClusterSnapshot(params)
	if err != nil {
		return err
	}
	d.SetId(d.Get("db_cluster_snapshot_identifier").(string))

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"available"},
		Refresh:    resourceAwsDbClusterSnapshotStateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutRead),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return resourceAwsDbClusterSnapshotRead(d, meta)
}

func resourceAwsDbClusterSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	params := &rds.DescribeDBClusterSnapshotsInput{
		DBClusterSnapshotIdentifier: aws.String(d.Id()),
	}
	resp, err := conn.DescribeDBClusterSnapshots(params)
	if err != nil {
		return err
	}

	snapshot := resp.DBClusterSnapshots[0]

	d.Set("allocated_storage", snapshot.AllocatedStorage)
	d.Set("availability_zones", flattenStringList(snapshot.AvailabilityZones))

	d.Set("db_cluster_snapshot_arn", snapshot.DBClusterSnapshotArn)
	d.Set("storage_encrypted", snapshot.StorageEncrypted)
	d.Set("engine", snapshot.Engine)
	d.Set("engine_version", snapshot.EngineVersion)
	d.Set("kms_key_id", snapshot.KmsKeyId)
	d.Set("license_model", snapshot.LicenseModel)
	d.Set("port", snapshot.Port)
	d.Set("source_db_cluster_snapshot_arn", snapshot.SourceDBClusterSnapshotArn)
	d.Set("snapshot_type", snapshot.SnapshotType)
	d.Set("status", snapshot.Status)
	d.Set("vpc_id", snapshot.VpcId)

	return nil
}

func resourceAwsDbClusterSnapshotDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	params := &rds.DeleteDBClusterSnapshotInput{
		DBClusterSnapshotIdentifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteDBClusterSnapshot(params)
	if err != nil {
		return err
	}

	return nil
}

func resourceAwsDbClusterSnapshotStateRefreshFunc(
	d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		conn := meta.(*AWSClient).rdsconn

		opts := &rds.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(d.Id()),
		}

		log.Printf("[DEBUG] DB Cluster Snapshot describe configuration: %#v", opts)

		resp, err := conn.DescribeDBClusterSnapshots(opts)
		if err != nil {
			snapshoterr, ok := err.(awserr.Error)
			if ok && snapshoterr.Code() == "DBClusterSnapshotNotFound" {
				return nil, "", nil
			}
			return nil, "", fmt.Errorf("Error retrieving DB Cluster Snapshots: %s", err)
		}

		if len(resp.DBClusterSnapshots) != 1 {
			return nil, "", fmt.Errorf("No snapshots returned for %s", d.Id())
		}

		snapshot := resp.DBClusterSnapshots[0]

		return resp, *snapshot.Status, nil
	}
}
