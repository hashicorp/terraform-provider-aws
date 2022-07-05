package ec2

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
		Create: resourceSecurityGroupRuleCreate,
		Read:   resourceSecurityGroupRuleRead,
		Update: resourceSecurityGroupRuleUpdate,
		Delete: resourceSecurityGroupRuleDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				importParts, err := validateSecurityGroupRuleImportString(d.Id())
				if err != nil {
					return nil, err
				}
				if err := populateSecurityGroupRuleFromImport(d, importParts); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
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
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
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

func resourceSecurityGroupRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	securityGroupID := d.Get("security_group_id").(string)

	conns.GlobalMutexKV.Lock(securityGroupID)
	defer conns.GlobalMutexKV.Unlock(securityGroupID)

	sg, err := FindSecurityGroupByID(conn, securityGroupID)

	if err != nil {
		return fmt.Errorf("reading Security Group (%s): %w", securityGroupID, err)
	}

	ipPermission := expandIpPermission(d, sg)
	ruleType := d.Get("type").(string)
	isVPC := aws.StringValue(sg.VpcId) != ""

	switch ruleType {
	case securityGroupRuleTypeIngress:
		input := &ec2.AuthorizeSecurityGroupIngressInput{
			IpPermissions: []*ec2.IpPermission{ipPermission},
		}

		if isVPC {
			input.GroupId = sg.GroupId
		} else {
			input.GroupName = sg.GroupName
		}

		_, err = conn.AuthorizeSecurityGroupIngress(input)

	case securityGroupRuleTypeEgress:
		input := &ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       sg.GroupId,
			IpPermissions: []*ec2.IpPermission{ipPermission},
		}

		_, err = conn.AuthorizeSecurityGroupEgress(input)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPermissionDuplicate) {
		return fmt.Errorf(`[WARN] A duplicate Security Group rule was found on (%s). This may be
a side effect of a now-fixed Terraform issue causing two security groups with
identical attributes but different source_security_group_ids to overwrite each
other in the state. See https://github.com/hashicorp/terraform/pull/2376 for more
information and instructions for recovery. Error: %w`, securityGroupID, err)
	}

	if err != nil {
		return fmt.Errorf("authorizing Security Group (%s) Rule (%s): %w", securityGroupID, ruleType, err)
	}

	id := IPPermissionIDHash(securityGroupID, ruleType, ipPermission)

	_, err = tfresource.RetryWhenNotFound(5*time.Minute, func() (interface{}, error) {
		sg, err := FindSecurityGroupByID(conn, securityGroupID)

		if err != nil {
			return nil, err
		}

		var rules []*ec2.IpPermission

		if ruleType == securityGroupRuleTypeIngress {
			rules = sg.IpPermissions
		} else {
			rules = sg.IpPermissionsEgress
		}

		rule := findRuleMatch(ipPermission, rules, isVPC)

		if rule == nil {
			return nil, &resource.NotFoundError{}
		}

		return rule, nil
	})

	if err != nil {
		return fmt.Errorf("waiting for Security Group (%s) Rule (%s): %w", securityGroupID, id, err)
	}

	d.SetId(id)

	return nil
}

func resourceSecurityGroupRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	sg_id := d.Get("security_group_id").(string)
	sg, err := FindSecurityGroupByID(conn, sg_id)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Group (%s) not found, removing Rule (%s) from state", sg_id, d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error finding Security Group (%s) for Rule (%s): %w", sg_id, d.Id(), err)
	}

	isVPC := aws.StringValue(sg.VpcId) != ""

	var rule *ec2.IpPermission
	var rules []*ec2.IpPermission
	ruleType := d.Get("type").(string)
	switch ruleType {
	case securityGroupRuleTypeIngress:
		rules = sg.IpPermissions
	default:
		rules = sg.IpPermissionsEgress
	}
	log.Printf("[DEBUG] Rules %v", rules)

	p, err := expandIPPerm(d, sg)
	if err != nil {
		return err
	}

	if !d.IsNewResource() && len(rules) == 0 {
		log.Printf("[WARN] No %s rules were found for Security Group (%s) looking for Security Group Rule (%s)", ruleType, aws.StringValue(sg.GroupName), d.Id())
		d.SetId("")
		return nil
	}

	rule = findRuleMatch(p, rules, isVPC)

	if !d.IsNewResource() && rule == nil {
		log.Printf("[DEBUG] Unable to find matching %s Security Group Rule (%s) for Group %s", ruleType, d.Id(), sg_id)
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] Found rule for Security Group Rule (%s): %s", d.Id(), rule)

	d.Set("type", ruleType)

	setFromIPPerm(d, sg, p)

	d.Set("description", descriptionFromIPPerm(d, rule))

	if strings.Contains(d.Id(), "_") {
		// import so fix the id
		id := IPPermissionIDHash(sg_id, ruleType, p)
		d.SetId(id)
	}

	return nil
}

func resourceSecurityGroupRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("description") {
		if err := resourceSecurityGroupRuleDescriptionUpdate(conn, d); err != nil {
			return err
		}
	}

	return resourceSecurityGroupRuleRead(d, meta)
}

func resourceSecurityGroupRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	securityGroupID := d.Get("security_group_id").(string)

	conns.GlobalMutexKV.Lock(securityGroupID)
	defer conns.GlobalMutexKV.Unlock(securityGroupID)

	sg, err := FindSecurityGroupByID(conn, securityGroupID)

	if err != nil {
		return fmt.Errorf("reading Security Group (%s): %w", securityGroupID, err)
	}

	ipPermission := expandIpPermission(d, sg)
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

		_, err = conn.RevokeSecurityGroupIngress(input)

	case securityGroupRuleTypeEgress:
		input := &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       sg.GroupId,
			IpPermissions: []*ec2.IpPermission{ipPermission},
		}

		_, err = conn.RevokeSecurityGroupEgress(input)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPermissionNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("revoking Security Group (%s) Rule (%s): %w", securityGroupID, ruleType, err)
	}

	return nil
}

// ByGroupPair implements sort.Interface for []*ec2.UserIDGroupPairs based on
// GroupID or GroupName field (only one should be set).
type ByGroupPair []*ec2.UserIdGroupPair

func (b ByGroupPair) Len() int      { return len(b) }
func (b ByGroupPair) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByGroupPair) Less(i, j int) bool {
	if b[i].GroupId != nil && b[j].GroupId != nil {
		return aws.StringValue(b[i].GroupId) < aws.StringValue(b[j].GroupId)
	}
	if b[i].GroupName != nil && b[j].GroupName != nil {
		return aws.StringValue(b[i].GroupName) < aws.StringValue(b[j].GroupName)
	}

	//lintignore:R009
	panic("mismatched security group rules, may be a terraform bug")
}

func findRuleMatch(p *ec2.IpPermission, rules []*ec2.IpPermission, isVPC bool) *ec2.IpPermission {
	var rule *ec2.IpPermission

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
					}
				} else {
					if v1.GroupName == nil || v2.GroupName == nil {
						continue
					}
					if aws.StringValue(v1.GroupName) == aws.StringValue(v2.GroupName) {
						remaining--
					}
				}
			}
		}

		if remaining > 0 {
			continue
		}

		rule = r
	}

	return rule
}

func IPPermissionIDHash(sg_id, ruleType string, ip *ec2.IpPermission) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", sg_id))
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
		sort.Sort(ByGroupPair(ip.UserIdGroupPairs))
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

