// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_cluster_snapshot", name="Cluster Snapshot")
// @Tags(identifierAttribute="arn")
func resourceClusterSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterSnapshotCreate,
		ReadWithoutTimeout:   resourceClusterSnapshotRead,
		UpdateWithoutTimeout: resourceClusterSnapshotUpdate,
		DeleteWithoutTimeout: resourceClusterSnapshotDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrClusterIdentifier: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"manual_snapshot_retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"owner_account": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	input := redshift.CreateClusterSnapshotInput{
		SnapshotIdentifier: aws.String(d.Get("snapshot_identifier").(string)),
		ClusterIdentifier:  aws.String(d.Get(names.AttrClusterIdentifier).(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("manual_snapshot_retention_period"); ok {
		input.ManualSnapshotRetentionPeriod = aws.Int64(int64(v.(int)))
	}

	output, err := conn.CreateClusterSnapshotWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Cluster Snapshot : %s", err)
	}

	d.SetId(aws.StringValue(output.Snapshot.SnapshotIdentifier))

	if _, err := waitClusterSnapshotCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster Snapshot (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceClusterSnapshotRead(ctx, d, meta)...)
}

func resourceClusterSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	snapshot, err := findClusterSnapshotByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Cluster Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Cluster Snapshot (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, snapshot.SnapshotArn)
	d.Set(names.AttrClusterIdentifier, snapshot.ClusterIdentifier)
	d.Set(names.AttrKMSKeyID, snapshot.KmsKeyId)
	d.Set("manual_snapshot_retention_period", snapshot.ManualSnapshotRetentionPeriod)
	d.Set("owner_account", snapshot.OwnerAccount)
	d.Set("snapshot_identifier", snapshot.SnapshotIdentifier)

	setTagsOut(ctx, snapshot.Tags)

	return diags
}

func resourceClusterSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &redshift.ModifyClusterSnapshotInput{
			ManualSnapshotRetentionPeriod: aws.Int64(int64(d.Get("manual_snapshot_retention_period").(int))),
			SnapshotIdentifier:            aws.String(d.Id()),
		}

		_, err := conn.ModifyClusterSnapshotWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Cluster Snapshot (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterSnapshotRead(ctx, d, meta)...)
}

func resourceClusterSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	log.Printf("[DEBUG] Deleting Redshift Cluster Snapshot: %s", d.Id())
	_, err := conn.DeleteClusterSnapshotWithContext(ctx, &redshift.DeleteClusterSnapshotInput{
		SnapshotIdentifier: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterSnapshotNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Cluster Snapshot (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterSnapshotDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster Snapshot (%s) delete: %s", d.Id(), err)
	}

	return diags
}
