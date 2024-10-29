// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_client_vpn_authorization_rule", name="Client VPN Authorization Rule")
func resourceClientVPNAuthorizationRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClientVPNAuthorizationRuleCreate,
		ReadWithoutTimeout:   resourceClientVPNAuthorizationRuleRead,
		DeleteWithoutTimeout: resourceClientVPNAuthorizationRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
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
			names.AttrDescription: {
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

func resourceClientVPNAuthorizationRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID := d.Get("client_vpn_endpoint_id").(string)
	targetNetworkCIDR := d.Get("target_network_cidr").(string)
	input := &ec2.AuthorizeClientVpnIngressInput{
		ClientToken:         aws.String(sdkid.UniqueId()),
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

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	id := clientVPNAuthorizationRuleCreateResourceID(endpointID, targetNetworkCIDR, accessGroupID)
	_, err := conn.AuthorizeClientVpnIngress(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "authorizing EC2 Client VPN Authorization Rule (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitClientVPNAuthorizationRuleCreated(ctx, conn, endpointID, targetNetworkCIDR, accessGroupID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Client VPN Authorization Rule (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceClientVPNAuthorizationRuleRead(ctx, d, meta)...)
}

func resourceClientVPNAuthorizationRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID, targetNetworkCIDR, accessGroupID, err := clientVPNAuthorizationRuleParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	rule, err := findClientVPNAuthorizationRuleByThreePartKey(ctx, conn, endpointID, targetNetworkCIDR, accessGroupID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Client VPN Authorization Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Client VPN Authorization Rule (%s): %s", d.Id(), err)
	}

	d.Set("access_group_id", rule.GroupId)
	d.Set("authorize_all_groups", rule.AccessAll)
	d.Set("client_vpn_endpoint_id", rule.ClientVpnEndpointId)
	d.Set(names.AttrDescription, rule.Description)
	d.Set("target_network_cidr", rule.DestinationCidr)

	return diags
}

func resourceClientVPNAuthorizationRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID, targetNetworkCIDR, accessGroupID, err := clientVPNAuthorizationRuleParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
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
	_, err = conn.RevokeClientVpnIngress(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound, errCodeInvalidClientVPNAuthorizationRuleNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Client VPN Authorization Rule (%s): %s", d.Id(), err)
	}

	if _, err := waitClientVPNAuthorizationRuleDeleted(ctx, conn, endpointID, targetNetworkCIDR, accessGroupID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Client VPN Authorization Rule (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const clientVPNAuthorizationRuleIDSeparator = ","

func clientVPNAuthorizationRuleCreateResourceID(endpointID, targetNetworkCIDR, accessGroupID string) string {
	parts := []string{endpointID, targetNetworkCIDR}
	if accessGroupID != "" {
		parts = append(parts, accessGroupID)
	}
	id := strings.Join(parts, clientVPNAuthorizationRuleIDSeparator)

	return id
}

func clientVPNAuthorizationRuleParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, clientVPNAuthorizationRuleIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], "", nil
	}

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected endpoint-id%[2]starget-network-cidr or endpoint-id%[2]starget-network-cidr%[2]sgroup-id", id, clientVPNAuthorizationRuleIDSeparator)
}
