package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVPCDHCPOptionsAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCDHCPOptionsAssociationPut,
		Read:   resourceVPCDHCPOptionsAssociationRead,
		Update: resourceVPCDHCPOptionsAssociationPut,
		Delete: resourceVPCDHCPOptionsAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: resourceVPCDHCPOptionsAssociationImport,
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

func resourceVPCDHCPOptionsAssociationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	dhcpOptionsID := d.Get("dhcp_options_id").(string)
	vpcID := d.Get("vpc_id").(string)
	id := VPCDHCPOptionsAssociationCreateResourceID(dhcpOptionsID, vpcID)
	input := &ec2.AssociateDhcpOptionsInput{
		DhcpOptionsId: aws.String(dhcpOptionsID),
		VpcId:         aws.String(vpcID),
	}

	log.Printf("[DEBUG] Creating EC2 VPC DHCP Options Set Association: %s", input)
	_, err := conn.AssociateDhcpOptions(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 VPC DHCP Options Set Association (%s): %w", id, err)
	}

	d.SetId(id)

	return resourceVPCDHCPOptionsAssociationRead(d, meta)
}

func resourceVPCDHCPOptionsAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	dhcpOptionsID, vpcID, err := VPCDHCPOptionsAssociationParseResourceID(d.Id())

	if err != nil {
		return err
	}

	_, err = tfresource.RetryWhenNewResourceNotFound(propagationTimeout, func() (interface{}, error) {
		return nil, FindVPCDHCPOptionsAssociation(conn, vpcID, dhcpOptionsID)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC DHCP Options Set Association %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC DHCP Options Set Association (%s): %w", d.Id(), err)
	}

	d.Set("dhcp_options_id", dhcpOptionsID)
	d.Set("vpc_id", vpcID)

	return nil
}

func resourceVPCDHCPOptionsAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	dhcpOptionsID, vpcID, err := VPCDHCPOptionsAssociationParseResourceID(d.Id())

	if err != nil {
		return err
	}

	if dhcpOptionsID == DefaultDHCPOptionsID {
		return nil
	}

	// AWS does not provide an API to disassociate a DHCP Options set from a VPC.
	// So, we do this by setting the VPC to the default DHCP Options Set.

	log.Printf("[DEBUG] Deleting EC2 VPC DHCP Options Set Association: %s", d.Id())
	_, err = conn.AssociateDhcpOptions(&ec2.AssociateDhcpOptionsInput{
		DhcpOptionsId: aws.String(DefaultDHCPOptionsID),
		VpcId:         aws.String(vpcID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating EC2 DHCP Options Set (%s) from VPC (%s): %w", dhcpOptionsID, vpcID, err)
	}

	return err
}

func resourceVPCDHCPOptionsAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpc, err := FindVPCByID(conn, d.Id())

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
