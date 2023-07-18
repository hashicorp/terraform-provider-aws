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
func ResourceSnapshotCopyGrant() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSnapshotCopyGrantCreate,
		ReadWithoutTimeout:   resourceSnapshotCopyGrantRead,
		UpdateWithoutTimeout: resourceSnapshotCopyGrantUpdate,
		DeleteWithoutTimeout: resourceSnapshotCopyGrantDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_copy_grant_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
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

	grantName := d.Get("snapshot_copy_grant_name").(string)

	input := redshift.CreateSnapshotCopyGrantInput{
		SnapshotCopyGrantName: aws.String(grantName),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	log.Printf("[DEBUG]: Adding new Redshift SnapshotCopyGrant: %s", input)

	var out *redshift.CreateSnapshotCopyGrantOutput
	var err error

	out, err = conn.CreateSnapshotCopyGrantWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Snapshot Copy Grant (%s): %s", grantName, err)
	}

	log.Printf("[DEBUG] Created new Redshift SnapshotCopyGrant: %s", *out.SnapshotCopyGrant.SnapshotCopyGrantName)
	d.SetId(grantName)

	_, err = tfresource.RetryWhenNotFound(ctx, 3*time.Minute, func() (any, error) {
		return findSnapshotCopyGrant(ctx, conn, grantName)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Snapshot Copy Grant (%s): waiting for completion: %s", grantName, err)
	}

	return append(diags, resourceSnapshotCopyGrantRead(ctx, d, meta)...)
}

func resourceSnapshotCopyGrantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	grantName := d.Id()

	grant, err := findSnapshotCopyGrant(ctx, conn, grantName)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Snapshot Copy Grant (%s) not found, removing from state", grantName)
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Snapshot Copy Grant (%s): %s", grantName, err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "redshift",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("snapshotcopygrant:%s", grantName),
	}.String()

	d.Set("arn", arn)

	d.Set("kms_key_id", grant.KmsKeyId)
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

	grantName := d.Id()

	deleteInput := redshift.DeleteSnapshotCopyGrantInput{
		SnapshotCopyGrantName: aws.String(grantName),
	}

	log.Printf("[DEBUG] Deleting snapshot copy grant: %s", grantName)
	_, err := conn.DeleteSnapshotCopyGrantWithContext(ctx, &deleteInput)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotCopyGrantNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Snapshot Copy Grant (%s): %s", d.Id(), err)
	}

	if err := WaitForSnapshotCopyGrantToBeDeleted(ctx, conn, grantName); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Snapshot Copy Grant (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

// Used by the tests as well
func WaitForSnapshotCopyGrantToBeDeleted(ctx context.Context, conn *redshift.Redshift, grantName string) error {
	_, err := tfresource.RetryUntilNotFound(ctx, 3*time.Minute, func() (any, error) {
		return findSnapshotCopyGrant(ctx, conn, grantName)
	})
	return err
}

func findSnapshotCopyGrant(ctx context.Context, conn *redshift.Redshift, grantName string) (*redshift.SnapshotCopyGrant, error) {
	input := redshift.DescribeSnapshotCopyGrantsInput{
		SnapshotCopyGrantName: aws.String(grantName),
	}

	out, err := conn.DescribeSnapshotCopyGrantsWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotCopyGrantNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if out == nil || len(out.SnapshotCopyGrants) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}
	if l := len(out.SnapshotCopyGrants); l > 1 {
		return nil, tfresource.NewTooManyResultsError(1, nil)
	}

	return out.SnapshotCopyGrants[0], nil
}
