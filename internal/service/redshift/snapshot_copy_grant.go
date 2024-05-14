// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSnapshotCopyGrantCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	name := d.Get("snapshot_copy_grant_name").(string)
	input := &redshift.CreateSnapshotCopyGrantInput{
		SnapshotCopyGrantName: aws.String(name),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	_, err := conn.CreateSnapshotCopyGrantWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Snapshot Copy Grant (%s): %s", name, err)
	}

	d.SetId(name)

	_, err = tfresource.RetryWhenNotFound(ctx, 3*time.Minute, func() (any, error) {
		return findSnapshotCopyGrantByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Snapshot Copy Grant (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceSnapshotCopyGrantRead(ctx, d, meta)...)
}

func resourceSnapshotCopyGrantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	grant, err := findSnapshotCopyGrantByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Snapshot Copy Grant (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Snapshot Copy Grant (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "redshift",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("snapshotcopygrant:%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrKMSKeyID, grant.KmsKeyId)
	d.Set("snapshot_copy_grant_name", grant.SnapshotCopyGrantName)

	setTagsOut(ctx, grant.Tags)

	return diags
}

func resourceSnapshotCopyGrantUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceSnapshotCopyGrantRead(ctx, d, meta)...)
}

func resourceSnapshotCopyGrantDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	log.Printf("[DEBUG] Deleting Redshift Snapshot Copy Grant: %s", d.Id())
	_, err := conn.DeleteSnapshotCopyGrantWithContext(ctx, &redshift.DeleteSnapshotCopyGrantInput{
		SnapshotCopyGrantName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotCopyGrantNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Snapshot Copy Grant (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, 3*time.Minute, func() (any, error) {
		return findSnapshotCopyGrantByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Snapshot Copy Grant (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findSnapshotCopyGrantByName(ctx context.Context, conn *redshift.Redshift, name string) (*redshift.SnapshotCopyGrant, error) {
	input := &redshift.DescribeSnapshotCopyGrantsInput{
		SnapshotCopyGrantName: aws.String(name),
	}

	output, err := conn.DescribeSnapshotCopyGrantsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotCopyGrantNotFoundFault) {
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

	return tfresource.AssertSinglePtrResult(output.SnapshotCopyGrants)
}
