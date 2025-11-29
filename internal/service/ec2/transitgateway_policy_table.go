// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway_policy_table", name="Transit Gateway Policy Table")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceTransitGatewayPolicyTable() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayPolicyTableCreate,
		ReadWithoutTimeout:   resourceTransitGatewayPolicyTableRead,
		UpdateWithoutTimeout: resourceTransitGatewayPolicyTableUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayPolicyTableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrTransitGatewayID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceTransitGatewayPolicyTableCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayID := d.Get(names.AttrTransitGatewayID).(string)
	input := &ec2.CreateTransitGatewayPolicyTableInput{
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeTransitGatewayPolicyTable),
		TransitGatewayId:  aws.String(transitGatewayID),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Policy Table: %+v", input)
	output, err := conn.CreateTransitGatewayPolicyTable(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway (%s) Policy Table: %s", transitGatewayID, err)
	}

	d.SetId(aws.ToString(output.TransitGatewayPolicyTable.TransitGatewayPolicyTableId))

	if _, err := waitTransitGatewayPolicyTableCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Policy Table (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayPolicyTableRead(ctx, d, meta)...)
}

func resourceTransitGatewayPolicyTableRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.EC2Client(ctx)

	transitGatewayPolicyTable, err := findTransitGatewayPolicyTableByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Policy Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Policy Table (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, transitGatewayPolicyTableARN(ctx, c, d.Id()))
	d.Set(names.AttrState, transitGatewayPolicyTable.State)
	d.Set(names.AttrTransitGatewayID, transitGatewayPolicyTable.TransitGatewayId)

	setTagsOut(ctx, transitGatewayPolicyTable.Tags)

	return diags
}

func resourceTransitGatewayPolicyTableUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceTransitGatewayPolicyTableRead(ctx, d, meta)...)
}

func resourceTransitGatewayPolicyTableDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Policy Table: %s", d.Id())
	input := ec2.DeleteTransitGatewayPolicyTableInput{
		TransitGatewayPolicyTableId: aws.String(d.Id()),
	}
	_, err := conn.DeleteTransitGatewayPolicyTable(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayPolicyTableIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Policy Table (%s): %s", d.Id(), err)
	}

	if _, err := waitTransitGatewayPolicyTableDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Policy Table (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func transitGatewayPolicyTableARN(ctx context.Context, c *conns.AWSClient, policyTableID string) string {
	return c.RegionalARN(ctx, names.EC2, "transit-gateway-policy-table/"+policyTableID)
}
