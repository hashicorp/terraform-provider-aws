// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package redshift

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_snapshot_copy_grant", name="Snapshot Copy Grant")
// @Tags(identifierAttribute="arn")
func resourceSnapshotCopyGrant() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSnapshotCopyGrantCreate,
		ReadWithoutTimeout:   resourceSnapshotCopyGrantRead,
		UpdateWithoutTimeout: resourceSnapshotCopyGrantUpdate,
		DeleteWithoutTimeout: resourceSnapshotCopyGrantDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"snapshot_copy_grant_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceSnapshotCopyGrantCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	name := d.Get("snapshot_copy_grant_name").(string)
	input := redshift.CreateSnapshotCopyGrantInput{
		SnapshotCopyGrantName: aws.String(name),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	_, err := conn.CreateSnapshotCopyGrant(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Snapshot Copy Grant (%s): %s", name, err)
	}

	d.SetId(name)

	const (
		timeout = 3 * time.Minute
	)
	_, err = tfresource.RetryWhenNotFound(ctx, timeout, func(ctx context.Context) (any, error) {
		return findSnapshotCopyGrantByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Snapshot Copy Grant (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceSnapshotCopyGrantRead(ctx, d, meta)...)
}

func resourceSnapshotCopyGrantRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.RedshiftClient(ctx)

	grant, err := findSnapshotCopyGrantByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Redshift Snapshot Copy Grant (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Snapshot Copy Grant (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, snapshotCopyGrantARN(ctx, c, d.Id()))
	d.Set(names.AttrKMSKeyID, grant.KmsKeyId)
	d.Set("snapshot_copy_grant_name", grant.SnapshotCopyGrantName)

	setTagsOut(ctx, grant.Tags)

	return diags
}

func resourceSnapshotCopyGrantUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceSnapshotCopyGrantRead(ctx, d, meta)...)
}

func resourceSnapshotCopyGrantDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	log.Printf("[DEBUG] Deleting Redshift Snapshot Copy Grant: %s", d.Id())
	input := redshift.DeleteSnapshotCopyGrantInput{
		SnapshotCopyGrantName: aws.String(d.Id()),
	}
	_, err := conn.DeleteSnapshotCopyGrant(ctx, &input)

	if errs.IsA[*awstypes.SnapshotCopyGrantNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Snapshot Copy Grant (%s): %s", d.Id(), err)
	}

	const (
		timeout = 3 * time.Minute
	)
	_, err = tfresource.RetryUntilNotFound(ctx, timeout, func(ctx context.Context) (any, error) {
		return findSnapshotCopyGrantByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Snapshot Copy Grant (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func snapshotCopyGrantARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.RegionalARN(ctx, names.Redshift, "snapshotcopygrant:"+id)
}
