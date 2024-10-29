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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dax_subnet_group")
func ResourceSubnetGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSubnetGroupCreate,
		ReadWithoutTimeout:   resourceSubnetGroupRead,
		UpdateWithoutTimeout: resourceSubnetGroupUpdate,
		DeleteWithoutTimeout: resourceSubnetGroupDelete,

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
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSubnetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXClient(ctx)

	input := &dax.CreateSubnetGroupInput{
		SubnetGroupName: aws.String(d.Get(names.AttrName).(string)),
		SubnetIds:       flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
	}
	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.CreateSubnetGroup(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DAX Subnet Group (%s): %s", d.Get(names.AttrName).(string), err)
	}

	d.SetId(d.Get(names.AttrName).(string))
	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXClient(ctx)

	resp, err := conn.DescribeSubnetGroups(ctx, &dax.DescribeSubnetGroupsInput{
		SubnetGroupNames: []string{d.Id()},
	})

	if errs.IsA[*awstypes.SubnetGroupNotFoundFault](err) {
		log.Printf("[WARN] DAX Subnet Group %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DAX Subnet Group (%s): %s", d.Id(), err)
	}
	sg := resp.SubnetGroups[0]

	d.Set(names.AttrName, sg.SubnetGroupName)
	d.Set(names.AttrDescription, sg.Description)
	subnetIDs := make([]*string, 0, len(sg.Subnets))
	for _, v := range sg.Subnets {
		subnetIDs = append(subnetIDs, v.SubnetIdentifier)
	}
	d.Set(names.AttrSubnetIDs, flex.FlattenStringList(subnetIDs))
	d.Set(names.AttrVPCID, sg.VpcId)

	return diags
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXClient(ctx)

	input := &dax.UpdateSubnetGroupInput{
		SubnetGroupName: aws.String(d.Id()),
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange(names.AttrSubnetIDs) {
		input.SubnetIds = flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set))
	}

	_, err := conn.UpdateSubnetGroup(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating DAX Subnet Group (%s): %s", d.Id(), err)
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXClient(ctx)

	input := &dax.DeleteSubnetGroupInput{
		SubnetGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteSubnetGroup(ctx, input)

	if errs.IsA[*awstypes.SubnetGroupNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DAX Subnet Group (%s): %s", d.Id(), err)
	}

	return diags
}
