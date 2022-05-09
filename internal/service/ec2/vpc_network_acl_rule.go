package ec2

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceNetworkACLRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkACLRuleCreate,
		Read:   resourceNetworkACLRuleRead,
		Delete: resourceNetworkACLRuleDelete,

		Importer: &schema.ResourceImporter{
			State: resourceNetworkACLRuleImport,
		},

		Schema: map[string]*schema.Schema{
			"cidr_block": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"cidr_block", "ipv6_cidr_block"},
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
				ExactlyOneOf: []string{"cidr_block", "ipv6_cidr_block"},
			},
			"network_acl_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"protocol": {
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
				ValidateFunc: validation.StringInSlice(ec2.RuleAction_Values(), true),
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

func resourceNetworkACLRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	protocol := d.Get("protocol").(string)
	protocolNumber, err := networkACLProtocolNumber(protocol)

	if err != nil {
		return err
	}

	egress := d.Get("egress").(bool)
	naclID := d.Get("network_acl_id").(string)
	ruleNumber := d.Get("rule_number").(int)

	input := &ec2.CreateNetworkAclEntryInput{
		Egress:       aws.Bool(egress),
		NetworkAclId: aws.String(naclID),
		PortRange: &ec2.PortRange{
			From: aws.Int64(int64(d.Get("from_port").(int))),
			To:   aws.Int64(int64(d.Get("to_port").(int))),
		},
		Protocol:   aws.String(strconv.Itoa(protocolNumber)),
		RuleAction: aws.String(d.Get("rule_action").(string)),
		RuleNumber: aws.Int64(int64(ruleNumber)),
	}

	if v, ok := d.GetOk("cidr_block"); ok {
		input.CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_cidr_block"); ok {
		input.Ipv6CidrBlock = aws.String(v.(string))
	}

	// Specify additional required fields for ICMP. For the list
	// of ICMP codes and types, see: https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml
	if protocolNumber == 1 || protocolNumber == 58 {
		input.IcmpTypeCode = &ec2.IcmpTypeCode{
			Code: aws.Int64(int64(d.Get("icmp_code").(int))),
			Type: aws.Int64(int64(d.Get("icmp_type").(int))),
		}
	}

	log.Printf("[DEBUG] Creating EC2 Network ACL Rule: %s", input)
	_, err = conn.CreateNetworkAclEntry(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Network ACL (%s) Rule (egress:%t)(%d): %w", naclID, egress, ruleNumber, err)
	}

	d.SetId(NetworkACLRuleCreateResourceID(naclID, ruleNumber, egress, protocol))

	return resourceNetworkACLRuleRead(d, meta)
}

func resourceNetworkACLRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	egress := d.Get("egress").(bool)
	naclID := d.Get("network_acl_id").(string)
	ruleNumber := d.Get("rule_number").(int)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(propagationTimeout, func() (interface{}, error) {
		return FindNetworkACLEntryByThreePartKey(conn, naclID, egress, ruleNumber)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network ACL Rule %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Network ACL Rule (%s): %w", d.Id(), err)
	}

	naclEntry := outputRaw.(*ec2.NetworkAclEntry)

	d.Set("cidr_block", naclEntry.CidrBlock)
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

	if v := aws.StringValue(naclEntry.Protocol); v != "" {
		// The AWS network ACL API only speaks protocol numbers, and
		// that's all we record.
		protocolNumber, err := networkACLProtocolNumber(v)

		if err != nil {
			return err
		}

		d.Set("protocol", strconv.Itoa(protocolNumber))
	} else {
		d.Set("protocol", nil)
	}

	return nil
}

func resourceNetworkACLRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting EC2 Network ACL Rule: %s", d.Id())
	_, err := conn.DeleteNetworkAclEntry(&ec2.DeleteNetworkAclEntryInput{
		Egress:       aws.Bool(d.Get("egress").(bool)),
		NetworkAclId: aws.String(d.Get("network_acl_id").(string)),
		RuleNumber:   aws.Int64(int64(d.Get("rule_number").(int))),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidNetworkAclEntryNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Network ACL Rule (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceNetworkACLRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), NetworkACLRuleImportIDSeparator)

	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%[1]s), expected NETWORK_ACL_ID%[2]sRULE_NUMBER%[2]sPROTOCOL%[2]sEGRESS", d.Id(), NetworkACLRuleImportIDSeparator)
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

	d.SetId(NetworkACLRuleCreateResourceID(naclID, ruleNumber, egress, protocol))
	d.Set("egress", egress)
	d.Set("network_acl_id", naclID)
	d.Set("rule_number", ruleNumber)

	return []*schema.ResourceData{d}, nil
}

const NetworkACLRuleImportIDSeparator = ":"

func NetworkACLRuleCreateResourceID(naclID string, ruleNumber int, egress bool, protocol string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", naclID))
	buf.WriteString(fmt.Sprintf("%d-", ruleNumber))
	buf.WriteString(fmt.Sprintf("%t-", egress))
	buf.WriteString(fmt.Sprintf("%s-", protocol))
	return fmt.Sprintf("nacl-%d", create.StringHashcode(buf.String()))
}
