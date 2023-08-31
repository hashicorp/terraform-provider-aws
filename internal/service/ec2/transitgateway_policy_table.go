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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway_policy_table", name="Transit Gateway Policy Table")
// @Tags(identifierAttribute="id")
func ResourceTransitGatewayPolicyTable() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayPolicyTableCreate,
		ReadWithoutTimeout:   resourceTransitGatewayPolicyTableRead,
		UpdateWithoutTimeout: resourceTransitGatewayPolicyTableUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayPolicyTableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"transit_gateway_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceTransitGatewayPolicyTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	transitGatewayID := d.Get("transit_gateway_id").(string)
	input := &ec2.CreateTransitGatewayPolicyTableInput{
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeTransitGatewayPolicyTable),
		TransitGatewayId:  aws.String(transitGatewayID),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Policy Table: %s", input)
	output, err := conn.CreateTransitGatewayPolicyTableWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway (%s) Policy Table: %s", transitGatewayID, err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayPolicyTable.TransitGatewayPolicyTableId))

	if _, err := WaitTransitGatewayPolicyTableCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Policy Table (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayPolicyTableRead(ctx, d, meta)...)
}

func resourceTransitGatewayPolicyTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	transitGatewayPolicyTable, err := FindTransitGatewayPolicyTableByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Policy Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Policy Table (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("transit-gateway-policy-table/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("state", transitGatewayPolicyTable.State)
	d.Set("transit_gateway_id", transitGatewayPolicyTable.TransitGatewayId)

	setTagsOut(ctx, transitGatewayPolicyTable.Tags)

	return diags
}

func resourceTransitGatewayPolicyTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceTransitGatewayPolicyTableRead(ctx, d, meta)...)
}

func resourceTransitGatewayPolicyTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Policy Table: %s", d.Id())
	_, err := conn.DeleteTransitGatewayPolicyTableWithContext(ctx, &ec2.DeleteTransitGatewayPolicyTableInput{
		TransitGatewayPolicyTableId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayPolicyTableIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Policy Table (%s): %s", d.Id(), err)
	}

	if _, err := WaitTransitGatewayPolicyTableDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Policy Table (%s) delete: %s", d.Id(), err)
	}

	return diags
}
