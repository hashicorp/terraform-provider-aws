package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsEc2ClientVpnAuthorizationRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2ClientVpnAuthorizationRuleCreate,
		Read:   resourceAwsEc2ClientVpnAuthorizationRuleRead,
		Delete: resourceAwsEc2ClientVpnAuthorizationRuleDelete,

		Schema: map[string]*schema.Schema{
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
			"access_group_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"access_group_id", "authorize_all_groups"},
			},
			"authorize_all_groups": {
				Type:         schema.TypeBool,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"access_group_id", "authorize_all_groups"},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsEc2ClientVpnAuthorizationRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointID := d.Get("client_vpn_endpoint_id").(string)
	targetNetworkCidr := d.Get("target_network_cidr").(string)

	input := &ec2.AuthorizeClientVpnIngressInput{
		ClientVpnEndpointId: aws.String(endpointID),
		TargetNetworkCidr:   aws.String(targetNetworkCidr),
	}

	var accessGroupID string
	if v, ok := d.GetOk("access_group_id"); ok {
		accessGroupID = v.(string)
		input.AccessGroupId = aws.String(accessGroupID)
	}

	if v, ok := d.GetOk("authorize_all_groups"); ok {
		input.AuthorizeAllGroups = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Client VPN authorization rule: %#v", input)
	_, err := conn.AuthorizeClientVpnIngress(input)
	if err != nil {
		return fmt.Errorf("Error creating Client VPN authorization rule: %s", err)
	}

	// TODO wait

	d.SetId(ec2ClientVpnAuthorizationRuleCreateID(endpointID, targetNetworkCidr, accessGroupID))

	// stateConf := &resource.StateChangeConf{
	// 	Pending: []string{ec2.AssociationStatusCodeAssociating},
	// 	Target:  []string{ec2.AssociationStatusCodeAssociated},
	// 	Refresh: clientVpnNetworkAssociationRefreshFunc(conn, d.Id(), d.Get("client_vpn_endpoint_id").(string)),
	// 	Timeout: d.Timeout(schema.TimeoutCreate),
	// }

	// log.Printf("[DEBUG] Waiting for Client VPN endpoint to associate with target network: %s", d.Id())
	// _, err = stateConf.WaitForState()
	// if err != nil {
	// 	return fmt.Errorf("Error waiting for Client VPN endpoint to associate with target network: %s", err)
	// }

	return resourceAwsEc2ClientVpnAuthorizationRuleRead(d, meta)
}

const errCodeClientVpnEndpointAuthorizationRuleNotFound = "InvalidClientVpnEndpointAuthorizationRuleNotFound"

func resourceAwsEc2ClientVpnAuthorizationRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DescribeClientVpnAuthorizationRulesInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
	}

	result, err := conn.DescribeClientVpnAuthorizationRules(input)

	if isAWSErr(err, errCodeClientVpnEndpointAuthorizationRuleNotFound, "") {
		log.Printf("[WARN] EC2 Client VPN authorization rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error reading Client VPN authorization rule: %w", err)
	}

	if result == nil || len(result.AuthorizationRules) == 0 || result.AuthorizationRules[0] == nil {
		log.Printf("[WARN] EC2 Client VPN authorization rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("client_vpn_endpoint_id", result.AuthorizationRules[0].ClientVpnEndpointId)
	d.Set("target_network_cidr", result.AuthorizationRules[0].DestinationCidr)
	// TODO: Fill in Groups
	d.Set("description", result.AuthorizationRules[0].Description)

	return nil
}

func resourceAwsEc2ClientVpnAuthorizationRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.RevokeClientVpnIngressInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
		TargetNetworkCidr:   aws.String(d.Get("target_network_cidr").(string)),
	}
	if v, ok := d.GetOk("access_group_id"); ok {
		input.AccessGroupId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("authorize_all_groups"); ok {
		input.RevokeAllGroups = aws.Bool(v.(bool))
	}

	_, err := conn.RevokeClientVpnIngress(input)

	if isAWSErr(err, errCodeClientVpnEndpointAuthorizationRuleNotFound, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Client VPN authorization rule: %w", err)
	}

	// WAIT
	// stateConf := &resource.StateChangeConf{
	// 	Pending: []string{ec2.AssociationStatusCodeDisassociating},
	// 	Target:  []string{ec2.AssociationStatusCodeDisassociated},
	// 	Refresh: clientVpnNetworkAssociationRefreshFunc(conn, d.Id(), d.Get("client_vpn_endpoint_id").(string)),
	// 	Timeout: d.Timeout(schema.TimeoutDelete),
	// }

	// log.Printf("[DEBUG] Waiting to revoke Client VPN authorization rule (%s)", d.Id())
	// _, err = stateConf.WaitForState()
	// if err != nil {
	// 	return fmt.Errorf("error waiting to revoke Client VPN authorization rule (%s): %w", d.Id(), err)
	// }

	return nil
}

// func clientVpnNetworkAssociationRefreshFunc(conn *ec2.EC2, cvnaID string, cvepID string) resource.StateRefreshFunc {
// 	return func() (interface{}, string, error) {
// 		resp, err := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
// 			ClientVpnEndpointId: aws.String(cvepID),
// 			AssociationIds:      []*string{aws.String(cvnaID)},
// 		})

// 		if isAWSErr(err, "InvalidClientVpnAssociationId.NotFound", "") || isAWSErr(err, clientVpnEndpointIdNotFoundError, "") {
// 			return 42, ec2.AssociationStatusCodeDisassociated, nil
// 		}

// 		if err != nil {
// 			return nil, "", err
// 		}

// 		if resp == nil || len(resp.ClientVpnTargetNetworks) == 0 || resp.ClientVpnTargetNetworks[0] == nil {
// 			return 42, ec2.AssociationStatusCodeDisassociated, nil
// 		}

// 		return resp.ClientVpnTargetNetworks[0], aws.StringValue(resp.ClientVpnTargetNetworks[0].Status.Code), nil
// 	}
// }

const ec2ClientVpnAuthorizationRuleIDSeparator = ","

func ec2ClientVpnAuthorizationRuleCreateID(endpointID, targetNetworkCidr, accessGroupID string) string {
	parts := []string{endpointID, targetNetworkCidr}
	if accessGroupID != "" {
		parts = append(parts, accessGroupID)
	}
	id := strings.Join(parts, ec2ClientVpnAuthorizationRuleIDSeparator)
	return id
}

func ec2ClientVpnAuthorizationRuleParseID(id string) (string, string, string, error) {
	parts := strings.Split(id, ec2ClientVpnAuthorizationRuleIDSeparator)
	if len(parts) == 2 {
		return parts[0], parts[1], "", nil
	}
	if len(parts) == 3 {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "",
		fmt.Errorf("unexpected format for ID (%q), expected endpoint-id"+ec2ClientVpnAuthorizationRuleIDSeparator+
			"target-network-cidr or endpoint-id"+ec2ClientVpnAuthorizationRuleIDSeparator+"target-network-cidr"+
			ec2ClientVpnAuthorizationRuleIDSeparator+"group-id", id)
}
