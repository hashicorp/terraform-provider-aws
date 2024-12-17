// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ce_cost_allocation_tag", name="Cost Allocation Tag")
func resourceCostAllocationTag() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCostAllocationTagUpdate,
		ReadWithoutTimeout:   resourceCostAllocationTagRead,
		UpdateWithoutTimeout: resourceCostAllocationTagUpdate,
		DeleteWithoutTimeout: resourceCostAllocationTagDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			names.AttrStatus: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.CostAllocationTagStatus](),
			},
			"tag_key": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCostAllocationTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CEClient(ctx)

	tag, err := findCostAllocationTagByTagKey(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cost Explorer Cost Allocation Tag (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cost Explorer Cost Allocation Tag (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrStatus, tag.Status)
	d.Set("tag_key", tag.TagKey)
	d.Set(names.AttrType, tag.Type)

	return diags
}

func resourceCostAllocationTagUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CEClient(ctx)

	tagKey := d.Get("tag_key").(string)

	if err := updateCostAllocationTagStatus(ctx, conn, tagKey, awstypes.CostAllocationTagStatus(d.Get(names.AttrStatus).(string))); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Cost Explorer Cost Allocation Tag (%s): %s", tagKey, err)
	}

	d.SetId(tagKey)

	return append(diags, resourceCostAllocationTagRead(ctx, d, meta)...)
}

func resourceCostAllocationTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CEClient(ctx)

	log.Printf("[DEBUG] Deleting Cost Explorer Cost Allocation Tag: %s", d.Id())
	if err := updateCostAllocationTagStatus(ctx, conn, d.Id(), awstypes.CostAllocationTagStatusInactive); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cost Explorer Cost Allocation Tag (%s): %s", d.Id(), err)
	}

	return diags
}

func updateCostAllocationTagStatus(ctx context.Context, conn *costexplorer.Client, tagKey string, status awstypes.CostAllocationTagStatus) error {
	input := &costexplorer.UpdateCostAllocationTagsStatusInput{
		CostAllocationTagsStatus: []awstypes.CostAllocationTagStatusEntry{{
			Status: status,
			TagKey: aws.String(tagKey),
		}},
	}

	_, err := conn.UpdateCostAllocationTagsStatus(ctx, input)

	return err
}

func findCostAllocationTagByTagKey(ctx context.Context, conn *costexplorer.Client, tagKey string) (*awstypes.CostAllocationTag, error) {
	input := &costexplorer.ListCostAllocationTagsInput{
		TagKeys:    []string{tagKey},
		MaxResults: aws.Int32(1),
	}

	output, err := conn.ListCostAllocationTags(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CostAllocationTags) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return &output.CostAllocationTags[0], nil
}
