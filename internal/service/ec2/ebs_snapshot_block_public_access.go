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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ebs_snapshot_block_public_access", name="EBS Snapshot Block Public Access")
func resourceEBSSnapshotBlockPublicAccess() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEBSSnapshotBlockPublicAccessPut,
		ReadWithoutTimeout:   resourceEBSSnapshotBlockPublicAccessRead,
		UpdateWithoutTimeout: resourceEBSSnapshotBlockPublicAccessPut,
		DeleteWithoutTimeout: resourceEBSSnapshotBlockPublicAccessDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrState: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.SnapshotBlockPublicAccessState](),
			},
		},
	}
}

func resourceEBSSnapshotBlockPublicAccessPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	state := d.Get(names.AttrState).(string)
	input := ec2.EnableSnapshotBlockPublicAccessInput{
		State: types.SnapshotBlockPublicAccessState(state),
	}

	_, err := conn.EnableSnapshotBlockPublicAccess(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling EBS Snapshot Block Public Access (%s): %s", state, err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).Region(ctx))
	}

	return append(diags, resourceEBSSnapshotBlockPublicAccessRead(ctx, d, meta)...)
}

func resourceEBSSnapshotBlockPublicAccessRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := ec2.GetSnapshotBlockPublicAccessStateInput{}
	output, err := conn.GetSnapshotBlockPublicAccessState(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS Snapshot Block Public Access: %s", err)
	}

	d.Set(names.AttrState, output.State)

	return diags
}

func resourceEBSSnapshotBlockPublicAccessDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	// Removing the resource disables blocking of EBS snapshot sharing.
	input := ec2.DisableSnapshotBlockPublicAccessInput{}
	_, err := conn.DisableSnapshotBlockPublicAccess(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling EBS Snapshot Block Public Access: %s", err)
	}

	return diags
}
