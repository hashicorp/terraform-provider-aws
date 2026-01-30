// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_image_block_public_access", name="Image Block Public Access")
// @SingletonIdentity
// @IdentityVersion(1, sdkV2IdentityUpgraders="imageBlockPublicAccessIdentityUpgradeV0")
// @V60SDKv2Fix
// @NoImport
// @Testing(checkDestroyNoop=true)
// @Testing(hasExistsFunction=false)
// @Testing(generator=false)
// Generated tests have several issues: (todo: list them)
// @Testing(identityTest=false)
// @Testing(identityVersion="0;v6.0.0")
// @Testing(identityVersion="1;v6.21.0")
func resourceImageBlockPublicAccess() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceImageBlockPublicAccessPut,
		ReadWithoutTimeout:   resourceImageBlockPublicAccessRead,
		UpdateWithoutTimeout: resourceImageBlockPublicAccessPut,
		DeleteWithoutTimeout: schema.NoopContext,

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrState: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(imageBlockPublicAccessState_Values(), false),
			},
		},
	}
}

func resourceImageBlockPublicAccessPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	state := d.Get(names.AttrState).(string)

	if slices.Contains(imageBlockPublicAccessEnabledState_Values(), state) {
		input := ec2.EnableImageBlockPublicAccessInput{
			ImageBlockPublicAccessState: awstypes.ImageBlockPublicAccessEnabledState(state),
		}

		_, err := conn.EnableImageBlockPublicAccess(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling EC2 Image Block Public Access: %s", err)
		}
	} else {
		input := ec2.DisableImageBlockPublicAccessInput{}

		_, err := conn.DisableImageBlockPublicAccess(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling EC2 Image Block Public Access: %s", err)
		}
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).AccountID(ctx))
	}

	if err := waitImageBlockPublicAccessState(ctx, conn, state, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Image Block Public Access state (%s): %s", state, err)
	}

	return append(diags, resourceImageBlockPublicAccessRead(ctx, d, meta)...)
}

func resourceImageBlockPublicAccessRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	output, err := findImageBlockPublicAccessState(ctx, conn)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EC2 Image Block Public Access %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Image Block Public Access (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrState, output)

	return diags
}

var imageBlockPublicAccessIdentityUpgradeV0 = schema.IdentityUpgrader{
	Version: 0,
	Upgrade: func(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
		rawState[names.AttrRegion] = meta.(*conns.AWSClient).Region(ctx)
		return rawState, nil
	},
}

func imageBlockPublicAccessDisabledState_Values() []string {
	return enum.Values[awstypes.ImageBlockPublicAccessDisabledState]()
}

func imageBlockPublicAccessEnabledState_Values() []string {
	return enum.Values[awstypes.ImageBlockPublicAccessEnabledState]()
}

func imageBlockPublicAccessState_Values() []string {
	return append(imageBlockPublicAccessEnabledState_Values(), imageBlockPublicAccessDisabledState_Values()...)
}
