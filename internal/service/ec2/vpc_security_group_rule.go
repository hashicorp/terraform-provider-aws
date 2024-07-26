// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_security_group_rule", name="Security Group Rule")
func resourceSecurityGroupRule() *schema.Resource {
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
		MigrateState:  securityGroupRuleMigrateState,

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
			names.AttrDescription: {
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
					protocol := protocolForValue(d.Get(names.AttrProtocol).(string))
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
			names.AttrProtocol: {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				StateFunc: protocolStateFunc,
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
					protocol := protocolForValue(d.Get(names.AttrProtocol).(string))
					if protocol == "-1" && old == "0" {
						return true
					}
					return false
				},
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[securityGroupRuleType](),
			},
		},
	}
}

func resourceSecurityGroupRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	securityGroupID := d.Get("security_group_id").(string)

	conns.GlobalMutexKV.Lock(securityGroupID)
	defer conns.GlobalMutexKV.Unlock(securityGroupID)

	sg, err := findSecurityGroupByID(ctx, conn, securityGroupID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Group (%s): %s", securityGroupID, err)
	}

	ipPermission := expandIPPermission(d, sg)
	ruleType := securityGroupRuleType(d.Get(names.AttrType).(string))
	id := securityGroupRuleCreateID(securityGroupID, string(ruleType), &ipPermission)

	switch ruleType {
	case securityGroupRuleTypeIngress:
		input := &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       sg.GroupId,
			IpPermissions: []awstypes.IpPermission{ipPermission},
		}
		var output *ec2.AuthorizeSecurityGroupIngressOutput

		output, err = conn.AuthorizeSecurityGroupIngress(ctx, input)

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
			IpPermissions: []awstypes.IpPermission{ipPermission},
		}
		var output *ec2.AuthorizeSecurityGroupEgressOutput

		output, err = conn.AuthorizeSecurityGroupEgress(ctx, input)

		if err == nil {
			if len(output.SecurityGroupRules) == 1 {
				d.Set("security_group_rule_id", output.SecurityGroupRules[0].SecurityGroupRuleId)
			} else {
				d.Set("security_group_rule_id", nil)
			}
		}
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPermissionDuplicate) {
		return sdkdiag.AppendErrorf(diags, `[WARN] A duplicate Security Group rule was found on (%s). This may be
a side effect of a now-fixed Terraform issue causing two security groups with
identical attributes but different source_security_group_ids to overwrite each
other in the state. See https://github.com/hashicorp/terraform/pull/2376 for more
information and instructions for recovery. Error: %s`, securityGroupID, err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "authorizing Security Group (%s) Rule (%s): %s", securityGroupID, id, err)
	}

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		sg, err := findSecurityGroupByID(ctx, conn, securityGroupID)

		if err != nil {
			return nil, err
		}

		var rules []awstypes.IpPermission

		if ruleType == securityGroupRuleTypeIngress {
			rules = sg.IpPermissions
		} else {
			rules = sg.IpPermissionsEgress
		}

		rule, _ := findRuleMatch(ipPermission, rules)

		if rule == nil {
			return nil, &retry.NotFoundError{}
		}

		return rule, nil
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Group (%s) Rule (%s) create: %s", securityGroupID, id, err)
	}

	d.SetId(id)

	return diags
}

func resourceSecurityGroupRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	securityGroupID := d.Get("security_group_id").(string)
	ruleType := securityGroupRuleType(d.Get(names.AttrType).(string))

	sg, err := findSecurityGroupByID(ctx, conn, securityGroupID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Group (%s) not found, removing from state", securityGroupID)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Group (%s): %s", securityGroupID, err)
	}

	ipPermission := expandIPPermission(d, sg)

	var rules []awstypes.IpPermission

	if ruleType == securityGroupRuleTypeIngress {
		rules = sg.IpPermissions
	} else {
		rules = sg.IpPermissionsEgress
	}

	rule, description := findRuleMatch(ipPermission, rules)

	if rule == nil {
		if !d.IsNewResource() {
			log.Printf("[WARN] Security Group (%s) Rule (%s) not found, removing from state", securityGroupID, d.Id())
			d.SetId("")
			return diags
		}

		// Shouldn't reach here as we aren't called from resourceSecurityGroupRuleCreate.
		return sdkdiag.AppendErrorf(diags, "reading Security Group (%s) Rule (%s): %s", securityGroupID, d.Id(), &retry.NotFoundError{})
	}

	flattenIpPermission(d, &ipPermission)

	if description != nil { // nosemgrep: ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set(names.AttrDescription, description)
	}
	d.Set(names.AttrType, ruleType)

	if strings.Contains(d.Id(), securityGroupRuleIDSeparator) {
		// import so fix the id
		id := securityGroupRuleCreateID(securityGroupID, string(ruleType), &ipPermission)
		d.SetId(id)
	}

	// Attempt to find the single matching AWS Security Group Rule resource ID.
	securityGroupRules, err := findSecurityGroupRulesBySecurityGroupID(ctx, conn, securityGroupID)

	// Ignore UnsupportedOperation errors for AWS China and GovCloud (US).
	if tfawserr.ErrCodeEquals(err, errCodeUnsupportedOperation) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Group (%s) Rules: %s", securityGroupID, err)
	}

	d.Set("security_group_rule_id", findSecurityGroupRuleMatch(ipPermission, securityGroupRules, ruleType))

	return diags
}

func resourceSecurityGroupRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChange(names.AttrDescription) {
		securityGroupID := d.Get("security_group_id").(string)

		conns.GlobalMutexKV.Lock(securityGroupID)
		defer conns.GlobalMutexKV.Unlock(securityGroupID)

		sg, err := findSecurityGroupByID(ctx, conn, securityGroupID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Security Group (%s): %s", securityGroupID, err)
		}

		ipPermission := expandIPPermission(d, sg)
		ruleType := securityGroupRuleType(d.Get(names.AttrType).(string))

		switch ruleType {
		case securityGroupRuleTypeIngress:
			input := &ec2.UpdateSecurityGroupRuleDescriptionsIngressInput{
				GroupId:       sg.GroupId,
				IpPermissions: []awstypes.IpPermission{ipPermission},
			}

			_, err = conn.UpdateSecurityGroupRuleDescriptionsIngress(ctx, input)

		case securityGroupRuleTypeEgress:
			input := &ec2.UpdateSecurityGroupRuleDescriptionsEgressInput{
				GroupId:       sg.GroupId,
				IpPermissions: []awstypes.IpPermission{ipPermission},
			}

			_, err = conn.UpdateSecurityGroupRuleDescriptionsEgress(ctx, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Security Group (%s) Rule (%s) description: %s", securityGroupID, d.Id(), err)
		}
	}

	return append(diags, resourceSecurityGroupRuleRead(ctx, d, meta)...)
}

func resourceSecurityGroupRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	securityGroupID := d.Get("security_group_id").(string)

	conns.GlobalMutexKV.Lock(securityGroupID)
	defer conns.GlobalMutexKV.Unlock(securityGroupID)

	sg, err := findSecurityGroupByID(ctx, conn, securityGroupID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Group (%s): %s", securityGroupID, err)
	}

	ipPermission := expandIPPermission(d, sg)
	ruleType := securityGroupRuleType(d.Get(names.AttrType).(string))

	switch ruleType {
	case securityGroupRuleTypeIngress:
		input := &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       sg.GroupId,
			IpPermissions: []awstypes.IpPermission{ipPermission},
		}

		_, err = conn.RevokeSecurityGroupIngress(ctx, input)

	case securityGroupRuleTypeEgress:
		input := &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       sg.GroupId,
			IpPermissions: []awstypes.IpPermission{ipPermission},
		}

		_, err = conn.RevokeSecurityGroupEgress(ctx, input)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPermissionNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "revoking Security Group (%s) Rule (%s): %s", securityGroupID, d.Id(), err)
	}

	return diags
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
	ruleType := securityGroupRuleType(parts[1])
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

	protocolName := protocolForValue(protocol)
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
	d.Set(names.AttrType, ruleType)
	d.Set(names.AttrProtocol, protocolName)
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

