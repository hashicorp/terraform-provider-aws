package ec2

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSecurityGroupRule() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityGroupRuleCreate,
		ReadWithoutTimeout:   resourceSecurityGroupRuleRead,
		UpdateWithoutTimeout: resourceSecurityGroupRuleUpdate,
		DeleteWithoutTimeout: resourceSecurityGroupRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceSecurityGroupRuleImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},

		SchemaVersion: 2,
		MigrateState:  SecurityGroupRuleMigrateState,

		Schema: map[string]*schema.Schema{
			"cidr_blocks": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
				},
				ConflictsWith: []string{"source_security_group_id", "self"},
				AtLeastOneOf:  []string{"cidr_blocks", "ipv6_cidr_blocks", "prefix_list_ids", "self", "source_security_group_id"},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validSecurityGroupRuleDescription,
			},
			"from_port": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
				// Support existing configurations that have non-zero from_port and to_port defined with all protocols
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					protocol := ProtocolForValue(d.Get("protocol").(string))
					if protocol == "-1" && old == "0" {
						return true
					}
					return false
				},
			},
			"ipv6_cidr_blocks": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidIPv6CIDRNetworkAddress,
				},
				ConflictsWith: []string{"source_security_group_id", "self"},
				AtLeastOneOf:  []string{"cidr_blocks", "ipv6_cidr_blocks", "prefix_list_ids", "self", "source_security_group_id"},
			},
			"prefix_list_ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
				AtLeastOneOf: []string{"cidr_blocks", "ipv6_cidr_blocks", "prefix_list_ids", "self", "source_security_group_id"},
			},
			"protocol": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				StateFunc: ProtocolStateFunc,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_rule_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"self": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ForceNew:      true,
				ConflictsWith: []string{"cidr_blocks", "ipv6_cidr_blocks", "source_security_group_id"},
				AtLeastOneOf:  []string{"cidr_blocks", "ipv6_cidr_blocks", "prefix_list_ids", "self", "source_security_group_id"},
			},
			"source_security_group_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ConflictsWith: []string{"cidr_blocks", "ipv6_cidr_blocks", "self"},
				AtLeastOneOf:  []string{"cidr_blocks", "ipv6_cidr_blocks", "prefix_list_ids", "self", "source_security_group_id"},
			},
			"to_port": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
				// Support existing configurations that have non-zero from_port and to_port defined with all protocols
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					protocol := ProtocolForValue(d.Get("protocol").(string))
					if protocol == "-1" && old == "0" {
						return true
					}
					return false
				},
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(securityGroupRuleType_Values(), false),
			},
		},
	}
}

func resourceSecurityGroupRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()
	securityGroupID := d.Get("security_group_id").(string)

	conns.GlobalMutexKV.Lock(securityGroupID)
	defer conns.GlobalMutexKV.Unlock(securityGroupID)

	sg, err := FindSecurityGroupByID(ctx, conn, securityGroupID)

	if err != nil {
		return diag.Errorf("reading Security Group (%s): %s", securityGroupID, err)
	}

	ipPermission := expandIPPermission(d, sg)
	ruleType := d.Get("type").(string)
	isVPC := aws.StringValue(sg.VpcId) != ""
	id := SecurityGroupRuleCreateID(securityGroupID, ruleType, ipPermission)

	switch ruleType {
	case securityGroupRuleTypeIngress:
		input := &ec2.AuthorizeSecurityGroupIngressInput{
			IpPermissions: []*ec2.IpPermission{ipPermission},
		}
		var output *ec2.AuthorizeSecurityGroupIngressOutput

		if isVPC {
			input.GroupId = sg.GroupId
		} else {
			input.GroupName = sg.GroupName
		}

		output, err = conn.AuthorizeSecurityGroupIngressWithContext(ctx, input)

		if err == nil {
			if len(output.SecurityGroupRules) == 1 {
				d.Set("security_group_rule_id", output.SecurityGroupRules[0].SecurityGroupRuleId)
			} else {
				d.Set("security_group_rule_id", nil)
			}
		}

	case securityGroupRuleTypeEgress:
		input := &ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       sg.GroupId,
			IpPermissions: []*ec2.IpPermission{ipPermission},
		}
		var output *ec2.AuthorizeSecurityGroupEgressOutput

		output, err = conn.AuthorizeSecurityGroupEgressWithContext(ctx, input)

		if err == nil {
			if len(output.SecurityGroupRules) == 1 {
				d.Set("security_group_rule_id", output.SecurityGroupRules[0].SecurityGroupRuleId)
			} else {
				d.Set("security_group_rule_id", nil)
			}
		}
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPermissionDuplicate) {
		return diag.Errorf(`[WARN] A duplicate Security Group rule was found on (%s). This may be
a side effect of a now-fixed Terraform issue causing two security groups with
identical attributes but different source_security_group_ids to overwrite each
other in the state. See https://github.com/hashicorp/terraform/pull/2376 for more
information and instructions for recovery. Error: %s`, securityGroupID, err)
	}

	if err != nil {
		return diag.Errorf("authorizing Security Group (%s) Rule (%s): %s", securityGroupID, id, err)
	}

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		sg, err := FindSecurityGroupByID(ctx, conn, securityGroupID)

		if err != nil {
			return nil, err
		}

		var rules []*ec2.IpPermission

		if ruleType == securityGroupRuleTypeIngress {
			rules = sg.IpPermissions
		} else {
			rules = sg.IpPermissionsEgress
		}

		rule, _ := findRuleMatch(ipPermission, rules, isVPC)

		if rule == nil {
			return nil, &resource.NotFoundError{}
		}

		return rule, nil
	})

	if err != nil {
		return diag.Errorf("waiting for Security Group (%s) Rule (%s) create: %s", securityGroupID, id, err)
	}

	d.SetId(id)

	return nil
}

func resourceSecurityGroupRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()
	securityGroupID := d.Get("security_group_id").(string)
	ruleType := d.Get("type").(string)

	sg, err := FindSecurityGroupByID(ctx, conn, securityGroupID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Group (%s) not found, removing from state", securityGroupID)
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Security Group (%s): %s", securityGroupID, err)
	}

	ipPermission := expandIPPermission(d, sg)
	isVPC := aws.StringValue(sg.VpcId) != ""

	var rules []*ec2.IpPermission

	if ruleType == securityGroupRuleTypeIngress {
		rules = sg.IpPermissions
	} else {
		rules = sg.IpPermissionsEgress
	}

	rule, description := findRuleMatch(ipPermission, rules, isVPC)

	if rule == nil {
		if !d.IsNewResource() {
			log.Printf("[WARN] Security Group (%s) Rule (%s) not found, removing from state", securityGroupID, d.Id())
			d.SetId("")
			return nil
		}

		// Shouldn't reach here as we aren't called from resourceSecurityGroupRuleCreate.
		return diag.Errorf("reading Security Group (%s) Rule (%s): %s", securityGroupID, d.Id(), &resource.NotFoundError{})
	}

	flattenIpPermission(d, ipPermission, isVPC)
	d.Set("description", description)
	d.Set("type", ruleType)

	if strings.Contains(d.Id(), securityGroupRuleIDSeparator) {
		// import so fix the id
		id := SecurityGroupRuleCreateID(securityGroupID, ruleType, ipPermission)
		d.SetId(id)
	}

	if isVPC {
		// Attempt to find the single matching AWS Security Group Rule resource ID.
		securityGroupRules, err := FindSecurityGroupRulesBySecurityGroupID(ctx, conn, securityGroupID)

		if err != nil {
			return diag.Errorf("reading Security Group (%s) Rules: %s", securityGroupID, err)
		}

		d.Set("security_group_rule_id", findSecurityGroupRuleMatch(ipPermission, securityGroupRules, ruleType))
	}

	return nil
}

func resourceSecurityGroupRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChange("description") {
		securityGroupID := d.Get("security_group_id").(string)

		conns.GlobalMutexKV.Lock(securityGroupID)
		defer conns.GlobalMutexKV.Unlock(securityGroupID)

		sg, err := FindSecurityGroupByID(ctx, conn, securityGroupID)

		if err != nil {
			return diag.Errorf("reading Security Group (%s): %s", securityGroupID, err)
		}

		ipPermission := expandIPPermission(d, sg)
		ruleType := d.Get("type").(string)
		isVPC := aws.StringValue(sg.VpcId) != ""

		switch ruleType {
		case securityGroupRuleTypeIngress:
			input := &ec2.UpdateSecurityGroupRuleDescriptionsIngressInput{
				IpPermissions: []*ec2.IpPermission{ipPermission},
			}

			if isVPC {
				input.GroupId = sg.GroupId
			} else {
				input.GroupName = sg.GroupName
			}

			_, err = conn.UpdateSecurityGroupRuleDescriptionsIngressWithContext(ctx, input)

		case securityGroupRuleTypeEgress:
			input := &ec2.UpdateSecurityGroupRuleDescriptionsEgressInput{
				GroupId:       sg.GroupId,
				IpPermissions: []*ec2.IpPermission{ipPermission},
			}

			_, err = conn.UpdateSecurityGroupRuleDescriptionsEgressWithContext(ctx, input)
		}

		if err != nil {
			return diag.Errorf("updating Security Group (%s) Rule (%s) description: %s", securityGroupID, d.Id(), err)
		}
	}

	return resourceSecurityGroupRuleRead(ctx, d, meta)
}

func resourceSecurityGroupRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()
	securityGroupID := d.Get("security_group_id").(string)

	conns.GlobalMutexKV.Lock(securityGroupID)
	defer conns.GlobalMutexKV.Unlock(securityGroupID)

	sg, err := FindSecurityGroupByID(ctx, conn, securityGroupID)

	if err != nil {
		return diag.Errorf("reading Security Group (%s): %s", securityGroupID, err)
	}

	ipPermission := expandIPPermission(d, sg)
	ruleType := d.Get("type").(string)
	isVPC := aws.StringValue(sg.VpcId) != ""

	switch ruleType {
	case securityGroupRuleTypeIngress:
		input := &ec2.RevokeSecurityGroupIngressInput{
			IpPermissions: []*ec2.IpPermission{ipPermission},
		}

		if isVPC {
			input.GroupId = sg.GroupId
		} else {
			input.GroupName = sg.GroupName
		}

		_, err = conn.RevokeSecurityGroupIngressWithContext(ctx, input)

	case securityGroupRuleTypeEgress:
		input := &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       sg.GroupId,
			IpPermissions: []*ec2.IpPermission{ipPermission},
		}

		_, err = conn.RevokeSecurityGroupEgressWithContext(ctx, input)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPermissionNotFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("revoking Security Group (%s) Rule (%s): %s", securityGroupID, d.Id(), err)
	}

	return nil
}

func resourceSecurityGroupRuleImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	invalidIDError := func(msg string) error {
		return fmt.Errorf("unexpected format for ID (%q), expected SECURITYGROUPID_TYPE_PROTOCOL_FROMPORT_TOPORT_SOURCE[_SOURCE]*: %s", d.Id(), msg)
	}

	// example: sg-09a093729ef9382a6_ingress_tcp_8000_8000_10.0.3.0/24
	// example: sg-09a093729ef9382a6_ingress_92_0_65536_10.0.3.0/24_10.0.4.0/24
	// example: sg-09a093729ef9382a6_egress_tcp_8000_8000_10.0.3.0/24
	// example: sg-09a093729ef9382a6_egress_tcp_8000_8000_pl-34800000
	// example: sg-09a093729ef9382a6_ingress_all_0_65536_sg-08123412342323
	// example: sg-09a093729ef9382a6_ingress_tcp_100_121_10.1.0.0/16_2001:db8::/48_10.2.0.0/16_2002:db8::/48
	parts := strings.Split(d.Id(), securityGroupRuleIDSeparator)

	if len(parts) < 6 {
		return nil, invalidIDError("too few parts")
	}

	securityGroupID := parts[0]
	ruleType := parts[1]
	protocol := parts[2]
	fromPort := parts[3]
	toPort := parts[4]
	sources := parts[5:]

	if !strings.HasPrefix(securityGroupID, "sg-") {
		return nil, invalidIDError("invalid Security Group ID")
	}

	if ruleType != securityGroupRuleTypeIngress && ruleType != securityGroupRuleTypeEgress {
		return nil, invalidIDError("expecting 'ingress' or 'egress'")
	}

	if _, ok := securityGroupProtocolIntegers[protocol]; !ok {
		if _, err := strconv.Atoi(protocol); err != nil {
			return nil, invalidIDError("protocol must be tcp/udp/icmp/icmpv6/all or a number")
		}
	}

	protocolName := ProtocolForValue(protocol)
	if protocolName == "icmp" || protocolName == "icmpv6" {
		if v, err := strconv.Atoi(fromPort); err != nil || v < -1 || v > 255 {
			return nil, invalidIDError("invalid icmp type")
		} else if v, err := strconv.Atoi(toPort); err != nil || v < -1 || v > 255 {
			return nil, invalidIDError("invalid icmp code")
		}
	} else {
		if p1, err := strconv.Atoi(fromPort); err != nil {
			return nil, invalidIDError("invalid from_port")
		} else if p2, err := strconv.Atoi(toPort); err != nil {
			return nil, invalidIDError("invalid to_port")
		} else if p2 < p1 {
			return nil, invalidIDError("to_port lower than from_port")
		}
	}

	for _, v := range sources {
		// will be properly validated later
		if v != "self" && !strings.Contains(v, "sg-") && !strings.Contains(v, "pl-") && !strings.Contains(v, ":") && !strings.Contains(v, ".") {
			return nil, invalidIDError("source must be cidr, ipv6cidr, prefix list, 'self', or a Security Group ID")
		}
	}

	d.Set("security_group_id", securityGroupID)
	d.Set("type", ruleType)
	d.Set("protocol", protocolName)
	if v, err := strconv.Atoi(fromPort); err == nil {
		d.Set("from_port", v)
	}
	if v, err := strconv.Atoi(toPort); err == nil {
		d.Set("to_port", v)
	}
	d.Set("self", false)

	var cidrBlocks, ipv6CIDRBlocks, prefixListIDs []string

	for _, v := range sources {
		switch {
		case v == "self":
			d.Set("self", true)
		case strings.Contains(v, "sg-"):
			d.Set("source_security_group_id", v)
		case strings.Contains(v, ":"):
			ipv6CIDRBlocks = append(ipv6CIDRBlocks, v)
		case strings.Contains(v, "pl-"):
			prefixListIDs = append(prefixListIDs, v)
		default:
			cidrBlocks = append(cidrBlocks, v)
		}
	}

	d.Set("cidr_blocks", cidrBlocks)
	d.Set("ipv6_cidr_blocks", ipv6CIDRBlocks)
	d.Set("prefix_list_ids", prefixListIDs)

	return []*schema.ResourceData{d}, nil
}

func findRuleMatch(p *ec2.IpPermission, rules []*ec2.IpPermission, isVPC bool) (*ec2.IpPermission, string) {
	var rule *ec2.IpPermission
	var description string

	for _, r := range rules {
		if p.ToPort != nil && r.ToPort != nil && aws.Int64Value(p.ToPort) != aws.Int64Value(r.ToPort) {
			continue
		}

		if p.FromPort != nil && r.FromPort != nil && aws.Int64Value(p.FromPort) != aws.Int64Value(r.FromPort) {
			continue
		}

		if p.IpProtocol != nil && r.IpProtocol != nil && aws.StringValue(p.IpProtocol) != aws.StringValue(r.IpProtocol) {
			continue
		}

		remaining := len(p.IpRanges)
		for _, v1 := range p.IpRanges {
			for _, v2 := range r.IpRanges {
				if v1.CidrIp == nil || v2.CidrIp == nil {
					continue
				}
				if aws.StringValue(v1.CidrIp) == aws.StringValue(v2.CidrIp) {
					remaining--

					if v := aws.StringValue(v2.Description); v != "" && description == "" {
						description = v
					}
				}
			}
		}

		if remaining > 0 {
			continue
		}

		remaining = len(p.Ipv6Ranges)
		for _, v1 := range p.Ipv6Ranges {
			for _, v2 := range r.Ipv6Ranges {
				if v1.CidrIpv6 == nil || v2.CidrIpv6 == nil {
					continue
				}
				if aws.StringValue(v1.CidrIpv6) == aws.StringValue(v2.CidrIpv6) {
					remaining--

					if v := aws.StringValue(v2.Description); v != "" && description == "" {
						description = v
					}
				}
			}
		}

		if remaining > 0 {
			continue
		}

		remaining = len(p.PrefixListIds)
		for _, v1 := range p.PrefixListIds {
			for _, v2 := range r.PrefixListIds {
				if v1.PrefixListId == nil || v2.PrefixListId == nil {
					continue
				}
				if aws.StringValue(v1.PrefixListId) == aws.StringValue(v2.PrefixListId) {
					remaining--

					if v := aws.StringValue(v2.Description); v != "" && description == "" {
						description = v
					}
				}
			}
		}

		if remaining > 0 {
			continue
		}

		remaining = len(p.UserIdGroupPairs)
		for _, v1 := range p.UserIdGroupPairs {
			for _, v2 := range r.UserIdGroupPairs {
				if isVPC {
					if v1.GroupId == nil || v2.GroupId == nil {
						continue
					}
					if aws.StringValue(v1.GroupId) == aws.StringValue(v2.GroupId) {
						remaining--

						if v := aws.StringValue(v2.Description); v != "" && description == "" {
							description = v
						}
					}
				} else {
					if v1.GroupName == nil || v2.GroupName == nil {
						continue
					}
					if aws.StringValue(v1.GroupName) == aws.StringValue(v2.GroupName) {
						remaining--

						if v := aws.StringValue(v2.Description); v != "" && description == "" {
							description = v
						}
					}
				}
			}
		}

		if remaining > 0 {
			description = ""

			continue
		}

		rule = r
	}

	return rule, description
}

