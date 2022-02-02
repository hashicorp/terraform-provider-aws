package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceClientVPNSecurityGroupsAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceClientVPNSecurityGroupsAssociationCreate,
		Read:   resourceClientVPNSecurityGroupsAssociationRead,
		Update: resourceClientVPNSecurityGroupsAssociationUpdate,
		Delete: resourceClientVPNSecurityGroupsAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				MinItems: 1,
				MaxItems: 5,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceClientVPNSecurityGroupsAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID := d.Get("client_vpn_endpoint_id").(string)
	vpcID := d.Get("vpc_id").(string)
	id := ClientVPNSecurityGroupsAssociationCreateResourceID(endpointID, vpcID)
	input := &ec2.ApplySecurityGroupsToClientVpnTargetNetworkInput{
		ClientVpnEndpointId: aws.String(endpointID),
		SecurityGroupIds:    flex.ExpandStringSet(d.Get("security_group_ids").(*schema.Set)),
		VpcId:               aws.String(vpcID),
	}

	log.Printf("[DEBUG] Creating EC2 Client VPN Security Groups Association: %s", input)
	_, err := conn.ApplySecurityGroupsToClientVpnTargetNetwork(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Client VPN Security Groups Association (%s): %w", id, err)
	}

	d.SetId(id)

	return resourceClientVPNSecurityGroupsAssociationRead(d, meta)
}

func resourceClientVPNSecurityGroupsAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID, vpcID, err := ClientVPNSecurityGroupsAssociationParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &ec2.DescribeClientVpnTargetNetworksInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters: BuildAttributeFilterList(map[string]string{
			"vpc-id": vpcID,
		}),
	}

	_, err = FindClientVPNNetworkAssociations(conn, input)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Client VPN Security Groups Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Client VPN Security Groups Association (%s): %w", d.Id(), err)
	}

	d.Set("client_vpn_endpoint_id", endpointID)
	d.Set("vpc_id", vpcID)

	return nil
}

func resourceClientVPNSecurityGroupsAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID, vpcID, err := ClientVPNSecurityGroupsAssociationParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &ec2.ApplySecurityGroupsToClientVpnTargetNetworkInput{
		ClientVpnEndpointId: aws.String(endpointID),
		SecurityGroupIds:    flex.ExpandStringSet(d.Get("security_group_ids").(*schema.Set)),
		VpcId:               aws.String(vpcID),
	}

	log.Printf("[DEBUG] Updating EC2 Client VPN Security Groups Association: %s", input)
	_, err = conn.ApplySecurityGroupsToClientVpnTargetNetwork(input)

	if err != nil {
		return fmt.Errorf("error updating EC2 Client VPN Security Groups Association (%s): %w", d.Id(), err)
	}

	return resourceClientVPNSecurityGroupsAssociationRead(d, meta)
}

func resourceClientVPNSecurityGroupsAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID, vpcID, err := ClientVPNSecurityGroupsAssociationParseResourceID(d.Id())

	if err != nil {
		return err
	}

	defaultSecurityGroup, err := FindVPCDefaultSecurityGroup(conn, vpcID)

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s) default Security Group: %w", vpcID, err)
	}

	log.Printf("[DEBUG] Deleting EC2 Client VPN Security Groups Association: %s", d.Id())
	_, err = conn.ApplySecurityGroupsToClientVpnTargetNetwork(&ec2.ApplySecurityGroupsToClientVpnTargetNetworkInput{
		ClientVpnEndpointId: aws.String(endpointID),
		SecurityGroupIds:    []*string{defaultSecurityGroup.GroupId},
		VpcId:               aws.String(vpcID),
	})

	if err != nil {
		return fmt.Errorf("error updating EC2 Client VPN Security Groups Association (%s): %w", d.Id(), err)
	}

	return nil
}

const clientVPNSecurityGroupsAssociationIDSeparator = "/"

func ClientVPNSecurityGroupsAssociationCreateResourceID(endpointID, vpcID string) string {
	parts := []string{endpointID, vpcID}
	id := strings.Join(parts, clientVPNSecurityGroupsAssociationIDSeparator)

	return id
}

func ClientVPNSecurityGroupsAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, clientVPNAuthorizationRuleIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected EndpointID%[2]sVPCID", id, clientVPNAuthorizationRuleIDSeparator)
}
