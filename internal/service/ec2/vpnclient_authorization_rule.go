package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceClientVPNAuthorizationRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceClientVPNAuthorizationRuleCreate,
		Read:   resourceClientVPNAuthorizationRuleRead,
		Delete: resourceClientVPNAuthorizationRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(ClientVPNAuthorizationRuleCreatedTimeout),
			Delete: schema.DefaultTimeout(ClientVPNAuthorizationRuleDeletedTimeout),
		},

		Schema: map[string]*schema.Schema{
			"access_group_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringDoesNotContainAny(","),
				ExactlyOneOf: []string{"access_group_id", "authorize_all_groups"},
			},
			"authorize_all_groups": {
				Type:         schema.TypeBool,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"access_group_id", "authorize_all_groups"},
			},
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"target_network_cidr": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidCIDRNetworkAddress,
			},
		},
	}
}

func resourceClientVPNAuthorizationRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID := d.Get("client_vpn_endpoint_id").(string)
	targetNetworkCIDR := d.Get("target_network_cidr").(string)

	input := &ec2.AuthorizeClientVpnIngressInput{
		ClientVpnEndpointId: aws.String(endpointID),
		TargetNetworkCidr:   aws.String(targetNetworkCIDR),
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

	id := ClientVPNAuthorizationRuleCreateResourceID(endpointID, targetNetworkCIDR, accessGroupID)

	log.Printf("[DEBUG] Creating EC2 Client VPN Authorization Rule: %s", input)
	_, err := conn.AuthorizeClientVpnIngress(input)

	if err != nil {
		return fmt.Errorf("error authorizing EC2 Client VPN Authorization Rule (%s): %w", id, err)
	}

	d.SetId(id)

	if _, err := WaitClientVPNAuthorizationRuleCreated(conn, endpointID, targetNetworkCIDR, accessGroupID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for EC2 Client VPN Authorization Rule (%s) create: %w", d.Id(), err)
	}

	return resourceClientVPNAuthorizationRuleRead(d, meta)
}

func resourceClientVPNAuthorizationRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID, targetNetworkCIDR, accessGroupID, err := ClientVPNAuthorizationRuleParseResourceID(d.Id())

	if err != nil {
		return err
	}

	rule, err := FindClientVPNAuthorizationRuleByThreePartKey(conn, endpointID, targetNetworkCIDR, accessGroupID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Client VPN Authorization Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Client VPN Authorization Rule (%s): %w", d.Id(), err)
	}

	d.Set("access_group_id", rule.GroupId)
	d.Set("authorize_all_groups", rule.AccessAll)
	d.Set("client_vpn_endpoint_id", rule.ClientVpnEndpointId)
	d.Set("description", rule.Description)
	d.Set("target_network_cidr", rule.DestinationCidr)

	return nil
}

func resourceClientVPNAuthorizationRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID, targetNetworkCIDR, accessGroupID, err := ClientVPNAuthorizationRuleParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &ec2.RevokeClientVpnIngressInput{
		ClientVpnEndpointId: aws.String(endpointID),
		RevokeAllGroups:     aws.Bool(d.Get("authorize_all_groups").(bool)),
		TargetNetworkCidr:   aws.String(targetNetworkCIDR),
	}
	if accessGroupID != "" {
		input.AccessGroupId = aws.String(accessGroupID)
	}

	log.Printf("[DEBUG] Deleting EC2 Client VPN Authorization Rule: %s", d.Id())
	_, err = conn.RevokeClientVpnIngress(input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound, errCodeInvalidClientVPNAuthorizationRuleNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error revoking EC2 Client VPN Authorization Rule (%s): %w", d.Id(), err)
	}

	if _, err := WaitClientVPNAuthorizationRuleDeleted(conn, endpointID, targetNetworkCIDR, accessGroupID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for EC2 Client VPN Authorization Rule (%s) delete: %w", d.Id(), err)
	}

	return nil
}

const clientVPNAuthorizationRuleIDSeparator = ","

func ClientVPNAuthorizationRuleCreateResourceID(endpointID, targetNetworkCIDR, accessGroupID string) string {
	parts := []string{endpointID, targetNetworkCIDR}
	if accessGroupID != "" {
		parts = append(parts, accessGroupID)
	}
	id := strings.Join(parts, clientVPNAuthorizationRuleIDSeparator)

	return id
}

func ClientVPNAuthorizationRuleParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, clientVPNAuthorizationRuleIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], "", nil
	}

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected endpoint-id%[2]starget-network-cidr or endpoint-id%[2]starget-network-cidr%[2]sgroup-id", id, clientVPNAuthorizationRuleIDSeparator)
}
