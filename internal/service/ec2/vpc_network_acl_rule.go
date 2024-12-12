// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_network_acl_rule", name="Network ACL Rule")
func resourceNetworkACLRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkACLRuleCreate,
		ReadWithoutTimeout:   resourceNetworkACLRuleRead,
		DeleteWithoutTimeout: resourceNetworkACLRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkACLRuleImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCIDRBlock: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{names.AttrCIDRBlock, "ipv6_cidr_block"},
			},
			"egress": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"from_port": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"icmp_code": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"icmp_type": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"ipv6_cidr_block": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{names.AttrCIDRBlock, "ipv6_cidr_block"},
			},
			"network_acl_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrProtocol: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if v, ok := ianaProtocolAToI[old]; ok {
						old = strconv.Itoa(v)
					}
					if v, ok := ianaProtocolAToI[new]; ok {
						new = strconv.Itoa(v)
					}

					return old == new
				},
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					_, err := networkACLProtocolNumber(v.(string))

					if err != nil {
						errors = append(errors, fmt.Errorf("%q : %w", k, err))
					}

					return
				},
			},
			"rule_action": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
				ValidateDiagFunc: enum.Validate[awstypes.RuleAction](),
			},
			"rule_number": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"to_port": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNetworkACLRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	protocol := d.Get(names.AttrProtocol).(string)
	protocolNumber, err := networkACLProtocolNumber(protocol)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Network ACL Rule: %s", err)
	}

	naclID, egress, ruleNumber := d.Get("network_acl_id").(string), d.Get("egress").(bool), d.Get("rule_number").(int)

	// CreateNetworkAclEntry succeeds if there is an existing rule with identical attributes.
	_, err = findNetworkACLEntryByThreePartKey(ctx, conn, naclID, egress, ruleNumber)

	switch {
	case err == nil:
		return sdkdiag.AppendFromErr(diags, networkACLEntryAlreadyExistsError(naclID, egress, ruleNumber))
	case tfresource.NotFound(err):
		break
	default:
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network ACL Rule: %s", err)
	}

	input := &ec2.CreateNetworkAclEntryInput{
		Egress:       aws.Bool(egress),
		NetworkAclId: aws.String(naclID),
		PortRange: &awstypes.PortRange{
			From: aws.Int32(int32(d.Get("from_port").(int))),
			To:   aws.Int32(int32(d.Get("to_port").(int))),
		},
		Protocol:   aws.String(strconv.Itoa(protocolNumber)),
		RuleAction: awstypes.RuleAction(d.Get("rule_action").(string)),
		RuleNumber: aws.Int32(int32(ruleNumber)),
	}

	if v, ok := d.GetOk(names.AttrCIDRBlock); ok {
		input.CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_cidr_block"); ok {
		input.Ipv6CidrBlock = aws.String(v.(string))
	}

	// Specify additional required fields for ICMP. For the list
	// of ICMP codes and types, see: https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml
	if protocolNumber == 1 || protocolNumber == 58 {
		input.IcmpTypeCode = &awstypes.IcmpTypeCode{
			Code: aws.Int32(int32(d.Get("icmp_code").(int))),
			Type: aws.Int32(int32(d.Get("icmp_type").(int))),
		}
	}

	log.Printf("[DEBUG] Creating EC2 Network ACL Rule: %#v", input)
	_, err = conn.CreateNetworkAclEntry(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Network ACL (%s) Rule (egress: %t)(%d): %s", naclID, egress, ruleNumber, err)
	}

	d.SetId(networkACLRuleCreateResourceID(naclID, ruleNumber, egress, protocol))

	return append(diags, resourceNetworkACLRuleRead(ctx, d, meta)...)
}

func resourceNetworkACLRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	egress := d.Get("egress").(bool)
	naclID := d.Get("network_acl_id").(string)
	ruleNumber := d.Get("rule_number").(int)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findNetworkACLEntryByThreePartKey(ctx, conn, naclID, egress, ruleNumber)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network ACL Rule %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network ACL Rule (%s): %s", d.Id(), err)
	}

	naclEntry := outputRaw.(*awstypes.NetworkAclEntry)

	d.Set(names.AttrCIDRBlock, naclEntry.CidrBlock)
	d.Set("egress", naclEntry.Egress)
	d.Set("ipv6_cidr_block", naclEntry.Ipv6CidrBlock)
	if naclEntry.IcmpTypeCode != nil {
		d.Set("icmp_code", naclEntry.IcmpTypeCode.Code)
		d.Set("icmp_type", naclEntry.IcmpTypeCode.Type)
	}
	if naclEntry.PortRange != nil {
		d.Set("from_port", naclEntry.PortRange.From)
		d.Set("to_port", naclEntry.PortRange.To)
	}
	d.Set("rule_action", naclEntry.RuleAction)
	d.Set("rule_number", naclEntry.RuleNumber)

	if v := aws.ToString(naclEntry.Protocol); v != "" {
		// The AWS network ACL API only speaks protocol numbers, and
		// that's all we record.
		protocolNumber, err := networkACLProtocolNumber(v)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Network ACL Rule (%s): %s", d.Id(), err)
		}

		d.Set(names.AttrProtocol, strconv.Itoa(protocolNumber))
	} else {
		d.Set(names.AttrProtocol, nil)
	}

	return diags
}

func resourceNetworkACLRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 Network ACL Rule: %s", d.Id())
	_, err := conn.DeleteNetworkAclEntry(ctx, &ec2.DeleteNetworkAclEntryInput{
		Egress:       aws.Bool(d.Get("egress").(bool)),
		NetworkAclId: aws.String(d.Get("network_acl_id").(string)),
		RuleNumber:   aws.Int32(int32(d.Get("rule_number").(int))),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkACLIDNotFound, errCodeInvalidNetworkACLEntryNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Network ACL Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceNetworkACLRuleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), networkACLRuleImportIDSeparator)

	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%[1]s), expected NETWORK_ACL_ID%[2]sRULE_NUMBER%[2]sPROTOCOL%[2]sEGRESS", d.Id(), networkACLRuleImportIDSeparator)
	}

	naclID := parts[0]
	ruleNumber, err := strconv.Atoi(parts[1])

	if err != nil {
		return nil, err
	}

	protocol := parts[2]
	egress, err := strconv.ParseBool(parts[3])

	if err != nil {
		return nil, err
	}

	d.SetId(networkACLRuleCreateResourceID(naclID, ruleNumber, egress, protocol))
	d.Set("egress", egress)
	d.Set("network_acl_id", naclID)
	d.Set("rule_number", ruleNumber)

	return []*schema.ResourceData{d}, nil
}

const networkACLRuleImportIDSeparator = ":"

func networkACLRuleCreateResourceID(naclID string, ruleNumber int, egress bool, protocol string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", naclID))
	buf.WriteString(fmt.Sprintf("%d-", ruleNumber))
	buf.WriteString(fmt.Sprintf("%t-", egress))
	buf.WriteString(fmt.Sprintf("%s-", protocol))
	return fmt.Sprintf("nacl-%d", create.StringHashcode(buf.String()))
}
