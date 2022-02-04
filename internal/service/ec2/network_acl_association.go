package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
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

	subnetID := d.Get("subnet_id").(string)

	association, err := FindNetworkACLAssociationBySubnetID(conn, subnetID)

	if err != nil {
		return fmt.Errorf("error reading EC2 Network ACL Association for EC2 Subnet (%s): %w", subnetID, err)
	}

	input := &ec2.ReplaceNetworkAclAssociationInput{
		AssociationId: association.NetworkAclAssociationId,
		NetworkAclId:  aws.String(d.Get("network_acl_id").(string)),
	}

	log.Printf("[DEBUG] Creating EC2 Network ACL Association: %s", input)
	output, err := conn.ReplaceNetworkAclAssociation(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Network ACL Association: %w", err)
	}

	d.SetId(aws.StringValue(output.NewAssociationId))

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

	_, err = conn.ReplaceNetworkAclAssociation(&ec2.ReplaceNetworkAclAssociationInput{
		AssociationId: aws.String(d.Id()),
		NetworkAclId:  defaultNACL.NetworkAclId,
	})

	if err != nil {
		return fmt.Errorf("error deleting EC2 Network ACL Association (%s): %w", d.Id(), err)
	}

	return nil
}
