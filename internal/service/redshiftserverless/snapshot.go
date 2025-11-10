// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshiftserverless_snapshot", name="Snapshot")
func resourceSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSnapshotCreate,
		ReadWithoutTimeout:   resourceSnapshotRead,
		UpdateWithoutTimeout: resourceSnapshotUpdate,
		DeleteWithoutTimeout: resourceSnapshotDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"accounts_with_provisioned_restore_access": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"accounts_with_restore_access": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"admin_username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"namespace_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"namespace_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner_account": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRetentionPeriod: {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"snapshot_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	input := redshiftserverless.CreateSnapshotInput{
		NamespaceName: aws.String(d.Get("namespace_name").(string)),
		SnapshotName:  aws.String(d.Get("snapshot_name").(string)),
	}

	if v, ok := d.GetOk(names.AttrRetentionPeriod); ok {
		input.RetentionPeriod = aws.Int32(int32(v.(int)))
	}

	out, err := conn.CreateSnapshot(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Serverless Snapshot : %s", err)
	}

	d.SetId(aws.ToString(out.Snapshot.SnapshotName))

	if _, err := waitSnapshotAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Serverless Snapshot (%s) to be Available: %s", d.Id(), err)
	}

	return append(diags, resourceSnapshotRead(ctx, d, meta)...)
}

func resourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	out, err := findSnapshotByName(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Serverless Snapshot (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.SnapshotArn)
	d.Set("snapshot_name", out.SnapshotName)
	d.Set("namespace_name", out.NamespaceName)
	d.Set("namespace_arn", out.NamespaceArn)
	d.Set(names.AttrRetentionPeriod, out.SnapshotRetentionPeriod)
	d.Set("admin_username", out.AdminUsername)
	d.Set(names.AttrKMSKeyID, out.KmsKeyId)
	d.Set("owner_account", out.OwnerAccount)
	d.Set("accounts_with_provisioned_restore_access", flex.FlattenStringValueSet(out.AccountsWithRestoreAccess))
	d.Set("accounts_with_restore_access", flex.FlattenStringValueSet(out.AccountsWithRestoreAccess))

	return diags
}

func resourceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	input := &redshiftserverless.UpdateSnapshotInput{
		SnapshotName:    aws.String(d.Id()),
		RetentionPeriod: aws.Int32(int32(d.Get(names.AttrRetentionPeriod).(int))),
	}

	_, err := conn.UpdateSnapshot(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Redshift Serverless Snapshot (%s): %s", d.Id(), err)
	}

	return append(diags, resourceSnapshotRead(ctx, d, meta)...)
}

func resourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	log.Printf("[DEBUG] Deleting Redshift Serverless Snapshot: %s", d.Id())
	_, err := conn.DeleteSnapshot(ctx, &redshiftserverless.DeleteSnapshotInput{
		SnapshotName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Serverless Snapshot (%s): %s", d.Id(), err)
	}

	if _, err := waitSnapshotDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Serverless Snapshot (%s) to be Deleted: %s", d.Id(), err)
	}

	return diags
}

func findSnapshotByName(ctx context.Context, conn *redshiftserverless.Client, name string) (*awstypes.Snapshot, error) {
	input := &redshiftserverless.GetSnapshotInput{
		SnapshotName: aws.String(name),
	}

	output, err := conn.GetSnapshot(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.ResourceNotFoundException](err, "snapshot") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Snapshot, nil
}

func statusSnapshot(ctx context.Context, conn *redshiftserverless.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findSnapshotByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitSnapshotAvailable(ctx context.Context, conn *redshiftserverless.Client, name string) (*awstypes.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SnapshotStatusCreating),
		Target:  enum.Slice(awstypes.SnapshotStatusAvailable),
		Refresh: statusSnapshot(ctx, conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Snapshot); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotDeleted(ctx context.Context, conn *redshiftserverless.Client, name string) (*awstypes.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SnapshotStatusAvailable),
		Target:  []string{},
		Refresh: statusSnapshot(ctx, conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Snapshot); ok {
		return output, err
	}

	return nil, err
}
