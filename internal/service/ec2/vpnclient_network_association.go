// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_client_vpn_network_association", name="Client VPN Network Association")
func resourceClientVPNNetworkAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClientVPNNetworkAssociationCreate,
		ReadWithoutTimeout:   resourceClientVPNNetworkAssociationRead,
		DeleteWithoutTimeout: resourceClientVPNNetworkAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceClientVPNNetworkAssociationImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAssociationID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceClientVPNNetworkAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID := d.Get("client_vpn_endpoint_id").(string)
	input := &ec2.AssociateClientVpnTargetNetworkInput{
		ClientToken:         aws.String(id.UniqueId()),
		ClientVpnEndpointId: aws.String(endpointID),
		SubnetId:            aws.String(d.Get(names.AttrSubnetID).(string)),
	}

	output, err := conn.AssociateClientVpnTargetNetwork(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Client VPN Network Association: %s", err)
	}

	d.SetId(aws.ToString(output.AssociationId))

	if _, err := waitClientVPNNetworkAssociationCreated(ctx, conn, d.Id(), endpointID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Client VPN Network Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceClientVPNNetworkAssociationRead(ctx, d, meta)...)
}

func resourceClientVPNNetworkAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID := d.Get("client_vpn_endpoint_id").(string)
	network, err := findClientVPNNetworkAssociationByTwoPartKey(ctx, conn, d.Id(), endpointID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Client VPN Network Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Client VPN Network Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAssociationID, network.AssociationId)
	d.Set("client_vpn_endpoint_id", network.ClientVpnEndpointId)
	d.Set(names.AttrSubnetID, network.TargetNetworkId)
	d.Set(names.AttrVPCID, network.VpcId)

	return diags
}

func resourceClientVPNNetworkAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID := d.Get("client_vpn_endpoint_id").(string)

	log.Printf("[DEBUG] Deleting EC2 Client VPN Network Association: %s", d.Id())
	_, err := conn.DisassociateClientVpnTargetNetwork(ctx, &ec2.DisassociateClientVpnTargetNetworkInput{
		ClientVpnEndpointId: aws.String(endpointID),
		AssociationId:       aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNAssociationIdNotFound, errCodeInvalidClientVPNEndpointIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating EC2 Client VPN Network Association (%s): %s", d.Id(), err)
	}

	if _, err := waitClientVPNNetworkAssociationDeleted(ctx, conn, d.Id(), endpointID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Client VPN Network Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceClientVPNNetworkAssociationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ",")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("unexpected format for ID (%[1]s), expected EndpointID%[2]sAssociationID", d.Id(), ",")
	}

	d.SetId(parts[1])
	d.Set("client_vpn_endpoint_id", parts[0])

	return []*schema.ResourceData{d}, nil
}
