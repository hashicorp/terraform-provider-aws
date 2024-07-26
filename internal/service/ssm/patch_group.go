// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	patchGroupResourceIDPartCount = 2
)

// @SDKResource("aws_ssm_patch_group", name="Patch Group")
func resourcePatchGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePatchGroupCreate,
		ReadWithoutTimeout:   resourcePatchGroupRead,
		DeleteWithoutTimeout: resourcePatchGroupDelete,

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourcePatchGroupV0().CoreConfigSchema().ImpliedType(),
				Upgrade: patchGroupStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"baseline_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"patch_group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePatchGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	baselineID := d.Get("baseline_id").(string)
	patchGroup := d.Get("patch_group").(string)
	id := errs.Must(flex.FlattenResourceId([]string{patchGroup, baselineID}, patchGroupResourceIDPartCount, false))
	input := &ssm.RegisterPatchBaselineForPatchGroupInput{
		BaselineId: aws.String(baselineID),
		PatchGroup: aws.String(patchGroup),
	}

	_, err := conn.RegisterPatchBaselineForPatchGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSM Patch Group (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourcePatchGroupRead(ctx, d, meta)...)
}

func resourcePatchGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), patchGroupResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	patchGroup, baselineID := parts[0], parts[1]
	group, err := findPatchGroupByTwoPartKey(ctx, conn, patchGroup, baselineID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Patch Group %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Patch Group (%s): %s", d.Id(), err)
	}

	var groupBaselineID string
	if group.BaselineIdentity != nil {
		groupBaselineID = aws.ToString(group.BaselineIdentity.BaselineId)
	}
	d.Set("baseline_id", groupBaselineID)
	d.Set("patch_group", group.PatchGroup)

	return diags
}

func resourcePatchGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), patchGroupResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	patchGroup, baselineID := parts[0], parts[1]

	log.Printf("[WARN] Deleting SSM Patch Group: %s", d.Id())
	_, err = conn.DeregisterPatchBaselineForPatchGroup(ctx, &ssm.DeregisterPatchBaselineForPatchGroupInput{
		BaselineId: aws.String(baselineID),
		PatchGroup: aws.String(patchGroup),
	})

	if errs.IsA[*awstypes.DoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSM Patch Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findPatchGroupByTwoPartKey(ctx context.Context, conn *ssm.Client, patchGroup, baselineID string) (*awstypes.PatchGroupPatchBaselineMapping, error) {
	input := &ssm.DescribePatchGroupsInput{}

	return findPatchGroup(ctx, conn, input, func(v *awstypes.PatchGroupPatchBaselineMapping) bool {
		if aws.ToString(v.PatchGroup) == patchGroup {
			if v.BaselineIdentity != nil && aws.ToString(v.BaselineIdentity.BaselineId) == baselineID {
				return true
			}
		}

		return false
	})
}

func findPatchGroup(ctx context.Context, conn *ssm.Client, input *ssm.DescribePatchGroupsInput, filter tfslices.Predicate[*awstypes.PatchGroupPatchBaselineMapping]) (*awstypes.PatchGroupPatchBaselineMapping, error) {
	output, err := findPatchGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPatchGroups(ctx context.Context, conn *ssm.Client, input *ssm.DescribePatchGroupsInput, filter tfslices.Predicate[*awstypes.PatchGroupPatchBaselineMapping]) ([]awstypes.PatchGroupPatchBaselineMapping, error) {
	var output []awstypes.PatchGroupPatchBaselineMapping

	pages := ssm.NewDescribePatchGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Mappings {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
