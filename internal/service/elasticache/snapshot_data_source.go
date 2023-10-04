package elasticache

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_elasticache_snapshot")
func DataSourceSnapshot() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSnapshotRead,

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
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
			},
			"replication_group_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"snapshot_source": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"snapshot_status": {
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

func dataSourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	snapshotName, snapshotNameOk := d.GetOk("snapshot_name")
	replicationGroupId, replicationGroupIdOk := d.GetOk("replication_group_id")
	cacheClusterId, cacheClusterIdOk := d.GetOk("cluster_id")

	if !snapshotNameOk && !replicationGroupIdOk && !cacheClusterIdOk {
		return sdkdiag.AppendErrorf(diags, "One of snapshot_name, cluster_id or replication_group_id must be specified")
	}

	params := &elasticache.DescribeSnapshotsInput{}

	if v, ok := d.GetOk("snapshot_source"); ok {
		params.SnapshotSource = aws.String(v.(string))
	}

	if snapshotNameOk {
		params.SnapshotName = aws.String(snapshotName.(string))
	}
	if replicationGroupIdOk {
		params.ReplicationGroupId = aws.String(replicationGroupId.(string))
	}
	if cacheClusterIdOk {
		params.CacheClusterId = aws.String(cacheClusterId.(string))
	}

	log.Printf("[DEBUG] Reading Elasticache Snapshot: %s", params)
	resp, err := conn.DescribeSnapshots(params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error retrieving snapshot details: %s", err)
	}

	if len(resp.Snapshots) < 1 {
		return sdkdiag.AppendErrorf(diags, "No snapshots found matching the specified criteria")
	}

	var snapshot *elasticache.Snapshot
	if len(resp.Snapshots) > 1 {
		recent := d.Get("most_recent").(bool)
		log.Printf("[DEBUG] aws_elasticache_snapshot - multiple results found and `most_recent` is set to: %t", recent)
		if recent {
			snapshot = mostRecentSnapshot(resp.Snapshots)
		} else {
			return sdkdiag.AppendErrorf(diags, "Your query returned more than one result. Please try a more specific search criteria.")
		}
	} else {
		snapshot = resp.Snapshots[0]
	}

	return snapshotDescriptionAttributes(diags, d, snapshot)
}

type elasticacheSnapshotSort []*elasticache.Snapshot

func (a elasticacheSnapshotSort) Len() int      { return len(a) }
func (a elasticacheSnapshotSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a elasticacheSnapshotSort) Less(i, j int) bool {
	// Snapshot creation can be in progress
	if a[i].NodeSnapshots[0].SnapshotCreateTime == nil {
		return true
	}
	if a[j].NodeSnapshots[0].SnapshotCreateTime == nil {
		return false
	}

	return (*a[i].NodeSnapshots[0].SnapshotCreateTime).Before(*a[j].NodeSnapshots[0].SnapshotCreateTime)
}
func mostRecentSnapshot(snapshots []*elasticache.Snapshot) *elasticache.Snapshot {
	sortedSnapshots := snapshots
	sort.Sort(elasticacheSnapshotSort(sortedSnapshots))
	return sortedSnapshots[len(sortedSnapshots)-1]
}

func snapshotDescriptionAttributes(diags diag.Diagnostics, d *schema.ResourceData, snapshot *elasticache.Snapshot) diag.Diagnostics {
	d.SetId(aws.StringValue(snapshot.SnapshotName))
	d.Set("arn", snapshot.ARN)
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