func findRuleMatch(p awstypes.IpPermission, rules []awstypes.IpPermission) (*awstypes.IpPermission, *string) {
	var rule awstypes.IpPermission
	var description *string

	for _, r := range rules {
		if p.ToPort != nil && r.ToPort != nil && aws.ToInt32(p.ToPort) != aws.ToInt32(r.ToPort) {
			continue
		}

		if p.FromPort != nil && r.FromPort != nil && aws.ToInt32(p.FromPort) != aws.ToInt32(r.FromPort) {
			continue
		}

		if p.IpProtocol != nil && r.IpProtocol != nil && aws.ToString(p.IpProtocol) != aws.ToString(r.IpProtocol) {
			continue
		}

		remaining := len(p.IpRanges)
		for _, v1 := range p.IpRanges {
			for _, v2 := range r.IpRanges {
				if v1.CidrIp == nil || v2.CidrIp == nil {
					continue
				}
				if aws.ToString(v1.CidrIp) == aws.ToString(v2.CidrIp) {
					remaining--

					if v := v2.Description; v != nil && description == nil {
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
				if aws.ToString(v1.CidrIpv6) == aws.ToString(v2.CidrIpv6) {
					remaining--

					if v := v2.Description; v != nil && description == nil {
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
				if aws.ToString(v1.PrefixListId) == aws.ToString(v2.PrefixListId) {
					remaining--

					if v := v2.Description; v != nil && description == nil {
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
				if v1.GroupId == nil || v2.GroupId == nil {
					continue
				}
				if aws.ToString(v1.GroupId) == aws.ToString(v2.GroupId) {
					remaining--

					if v := v2.Description; v != nil && description == nil {
						description = v
					}
				}
			}
		}

		if remaining > 0 {
			description = nil

			continue
		}

		rule = r
	}

	return &rule, description
}

func findSecurityGroupRuleMatch(p awstypes.IpPermission, securityGroupRules []awstypes.SecurityGroupRule, ruleType securityGroupRuleType) string {
	for _, r := range securityGroupRules {
		if ruleType == securityGroupRuleTypeIngress && aws.ToBool(r.IsEgress) {
			continue
		}

		if p.ToPort != nil && r.ToPort != nil && aws.ToInt32(p.ToPort) != aws.ToInt32(r.ToPort) {
			continue
		}

		if p.FromPort != nil && r.FromPort != nil && aws.ToInt32(p.FromPort) != aws.ToInt32(r.FromPort) {
			continue
		}

		if p.IpProtocol != nil && r.IpProtocol != nil && aws.ToString(p.IpProtocol) != aws.ToString(r.IpProtocol) {
			continue
		}

		// SecurityGroupRule has only a single source or destination set.
		if r.CidrIpv4 != nil {
			if len(p.IpRanges) == 1 && aws.ToString(p.IpRanges[0].CidrIp) == aws.ToString(r.CidrIpv4) {
				if len(p.Ipv6Ranges) == 0 && len(p.PrefixListIds) == 0 && len(p.UserIdGroupPairs) == 0 {
					return aws.ToString(r.SecurityGroupRuleId)
				}
			}
		} else if r.CidrIpv6 != nil {
			if len(p.Ipv6Ranges) == 1 && aws.ToString(p.Ipv6Ranges[0].CidrIpv6) == aws.ToString(r.CidrIpv6) {
				if len(p.IpRanges) == 0 && len(p.PrefixListIds) == 0 && len(p.UserIdGroupPairs) == 0 {
					return aws.ToString(r.SecurityGroupRuleId)
				}
			}
		} else if r.PrefixListId != nil {
			if len(p.PrefixListIds) == 1 && aws.ToString(p.PrefixListIds[0].PrefixListId) == aws.ToString(r.PrefixListId) {
				if len(p.IpRanges) == 0 && len(p.Ipv6Ranges) == 0 && len(p.UserIdGroupPairs) == 0 {
					return aws.ToString(r.SecurityGroupRuleId)
				}
			}
		} else if r.ReferencedGroupInfo != nil {
			if len(p.UserIdGroupPairs) == 1 && aws.ToString(p.UserIdGroupPairs[0].GroupId) == aws.ToString(r.ReferencedGroupInfo.GroupId) {
				if len(p.IpRanges) == 0 && len(p.Ipv6Ranges) == 0 && len(p.PrefixListIds) == 0 {
					return aws.ToString(r.SecurityGroupRuleId)
				}
			}
		}
	}

	return ""
}

const securityGroupRuleIDSeparator = "_"

// byGroupPair implements sort.Interface for []*ec2.UserIDGroupPairs based on
// GroupID or GroupName field (only one should be set).
type byGroupPair []awstypes.UserIdGroupPair

func (b byGroupPair) Len() int      { return len(b) }
func (b byGroupPair) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byGroupPair) Less(i, j int) bool {
	if b[i].GroupId != nil && b[j].GroupId != nil {
		return aws.ToString(b[i].GroupId) < aws.ToString(b[j].GroupId)
	}
	if b[i].GroupName != nil && b[j].GroupName != nil {
		return aws.ToString(b[i].GroupName) < aws.ToString(b[j].GroupName)
	}

	//lintignore:R009
	panic("mismatched security group rules, may be a terraform bug")
}

func securityGroupRuleCreateID(securityGroupID, ruleType string, ip *awstypes.IpPermission) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s-", securityGroupID))
	if aws.ToInt32(ip.FromPort) > 0 {
		buf.WriteString(fmt.Sprintf("%d-", *ip.FromPort))
	}
	if aws.ToInt32(ip.ToPort) > 0 {
		buf.WriteString(fmt.Sprintf("%d-", *ip.ToPort))
	}
	buf.WriteString(fmt.Sprintf("%s-", *ip.IpProtocol))
	buf.WriteString(fmt.Sprintf("%s-", ruleType))

	// We need to make sure to sort the strings below so that we always
	// generate the same hash code no matter what is in the set.
	if len(ip.IpRanges) > 0 {
		s := make([]string, len(ip.IpRanges))
		for i, r := range ip.IpRanges {
			s[i] = aws.ToString(r.CidrIp)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	if len(ip.Ipv6Ranges) > 0 {
		s := make([]string, len(ip.Ipv6Ranges))
		for i, r := range ip.Ipv6Ranges {
			s[i] = aws.ToString(r.CidrIpv6)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	if len(ip.PrefixListIds) > 0 {
		s := make([]string, len(ip.PrefixListIds))
		for i, pl := range ip.PrefixListIds {
			s[i] = aws.ToString(pl.PrefixListId)
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

func expandIPPermission(d *schema.ResourceData, sg *awstypes.SecurityGroup) awstypes.IpPermission { // nosemgrep:ci.caps5-in-func-name
	apiObject := awstypes.IpPermission{
		IpProtocol: aws.String(protocolForValue(d.Get(names.AttrProtocol).(string))),
	}

	// InvalidParameterValue: When protocol is ALL, you cannot specify from-port.
	if v := aws.ToString(apiObject.IpProtocol); v != "-1" {
		apiObject.FromPort = aws.Int32(int32(d.Get("from_port").(int)))
		apiObject.ToPort = aws.Int32(int32(d.Get("to_port").(int)))
	}

	if v, ok := d.GetOk("cidr_blocks"); ok && len(v.([]interface{})) > 0 {
		for _, v := range v.([]interface{}) {
			apiObject.IpRanges = append(apiObject.IpRanges, awstypes.IpRange{
				CidrIp: aws.String(v.(string)),
			})
		}
	}

	if v, ok := d.GetOk("ipv6_cidr_blocks"); ok && len(v.([]interface{})) > 0 {
		for _, v := range v.([]interface{}) {
			apiObject.Ipv6Ranges = append(apiObject.Ipv6Ranges, awstypes.Ipv6Range{
				CidrIpv6: aws.String(v.(string)),
			})
		}
	}

	if v, ok := d.GetOk("prefix_list_ids"); ok && len(v.([]interface{})) > 0 {
		for _, v := range v.([]interface{}) {
			apiObject.PrefixListIds = append(apiObject.PrefixListIds, awstypes.PrefixListId{
				PrefixListId: aws.String(v.(string)),
			})
		}
	}

	var self string

	if _, ok := d.GetOk("self"); ok {
		self = aws.ToString(sg.GroupId)
		apiObject.UserIdGroupPairs = append(apiObject.UserIdGroupPairs, awstypes.UserIdGroupPair{
			GroupId: aws.String(self),
		})
	}

	if v, ok := d.GetOk("source_security_group_id"); ok {
		if v := v.(string); v != self {
			// [OwnerID/]SecurityGroupID.
			if parts := strings.Split(v, "/"); len(parts) == 1 {
				apiObject.UserIdGroupPairs = append(apiObject.UserIdGroupPairs, awstypes.UserIdGroupPair{
					GroupId: aws.String(v),
				})
			} else {
				apiObject.UserIdGroupPairs = append(apiObject.UserIdGroupPairs, awstypes.UserIdGroupPair{
					GroupId: aws.String(parts[1]),
					UserId:  aws.String(parts[0]),
				})
			}
		}
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		description := v.(string)
		for i := range apiObject.IpRanges {
			apiObject.IpRanges[i].Description = aws.String(description)
		}

		for i := range apiObject.Ipv6Ranges {
			apiObject.Ipv6Ranges[i].Description = aws.String(description)
		}

		for i := range apiObject.PrefixListIds {
			apiObject.PrefixListIds[i].Description = aws.String(description)
		}

		for i := range apiObject.UserIdGroupPairs {
			apiObject.UserIdGroupPairs[i].Description = aws.String(description)
		}
	}

	return apiObject
}

func flattenIpPermission(d *schema.ResourceData, apiObject *awstypes.IpPermission) { // nosemgrep:ci.caps5-in-func-name
	if apiObject == nil {
		return
	}

	d.Set("from_port", apiObject.FromPort)
	d.Set(names.AttrProtocol, apiObject.IpProtocol)
	d.Set("to_port", apiObject.ToPort)

	if v := apiObject.IpRanges; len(v) > 0 {
		var ipRanges []string

		for _, v := range v {
			ipRanges = append(ipRanges, aws.ToString(v.CidrIp))
		}

		d.Set("cidr_blocks", ipRanges)
	}

	if v := apiObject.Ipv6Ranges; len(v) > 0 {
		var ipv6Ranges []string

		for _, v := range v {
			ipv6Ranges = append(ipv6Ranges, aws.ToString(v.CidrIpv6))
		}

		d.Set("ipv6_cidr_blocks", ipv6Ranges)
	}

	if v := apiObject.PrefixListIds; len(v) > 0 {
		var prefixListIDs []string

		for _, v := range v {
			prefixListIDs = append(prefixListIDs, aws.ToString(v.PrefixListId))
		}

		d.Set("prefix_list_ids", prefixListIDs)
	}

	if v := apiObject.UserIdGroupPairs; len(v) > 0 {
		v := v[0]

		if old, ok := d.GetOk("source_security_group_id"); ok {
			// [OwnerID/]SecurityGroupID.
			if parts := strings.Split(old.(string), "/"); len(parts) == 1 || aws.ToString(v.UserId) == "" {
				d.Set("source_security_group_id", v.GroupId)
			} else {
				d.Set("source_security_group_id", strings.Join([]string{aws.ToString(v.UserId), aws.ToString(v.GroupId)}, "/"))
			}
		}
	}
}
