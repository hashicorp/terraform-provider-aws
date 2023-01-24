package neptune

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func ResourceClusterSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterSnapshotCreate,
		ReadWithoutTimeout:   resourceClusterSnapshotRead,
		DeleteWithoutTimeout: resourceClusterSnapshotDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
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
			"storage_encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceClusterSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn()

	input := &neptune.CreateDBClusterSnapshotInput{
		DBClusterIdentifier:         aws.String(d.Get("db_cluster_identifier").(string)),
		DBClusterSnapshotIdentifier: aws.String(d.Get("db_cluster_snapshot_identifier").(string)),
	}

	log.Printf("[DEBUG] Creating Neptune DB Cluster Snapshot: %s", input)
	_, err := conn.CreateDBClusterSnapshotWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune DB Cluster Snapshot: %s", err)
	}
	d.SetId(d.Get("db_cluster_snapshot_identifier").(string))

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"available"},
		Refresh:    resourceClusterSnapshotStateRefreshFunc(ctx, d.Id(), conn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      5 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune DB Cluster Snapshot %q to create: %s", d.Id(), err)
	}

	return append(diags, resourceClusterSnapshotRead(ctx, d, meta)...)
}

func resourceClusterSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn()

	input := &neptune.DescribeDBClusterSnapshotsInput{
		DBClusterSnapshotIdentifier: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Neptune DB Cluster Snapshot: %s", input)
	output, err := conn.DescribeDBClusterSnapshotsWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterSnapshotNotFoundFault) {
			log.Printf("[WARN] Neptune DB Cluster Snapshot %q not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Neptune DB Cluster Snapshot %q: %s", d.Id(), err)
	}

	if output == nil || len(output.DBClusterSnapshots) == 0 || output.DBClusterSnapshots[0] == nil || aws.StringValue(output.DBClusterSnapshots[0].DBClusterSnapshotIdentifier) != d.Id() {
		log.Printf("[WARN] Neptune DB Cluster Snapshot %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	snapshot := output.DBClusterSnapshots[0]

	d.Set("allocated_storage", snapshot.AllocatedStorage)
	if err := d.Set("availability_zones", flex.FlattenStringList(snapshot.AvailabilityZones)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting availability_zones: %s", err)
	}
	d.Set("db_cluster_identifier", snapshot.DBClusterIdentifier)
	d.Set("db_cluster_snapshot_arn", snapshot.DBClusterSnapshotArn)
	d.Set("db_cluster_snapshot_identifier", snapshot.DBClusterSnapshotIdentifier)
	d.Set("engine_version", snapshot.EngineVersion)
	d.Set("engine", snapshot.Engine)
	d.Set("kms_key_id", snapshot.KmsKeyId)
	d.Set("license_model", snapshot.LicenseModel)
	d.Set("port", snapshot.Port)
	d.Set("snapshot_type", snapshot.SnapshotType)
	d.Set("source_db_cluster_snapshot_arn", snapshot.SourceDBClusterSnapshotArn)
	d.Set("status", snapshot.Status)
	d.Set("storage_encrypted", snapshot.StorageEncrypted)
	d.Set("vpc_id", snapshot.VpcId)

	return diags
}

func resourceClusterSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn()

	input := &neptune.DeleteDBClusterSnapshotInput{
		DBClusterSnapshotIdentifier: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Neptune DB Cluster Snapshot: %s", input)
	_, err := conn.DeleteDBClusterSnapshotWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterSnapshotNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Neptune DB Cluster Snapshot %q: %s", d.Id(), err)
	}

	return diags
}

func resourceClusterSnapshotStateRefreshFunc(ctx context.Context, dbClusterSnapshotIdentifier string, conn *neptune.Neptune) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &neptune.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(dbClusterSnapshotIdentifier),
		}

		log.Printf("[DEBUG] Reading Neptune DB Cluster Snapshot: %s", input)
		output, err := conn.DescribeDBClusterSnapshotsWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterSnapshotNotFoundFault) {
				return nil, "", nil
			}
			return nil, "", fmt.Errorf("Error retrieving DB Cluster Snapshots: %s", err)
		}

		if output == nil || len(output.DBClusterSnapshots) == 0 || output.DBClusterSnapshots[0] == nil {
			return nil, "", fmt.Errorf("No snapshots returned for %s", dbClusterSnapshotIdentifier)
		}

		snapshot := output.DBClusterSnapshots[0]

		return output, aws.StringValue(snapshot.Status), nil
	}
}
