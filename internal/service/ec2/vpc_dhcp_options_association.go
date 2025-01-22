// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_dhcp_options_association", name="VPC DHCP Options Association")
func resourceVPCDHCPOptionsAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCDHCPOptionsAssociationPut,
		ReadWithoutTimeout:   resourceVPCDHCPOptionsAssociationRead,
		UpdateWithoutTimeout: resourceVPCDHCPOptionsAssociationPut,
		DeleteWithoutTimeout: resourceVPCDHCPOptionsAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCDHCPOptionsAssociationImport,
		},

		Schema: map[string]*schema.Schema{
			"dhcp_options_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCDHCPOptionsAssociationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	dhcpOptionsID := d.Get("dhcp_options_id").(string)
	vpcID := d.Get(names.AttrVPCID).(string)
	id := vpcDHCPOptionsAssociationCreateResourceID(dhcpOptionsID, vpcID)
	input := &ec2.AssociateDhcpOptionsInput{
		DhcpOptionsId: aws.String(dhcpOptionsID),
		VpcId:         aws.String(vpcID),
	}

	log.Printf("[DEBUG] Creating EC2 VPC DHCP Options Set Association: %#v", input)
	_, err := conn.AssociateDhcpOptions(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPC DHCP Options Set Association (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceVPCDHCPOptionsAssociationRead(ctx, d, meta)...)
}

func resourceVPCDHCPOptionsAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	dhcpOptionsID, vpcID, err := vpcDHCPOptionsAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC DHCP Options Set Association (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return nil, findVPCDHCPOptionsAssociation(ctx, conn, vpcID, dhcpOptionsID)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC DHCP Options Set Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC DHCP Options Set Association (%s): %s", d.Id(), err)
	}

	d.Set("dhcp_options_id", dhcpOptionsID)
	d.Set(names.AttrVPCID, vpcID)

	return diags
}

func resourceVPCDHCPOptionsAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	dhcpOptionsID, vpcID, err := vpcDHCPOptionsAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if dhcpOptionsID == defaultDHCPOptionsID {
		return diags
	}

	// AWS does not provide an API to disassociate a DHCP Options set from a VPC.
	// So, we do this by setting the VPC to the default DHCP Options Set.

	log.Printf("[DEBUG] Deleting EC2 VPC DHCP Options Set Association: %s", d.Id())
	_, err = conn.AssociateDhcpOptions(ctx, &ec2.AssociateDhcpOptionsInput{
		DhcpOptionsId: aws.String(defaultDHCPOptionsID),
		VpcId:         aws.String(vpcID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating EC2 DHCP Options Set (%s) from VPC (%s): %s", dhcpOptionsID, vpcID, err)
	}

	return diags
}

func resourceVPCDHCPOptionsAssociationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpc, err := findVPCByID(ctx, conn, d.Id())

	if err != nil {
		return nil, fmt.Errorf("reading EC2 VPC (%s): %w", d.Id(), err)
	}

	dhcpOptionsID := aws.ToString(vpc.DhcpOptionsId)
	vpcID := aws.ToString(vpc.VpcId)

	d.SetId(vpcDHCPOptionsAssociationCreateResourceID(dhcpOptionsID, vpcID))
	d.Set("dhcp_options_id", dhcpOptionsID)
	d.Set(names.AttrVPCID, vpcID)

	return []*schema.ResourceData{d}, nil
}

const vpcDHCPOptionsAssociationResourceIDSeparator = "-"

func vpcDHCPOptionsAssociationCreateResourceID(dhcpOptionsID, vpcID string) string {
	parts := []string{dhcpOptionsID, vpcID}
	id := strings.Join(parts, vpcDHCPOptionsAssociationResourceIDSeparator)

	return id
}

func vpcDHCPOptionsAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, vpcDHCPOptionsAssociationResourceIDSeparator)

	// The DHCP Options ID either contains '-' or is the special value "default".
	// The VPC ID contains '-'.
	switch n := len(parts); n {
	case 3:
		if parts[0] == defaultDHCPOptionsID && parts[1] != "" && parts[2] != "" {
			return parts[0], strings.Join([]string{parts[1], parts[2]}, vpcDHCPOptionsAssociationResourceIDSeparator), nil
		}
	case 4:
		if parts[0] != "" && parts[1] != "" && parts[2] != "" && parts[3] != "" {
			return strings.Join([]string{parts[0], parts[1]}, vpcDHCPOptionsAssociationResourceIDSeparator), strings.Join([]string{parts[2], parts[3]}, vpcDHCPOptionsAssociationResourceIDSeparator), nil
		}
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DHCPOptionsID%[2]sVPCID", id, vpcDHCPOptionsAssociationResourceIDSeparator)
}