func findSecurityGroupRuleMatch(p *ec2.IpPermission, securityGroupRules []*ec2.SecurityGroupRule, ruleType string) string {
	for _, r := range securityGroupRules {
		if ruleType == securityGroupRuleTypeIngress && aws.BoolValue(r.IsEgress) {
			continue
		}

		if p.ToPort != nil && r.ToPort != nil && aws.Int64Value(p.ToPort) != aws.Int64Value(r.ToPort) {
			continue
		}

		if p.FromPort != nil && r.FromPort != nil && aws.Int64Value(p.FromPort) != aws.Int64Value(r.FromPort) {
			continue
		}

		if p.IpProtocol != nil && r.IpProtocol != nil && aws.StringValue(p.IpProtocol) != aws.StringValue(r.IpProtocol) {
			continue
		}

		// SecurityGroupRule has only a single source or destination set.
		if r.CidrIpv4 != nil {
			if len(p.IpRanges) == 1 && aws.StringValue(p.IpRanges[0].CidrIp) == aws.StringValue(r.CidrIpv4) {
				if len(p.Ipv6Ranges) == 0 && len(p.PrefixListIds) == 0 && len(p.UserIdGroupPairs) == 0 {
					return aws.StringValue(r.SecurityGroupRuleId)
				}
			}
		} else if r.CidrIpv6 != nil {
			if len(p.Ipv6Ranges) == 1 && aws.StringValue(p.Ipv6Ranges[0].CidrIpv6) == aws.StringValue(r.CidrIpv6) {
				if len(p.IpRanges) == 0 && len(p.PrefixListIds) == 0 && len(p.UserIdGroupPairs) == 0 {
					return aws.StringValue(r.SecurityGroupRuleId)
				}
			}
		} else if r.PrefixListId != nil {
			if len(p.PrefixListIds) == 1 && aws.StringValue(p.PrefixListIds[0].PrefixListId) == aws.StringValue(r.PrefixListId) {
				if len(p.IpRanges) == 0 && len(p.Ipv6Ranges) == 0 && len(p.UserIdGroupPairs) == 0 {
					return aws.StringValue(r.SecurityGroupRuleId)
				}
			}
		} else if r.ReferencedGroupInfo != nil {
			if len(p.UserIdGroupPairs) == 1 && aws.StringValue(p.UserIdGroupPairs[0].GroupId) == aws.StringValue(r.ReferencedGroupInfo.GroupId) {
				if len(p.IpRanges) == 0 && len(p.Ipv6Ranges) == 0 && len(p.PrefixListIds) == 0 {
					return aws.StringValue(r.SecurityGroupRuleId)
				}
			}
		}
	}

	return ""
}

