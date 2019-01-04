package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsEc2ClientVpnNetworkAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2ClientVpnNetworkAssociationCreate,
		Read:   resourceAwsEc2ClientVpnNetworkAssociationRead,
		Delete: resourceAwsEc2ClientVpnNetworkAssociationDelete,
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
			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsEc2ClientVpnNetworkAssociationCreate(d *schema.ResourceData, meta interface{}) error {
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

	stateConf := &resource.StateChangeConf{
		Pending: []string{"associating"},
		Target:  []string{"associated"},
		Refresh: clientVpnNetworkAssociationRefreshFunc(conn, d.Id(), d.Get("client_vpn_endpoint_id").(string)),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	log.Printf("[DEBUG] Waiting for Client VPN endpoint to associate with target network: %s", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Client VPN endpoint to associate with target network: %s", err)
	}

	return resourceAwsEc2ClientVpnNetworkAssociationRead(d, meta)
}

func resourceAwsEc2ClientVpnNetworkAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	var err error

	result, err := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
		AssociationIds:      []*string{aws.String(d.Id())},
	})

	if err != nil {
		return fmt.Errorf("Error reading Client VPN network association: %s", err)
	}

	d.Set("client_vpn_endpoint_id", result.ClientVpnTargetNetworks[0].ClientVpnEndpointId)
	d.Set("security_groups", result.ClientVpnTargetNetworks[0].SecurityGroups)
	d.Set("status", result.ClientVpnTargetNetworks[0].Status)
	d.Set("subnet_id", result.ClientVpnTargetNetworks[0].TargetNetworkId)
	d.Set("vpc_id", result.ClientVpnTargetNetworks[0].VpcId)

	return nil
}

func resourceAwsEc2ClientVpnNetworkAssociationDelete(d *schema.ResourceData, meta interface{}) error {
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

func clientVpnNetworkAssociationRefreshFunc(conn *ec2.EC2, cvnaID string, cvepID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: aws.String(cvepID),
			AssociationIds:      []*string{aws.String(cvnaID)},
		})
		if err != nil {
			return nil, "", err
		}

		return resp.ClientVpnTargetNetworks[0], aws.StringValue(resp.ClientVpnTargetNetworks[0].Status.Code), nil
	}
}
