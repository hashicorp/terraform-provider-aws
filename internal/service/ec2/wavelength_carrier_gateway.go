// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_carrier_gateway, name="Carrier Gateway")
// @Tags(identifierAttribute="id")
func ResourceCarrierGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCarrierGatewayCreate,
		ReadWithoutTimeout:   resourceCarrierGatewayRead,
		UpdateWithoutTimeout: resourceCarrierGatewayUpdate,
		DeleteWithoutTimeout: resourceCarrierGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceCarrierGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.CreateCarrierGatewayInput{
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeCarrierGateway),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	output, err := conn.CreateCarrierGatewayWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Carrier Gateway: %s", err)
	}

	d.SetId(aws.StringValue(output.CarrierGateway.CarrierGatewayId))

	_, err = WaitCarrierGatewayCreated(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Carrier Gateway (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceCarrierGatewayRead(ctx, d, meta)...)
}

func resourceCarrierGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	carrierGateway, err := FindCarrierGatewayByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Carrier Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Carrier Gateway (%s): %s", d.Id(), err)
	}

	ownerID := aws.StringValue(carrierGateway.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("carrier-gateway/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("owner_id", ownerID)
	d.Set("vpc_id", carrierGateway.VpcId)

	setTagsOut(ctx, carrierGateway.Tags)

	return diags
}

func resourceCarrierGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceCarrierGatewayRead(ctx, d, meta)...)
}

func resourceCarrierGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[INFO] Deleting EC2 Carrier Gateway (%s)", d.Id())
	_, err := conn.DeleteCarrierGatewayWithContext(ctx, &ec2.DeleteCarrierGatewayInput{
		CarrierGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidCarrierGatewayIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Carrier Gateway (%s): %s", d.Id(), err)
	}

	_, err = WaitCarrierGatewayDeleted(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Carrier Gateway (%s) delete: %s", d.Id(), err)
	}

	return diags
}
