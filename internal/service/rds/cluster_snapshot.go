package rds

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const clusterSnapshotCreateTimeout = 2 * time.Minute

// @SDKResource("aws_db_cluster_snapshot", name="Cluster Snapshot")
// @Tags(identifierAttribute="db_cluster_snapshot_arn")
func ResourceClusterSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterSnapshotCreate,
		ReadWithoutTimeout:   resourceClusterSnapshotRead,
		DeleteWithoutTimeout: resourceClusterSnapshotDelete,
		UpdateWithoutTimeout: resourceClusterSnapshotUpdate,
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
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase alphanumeric characters and hyphens"),
					validation.StringMatch(regexp.MustCompile(`^[a-z]`), "must begin with a lowercase letter"),
					validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end with a hyphen"),
				),
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
			"storage_encrypted": {
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()

	input := &rds.CreateDBClusterSnapshotInput{
		DBClusterIdentifier:         aws.String(d.Get("db_cluster_identifier").(string)),
		DBClusterSnapshotIdentifier: aws.String(d.Get("db_cluster_snapshot_identifier").(string)),
		Tags:                        GetTagsIn(ctx),
	}

	err := retry.RetryContext(ctx, clusterSnapshotCreateTimeout, func() *retry.RetryError {
		_, err := conn.CreateDBClusterSnapshotWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, rds.ErrCodeInvalidDBClusterStateFault) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.CreateDBClusterSnapshotWithContext(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Cluster Snapshot: %s", err)
	}
	d.SetId(d.Get("db_cluster_snapshot_identifier").(string))

	stateConf := &retry.StateChangeConf{
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
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Cluster Snapshot %q to create: %s", d.Id(), err)
	}

	return append(diags, resourceClusterSnapshotRead(ctx, d, meta)...)
}

func resourceClusterSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()

	params := &rds.DescribeDBClusterSnapshotsInput{
		DBClusterSnapshotIdentifier: aws.String(d.Id()),
	}
	resp, err := conn.DescribeDBClusterSnapshotsWithContext(ctx, params)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterSnapshotNotFoundFault) {
			log.Printf("[WARN] RDS DB Cluster Snapshot %q not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Cluster Snapshot %q: %s", d.Id(), err)
	}

	if resp == nil || len(resp.DBClusterSnapshots) == 0 || resp.DBClusterSnapshots[0] == nil || aws.StringValue(resp.DBClusterSnapshots[0].DBClusterSnapshotIdentifier) != d.Id() {
		log.Printf("[WARN] RDS DB Cluster Snapshot %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	snapshot := resp.DBClusterSnapshots[0]

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

func resourceClusterSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceClusterSnapshotRead(ctx, d, meta)...)
}

func resourceClusterSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()

	log.Printf("[DEBUG] Deleting RDS DB Cluster Snapshot: %s", d.Id())
	_, err := conn.DeleteDBClusterSnapshotWithContext(ctx, &rds.DeleteDBClusterSnapshotInput{
		DBClusterSnapshotIdentifier: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterSnapshotNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Cluster Snapshot (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceClusterSnapshotStateRefreshFunc(ctx context.Context, dbClusterSnapshotIdentifier string, conn *rds.RDS) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		opts := &rds.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(dbClusterSnapshotIdentifier),
		}

		log.Printf("[DEBUG] DB Cluster Snapshot describe configuration: %#v", opts)

		resp, err := conn.DescribeDBClusterSnapshotsWithContext(ctx, opts)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterSnapshotNotFoundFault) {
				return nil, "", nil
			}
			return nil, "", fmt.Errorf("Error retrieving DB Cluster Snapshots: %s", err)
		}

		if resp == nil || len(resp.DBClusterSnapshots) == 0 || resp.DBClusterSnapshots[0] == nil {
			return nil, "", fmt.Errorf("No snapshots returned for %s", dbClusterSnapshotIdentifier)
		}

		snapshot := resp.DBClusterSnapshots[0]

		return resp, aws.StringValue(snapshot.Status), nil
	}
}