func expandIPPerm(d *schema.ResourceData, sg *ec2.SecurityGroup) (*ec2.IpPermission, error) {
	var perm ec2.IpPermission

	protocol := ProtocolForValue(d.Get("protocol").(string))
	perm.IpProtocol = aws.String(protocol)

	// InvalidParameterValue: When protocol is ALL, you cannot specify from-port.
	if protocol != "-1" {
		perm.FromPort = aws.Int64(int64(d.Get("from_port").(int)))
		perm.ToPort = aws.Int64(int64(d.Get("to_port").(int)))
	}

	// build a group map that behaves like a set
	groups := make(map[string]bool)
	if raw, ok := d.GetOk("source_security_group_id"); ok {
		groups[raw.(string)] = true
	}

	if _, ok := d.GetOk("self"); ok {
		if aws.StringValue(sg.VpcId) != "" {
			groups[*sg.GroupId] = true
		} else {
			groups[*sg.GroupName] = true
		}
	}

	description := d.Get("description").(string)

	if len(groups) > 0 {
		perm.UserIdGroupPairs = make([]*ec2.UserIdGroupPair, len(groups))
		// build string list of group name/ids
		var gl []string
		for k := range groups {
			gl = append(gl, k)
		}

		for i, name := range gl {
			ownerId, id := "", name
			if items := strings.Split(id, "/"); len(items) > 1 {
				ownerId, id = items[0], items[1]
			}

			perm.UserIdGroupPairs[i] = &ec2.UserIdGroupPair{
				GroupId: aws.String(id),
				UserId:  aws.String(ownerId),
			}

			if aws.StringValue(sg.VpcId) == "" {
				perm.UserIdGroupPairs[i].GroupId = nil
				perm.UserIdGroupPairs[i].GroupName = aws.String(id)
				perm.UserIdGroupPairs[i].UserId = nil
			}

			if description != "" {
				perm.UserIdGroupPairs[i].Description = aws.String(description)
			}
		}
	}

	if raw, ok := d.GetOk("cidr_blocks"); ok {
		list := raw.([]interface{})
		perm.IpRanges = make([]*ec2.IpRange, len(list))
		for i, v := range list {
			cidrIP, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("empty element found in cidr_blocks - consider using the compact function")
			}
			perm.IpRanges[i] = &ec2.IpRange{CidrIp: aws.String(cidrIP)}

			if description != "" {
				perm.IpRanges[i].Description = aws.String(description)
			}
		}
	}

	if raw, ok := d.GetOk("ipv6_cidr_blocks"); ok {
		list := raw.([]interface{})
		perm.Ipv6Ranges = make([]*ec2.Ipv6Range, len(list))
		for i, v := range list {
			cidrIP, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("empty element found in ipv6_cidr_blocks - consider using the compact function")
			}
			perm.Ipv6Ranges[i] = &ec2.Ipv6Range{CidrIpv6: aws.String(cidrIP)}

			if description != "" {
				perm.Ipv6Ranges[i].Description = aws.String(description)
			}
		}
	}

	if raw, ok := d.GetOk("prefix_list_ids"); ok {
		list := raw.([]interface{})
		perm.PrefixListIds = make([]*ec2.PrefixListId, len(list))
		for i, v := range list {
			prefixListID, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("empty element found in prefix_list_ids - consider using the compact function")
			}
			perm.PrefixListIds[i] = &ec2.PrefixListId{PrefixListId: aws.String(prefixListID)}

			if description != "" {
				perm.PrefixListIds[i].Description = aws.String(description)
			}
		}
	}

	return &perm, nil
}

func setFromIPPerm(d *schema.ResourceData, sg *ec2.SecurityGroup, rule *ec2.IpPermission) {
	isVPC := aws.StringValue(sg.VpcId) != ""

	d.Set("from_port", rule.FromPort)
	d.Set("to_port", rule.ToPort)
	d.Set("protocol", rule.IpProtocol)

	var cb []string
	for _, c := range rule.IpRanges {
		cb = append(cb, *c.CidrIp)
	}
	d.Set("cidr_blocks", cb)

	var ipv6 []string
	for _, ip := range rule.Ipv6Ranges {
		ipv6 = append(ipv6, *ip.CidrIpv6)
	}
	d.Set("ipv6_cidr_blocks", ipv6)

	var pl []string
	for _, p := range rule.PrefixListIds {
		pl = append(pl, *p.PrefixListId)
	}
	d.Set("prefix_list_ids", pl)

	if len(rule.UserIdGroupPairs) > 0 {
		s := rule.UserIdGroupPairs[0]

		if isVPC {
			if existingSourceSgId, ok := d.GetOk("source_security_group_id"); ok {
				sgIdComponents := strings.Split(existingSourceSgId.(string), "/")
				hasAccountIdPrefix := len(sgIdComponents) == 2

				if hasAccountIdPrefix && s.UserId != nil {
					// then ensure on refresh that we preserve the account id prefix
					d.Set("source_security_group_id", fmt.Sprintf("%s/%s", aws.StringValue(s.UserId), aws.StringValue(s.GroupId)))
				} else {
					d.Set("source_security_group_id", s.GroupId)
				}
			} else {
				d.Set("source_security_group_id", s.GroupId)
			}
		} else {
			d.Set("source_security_group_id", s.GroupName)
		}
	}
}

