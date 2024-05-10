// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_ebs_snapshot_block_public_access")
func ResourceEBSSnapshotBlockPublicAccess() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEBSSnapshotBlockPublicAccessCreate,
		ReadWithoutTimeout:   resourceEBSSnapshotBlockPublicAccessRead,
		UpdateWithoutTimeout: resourceEBSSnapshotBlockPublicAccessUpdate,
		DeleteWithoutTimeout: resourceEBSSnapshotBlockPublicAccessDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"state": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.SnapshotBlockPublicAccessState](),
			},
		},
	}
}

func resourceEBSSnapshotBlockPublicAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.EnableSnapshotBlockPublicAccessInput{
		State: types.SnapshotBlockPublicAccessState(d.Get("state").(string)),
	}

	_, err := conn.EnableSnapshotBlockPublicAccess(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling EBS snapshot block public access: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	return append(diags, resourceEBSSnapshotBlockPublicAccessRead(ctx, d, meta)...)
}

func resourceEBSSnapshotBlockPublicAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	output, err := conn.GetSnapshotBlockPublicAccessState(ctx, &ec2.GetSnapshotBlockPublicAccessStateInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS snapshot block public access: %s", err)
	}

	d.Set("state", string(output.State))

	return diags
}

func resourceEBSSnapshotBlockPublicAccessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.EnableSnapshotBlockPublicAccessInput{
		State: types.SnapshotBlockPublicAccessState(d.Get("state").(string)),
	}

	_, err := conn.EnableSnapshotBlockPublicAccess(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "update EBS snapshot block public access: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	return append(diags, resourceEBSSnapshotBlockPublicAccessRead(ctx, d, meta)...)
}

func resourceEBSSnapshotBlockPublicAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	_, err := conn.DisableSnapshotBlockPublicAccess(ctx, &ec2.DisableSnapshotBlockPublicAccessInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling EBS snapshot block public access: %s", err)
	}

	return diags
}