const securityGroupRuleIDSeparator = "_"

// byGroupPair implements sort.Interface for []*ec2.UserIDGroupPairs based on
// GroupID or GroupName field (only one should be set).
type byGroupPair []*ec2.UserIdGroupPair

func (b byGroupPair) Len() int      { return len(b) }
func (b byGroupPair) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byGroupPair) Less(i, j int) bool {
	if b[i].GroupId != nil && b[j].GroupId != nil {
		return aws.StringValue(b[i].GroupId) < aws.StringValue(b[j].GroupId)
	}
	if b[i].GroupName != nil && b[j].GroupName != nil {
		return aws.StringValue(b[i].GroupName) < aws.StringValue(b[j].GroupName)
	}

	//lintignore:R009
	panic("mismatched security group rules, may be a terraform bug")
}

func SecurityGroupRuleCreateID(securityGroupID, ruleType string, ip *ec2.IpPermission) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s-", securityGroupID))
	if aws.Int64Value(ip.FromPort) > 0 {
		buf.WriteString(fmt.Sprintf("%d-", *ip.FromPort))
	}
	if aws.Int64Value(ip.ToPort) > 0 {
		buf.WriteString(fmt.Sprintf("%d-", *ip.ToPort))
	}
	buf.WriteString(fmt.Sprintf("%s-", *ip.IpProtocol))
	buf.WriteString(fmt.Sprintf("%s-", ruleType))

	// We need to make sure to sort the strings below so that we always
	// generate the same hash code no matter what is in the set.
	if len(ip.IpRanges) > 0 {
		s := make([]string, len(ip.IpRanges))
		for i, r := range ip.IpRanges {
			s[i] = aws.StringValue(r.CidrIp)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	if len(ip.Ipv6Ranges) > 0 {
		s := make([]string, len(ip.Ipv6Ranges))
		for i, r := range ip.Ipv6Ranges {
			s[i] = aws.StringValue(r.CidrIpv6)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	if len(ip.PrefixListIds) > 0 {
		s := make([]string, len(ip.PrefixListIds))
		for i, pl := range ip.PrefixListIds {
			s[i] = aws.StringValue(pl.PrefixListId)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	if len(ip.UserIdGroupPairs) > 0 {
		sort.Sort(byGroupPair(ip.UserIdGroupPairs))
		for _, pair := range ip.UserIdGroupPairs {
			if pair.GroupId != nil {
				buf.WriteString(fmt.Sprintf("%s-", *pair.GroupId))
			} else {
				buf.WriteString("-")
			}
			if pair.GroupName != nil {
				buf.WriteString(fmt.Sprintf("%s-", *pair.GroupName))
			} else {
				buf.WriteString("-")
			}
		}
	}

	return fmt.Sprintf("sgrule-%d", create.StringHashcode(buf.String()))
}

func expandIPPermission(d *schema.ResourceData, sg *ec2.SecurityGroup) *ec2.IpPermission { // nosemgrep:ci.caps5-in-func-name
	apiObject := &ec2.IpPermission{
		IpProtocol: aws.String(ProtocolForValue(d.Get("protocol").(string))),
	}

	// InvalidParameterValue: When protocol is ALL, you cannot specify from-port.
	if v := aws.StringValue(apiObject.IpProtocol); v != "-1" {
		apiObject.FromPort = aws.Int64(int64(d.Get("from_port").(int)))
		apiObject.ToPort = aws.Int64(int64(d.Get("to_port").(int)))
	}

	if v, ok := d.GetOk("cidr_blocks"); ok && len(v.([]interface{})) > 0 {
		for _, v := range v.([]interface{}) {
			apiObject.IpRanges = append(apiObject.IpRanges, &ec2.IpRange{
				CidrIp: aws.String(v.(string)),
			})
		}
	}

	if v, ok := d.GetOk("ipv6_cidr_blocks"); ok && len(v.([]interface{})) > 0 {
		for _, v := range v.([]interface{}) {
			apiObject.Ipv6Ranges = append(apiObject.Ipv6Ranges, &ec2.Ipv6Range{
				CidrIpv6: aws.String(v.(string)),
			})
		}
	}

	if v, ok := d.GetOk("prefix_list_ids"); ok && len(v.([]interface{})) > 0 {
		for _, v := range v.([]interface{}) {
			apiObject.PrefixListIds = append(apiObject.PrefixListIds, &ec2.PrefixListId{
				PrefixListId: aws.String(v.(string)),
			})
		}
	}

	var self string
	vpc := aws.StringValue(sg.VpcId) != ""

	if _, ok := d.GetOk("self"); ok {
		if vpc {
			self = aws.StringValue(sg.GroupId)
			apiObject.UserIdGroupPairs = append(apiObject.UserIdGroupPairs, &ec2.UserIdGroupPair{
				GroupId: aws.String(self),
			})
		} else {
			self = aws.StringValue(sg.GroupName)
			apiObject.UserIdGroupPairs = append(apiObject.UserIdGroupPairs, &ec2.UserIdGroupPair{
				GroupName: aws.String(self),
			})
		}
	}

	if v, ok := d.GetOk("source_security_group_id"); ok {
		if v := v.(string); v != self {
			if vpc {
				// [OwnerID/]SecurityGroupID.
				if parts := strings.Split(v, "/"); len(parts) == 1 {
					apiObject.UserIdGroupPairs = append(apiObject.UserIdGroupPairs, &ec2.UserIdGroupPair{
						GroupId: aws.String(v),
					})
				} else {
					apiObject.UserIdGroupPairs = append(apiObject.UserIdGroupPairs, &ec2.UserIdGroupPair{
						GroupId: aws.String(parts[1]),
						UserId:  aws.String(parts[0]),
					})
				}
			} else {
				apiObject.UserIdGroupPairs = append(apiObject.UserIdGroupPairs, &ec2.UserIdGroupPair{
					GroupName: aws.String(v),
				})
			}
		}
	}

	if v, ok := d.GetOk("description"); ok {
		description := v.(string)

		for _, v := range apiObject.IpRanges {
			v.Description = aws.String(description)
		}

		for _, v := range apiObject.Ipv6Ranges {
			v.Description = aws.String(description)
		}

		for _, v := range apiObject.PrefixListIds {
			v.Description = aws.String(description)
		}

		for _, v := range apiObject.UserIdGroupPairs {
			v.Description = aws.String(description)
		}
	}

	return apiObject
}

func flattenIpPermission(d *schema.ResourceData, apiObject *ec2.IpPermission, isVPC bool) { // nosemgrep:ci.caps5-in-func-name
	if apiObject == nil {
		return
	}

	d.Set("from_port", apiObject.FromPort)
	d.Set("protocol", apiObject.IpProtocol)
	d.Set("to_port", apiObject.ToPort)

	if v := apiObject.IpRanges; len(v) > 0 {
		var ipRanges []string

		for _, v := range v {
			ipRanges = append(ipRanges, aws.StringValue(v.CidrIp))
		}

		d.Set("cidr_blocks", ipRanges)
	}

	if v := apiObject.Ipv6Ranges; len(v) > 0 {
		var ipv6Ranges []string

		for _, v := range v {
			ipv6Ranges = append(ipv6Ranges, aws.StringValue(v.CidrIpv6))
		}

		d.Set("ipv6_cidr_blocks", ipv6Ranges)
	}

	if v := apiObject.PrefixListIds; len(v) > 0 {
		var prefixListIDs []string

		for _, v := range v {
			prefixListIDs = append(prefixListIDs, aws.StringValue(v.PrefixListId))
		}

		d.Set("prefix_list_ids", prefixListIDs)
	}

	if v := apiObject.UserIdGroupPairs; len(v) > 0 {
		v := v[0]

		if isVPC {
			if old, ok := d.GetOk("source_security_group_id"); ok {
				// [OwnerID/]SecurityGroupID.
				if parts := strings.Split(old.(string), "/"); len(parts) == 1 || aws.StringValue(v.UserId) == "" {
					d.Set("source_security_group_id", v.GroupId)
				} else {
					d.Set("source_security_group_id", strings.Join([]string{aws.StringValue(v.UserId), aws.StringValue(v.GroupId)}, "/"))
				}
			}
		} else {
			d.Set("source_security_group_id", v.GroupName)
		}
	}
}
