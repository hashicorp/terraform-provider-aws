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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_memorydb_snapshot", name="Snapshot")
// @Tags(identifierAttribute="arn")
func resourceSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSnapshotCreate,
		ReadWithoutTimeout:   resourceSnapshotRead,
		UpdateWithoutTimeout: resourceSnapshotUpdate,
		DeleteWithoutTimeout: resourceSnapshotDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDescription: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEngine: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEngineVersion: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"maintenance_window": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"node_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"num_shards": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrParameterGroupName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrPort: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"snapshot_retention_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"snapshot_window": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrTopicARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrClusterName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrKMSKeyARN: {
				// The API will accept an ID, but return the ARN on every read.
				// For the sake of consistency, force everyone to use ARN-s.
				// To prevent confusion, the attribute is suffixed _arn rather
				// than the _id implied by the API.
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validateResourceName(snapshotNameMaxLength),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validateResourceNamePrefix(snapshotNameMaxLength - id.UniqueIDSuffixLength),
			},
			names.AttrSource: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &memorydb.CreateSnapshotInput{
		ClusterName:  aws.String(d.Get(names.AttrClusterName).(string)),
		SnapshotName: aws.String(name),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	_, err := conn.CreateSnapshot(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MemoryDB Snapshot (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitSnapshotAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Snapshot (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceSnapshotRead(ctx, d, meta)...)
}

func resourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	snapshot, err := findSnapshotByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MemoryDB Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MemoryDB Snapshot (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, snapshot.ARN)
	if err := d.Set("cluster_configuration", flattenClusterConfiguration(snapshot.ClusterConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cluster_configuration: %s", err)
	}
	d.Set(names.AttrClusterName, snapshot.ClusterConfiguration.Name)
	d.Set(names.AttrKMSKeyARN, snapshot.KmsKeyId)
	d.Set(names.AttrName, snapshot.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(snapshot.Name)))
	d.Set(names.AttrSource, snapshot.Source)

	return diags
}

func resourceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceSnapshotRead(ctx, d, meta)
}

func resourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	log.Printf("[DEBUG] Deleting MemoryDB Snapshot: (%s)", d.Id())
	_, err := conn.DeleteSnapshot(ctx, &memorydb.DeleteSnapshotInput{
		SnapshotName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.SnapshotNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MemoryDB Snapshot (%s): %s", d.Id(), err)
	}

	if _, err := waitSnapshotDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Snapshot (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findSnapshotByName(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.Snapshot, error) {
	input := &memorydb.DescribeSnapshotsInput{
		SnapshotName: aws.String(name),
	}

	return findSnapshot(ctx, conn, input)
}

func findSnapshot(ctx context.Context, conn *memorydb.Client, input *memorydb.DescribeSnapshotsInput) (*awstypes.Snapshot, error) {
	output, err := findSnapshots(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSnapshots(ctx context.Context, conn *memorydb.Client, input *memorydb.DescribeSnapshotsInput) ([]awstypes.Snapshot, error) {
	var output []awstypes.Snapshot

	pages := memorydb.NewDescribeSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.SnapshotNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Snapshots...)
	}

	return output, nil
}

func statusSnapshot(ctx context.Context, conn *memorydb.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findSnapshotByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitSnapshotAvailable(ctx context.Context, conn *memorydb.Client, name string, timeout time.Duration) (*awstypes.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{snapshotStatusCreating},
		Target:  []string{snapshotStatusAvailable},
		Refresh: statusSnapshot(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Snapshot); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotDeleted(ctx context.Context, conn *memorydb.Client, name string, timeout time.Duration) (*awstypes.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{snapshotStatusDeleting},
		Target:  []string{},
		Refresh: statusSnapshot(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Snapshot); ok {
		return output, err
	}

	return nil, err
}

func flattenClusterConfiguration(apiObject *awstypes.ClusterConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrDescription:        aws.ToString(apiObject.Description),
		names.AttrEngine:             aws.ToString(apiObject.Engine),
		names.AttrEngineVersion:      aws.ToString(apiObject.EngineVersion),
		"maintenance_window":         aws.ToString(apiObject.MaintenanceWindow),
		names.AttrName:               aws.ToString(apiObject.Name),
		"node_type":                  aws.ToString(apiObject.NodeType),
		"num_shards":                 aws.ToInt32(apiObject.NumShards),
		names.AttrParameterGroupName: aws.ToString(apiObject.ParameterGroupName),
		names.AttrPort:               aws.ToInt32(apiObject.Port),
		"snapshot_retention_limit":   aws.ToInt32(apiObject.SnapshotRetentionLimit),
		"snapshot_window":            aws.ToString(apiObject.SnapshotWindow),
		"subnet_group_name":          aws.ToString(apiObject.SubnetGroupName),
		names.AttrTopicARN:           aws.ToString(apiObject.TopicArn),
		names.AttrVPCID:              aws.ToString(apiObject.VpcId),
	}

	return []any{tfMap}
}