func descriptionFromIPPerm(d *schema.ResourceData, rule *ec2.IpPermission) string {
	// probe IpRanges
	cidrIps := make(map[string]bool)
	if raw, ok := d.GetOk("cidr_blocks"); ok {
		for _, v := range raw.([]interface{}) {
			cidrIps[v.(string)] = true
		}
	}

	if len(cidrIps) > 0 {
		for _, c := range rule.IpRanges {
			if _, ok := cidrIps[*c.CidrIp]; !ok {
				continue
			}

			if desc := aws.StringValue(c.Description); desc != "" {
				return desc
			}
		}
	}

	// probe Ipv6Ranges
	cidrIpv6s := make(map[string]bool)
	if raw, ok := d.GetOk("ipv6_cidr_blocks"); ok {
		for _, v := range raw.([]interface{}) {
			cidrIpv6s[v.(string)] = true
		}
	}

	if len(cidrIpv6s) > 0 {
		for _, ip := range rule.Ipv6Ranges {
			if _, ok := cidrIpv6s[*ip.CidrIpv6]; !ok {
				continue
			}

			if desc := aws.StringValue(ip.Description); desc != "" {
				return desc
			}
		}
	}

	// probe PrefixListIds
	listIds := make(map[string]bool)
	if raw, ok := d.GetOk("prefix_list_ids"); ok {
		for _, v := range raw.([]interface{}) {
			listIds[v.(string)] = true
		}
	}

	if len(listIds) > 0 {
		for _, p := range rule.PrefixListIds {
			if _, ok := listIds[*p.PrefixListId]; !ok {
				continue
			}

			if desc := aws.StringValue(p.Description); desc != "" {
				return desc
			}
		}
	}

	// probe UserIdGroupPairs
	if raw, ok := d.GetOk("source_security_group_id"); ok {
		components := strings.Split(raw.(string), "/")

		switch len(components) {
		case 2:
			userId := components[0]
			groupId := components[1]

			for _, gp := range rule.UserIdGroupPairs {
				if aws.StringValue(gp.GroupId) != groupId || aws.StringValue(gp.UserId) != userId {
					continue
				}

				if desc := aws.StringValue(gp.Description); desc != "" {
					return desc
				}
			}
		case 1:
			groupId := components[0]
			for _, gp := range rule.UserIdGroupPairs {
				if aws.StringValue(gp.GroupId) != groupId {
					continue
				}

				if desc := aws.StringValue(gp.Description); desc != "" {
					return desc
				}
			}
		}
	}

	return ""
}

func resourceSecurityGroupRuleDescriptionUpdate(conn *ec2.EC2, d *schema.ResourceData) error {
	sg_id := d.Get("security_group_id").(string)

	conns.GlobalMutexKV.Lock(sg_id)
	defer conns.GlobalMutexKV.Unlock(sg_id)

	sg, err := FindSecurityGroupByID(conn, sg_id)
	if err != nil {
		return err
	}

	perm, err := expandIPPerm(d, sg)
	if err != nil {
		return err
	}
	ruleType := d.Get("type").(string)
	switch ruleType {
	case securityGroupRuleTypeIngress:
		req := &ec2.UpdateSecurityGroupRuleDescriptionsIngressInput{
			GroupId:       sg.GroupId,
			IpPermissions: []*ec2.IpPermission{perm},
		}

		_, err = conn.UpdateSecurityGroupRuleDescriptionsIngress(req)

		if err != nil {
			return fmt.Errorf("Error updating security group %s rule description: %w", sg_id, err)
		}
	case securityGroupRuleTypeEgress:
		req := &ec2.UpdateSecurityGroupRuleDescriptionsEgressInput{
			GroupId:       sg.GroupId,
			IpPermissions: []*ec2.IpPermission{perm},
		}

		_, err = conn.UpdateSecurityGroupRuleDescriptionsEgress(req)

		if err != nil {
			return fmt.Errorf("Error updating security group %s rule description: %w", sg_id, err)
		}
	}

	return nil
}

