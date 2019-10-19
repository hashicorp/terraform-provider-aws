package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

func resourceAwsEc2ClientVpnRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2ClientVpnRouteCreate,
		Read:   resourceAwsEc2ClientVpnRouteRead,
		Delete: resourceAwsEc2ClientVpnRouteDelete,

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_vpn_association_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCIDRNetworkAddress,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2ClientVpnRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	assoc, err := getAwsEc2ClientNetworkAssociation(conn, d.Get("client_vpn_endpoint_id").(string), d.Get("client_vpn_association_id").(string))
	if err != nil {
		return fmt.Errorf("Error getting Client VPN network association: %s", err)
	}

	req := &ec2.CreateClientVpnRouteInput{
		ClientVpnEndpointId:  aws.String(d.Get("client_vpn_endpoint_id").(string)),
		TargetVpcSubnetId:    assoc.TargetNetworkId,
		DestinationCidrBlock: aws.String(d.Get("cidr_block").(string)),
		Description:          aws.String(d.Get("description").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String((v.(string)))
	}

	log.Printf("[DEBUG] Creating Client VPN Route: %#v", req)
	_, err = conn.CreateClientVpnRoute(req)
	if err != nil {
		return fmt.Errorf("Error creating Client VPN route: %s", err)
	}

	d.SetId(resource.UniqueId())

	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnRouteStatusCodeCreating},
		Target:  []string{ec2.ClientVpnRouteStatusCodeActive},
		Refresh: clientVpnRouteRefreshFunc(conn,
			*assoc.TargetNetworkId,
			d.Get("cidr_block").(string),
			d.Get("client_vpn_endpoint_id").(string)),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Client VPN route create: %s", err)
	}

	return resourceAwsEc2ClientVpnRouteRead(d, meta)
}

func resourceAwsEc2ClientVpnRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	assoc, err := getAwsEc2ClientNetworkAssociation(conn, d.Get("client_vpn_endpoint_id").(string), d.Get("client_vpn_association_id").(string))
	if err != nil {
		return fmt.Errorf("Error getting Client VPN network association: %s", err)
	}

	req := &ec2.DescribeClientVpnRoutesInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("destinationCidr"),
				Values: []*string{aws.String(d.Get("cidr_block").(string))},
			},
			{
				Name:   aws.String("targetSubnet"),
				Values: []*string{assoc.TargetNetworkId},
			},
		},
	}

	result, err := conn.DescribeClientVpnRoutes(req)
	if err != nil {
		log.Printf("[WARN] EC2 Client VPN route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if len(result.Routes) != 1 {
		log.Printf("[WARN] EC2 Client VPN route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("client_vpn_endpoint_id", result.Routes[0].ClientVpnEndpointId)
	d.Set("client_vpn_association_id", assoc.AssociationId)
	d.Set("subnet_id", assoc.TargetNetworkId)
	d.Set("cidr_block", result.Routes[0].DestinationCidr)

	return nil
}

func resourceAwsEc2ClientVpnRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.DeleteClientVpnRoute(&ec2.DeleteClientVpnRouteInput{
		ClientVpnEndpointId:  aws.String(d.Get("client_vpn_endpoint_id").(string)),
		DestinationCidrBlock: aws.String(d.Get("cidr_block").(string)),
		TargetVpcSubnetId:    aws.String(d.Get("subnet_id").(string)),
	})

	if err != nil {
		return fmt.Errorf("Error deleting Client VPN route: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnRouteStatusCodeDeleting},
		Target:  []string{"DELETED"},
		Refresh: clientVpnRouteRefreshFunc(conn,
			d.Get("subnet_id").(string),
			d.Get("cidr_block").(string),
			d.Get("client_vpn_endpoint_id").(string)),
		Timeout: d.Timeout(schema.TimeoutDelete),
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Client VPN route to delete: %s", err)
	}

	return nil
}

func clientVpnRouteRefreshFunc(conn *ec2.EC2, subnetId string, cidrBlock string, cvepID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		req := &ec2.DescribeClientVpnRoutesInput{
			ClientVpnEndpointId: aws.String(cvepID),
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("destinationCidr"),
					Values: []*string{aws.String(cidrBlock)},
				},
				{
					Name:   aws.String("targetSubnet"),
					Values: []*string{aws.String(subnetId)},
				},
			},
		}

		resp, err := conn.DescribeClientVpnRoutes(req)
		if err != nil {
			return nil, "", err
		}

		if resp == nil || len(resp.Routes) == 0 || resp.Routes[0] == nil {
			emptyResp := &ec2.DescribeClientVpnRoutesOutput{}
			return emptyResp, "DELETED", nil
		}

		return resp.Routes[0], aws.StringValue(resp.Routes[0].Status.Code), nil

	}
}

func getAwsEc2ClientNetworkAssociation(conn *ec2.EC2, cvepID string, assocId string) (*ec2.TargetNetwork, error) {
	emptyReturn := &ec2.TargetNetwork{}

	req := &ec2.DescribeClientVpnTargetNetworksInput{
		ClientVpnEndpointId: aws.String(cvepID),
		AssociationIds:      []*string{aws.String(assocId)},
	}
	resp, err := conn.DescribeClientVpnTargetNetworks(req)
	if err != nil {
		return emptyReturn, err
	}

	if len(resp.ClientVpnTargetNetworks) != 1 {
		return emptyReturn, fmt.Errorf("Error getting Client VPN network association: %s", assocId)
	}

	return resp.ClientVpnTargetNetworks[0], nil
}
