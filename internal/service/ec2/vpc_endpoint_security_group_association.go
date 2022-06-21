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

func ResourceVPCEndpointSecurityGroupAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCEndpointSecurityGroupAssociationCreate,
		Read:   resourceVPCEndpointSecurityGroupAssociationRead,
		Delete: resourceVPCEndpointSecurityGroupAssociationDelete,

		Schema: map[string]*schema.Schema{
			"replace_default_association": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCEndpointSecurityGroupAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcEndpointID := d.Get("vpc_endpoint_id").(string)
	securityGroupID := d.Get("security_group_id").(string)
	replaceDefaultAssociation := d.Get("replace_default_association").(bool)

	defaultSecurityGroupID := ""
	if replaceDefaultAssociation {
		vpcEndpoint, err := FindVPCEndpointByID(conn, vpcEndpointID)

		if err != nil {
			return fmt.Errorf("error reading VPC Endpoint (%s): %w", vpcEndpointID, err)
		}

		vpcID := aws.StringValue(vpcEndpoint.VpcId)

		defaultSecurityGroup, err := FindVPCDefaultSecurityGroup(conn, vpcID)

		if err != nil {
			return fmt.Errorf("error reading EC2 VPC (%s) default Security Group: %w", vpcID, err)
		}

		defaultSecurityGroupID = aws.StringValue(defaultSecurityGroup.GroupId)

		if defaultSecurityGroupID == securityGroupID {
			return fmt.Errorf("%s is the default Security Group for EC2 VPC (%s)", securityGroupID, vpcID)
		}

		foundDefaultAssociation := false

		for _, group := range vpcEndpoint.Groups {
			if aws.StringValue(group.GroupId) == defaultSecurityGroupID {
				foundDefaultAssociation = true
				break
			}
		}

		if !foundDefaultAssociation {
			return fmt.Errorf("no association of default Security Group (%s) with VPC Endpoint (%s)", defaultSecurityGroupID, vpcEndpointID)
		}
	}

	err := createVPCEndpointSecurityGroupAssociation(conn, vpcEndpointID, securityGroupID)

	if err != nil {
		return err
	}

	d.SetId(VPCEndpointSecurityGroupAssociationCreateID(vpcEndpointID, securityGroupID))

	if replaceDefaultAssociation {
		// Delete the existing VPC endpoint/default security group association.
		if err := deleteVPCEndpointSecurityGroupAssociation(conn, vpcEndpointID, defaultSecurityGroupID); err != nil {
			return err
		}
	}

	return resourceVPCEndpointSecurityGroupAssociationRead(d, meta)
}

func resourceVPCEndpointSecurityGroupAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcEndpointID := d.Get("vpc_endpoint_id").(string)
	securityGroupID := d.Get("security_group_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", vpcEndpointID, securityGroupID)

	err := FindVPCEndpointSecurityGroupAssociationExists(conn, vpcEndpointID, securityGroupID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Security Group Association (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading VPC Security Group Association (%s): %w", id, err)
	}

	return nil
}

func resourceVPCEndpointSecurityGroupAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcEndpointID := d.Get("vpc_endpoint_id").(string)
	securityGroupID := d.Get("security_group_id").(string)
	replaceDefaultAssociation := d.Get("replace_default_association").(bool)

	if replaceDefaultAssociation {
		vpcEndpoint, err := FindVPCEndpointByID(conn, vpcEndpointID)

		if err != nil {
			return fmt.Errorf("error reading VPC Endpoint (%s): %w", vpcEndpointID, err)
		}

		vpcID := aws.StringValue(vpcEndpoint.VpcId)

		defaultSecurityGroup, err := FindVPCDefaultSecurityGroup(conn, vpcID)

		if err != nil {
			return fmt.Errorf("error reading EC2 VPC (%s) default Security Group: %w", vpcID, err)
		}

		// Add back the VPC endpoint/default security group association.
		err = createVPCEndpointSecurityGroupAssociation(conn, vpcEndpointID, aws.StringValue(defaultSecurityGroup.GroupId))

		if err != nil {
			return err
		}
	}

	return deleteVPCEndpointSecurityGroupAssociation(conn, vpcEndpointID, securityGroupID)
}

// createVPCEndpointSecurityGroupAssociation creates the specified VPC endpoint/security group association.
func createVPCEndpointSecurityGroupAssociation(conn *ec2.EC2, vpcEndpointID, securityGroupID string) error {
	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:       aws.String(vpcEndpointID),
		AddSecurityGroupIds: aws.StringSlice([]string{securityGroupID}),
	}

	log.Printf("[DEBUG] Creating VPC Endpoint Security Group Association: %s", input)
	_, err := conn.ModifyVpcEndpoint(input)

	if err != nil {
		return fmt.Errorf("error creating VPC Endpoint (%s) Security Group (%s) Association: %w", vpcEndpointID, securityGroupID, err)
	}

	return nil
}

// deleteVPCEndpointSecurityGroupAssociation deletes the specified VPC endpoint/security group association.
func deleteVPCEndpointSecurityGroupAssociation(conn *ec2.EC2, vpcEndpointID, securityGroupID string) error {
	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:          aws.String(vpcEndpointID),
		RemoveSecurityGroupIds: aws.StringSlice([]string{securityGroupID}),
	}

	log.Printf("[DEBUG] Deleting VPC Endpoint Security Group Association: %s", input)
	_, err := conn.ModifyVpcEndpoint(input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointIDNotFound, errCodeInvalidGroupNotFound, errCodeInvalidParameter) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting VPC Endpoint (%s) Security Group (%s) Association: %w", vpcEndpointID, securityGroupID, err)
	}

	return nil
}
