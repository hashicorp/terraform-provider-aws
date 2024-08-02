// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fsx_backup", name="Backup")
// @Tags(identifierAttribute="arn")
func resourceBackup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBackupCreate,
		ReadWithoutTimeout:   resourceBackupRead,
		UpdateWithoutTimeout: resourceBackupUpdate,
		DeleteWithoutTimeout: resourceBackupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFileSystemID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
		),
	}
}

func resourceBackupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	input := &fsx.CreateBackupInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrFileSystemID); ok {
		input.FileSystemId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("volume_id"); ok {
		input.VolumeId = aws.String(v.(string))
	}

	if input.FileSystemId == nil && input.VolumeId == nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx Backup: %s", "must specify either file_system_id or volume_id")
	}

	if input.FileSystemId != nil && input.VolumeId != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx Backup: %s", "can only specify either file_system_id or volume_id")
	}

	output, err := conn.CreateBackup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx Backup: %s", err)
	}

	d.SetId(aws.ToString(output.Backup.BackupId))

	if _, err := waitBackupAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx Backup (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBackupRead(ctx, d, meta)...)
}

func resourceBackupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	backup, err := findBackupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx Backup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx Backup (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, backup.ResourceARN)
	if backup.FileSystem != nil {
		d.Set(names.AttrFileSystemID, backup.FileSystem.FileSystemId)
	}
	d.Set(names.AttrKMSKeyID, backup.KmsKeyId)
	d.Set(names.AttrOwnerID, backup.OwnerId)
	d.Set(names.AttrType, backup.Type)
	if backup.Volume != nil {
		d.Set("volume_id", backup.Volume.VolumeId)
	}

	setTagsOut(ctx, backup.Tags)

	return diags
}

func resourceBackupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceBackupRead(ctx, d, meta)...)
}

func resourceBackupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	log.Printf("[INFO] Deleting FSx Backup: %s", d.Id())
	_, err := conn.DeleteBackup(ctx, &fsx.DeleteBackupInput{
		BackupId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.BackupNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FSx Backup (%s): %s", d.Id(), err)
	}

	if _, err := waitBackupDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx Backup (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findBackupByID(ctx context.Context, conn *fsx.Client, id string) (*awstypes.Backup, error) {
	input := &fsx.DescribeBackupsInput{
		BackupIds: []string{id},
	}

	return findBackup(ctx, conn, input, tfslices.PredicateTrue[*awstypes.Backup]())
}

func findBackup(ctx context.Context, conn *fsx.Client, input *fsx.DescribeBackupsInput, filter tfslices.Predicate[*awstypes.Backup]) (*awstypes.Backup, error) {
	output, err := findBackups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findBackups(ctx context.Context, conn *fsx.Client, input *fsx.DescribeBackupsInput, filter tfslices.Predicate[*awstypes.Backup]) ([]awstypes.Backup, error) {
	var output []awstypes.Backup

	pages := fsx.NewDescribeBackupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.FileSystemNotFound](err) || errs.IsA[*awstypes.BackupNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Backups {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusBackup(ctx context.Context, conn *fsx.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findBackupByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Lifecycle), nil
	}
}

func waitBackupAvailable(ctx context.Context, conn *fsx.Client, id string) (*awstypes.Backup, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BackupLifecycleCreating, awstypes.BackupLifecyclePending, awstypes.BackupLifecycleTransferring),
		Target:  enum.Slice(awstypes.BackupLifecycleAvailable),
		Refresh: statusBackup(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Backup); ok {
		return output, err
	}

	return nil, err
}

func waitBackupDeleted(ctx context.Context, conn *fsx.Client, id string) (*awstypes.Backup, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FileSystemLifecycleDeleting),
		Target:  []string{},
		Refresh: statusBackup(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Backup); ok {
		return output, err
	}

	return nil, err
}
