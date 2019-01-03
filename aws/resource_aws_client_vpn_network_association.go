package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsClientVpnNetworkAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsClientVpnNetworkAssociationCreate,
		Read:   resourceAwsClientVpnNetworkAssociationRead,
		Delete: resourceAwsClientVpnNetworkAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsClientVpnNetworkAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.AssociateClientVpnTargetNetworkInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
		SubnetId:            aws.String(d.Get("subnet_id").(string)),
	}

	log.Printf("[DEBUG] Creating Client VPN network association: %#v", req)
	resp, err := conn.AssociateClientVpnTargetNetwork(req)
	if err != nil {
		return fmt.Errorf("Error creating Client VPN network association: %s", err)
	}

	d.SetId(*resp.AssociationId)

	return resourceAwsClientVpnEndpointRead(d, meta)
}

func resourceAwsClientVpnNetworkAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	var err error

	result, err := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
		AssociationIds:      []*string{aws.String(d.Id())},
	})

	if err != nil {
		return fmt.Errorf("Error reading Client VPN network association: %s", err)
	}

	d.Set("association_id", result.ClientVpnTargetNetworks[0].AssociationId)
	d.Set("client_vpn_endpoint_id", result.ClientVpnTargetNetworks[0].ClientVpnEndpointId)
	d.Set("security_groups", result.ClientVpnTargetNetworks[0].SecurityGroups)
	d.Set("status", result.ClientVpnTargetNetworks[0].Status)
	d.Set("target_network_id", result.ClientVpnTargetNetworks[0].TargetNetworkId)
	d.Set("vpc_id", result.ClientVpnTargetNetworks[0].VpcId)

	return nil
}

func resourceAwsClientVpnNetworkAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.DisassociateClientVpnTargetNetwork(&ec2.DisassociateClientVpnTargetNetworkInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
		AssociationId:       aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error deleting Client VPN network association: %s", err)
	}

	return nil
}
