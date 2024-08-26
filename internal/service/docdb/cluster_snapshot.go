// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_docdb_cluster_snapshot")
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
			names.AttrAvailabilityZones: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"db_cluster_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterIdentifier,
			},
			"db_cluster_snapshot_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_cluster_snapshot_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterSnapshotIdentifier,
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"snapshot_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_db_cluster_snapshot_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceClusterSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	clusterSnapshotID := d.Get("db_cluster_snapshot_identifier").(string)
	input := &docdb.CreateDBClusterSnapshotInput{
		DBClusterIdentifier:         aws.String(d.Get("db_cluster_identifier").(string)),
		DBClusterSnapshotIdentifier: aws.String(clusterSnapshotID),
	}

	_, err := conn.CreateDBClusterSnapshot(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DocumentDB Cluster Snapshot (%s): %s", clusterSnapshotID, err)
	}

	d.SetId(clusterSnapshotID)

	if _, err := waitClusterSnapshotCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster Snapshot (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceClusterSnapshotRead(ctx, d, meta)...)
}

func resourceClusterSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	snapshot, err := findClusterSnapshotByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DocumentDB Cluster Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Cluster Snapshot (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAvailabilityZones, snapshot.AvailabilityZones)
	d.Set("db_cluster_identifier", snapshot.DBClusterIdentifier)
	d.Set("db_cluster_snapshot_arn", snapshot.DBClusterSnapshotArn)
	d.Set("db_cluster_snapshot_identifier", snapshot.DBClusterSnapshotIdentifier)
	d.Set(names.AttrEngineVersion, snapshot.EngineVersion)
	d.Set(names.AttrEngine, snapshot.Engine)
	d.Set(names.AttrKMSKeyID, snapshot.KmsKeyId)
	d.Set(names.AttrPort, snapshot.Port)
	d.Set("snapshot_type", snapshot.SnapshotType)
	d.Set("source_db_cluster_snapshot_arn", snapshot.SourceDBClusterSnapshotArn)
	d.Set(names.AttrStatus, snapshot.Status)
	d.Set(names.AttrStorageEncrypted, snapshot.StorageEncrypted)
	d.Set(names.AttrVPCID, snapshot.VpcId)

	return diags
}

func resourceClusterSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	log.Printf("[DEBUG] Deleting DocumentDB Cluster Snapshot: %s", d.Id())
	_, err := conn.DeleteDBClusterSnapshot(ctx, &docdb.DeleteDBClusterSnapshotInput{
		DBClusterSnapshotIdentifier: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.DBClusterSnapshotNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DocumentDB Cluster Snapshot (%s): %s", d.Id(), err)
	}

	return diags
}

func findClusterSnapshotByID(ctx context.Context, conn *docdb.Client, id string) (*awstypes.DBClusterSnapshot, error) {
	input := &docdb.DescribeDBClusterSnapshotsInput{
		DBClusterSnapshotIdentifier: aws.String(id),
	}
	output, err := findClusterSnapshot(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DBClusterSnapshotIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findClusterSnapshot(ctx context.Context, conn *docdb.Client, input *docdb.DescribeDBClusterSnapshotsInput) (*awstypes.DBClusterSnapshot, error) {
	output, err := findClusterSnapshots(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClusterSnapshots(ctx context.Context, conn *docdb.Client, input *docdb.DescribeDBClusterSnapshotsInput) ([]awstypes.DBClusterSnapshot, error) {
	var output []awstypes.DBClusterSnapshot

	pages := docdb.NewDescribeDBClusterSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.DBClusterSnapshotNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusterSnapshots {
			if !reflect.ValueOf(v).IsZero() {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusClusterSnapshot(ctx context.Context, conn *docdb.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findClusterSnapshotByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitClusterSnapshotCreated(ctx context.Context, conn *docdb.Client, id string, timeout time.Duration) (*awstypes.DBClusterSnapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterSnapshotStatusCreating},
		Target:     []string{clusterSnapshotStatusAvailable},
		Refresh:    statusClusterSnapshot(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DBClusterSnapshot); ok {
		return output, err
	}

	return nil, err
}
