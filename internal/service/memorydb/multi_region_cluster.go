// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/memorydb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/memorydb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_memorydb_multi_region_cluster", name="MultiRegionCluster")
// @Tags(identifierAttribute="arn")
func resourceMultiRegionCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMultiRegionClusterCreate,
		ReadWithoutTimeout:   resourceMultiRegionClusterRead,
		UpdateWithoutTimeout: resourceMultiRegionClusterUpdate,
		DeleteWithoutTimeout: resourceMultiRegionClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "Managed by Terraform",
				},
				names.AttrEngine: {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: enum.Validate[clusterEngine](),
				},
				names.AttrEngineVersion: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"multi_region_cluster_name": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"multi_region_cluster_name_suffix": {
					Type:     schema.TypeString,
					Required: true,
				},
				"multi_region_parameter_group_name": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"node_type": {
					Type:     schema.TypeString,
					Required: true,
				},
				"num_shards": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"status": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"tls_enabled": {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true,
					Default:  true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},
	}
}

func resourceMultiRegionClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	name := create.Name(d.Get("multi_region_cluster_name").(string), "")
	input := &memorydb.CreateMultiRegionClusterInput{
		MultiRegionClusterName:       aws.String(name),
		MultiRegionClusterNameSuffix: aws.String(d.Get("multi_region_cluster_name_suffix").(string)),
		NodeType:                     aws.String(d.Get("node_type").(string)),
		TLSEnabled:                   aws.Bool(d.Get("tls_enabled").(bool)),
		Tags:                         getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEngine); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("multi_region_parameter_group_name"); ok {
		input.MultiRegionParameterGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("num_shards"); ok {
		input.Port = aws.Int32(int32(v.(int)))
	}

	_, err := conn.CreateMultiRegionCluster(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MemoryDB Multi Region cluster (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitMultiRegionClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Multi Region cluster (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceMultiRegionClusterRead(ctx, d, meta)...)
}

func resourceMultiRegionClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	cluster, err := findMultiRegionClusterByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MemoryDB Multi Region cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MemoryDB Multi Region cluster (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, cluster.ARN)
	d.Set(names.AttrDescription, cluster.Description)
	d.Set(names.AttrEngine, cluster.Engine)
	d.Set(names.AttrEngineVersion, cluster.EngineVersion)
	d.Set("multi_region_cluster_name", cluster.MultiRegionClusterName)
	d.Set("node_type", cluster.NodeType)
	d.Set("num_shards", cluster.NumberOfShards)
	d.Set("multi_region_parameter_group_name", cluster.MultiRegionParameterGroupName)
	d.Set("tls_enabled", cluster.TLSEnabled)
	d.Set(names.AttrStatus, cluster.Status)

	return diags
}

func resourceMultiRegionClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	if d.HasChangesExcept("final_snapshot_name", names.AttrTags, names.AttrTagsAll) {
		waitParameterGroupInSync := false
		waitSecurityGroupsActive := false

		input := &memorydb.UpdateMultiRegionClusterInput{
			MultiRegionClusterName: aws.String(d.Id()),
		}

		if d.HasChange("acl_name") {
			input.ACLName = aws.String(d.Get("acl_name").(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrEngine) {
			input.Engine = aws.String(d.Get(names.AttrEngine).(string))
		}

		if d.HasChange(names.AttrEngineVersion) {
			input.EngineVersion = aws.String(d.Get(names.AttrEngineVersion).(string))
		}

		if d.HasChange("maintenance_window") {
			input.MaintenanceWindow = aws.String(d.Get("maintenance_window").(string))
		}

		if d.HasChange("node_type") {
			input.NodeType = aws.String(d.Get("node_type").(string))
		}

		if d.HasChange("num_replicas_per_shard") {
			input.ReplicaConfiguration = &awstypes.ReplicaConfigurationRequest{
				ReplicaCount: int32(d.Get("num_replicas_per_shard").(int)),
			}
		}

		if d.HasChange("num_shards") {
			input.ShardConfiguration = &awstypes.ShardConfigurationRequest{
				ShardCount: int32(d.Get("num_shards").(int)),
			}
		}

		if d.HasChange(names.AttrParameterGroupName) {
			input.ParameterGroupName = aws.String(d.Get(names.AttrParameterGroupName).(string))
			waitParameterGroupInSync = true
		}

		if d.HasChange(names.AttrSecurityGroupIDs) {
			// UpdateMultiRegionCluster reads null and empty slice as "no change", so once
			// at least one security group is present, it's no longer possible
			// to remove all of them.

			v := d.Get(names.AttrSecurityGroupIDs).(*schema.Set)

			if v.Len() == 0 {
				return sdkdiag.AppendErrorf(diags, "unable to update MemoryDB MultiRegionCluster (%s): removing all security groups is not possible", d.Id())
			}

			input.SecurityGroupIds = flex.ExpandStringValueSet(v)
			waitSecurityGroupsActive = true
		}

		if d.HasChange("snapshot_retention_limit") {
			input.SnapshotRetentionLimit = aws.Int32(int32(d.Get("snapshot_retention_limit").(int)))
		}

		if d.HasChange("snapshot_window") {
			input.SnapshotWindow = aws.String(d.Get("snapshot_window").(string))
		}

		if d.HasChange(names.AttrSNSTopicARN) {
			v := d.Get(names.AttrSNSTopicARN).(string)
			input.SnsTopicArn = aws.String(v)
			if v == "" {
				input.SnsTopicStatus = aws.String(clusterSNSTopicStatusInactive)
			} else {
				input.SnsTopicStatus = aws.String(clusterSNSTopicStatusActive)
			}
		}

		_, err := conn.UpdateMultiRegionCluster(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MemoryDB MultiRegionCluster (%s): %s", d.Id(), err)
		}

		if _, err := waitMultiRegionClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB MultiRegionCluster (%s) update: %s", d.Id(), err)
		}

		if waitParameterGroupInSync {
			if _, err := waitMultiRegionClusterParameterGroupInSync(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB MultiRegionCluster (%s) parameter group: %s", d.Id(), err)
			}
		}

		if waitSecurityGroupsActive {
			if _, err := waitMultiRegionClusterSecurityGroupsActive(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB MultiRegionCluster (%s) security groups: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceMultiRegionClusterRead(ctx, d, meta)...)
}

func resourceMultiRegionClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	input := &memorydb.DeleteMultiRegionClusterInput{
		MultiRegionClusterName: aws.String(d.Id()),
	}

	if v := d.Get("final_snapshot_name"); v != nil && len(v.(string)) > 0 {
		input.FinalSnapshotName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting MemoryDB MultiRegionCluster: (%s)", d.Id())
	_, err := conn.DeleteMultiRegionCluster(ctx, input)

	if errs.IsA[*awstypes.MultiRegionClusterNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MemoryDB MultiRegionCluster (%s): %s", d.Id(), err)
	}

	if _, err := waitMultiRegionClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB MultiRegionCluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findMultiRegionClusterByName(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.MultiRegionCluster, error) {
	input := &memorydb.DescribeMultiRegionClustersInput{
		MultiRegionClusterName: aws.String(name),
		ShowClusterDetails:     aws.Bool(true),
	}

	return findMultiRegionCluster(ctx, conn, input)
}

func findMultiRegionCluster(ctx context.Context, conn *memorydb.Client, input *memorydb.DescribeMultiRegionClustersInput) (*awstypes.MultiRegionCluster, error) {
	output, err := findMultiRegionClusters(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findMultiRegionClusters(ctx context.Context, conn *memorydb.Client, input *memorydb.DescribeMultiRegionClustersInput) ([]awstypes.MultiRegionCluster, error) {
	var output []awstypes.MultiRegionCluster

	pages := memorydb.NewDescribeMultiRegionClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.MultiRegionClusterNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.MultiRegionClusters...)
	}

	return output, nil
}

func statusMultiRegionCluster(ctx context.Context, conn *memorydb.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findMultiRegionClusterByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func statusMultiRegionClusterParameterGroup(ctx context.Context, conn *memorydb.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findMultiRegionClusterByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.ParameterGroupStatus), nil
	}
}

func waitMultiRegionClusterAvailable(ctx context.Context, conn *memorydb.Client, name string, timeout time.Duration) (*awstypes.MultiRegionCluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterStatusCreating, clusterStatusUpdating, clusterStatusSnapshotting},
		Target:  []string{clusterStatusAvailable},
		Refresh: statusMultiRegionCluster(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.MultiRegionCluster); ok {
		return output, err
	}

	return nil, err
}

func waitMultiRegionClusterDeleted(ctx context.Context, conn *memorydb.Client, name string, timeout time.Duration) (*awstypes.MultiRegionCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterStatusDeleting},
		Target:  []string{},
		Refresh: statusMultiRegionCluster(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.MultiRegionCluster); ok {
		return output, err
	}

	return nil, err
}

func waitMultiRegionClusterParameterGroupInSync(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.MultiRegionCluster, error) {
	const (
		timeout = 60 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterParameterGroupStatusApplying},
		Target:  []string{clusterParameterGroupStatusInSync},
		Refresh: statusMultiRegionClusterParameterGroup(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.MultiRegionCluster); ok {
		return output, err
	}

	return nil, err
}
