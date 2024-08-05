// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_memorydb_snapshot", name="Snapshot")
// @Tags(identifierAttribute="arn")
func ResourceSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSnapshotCreate,
		ReadWithoutTimeout:   resourceSnapshotRead,
		UpdateWithoutTimeout: resourceSnapshotUpdate,
		DeleteWithoutTimeout: resourceSnapshotDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(snapshotAvailableTimeout),
			Delete: schema.DefaultTimeout(snapshotDeletedTimeout),
		},

		CustomizeDiff: verify.SetTagsDiff,

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

func resourceSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &memorydb.CreateSnapshotInput{
		ClusterName:  aws.String(d.Get(names.AttrClusterName).(string)),
		SnapshotName: aws.String(name),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating MemoryDB Snapshot: %s", input)
	_, err := conn.CreateSnapshotWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MemoryDB Snapshot (%s): %s", name, err)
	}

	if err := waitSnapshotAvailable(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Snapshot (%s) to be created: %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceSnapshotRead(ctx, d, meta)...)
}

func resourceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceSnapshotRead(ctx, d, meta)
}

func resourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	snapshot, err := FindSnapshotByName(ctx, conn, d.Id())

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
		return sdkdiag.AppendErrorf(diags, "failed to set cluster_configuration for MemoryDB Snapshot (%s): %s", d.Id(), err)
	}
	d.Set(names.AttrClusterName, snapshot.ClusterConfiguration.Name)
	d.Set(names.AttrKMSKeyARN, snapshot.KmsKeyId)
	d.Set(names.AttrName, snapshot.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.StringValue(snapshot.Name)))
	d.Set(names.AttrSource, snapshot.Source)

	return diags
}

func resourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	log.Printf("[DEBUG] Deleting MemoryDB Snapshot: (%s)", d.Id())
	_, err := conn.DeleteSnapshotWithContext(ctx, &memorydb.DeleteSnapshotInput{
		SnapshotName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeSnapshotNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MemoryDB Snapshot (%s): %s", d.Id(), err)
	}

	if err := waitSnapshotDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MemoryDB Snapshot (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func flattenClusterConfiguration(v *memorydb.ClusterConfiguration) []interface{} {
	if v == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrDescription:        aws.StringValue(v.Description),
		names.AttrEngineVersion:      aws.StringValue(v.EngineVersion),
		"maintenance_window":         aws.StringValue(v.MaintenanceWindow),
		names.AttrName:               aws.StringValue(v.Name),
		"node_type":                  aws.StringValue(v.NodeType),
		"num_shards":                 aws.Int64Value(v.NumShards),
		names.AttrParameterGroupName: aws.StringValue(v.ParameterGroupName),
		names.AttrPort:               aws.Int64Value(v.Port),
		"snapshot_retention_limit":   aws.Int64Value(v.SnapshotRetentionLimit),
		"snapshot_window":            aws.StringValue(v.SnapshotWindow),
		"subnet_group_name":          aws.StringValue(v.SubnetGroupName),
		names.AttrTopicARN:           aws.StringValue(v.TopicArn),
		names.AttrVPCID:              aws.StringValue(v.VpcId),
	}

	return []interface{}{m}
}
