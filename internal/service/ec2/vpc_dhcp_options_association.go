package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVPCDHCPOptionsAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCDHCPOptionsAssociationCreate,
		Read:   resourceVPCDHCPOptionsAssociationRead,
		Update: resourceVPCDHCPOptionsAssociationUpdate,
		Delete: resourceVPCDHCPOptionsAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsVpcDhcpOptionsAssociationImport,
		},

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"dhcp_options_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsVpcDhcpOptionsAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).EC2Conn
	// Provide the vpc_id as the id to import
	vpcRaw, _, err := VPCStateRefreshFunc(conn, d.Id())()
	if err != nil {
		return nil, err
	}
	if vpcRaw == nil {
		return nil, nil
	}
	vpc := vpcRaw.(*ec2.Vpc)
	if err = d.Set("vpc_id", vpc.VpcId); err != nil {
		return nil, err
	}
	if err = d.Set("dhcp_options_id", vpc.DhcpOptionsId); err != nil {
		return nil, err
	}
	d.SetId(fmt.Sprintf("%s-%s", aws.StringValue(vpc.DhcpOptionsId), aws.StringValue(vpc.VpcId)))
	return []*schema.ResourceData{d}, nil
}

func resourceVPCDHCPOptionsAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcId := d.Get("vpc_id").(string)
	optsID := d.Get("dhcp_options_id").(string)

	log.Printf("[INFO] Creating DHCP Options association: %s => %s", vpcId, optsID)

	if _, err := conn.AssociateDhcpOptions(&ec2.AssociateDhcpOptionsInput{
		DhcpOptionsId: aws.String(optsID),
		VpcId:         aws.String(vpcId),
	}); err != nil {
		return err
	}

	// Set the ID and return
	d.SetId(fmt.Sprintf("%s-%s", optsID, vpcId))

	log.Printf("[INFO] VPC DHCP Association ID: %s", d.Id())

	return resourceVPCDHCPOptionsAssociationRead(d, meta)
}

func resourceVPCDHCPOptionsAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	var vpc *ec2.Vpc

	err := resource.Retry(PropagationTimeout, func() *resource.RetryError {
		var err error

		vpc, err = FindVPCByID(conn, d.Get("vpc_id").(string))

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeInvalidVPCIDNotFound) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if d.IsNewResource() && aws.StringValue(vpc.DhcpOptionsId) != d.Get("dhcp_options_id").(string) {
			return resource.RetryableError(&resource.NotFoundError{
				LastError: fmt.Errorf("EC2 VPC DHCP Options Association (%s) not found", d.Id()),
			})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		vpc, err = FindVPCByID(conn, d.Get("vpc_id").(string))
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeInvalidVPCIDNotFound) {
		log.Printf("[WARN] EC2 VPC DHCP Options Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC DHCP Options Association (%s): %w", d.Id(), err)
	}

	if vpc == nil {
		return fmt.Errorf("error reading EC2 VPC DHCP Options Association (%s): empty response", d.Id())
	}

	d.Set("vpc_id", vpc.VpcId)
	d.Set("dhcp_options_id", vpc.DhcpOptionsId)

	return nil
}

// DHCP Options Asociations cannot be updated.
func resourceVPCDHCPOptionsAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceVPCDHCPOptionsAssociationCreate(d, meta)
}

const VPCDefaultOptionsID = "default"

// AWS does not provide an API to disassociate a DHCP Options set from a VPC.
// So, we do this by setting the VPC to the default DHCP Options Set.
func resourceVPCDHCPOptionsAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Disassociating DHCP Options Set %s from VPC %s...", d.Get("dhcp_options_id"), d.Get("vpc_id"))

	if d.Get("dhcp_options_id").(string) == VPCDefaultOptionsID {
		// definition of deleted is DhcpOptionsId being equal to "default", nothing to do
		return nil
	}

	_, err := conn.AssociateDhcpOptions(&ec2.AssociateDhcpOptionsInput{
		DhcpOptionsId: aws.String(VPCDefaultOptionsID),
		VpcId:         aws.String(d.Get("vpc_id").(string)),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVPCIDNotFound) {
		return nil
	}

	return err
}
