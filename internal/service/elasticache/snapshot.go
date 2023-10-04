package elasticache

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elasticache_snapshot", name="Snapshot")
// @Tags(identifierAttribute="arn")
func ResourceSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSnapshotCreate,
		ReadWithoutTimeout:   resourceSnapshotRead,
		DeleteWithoutTimeout: resourceSnapshotDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			// Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automatic_failover": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cluster_create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
			"node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"num_cache_nodes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"parameter_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"subnet_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"replication_group_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"snapshot_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_status": {
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

const (
	ResNameSnapshot = "Snapshot"
)

func resourceSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	snapshotName := d.Get("snapshot_name").(string)

	replicationGroupId, replicationGroupIdOk := d.GetOk("replication_group_id")
	cacheClusterId, cacheClusterIdOk := d.GetOk("cluster_id")

	if !replicationGroupIdOk && !cacheClusterIdOk {
		return sdkdiag.AppendErrorf(diags, "Only one of cluster_id or replication_group_id must be specified")
	}

	in := &elasticache.CreateSnapshotInput{
		SnapshotName: aws.String(snapshotName),
		Tags:         getTagsIn(ctx),
	}

	if cacheClusterIdOk {
		in.CacheClusterId = aws.String(cacheClusterId.(string))
	}
	if replicationGroupIdOk {
		in.ReplicationGroupId = aws.String(replicationGroupId.(string))
	}
	out, err := conn.CreateSnapshotWithContext(ctx, in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error creating AWS Elasticache Snapshot %s: %s", snapshotName, err)
	}
	d.SetId(aws.StringValue(out.Snapshot.SnapshotName))

	if _, err := waitSnapshotCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "Error waiting for AWS Elasticache Snapshot creation %s: %s", snapshotName, err)
	}

	return append(diags, resourceSnapshotRead(ctx, d, meta)...)
}

func resourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	snapshot, err := findSnapshotByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Unable to read Elasticace Snapshot %s: %s", d.Id(), err)
	}

	arn := aws.StringValue(snapshot.ARN)
	d.Set("arn", arn)
	d.Set("automatic_failover", snapshot.AutomaticFailover)
	d.Set("auto_minor_version_upgrade", snapshot.AutoMinorVersionUpgrade)
	d.Set("cluster_create_time", aws.TimeValue(snapshot.CacheClusterCreateTime).Format(time.RFC3339))
	d.Set("cluster_id", snapshot.CacheClusterId)
	d.Set("engine", snapshot.Engine)
	d.Set("engine_version", snapshot.EngineVersion)
	d.Set("kms_key_id", snapshot.KmsKeyId)
	d.Set("node_type", snapshot.CacheNodeType)
	d.Set("node_type", snapshot.CacheNodeType)
	d.Set("num_cache_nodes", snapshot.NumCacheNodes)
	d.Set("parameter_group_name", snapshot.CacheParameterGroupName)
	d.Set("port", snapshot.Port)
	d.Set("subnet_group_name", snapshot.CacheSubnetGroupName)
	d.Set("replication_group_id", snapshot.ReplicationGroupId)
	d.Set("replication_group_description", snapshot.ReplicationGroupDescription)
	d.Set("snapshot_name", snapshot.SnapshotName)
	d.Set("snapshot_source", snapshot.SnapshotSource)
	d.Set("snapshot_status", snapshot.SnapshotStatus)
	d.Set("vpc_id", snapshot.VpcId)

	return diags
}

func resourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	log.Printf("[INFO] Deleting ElastiCache Snapshot %s", d.Id())

	_, err := conn.DeleteSnapshot(&elasticache.DeleteSnapshotInput{
		SnapshotName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeSnapshotNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error deleting Elasticache Snapshot (%s): %s", d.Id(), err)
	}

	return diags
}

const (
	statusChangePending = "creating"
	statusDeleting      = "deleting"
	statusNormal        = "available"
)

func waitSnapshotCreated(ctx context.Context, conn *elasticache.ElastiCache, id string, timeout time.Duration) (*elasticache.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusSnapshot(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*elasticache.Snapshot); ok {
		return out, err
	}

	return nil, err
}

func statusSnapshot(conn *elasticache.ElastiCache, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findSnapshotByID(conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.StringValue(out.SnapshotStatus), nil
	}
}

func findSnapshotByID(conn *elasticache.ElastiCache, id string) (*elasticache.Snapshot, error) {
	in := &elasticache.DescribeSnapshotsInput{
		SnapshotName: aws.String(id),
	}
	out, err := conn.DescribeSnapshots(in)
	if len(out.Snapshots) == 0 {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Snapshots[0] == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Snapshots[0], nil
}