// validateSecurityGroupRuleImportString does minimal validation of import string without going to AWS
func validateSecurityGroupRuleImportString(importStr string) ([]string, error) {
	// example: sg-09a093729ef9382a6_ingress_tcp_8000_8000_10.0.3.0/24
	// example: sg-09a093729ef9382a6_ingress_92_0_65536_10.0.3.0/24_10.0.4.0/24
	// example: sg-09a093729ef9382a6_egress_tcp_8000_8000_10.0.3.0/24
	// example: sg-09a093729ef9382a6_egress_tcp_8000_8000_pl-34800000
	// example: sg-09a093729ef9382a6_ingress_all_0_65536_sg-08123412342323
	// example: sg-09a093729ef9382a6_ingress_tcp_100_121_10.1.0.0/16_2001:db8::/48_10.2.0.0/16_2002:db8::/48

	log.Printf("[DEBUG] Validating import string %s", importStr)

	importParts := strings.Split(strings.ToLower(importStr), "_")
	errStr := "unexpected format of import string (%q), expected SECURITYGROUPID_TYPE_PROTOCOL_FROMPORT_TOPORT_SOURCE[_SOURCE]*: %s"
	if len(importParts) < 6 {
		return nil, fmt.Errorf(errStr, importStr, "too few parts")
	}

	sgID := importParts[0]
	ruleType := importParts[1]
	protocol := importParts[2]
	fromPort := importParts[3]
	toPort := importParts[4]
	sources := importParts[5:]

	if !strings.HasPrefix(sgID, "sg-") {
		return nil, fmt.Errorf(errStr, importStr, "invalid security group ID")
	}

	if ruleType != securityGroupRuleTypeIngress && ruleType != securityGroupRuleTypeEgress {
		return nil, fmt.Errorf(errStr, importStr, "expecting 'ingress' or 'egress'")
	}

	if _, ok := securityGroupProtocolIntegers[protocol]; !ok {
		if _, err := strconv.Atoi(protocol); err != nil {
			return nil, fmt.Errorf(errStr, importStr, "protocol must be tcp/udp/icmp/icmpv6/all or a number")
		}
	}

	protocolName := ProtocolForValue(protocol)
	if protocolName == "icmp" || protocolName == "icmpv6" {
		if itype, err := strconv.Atoi(fromPort); err != nil || itype < -1 || itype > 255 {
			return nil, fmt.Errorf(errStr, importStr, "invalid icmp type")
		} else if icode, err := strconv.Atoi(toPort); err != nil || icode < -1 || icode > 255 {
			return nil, fmt.Errorf(errStr, importStr, "invalid icmp code")
		}
	} else {
		if p1, err := strconv.Atoi(fromPort); err != nil {
			return nil, fmt.Errorf(errStr, importStr, "invalid from_port")
		} else if p2, err := strconv.Atoi(toPort); err != nil {
			return nil, fmt.Errorf(errStr, importStr, "invalid to_port")
		} else if p2 < p1 {
			return nil, fmt.Errorf(errStr, importStr, "to_port lower than from_port")
		}
	}

	for _, source := range sources {
		// will be properly validated later
		if source != "self" && !strings.Contains(source, "sg-") && !strings.Contains(source, "pl-") && !strings.Contains(source, ":") && !strings.Contains(source, ".") {
			return nil, fmt.Errorf(errStr, importStr, "source must be cidr, ipv6cidr, prefix list, 'self', or a sg ID")
		}
	}

	log.Printf("[DEBUG] Validated import string %s", importStr)
	return importParts, nil
}

func populateSecurityGroupRuleFromImport(d *schema.ResourceData, importParts []string) error {
	log.Printf("[DEBUG] Populating resource data on import: %v", importParts)

	sgID := importParts[0]
	ruleType := importParts[1]
	protocol := importParts[2]
	fromPort, err := strconv.Atoi(importParts[3])
	if err != nil {
		return err
	}
	toPort, err := strconv.Atoi(importParts[4])
	if err != nil {
		return err
	}
	sources := importParts[5:]

	d.Set("security_group_id", sgID)

	if ruleType == securityGroupRuleTypeIngress {
		d.Set("type", ruleType)
	} else {
		d.Set("type", "egress")
	}

	d.Set("protocol", ProtocolForValue(protocol))
	d.Set("from_port", fromPort)
	d.Set("to_port", toPort)

	d.Set("self", false)
	var cidrs []string
	var prefixList []string
	var ipv6cidrs []string
	for _, source := range sources {
		if source == "self" {
			d.Set("self", true)
		} else if strings.Contains(source, "sg-") {
			d.Set("source_security_group_id", source)
		} else if strings.Contains(source, "pl-") {
			prefixList = append(prefixList, source)
		} else if strings.Contains(source, ":") {
			ipv6cidrs = append(ipv6cidrs, source)
		} else {
			cidrs = append(cidrs, source)
		}
	}
	d.Set("ipv6_cidr_blocks", ipv6cidrs)
	d.Set("cidr_blocks", cidrs)
	d.Set("prefix_list_ids", prefixList)

	return nil
}

func expandIpPermission(d *schema.ResourceData, sg *ec2.SecurityGroup) *ec2.IpPermission { // nosemgrep:caps5-in-func-name
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
