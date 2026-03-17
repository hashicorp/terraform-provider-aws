// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_carrier_gateway, name="Carrier Gateway")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceCarrierGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCarrierGatewayCreate,
		ReadWithoutTimeout:   resourceCarrierGatewayRead,
		UpdateWithoutTimeout: resourceCarrierGatewayUpdate,
		DeleteWithoutTimeout: resourceCarrierGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceCarrierGatewayCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := ec2.CreateCarrierGatewayInput{
		ClientToken:       aws.String(sdkid.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeCarrierGateway),
		VpcId:             aws.String(d.Get(names.AttrVPCID).(string)),
	}

	output, err := conn.CreateCarrierGateway(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Carrier Gateway: %s", err)
	}

	d.SetId(aws.ToString(output.CarrierGateway.CarrierGatewayId))

	if _, err := waitCarrierGatewayCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Carrier Gateway (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceCarrierGatewayRead(ctx, d, meta)...)
}

func resourceCarrierGatewayRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.EC2Client(ctx)

	carrierGateway, err := findCarrierGatewayByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EC2 Carrier Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Carrier Gateway (%s): %s", d.Id(), err)
	}

	ownerID := aws.ToString(carrierGateway.OwnerId)
	d.Set(names.AttrARN, carrierGatewayARN(ctx, c, ownerID, d.Id()))
	d.Set(names.AttrOwnerID, ownerID)
	d.Set(names.AttrVPCID, carrierGateway.VpcId)

	setTagsOut(ctx, carrierGateway.Tags)

	return diags
}

func resourceCarrierGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceCarrierGatewayRead(ctx, d, meta)...)
}

func resourceCarrierGatewayDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 Carrier Gateway (%s)", d.Id())
	input := ec2.DeleteCarrierGatewayInput{
		CarrierGatewayId: aws.String(d.Id()),
	}
	_, err := conn.DeleteCarrierGateway(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidCarrierGatewayIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Carrier Gateway (%s): %s", d.Id(), err)
	}

	if _, err := waitCarrierGatewayDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Carrier Gateway (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func carrierGatewayARN(ctx context.Context, c *conns.AWSClient, accountID, carrierGatewayID string) string {
	return c.RegionalARNWithAccount(ctx, names.EC2, accountID, "carrier-gateway/"+carrierGatewayID)
}
