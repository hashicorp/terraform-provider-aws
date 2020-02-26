package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

func resourceAwsEc2ClientVpnAuthorization() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2AuthorizeClientVpnIngress,
		Read:   resourceAwsEc2ReadClientVpnIngress,
		Delete: resourceAwsEc2RevokeClientVpnIngress,

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
			"target_network_cidr": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCIDRNetworkAddress,
			},
			"authorize_all_groups": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"access_group_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateAccessGroupId,
			},
		},
	}
}

func resourceAwsEc2AuthorizeClientVpnIngress(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	var err error

	req := &ec2.AuthorizeClientVpnIngressInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
		TargetNetworkCidr:   aws.String(d.Get("target_network_cidr").(string)),
		Description:         aws.String(d.Get("description").(string)),
	}

	if v, ok := d.GetOk("access_group_id"); ok {
		req.AccessGroupId = aws.String((v.(string)))
	} else if v, ok := d.GetOk("authorize_all_groups"); ok {
		req.AuthorizeAllGroups = aws.Bool((v.(bool)))
	}

	log.Printf("[DEBUG] Authorizing Client VPN: %#v", req)
	_, err = conn.AuthorizeClientVpnIngress(req)
	if err != nil {
		return fmt.Errorf("Error authorizing Client VPN: %s", err)
	}

	d.SetId(resource.UniqueId())

	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnAuthorizationRuleStatusCodeAuthorizing},
		Target:  []string{ec2.ClientVpnAuthorizationRuleStatusCodeActive},
		Refresh: clientVpnAuthorizationRefreshFunc(conn, d.Get("target_network_cidr").(string), d.Get("client_vpn_endpoint_id").(string)),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Client VPN authorize: %s", err)
	}

	return resourceAwsEc2ReadClientVpnIngress(d, meta)
}

func resourceAwsEc2ReadClientVpnIngress(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	filter := &ec2.Filter{
		Name:   aws.String("destination-cidr"),
		Values: []*string{aws.String(d.Get("target_network_cidr").(string))},
	}
	req := &ec2.DescribeClientVpnAuthorizationRulesInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
		Filters:             []*ec2.Filter{filter},
	}

	result, err := conn.DescribeClientVpnAuthorizationRules(req)
	if err != nil {
		log.Printf("[WARN] EC2 Client VPN authorization (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("client_vpn_endpoint_id", result.AuthorizationRules[0].ClientVpnEndpointId)
	d.Set("target_cidr_block", result.AuthorizationRules[0].DestinationCidr)
	d.Set("access_group_id", result.AuthorizationRules[0].GroupId)
	d.Set("authorize_all_groups", result.AuthorizationRules[0].AccessAll)

	return nil
}

func resourceAwsEc2RevokeClientVpnIngress(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.RevokeClientVpnIngressInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
		TargetNetworkCidr:   aws.String(d.Get("target_network_cidr").(string)),
	}

	if v, ok := d.GetOk("access_group_id"); ok {
		req.AccessGroupId = aws.String((v.(string)))
	} else if v, ok := d.GetOk("authorize_all_groups"); ok {
		req.RevokeAllGroups = aws.Bool((v.(bool)))
	}

	_, err := conn.RevokeClientVpnIngress(req)
	if err != nil {
		return fmt.Errorf("Error deleting Client VPN authorization: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnAuthorizationRuleStatusCodeRevoking},
		Target:  []string{"DELETED"},
		Refresh: clientVpnAuthorizationRefreshFunc(conn, d.Get("target_network_cidr").(string), d.Get("client_vpn_endpoint_id").(string)),
		Timeout: d.Timeout(schema.TimeoutDelete),
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Client VPN authorize to delete: %s", err)
	}

	return nil
}

func clientVpnAuthorizationRefreshFunc(conn *ec2.EC2, cidrBlock string, cvepID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		filter := &ec2.Filter{
			Name:   aws.String("destination-cidr"),
			Values: []*string{aws.String(cidrBlock)},
		}
		req := &ec2.DescribeClientVpnAuthorizationRulesInput{
			ClientVpnEndpointId: aws.String(cvepID),
			Filters:             []*ec2.Filter{filter},
		}

		resp, err := conn.DescribeClientVpnAuthorizationRules(req)

		if err != nil {
			return nil, "", err
		}

		if resp == nil || len(resp.AuthorizationRules) == 0 || resp.AuthorizationRules[0] == nil {
			emptyResp := &ec2.DescribeClientVpnAuthorizationRulesOutput{}
			return emptyResp, "DELETED", nil
		}

		return resp.AuthorizationRules[0], aws.StringValue(resp.AuthorizationRules[0].Status.Code), nil

	}
}
