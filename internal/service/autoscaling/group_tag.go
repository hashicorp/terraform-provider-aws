// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_autoscaling_group_tag", name="Group Tag")
func resourceGroupTag() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupTagCreate,
		ReadWithoutTimeout:   resourceGroupTagRead,
		UpdateWithoutTimeout: resourceGroupTagUpdate,
		DeleteWithoutTimeout: resourceGroupTagDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"autoscaling_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tag": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
						},
						"propagate_at_launch": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceGroupTagCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.semgrep.tags.calling-UpdateTags-in-resource-create
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	identifier := d.Get("autoscaling_group_name").(string)
	tags := d.Get("tag").([]interface{})
	key := tags[0].(map[string]interface{})[names.AttrKey].(string)

	if err := updateTags(ctx, conn, identifier, TagResourceTypeGroup, nil, tags); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AutoScaling Group (%s) tag (%s): %s", identifier, key, err)
	}

	d.SetId(tftags.SetResourceID(identifier, key))

	return append(diags, resourceGroupTagRead(ctx, d, meta)...)
}

func resourceGroupTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)
	identifier, key, err := tftags.GetResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AutoScaling Group (%s) tag (%s): %s", identifier, key, err)
	}

	value, err := findTag(ctx, conn, identifier, TagResourceTypeGroup, key)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AutoScaling Group (%s) tag (%s), removing from state", identifier, key)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AutoScaling Group (%s) tag (%s): %s", identifier, key, err)
	}

	d.Set("autoscaling_group_name", identifier)

	if err := d.Set("tag", []map[string]interface{}{{
		names.AttrKey:         key,
		names.AttrValue:       value.Value,
		"propagate_at_launch": value.AdditionalBoolFields["PropagateAtLaunch"],
	}}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tag: %s", err)
	}

	return diags
}

func resourceGroupTagUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)
	identifier, key, err := tftags.GetResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating AutoScaling Group Tag (%s): %s", d.Id(), err)
	}

	if err := updateTags(ctx, conn, identifier, TagResourceTypeGroup, nil, d.Get("tag")); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating AutoScaling Group (%s) tag (%s): %s", identifier, key, err)
	}

	return append(diags, resourceGroupTagRead(ctx, d, meta)...)
}

func resourceGroupTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)
	identifier, key, err := tftags.GetResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AutoScaling Group Tag (%s): %s", d.Id(), err)
	}

	if err := updateTags(ctx, conn, identifier, TagResourceTypeGroup, d.Get("tag"), nil); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AutoScaling Group (%s) tag (%s): %s", identifier, key, err)
	}

	return diags
}
