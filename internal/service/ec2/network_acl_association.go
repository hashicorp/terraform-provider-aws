package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceNetworkACLAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkACLAssociationCreate,
		Read:   resourceNetworkACLAssociationRead,
		Delete: resourceNetworkACLAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"network_acl_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNetworkACLAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	associationID, err := networkACLAssociationCreate(conn, d.Get("network_acl_id").(string), d.Get("subnet_id").(string))

	if err != nil {
		return err
	}

	d.SetId(associationID)

	return resourceNetworkACLAssociationRead(d, meta)
}

func resourceNetworkACLAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	association, err := FindNetworkACLAssociationByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network ACL Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Network ACL Association (%s): %w", d.Id(), err)
	}

	d.Set("network_acl_id", association.NetworkAclId)
	d.Set("subnet_id", association.SubnetId)

	return nil
}

func resourceNetworkACLAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.association-id": d.Id(),
		}),
	}

	nacl, err := FindNetworkACL(conn, input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Network ACL for Association (%s): %w", d.Id(), err)
	}

	vpcID := aws.StringValue(nacl.VpcId)
	defaultNACL, err := FindVPCDefaultNetworkACL(conn, vpcID)

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s) default NACL: %w", vpcID, err)
	}

	if err := networkACLAssociationDelete(conn, d.Id(), aws.StringValue(defaultNACL.NetworkAclId)); err != nil {
		return err
	}

	return nil
}

// networkACLAssociationCreate creates an association between the specified NACL and subnet.
// The subnet's current association is replaced and the new association's ID is returned.
func networkACLAssociationCreate(conn *ec2.EC2, naclID, subnetID string) (string, error) {
	association, err := FindNetworkACLAssociationBySubnetID(conn, subnetID)

	if err != nil {
		return "", fmt.Errorf("error reading EC2 Network ACL Association for EC2 Subnet (%s): %w", subnetID, err)
	}

	input := &ec2.ReplaceNetworkAclAssociationInput{
		AssociationId: association.NetworkAclAssociationId,
		NetworkAclId:  aws.String(naclID),
	}

	log.Printf("[DEBUG] Creating EC2 Network ACL Association: %s", input)
	output, err := conn.ReplaceNetworkAclAssociation(input)

	if err != nil {
		return "", fmt.Errorf("error creating EC2 Network ACL Association: %w", err)
	}

	return aws.StringValue(output.NewAssociationId), nil
}

// networkACLAssociationDelete deletes the specified association between a NACL and subnet.
// The subnet's current association is replaced by an association with the VPC's default NACL.
func networkACLAssociationDelete(conn *ec2.EC2, associationID, naclID string) error {
	input := &ec2.ReplaceNetworkAclAssociationInput{
		AssociationId: aws.String(associationID),
		NetworkAclId:  aws.String(naclID),
	}

	log.Printf("[DEBUG] Deleting EC2 Network ACL Association: %s", associationID)
	_, err := conn.ReplaceNetworkAclAssociation(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidAssociationIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Network ACL Association (%s): %w", associationID, err)
	}

	return nil
}
