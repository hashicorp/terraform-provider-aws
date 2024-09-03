// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dax

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dax"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dax/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dax_parameter_group")
func ResourceParameterGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceParameterGroupCreate,
		ReadWithoutTimeout:   resourceParameterGroupRead,
		UpdateWithoutTimeout: resourceParameterGroupUpdate,
		DeleteWithoutTimeout: resourceParameterGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrParameters: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXClient(ctx)

	input := &dax.CreateParameterGroupInput{
		ParameterGroupName: aws.String(d.Get(names.AttrName).(string)),
	}
	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.CreateParameterGroup(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DAX Parameter Group (%s): %s", d.Get(names.AttrName).(string), err)
	}

	d.SetId(d.Get(names.AttrName).(string))

	if len(d.Get(names.AttrParameters).(*schema.Set).List()) > 0 {
		return append(diags, resourceParameterGroupUpdate(ctx, d, meta)...)
	}
	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXClient(ctx)

	resp, err := conn.DescribeParameterGroups(ctx, &dax.DescribeParameterGroupsInput{
		ParameterGroupNames: []string{d.Id()},
	})

	if errs.IsA[*awstypes.ParameterGroupNotFoundFault](err) {
		log.Printf("[WARN] DAX ParameterGroup %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DAX Parameter Group (%s): %s", d.Id(), err)
	}

	if len(resp.ParameterGroups) == 0 {
		log.Printf("[WARN] DAX ParameterGroup %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	pg := resp.ParameterGroups[0]

	paramresp, err := conn.DescribeParameters(ctx, &dax.DescribeParametersInput{
		ParameterGroupName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ParameterGroupNotFoundFault](err) {
		log.Printf("[WARN] DAX ParameterGroup %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DAX Parameter Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, pg.ParameterGroupName)
	desc := pg.Description
	// default description is " "
	if desc != nil && aws.ToString(desc) == " " {
		*desc = ""
	}
	d.Set(names.AttrDescription, desc)
	d.Set(names.AttrParameters, flattenParameterGroupParameters(paramresp.Parameters))
	return diags
}

func resourceParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXClient(ctx)

	input := &dax.UpdateParameterGroupInput{
		ParameterGroupName: aws.String(d.Id()),
	}

	if d.HasChange(names.AttrParameters) {
		input.ParameterNameValues = expandParameterGroupParameterNameValue(
			d.Get(names.AttrParameters).(*schema.Set).List(),
		)
	}

	_, err := conn.UpdateParameterGroup(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating DAX Parameter Group (%s): %s", d.Id(), err)
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXClient(ctx)

	input := &dax.DeleteParameterGroupInput{
		ParameterGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteParameterGroup(ctx, input)

	if errs.IsA[*awstypes.ParameterGroupNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DAX Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}
