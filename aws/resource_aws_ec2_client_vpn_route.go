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
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_cidr_block": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"target_vpc_subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"origin": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsEc2ClientVpnRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.CreateClientVpnRouteInput{
		ClientVpnEndpointId:  aws.String(d.Get("client_vpn_endpoint_id").(string)),
		DestinationCidrBlock: aws.String(d.Get("destination_cidr_block").(string)),
		TargetVpcSubnetId:    aws.String(d.Get("target_vpc_subnet_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}

	_, err := conn.CreateClientVpnRoute(req)

	if err != nil {
		return fmt.Errorf("error creating client VPN route: %s", err)
	}

	d.SetId(resource.UniqueId())
	return resourceAwsEc2ClientVpnRouteRead(d, meta)
}

func resourceAwsEc2ClientVpnRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	resp, err := conn.DescribeClientVpnRoutes(&ec2.DescribeClientVpnRoutesInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
	})

	if isAWSErr(err, "InvalidClientVpnRouteNotFound", "") {
		log.Printf("[WARN] EC2 Client VPN Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading client VPN route: %s", err)
	}

	if resp == nil || len(resp.Routes) == 0 || resp.Routes[0] == nil {
		log.Printf("[WARN] EC2 Client VPN Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if resp.Routes[0].Status != nil && aws.StringValue(resp.Routes[0].Status.Code) == ec2.ClientVpnRouteStatusCodeDeleting {
		log.Printf("[WARN] EC2 Client VPN Route (%s) has been deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	_, exists := d.GetOk("client_vpn_endpoint_id")
	if !exists {
		d.Set("client_vpn_endpoint_id", resp.Routes[0].ClientVpnEndpointId)
	}
	_, exists = d.GetOk("destination_cidr_block")
	if !exists {
		d.Set("destination_cidr_block", resp.Routes[0].DestinationCidr)
	}
	_, exists = d.GetOk("description")
	if !exists {
		d.Set("description", resp.Routes[0].Description)
	}
	_, exists = d.GetOk("target_vpc_subnet_id")
	if !exists {
		d.Set("target_vpc_subnet_id", resp.Routes[0].TargetSubnet)
	}
	_, exists = d.GetOk("origin")
	if !exists {
		d.Set("origin", resp.Routes[0].Origin)
	}
	_, exists = d.GetOk("status")
	if !exists {
		d.Set("status", resp.Routes[0].Status)
	}
	_, exists = d.GetOk("type")
	if !exists {
		d.Set("type", resp.Routes[0].Type)
	}

	return nil
}

func resourceAwsEc2ClientVpnRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.DeleteClientVpnRoute(&ec2.DeleteClientVpnRouteInput{
		ClientVpnEndpointId:  aws.String(d.Get("client_vpn_endpoint_id").(string)),
		DestinationCidrBlock: aws.String(d.Get("destination_cidr_block").(string)),
		TargetVpcSubnetId:    aws.String(d.Get("target_vpc_subnet_id").(string)),
	})

	if isAWSErr(err, "InvalidClientVpnRouteNotFound", "") {
		log.Printf("[WARN] EC2 Client VPN Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting client VPN route: %s", err)
	}

	return nil
}
