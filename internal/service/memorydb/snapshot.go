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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"maintenance_window": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
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
						"parameter_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
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
						"topic_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"kms_key_arn": {
				// The API will accept an ID, but return the ARN on every read.
				// For the sake of consistency, force everyone to use ARN-s.
				// To prevent confusion, the attribute is suffixed _arn rather
				// than the _id implied by the API.
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateResourceName(snapshotNameMaxLength),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateResourceNamePrefix(snapshotNameMaxLength - id.UniqueIDSuffixLength),
			},
			"source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &memorydb.CreateSnapshotInput{
		ClusterName:  aws.String(d.Get("cluster_name").(string)),
		SnapshotName: aws.String(name),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating MemoryDB Snapshot: %s", input)
	_, err := conn.CreateSnapshotWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating MemoryDB Snapshot (%s): %s", name, err)
	}

	if err := waitSnapshotAvailable(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for MemoryDB Snapshot (%s) to be created: %s", name, err)
	}

	d.SetId(name)

	return resourceSnapshotRead(ctx, d, meta)
}

func resourceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceSnapshotRead(ctx, d, meta)
}

func resourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	snapshot, err := FindSnapshotByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MemoryDB Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading MemoryDB Snapshot (%s): %s", d.Id(), err)
	}

	d.Set("arn", snapshot.ARN)
	if err := d.Set("cluster_configuration", flattenClusterConfiguration(snapshot.ClusterConfiguration)); err != nil {
		return diag.Errorf("failed to set cluster_configuration for MemoryDB Snapshot (%s): %s", d.Id(), err)
	}
	d.Set("cluster_name", snapshot.ClusterConfiguration.Name)
	d.Set("kms_key_arn", snapshot.KmsKeyId)
	d.Set("name", snapshot.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(snapshot.Name)))
	d.Set("source", snapshot.Source)

	return nil
}

func resourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	log.Printf("[DEBUG] Deleting MemoryDB Snapshot: (%s)", d.Id())
	_, err := conn.DeleteSnapshotWithContext(ctx, &memorydb.DeleteSnapshotInput{
		SnapshotName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeSnapshotNotFoundFault) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting MemoryDB Snapshot (%s): %s", d.Id(), err)
	}

	if err := waitSnapshotDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for MemoryDB Snapshot (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func flattenClusterConfiguration(v *memorydb.ClusterConfiguration) []interface{} {
	if v == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"description":              aws.StringValue(v.Description),
		"engine_version":           aws.StringValue(v.EngineVersion),
		"maintenance_window":       aws.StringValue(v.MaintenanceWindow),
		"name":                     aws.StringValue(v.Name),
		"node_type":                aws.StringValue(v.NodeType),
		"num_shards":               aws.Int64Value(v.NumShards),
		"parameter_group_name":     aws.StringValue(v.ParameterGroupName),
		"port":                     aws.Int64Value(v.Port),
		"snapshot_retention_limit": aws.Int64Value(v.SnapshotRetentionLimit),
		"snapshot_window":          aws.StringValue(v.SnapshotWindow),
		"subnet_group_name":        aws.StringValue(v.SubnetGroupName),
		"topic_arn":                aws.StringValue(v.TopicArn),
		"vpc_id":                   aws.StringValue(v.VpcId),
	}

	return []interface{}{m}
}
