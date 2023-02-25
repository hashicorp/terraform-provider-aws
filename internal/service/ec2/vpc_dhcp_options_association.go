package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVPCDHCPOptionsAssociation() *schema.Resource {
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
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCDHCPOptionsAssociationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	dhcpOptionsID := d.Get("dhcp_options_id").(string)
	vpcID := d.Get("vpc_id").(string)
	id := VPCDHCPOptionsAssociationCreateResourceID(dhcpOptionsID, vpcID)
	input := &ec2.AssociateDhcpOptionsInput{
		DhcpOptionsId: aws.String(dhcpOptionsID),
		VpcId:         aws.String(vpcID),
	}

	log.Printf("[DEBUG] Creating EC2 VPC DHCP Options Set Association: %s", input)
	_, err := conn.AssociateDhcpOptionsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPC DHCP Options Set Association (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceVPCDHCPOptionsAssociationRead(ctx, d, meta)...)
}

func resourceVPCDHCPOptionsAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	dhcpOptionsID, vpcID, err := VPCDHCPOptionsAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC DHCP Options Set Association (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return nil, FindVPCDHCPOptionsAssociation(ctx, conn, vpcID, dhcpOptionsID)
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
	d.Set("vpc_id", vpcID)

	return diags
}

func resourceVPCDHCPOptionsAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	dhcpOptionsID, vpcID, err := VPCDHCPOptionsAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if dhcpOptionsID == DefaultDHCPOptionsID {
		return diags
	}

	// AWS does not provide an API to disassociate a DHCP Options set from a VPC.
	// So, we do this by setting the VPC to the default DHCP Options Set.

	log.Printf("[DEBUG] Deleting EC2 VPC DHCP Options Set Association: %s", d.Id())
	_, err = conn.AssociateDhcpOptionsWithContext(ctx, &ec2.AssociateDhcpOptionsInput{
		DhcpOptionsId: aws.String(DefaultDHCPOptionsID),
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
	conn := meta.(*conns.AWSClient).EC2Conn()

	vpc, err := FindVPCByID(ctx, conn, d.Id())

	if err != nil {
		return nil, fmt.Errorf("error reading EC2 VPC (%s): %w", d.Id(), err)
	}

	dhcpOptionsID := aws.StringValue(vpc.DhcpOptionsId)
	vpcID := aws.StringValue(vpc.VpcId)

	d.SetId(VPCDHCPOptionsAssociationCreateResourceID(dhcpOptionsID, vpcID))
	d.Set("dhcp_options_id", dhcpOptionsID)
	d.Set("vpc_id", vpcID)

	return []*schema.ResourceData{d}, nil
}

const vpcDHCPOptionsAssociationResourceIDSeparator = "-"

func VPCDHCPOptionsAssociationCreateResourceID(dhcpOptionsID, vpcID string) string {
	parts := []string{dhcpOptionsID, vpcID}
	id := strings.Join(parts, vpcDHCPOptionsAssociationResourceIDSeparator)

	return id
}

func VPCDHCPOptionsAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, vpcDHCPOptionsAssociationResourceIDSeparator)

	// The DHCP Options ID either contains '-' or is the special value "default".
	// The VPC ID contains '-'.
	switch n := len(parts); n {
	case 3:
		if parts[0] == DefaultDHCPOptionsID && parts[1] != "" && parts[2] != "" {
			return parts[0], strings.Join([]string{parts[1], parts[2]}, vpcDHCPOptionsAssociationResourceIDSeparator), nil
		}
	case 4:
		if parts[0] != "" && parts[1] != "" && parts[2] != "" && parts[3] != "" {
			return strings.Join([]string{parts[0], parts[1]}, vpcDHCPOptionsAssociationResourceIDSeparator), strings.Join([]string{parts[2], parts[3]}, vpcDHCPOptionsAssociationResourceIDSeparator), nil
		}
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DHCPOptionsID%[2]sVPCID", id, vpcDHCPOptionsAssociationResourceIDSeparator)
}
