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

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_security_group", name="Security Group")
// @Tags(identifierAttribute="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;types.SecurityGroup")
// @Testing(importIgnore="revoke_rules_on_delete")
func resourceSecurityGroup() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityGroupCreate,
		ReadWithoutTimeout:   resourceSecurityGroupRead,
		UpdateWithoutTimeout: resourceSecurityGroupUpdate,
		DeleteWithoutTimeout: resourceSecurityGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		SchemaVersion: 1,
		MigrateState:  securityGroupMigrateState,

		// Keep in sync with aws_default_security_group's schema.
		// See notes in vpc_default_security_group.go.
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "Managed by Terraform",
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"egress":  securityGroupRuleSetNestedBlock,
			"ingress": securityGroupRuleSetNestedBlock,
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 255),
					validation.StringDoesNotMatch(regexache.MustCompile(`^sg-`), "cannot begin with sg-"),
				),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 255-id.UniqueIDSuffixLength),
					validation.StringDoesNotMatch(regexache.MustCompile(`^sg-`), "cannot begin with sg-"),
				),
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revoke_rules_on_delete": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

// Security Group rule nested block definition.
// Used in aws_security_group and aws_default_security_group ingress and egress rule sets.
var (
	securityGroupRuleSetNestedBlock = &schema.Schema{
		Type:       schema.TypeSet,
		Optional:   true,
		Computed:   true,
		ConfigMode: schema.SchemaConfigModeAttr,
		Elem:       securityGroupRuleNestedBlock,
		Set:        securityGroupRuleHash,
	}

	securityGroupRuleNestedBlock = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cidr_blocks": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
				},
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validSecurityGroupRuleDescription,
			},
			"from_port": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"ipv6_cidr_blocks": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidIPv6CIDRNetworkAddress,
				},
			},
			"prefix_list_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrProtocol: {
				Type:      schema.TypeString,
				Required:  true,
				StateFunc: protocolStateFunc,
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString, // Required to ensure consistent hashing
			},
			"self": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"to_port": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
)

func resourceSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	inputC := &ec2.CreateSecurityGroupInput{
		GroupName:         aws.String(name),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeSecurityGroup),
	}

	if v := d.Get(names.AttrDescription); v != nil {
		inputC.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVPCID); ok {
		inputC.VpcId = aws.String(v.(string))
	}

	output, err := conn.CreateSecurityGroup(ctx, inputC)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.GroupId))

	// Wait for the security group to truly exist
	if _, err := waitSecurityGroupCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Group (%s) create: %s", d.Id(), err)
	}

	// AWS defaults all Security Groups to have an ALLOW ALL egress rule.
	// Here we revoke that rule, so users don't unknowingly have/use it.
	inputR := &ec2.RevokeSecurityGroupEgressInput{
		GroupId: aws.String(d.Id()),
		IpPermissions: []awstypes.IpPermission{
			{
				FromPort: aws.Int32(0),
				ToPort:   aws.Int32(0),
				IpRanges: []awstypes.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
				IpProtocol: aws.String("-1"),
			},
		},
	}

	if _, err := conn.RevokeSecurityGroupEgress(ctx, inputR); err != nil {
		return sdkdiag.AppendErrorf(diags, "revoking default IPv4 egress rule for Security Group (%s): %s", d.Id(), err)
	}

	inputR = &ec2.RevokeSecurityGroupEgressInput{
		GroupId: aws.String(d.Id()),
		IpPermissions: []awstypes.IpPermission{
			{
				FromPort: aws.Int32(0),
				ToPort:   aws.Int32(0),
				Ipv6Ranges: []awstypes.Ipv6Range{
					{
						CidrIpv6: aws.String("::/0"),
					},
				},
				IpProtocol: aws.String("-1"),
			},
		},
	}

	if _, err := conn.RevokeSecurityGroupEgress(ctx, inputR); err != nil {
		// If we have a NotFound or InvalidParameterValue, then we are trying to remove the default IPv6 egress of a non-IPv6 enabled SG.
		if !tfawserr.ErrCodeEquals(err, errCodeInvalidPermissionNotFound) && !tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "remote-ipv6-range") {
			return sdkdiag.AppendErrorf(diags, "revoking default IPv6 egress rule for Security Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSecurityGroupUpdate(ctx, d, meta)...)
}

func resourceSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	sg, err := findSecurityGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Group (%s): %s", d.Id(), err)
	}

	remoteIngressRules := securityGroupIPPermGather(d.Id(), sg.IpPermissions, sg.OwnerId)
	remoteEgressRules := securityGroupIPPermGather(d.Id(), sg.IpPermissionsEgress, sg.OwnerId)

	localIngressRules := d.Get("ingress").(*schema.Set).List()
	localEgressRules := d.Get("egress").(*schema.Set).List()

	// Loop through the local state of rules, doing a match against the remote
	// ruleSet we built above.
	ingressRules := matchRules("ingress", localIngressRules, remoteIngressRules)
	egressRules := matchRules("egress", localEgressRules, remoteEgressRules)

	ownerID := aws.ToString(sg.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("security-group/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())
	d.Set(names.AttrDescription, sg.Description)
	d.Set(names.AttrName, sg.GroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(sg.GroupName)))
	d.Set(names.AttrOwnerID, ownerID)
	d.Set(names.AttrVPCID, sg.VpcId)

	if err := d.Set("ingress", ingressRules); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ingress: %s", err)
	}

	if err := d.Set("egress", egressRules); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting egress: %s", err)
	}

	setTagsOut(ctx, sg.Tags)

	return diags
}

func resourceSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	group, err := findSecurityGroupByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Group (%s): %s", d.Id(), err)
	}

	err = updateSecurityGroupRules(ctx, conn, d, "ingress", group)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Group (%s) ingress rules: %s", d.Id(), err)
	}

	err = updateSecurityGroupRules(ctx, conn, d, "egress", group)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Group (%s) egress rules: %s", d.Id(), err)
	}

	return append(diags, resourceSecurityGroupRead(ctx, d, meta)...)
}

func resourceSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, d.Id())
	ctx = tflog.SetField(ctx, names.AttrVPCID, d.Get(names.AttrVPCID))

	if err := deleteLingeringENIs(ctx, meta.(*conns.AWSClient).EC2Client(ctx), "group-id", d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ENIs using Security Group (%s): %s", d.Id(), err)
	}

	// conditionally revoke rules first before attempting to delete the group
	if v := d.Get("revoke_rules_on_delete").(bool); v {
		err := forceRevokeSecurityGroupRules(ctx, conn, d.Id(), false)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	firstShortRetry := 1 * time.Minute
	remainingRetry := d.Timeout(schema.TimeoutDelete) - firstShortRetry
	if d.Timeout(schema.TimeoutDelete) < 1*time.Minute {
		remainingRetry = 30 * time.Second
	}

	tflog.Info(ctx, "Deleting EC2 Security Group")

	_, err := tfresource.RetryWhenAWSErrCodeEquals(
		ctx,
		firstShortRetry, // short initial attempt followed by full length attempt
		func() (interface{}, error) {
			return conn.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
				GroupId: aws.String(d.Id()),
			})
		},
		errCodeDependencyViolation, errCodeInvalidGroupInUse,
	)

	if tfawserr.ErrCodeEquals(err, errCodeDependencyViolation) || tfawserr.ErrCodeEquals(err, errCodeInvalidGroupInUse) {
		if v := d.Get("revoke_rules_on_delete").(bool); v {
			err := forceRevokeSecurityGroupRules(ctx, conn, d.Id(), true)

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		_, err = tfresource.RetryWhenAWSErrCodeEquals(
			ctx,
			remainingRetry,
			func() (interface{}, error) {
				return conn.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
					GroupId: aws.String(d.Id()),
				})
			},
			errCodeDependencyViolation, errCodeInvalidGroupInUse,
		)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Group (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findSecurityGroupByID(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Group (%s) delete: %s", d.Id(), err)
	}

	return diags
}

// forceRevokeSecurityGroupRules revokes all of the security group's ingress & egress rules
// AND rules in other security groups that depend on this security group. Trying to delete
// this security group with rules that originate in other groups but point here, will cause
// a DepedencyViolation error. searchAll = true means to search every security group
// looking for a rule depending on this security group. Otherwise, it will only look at
// groups that this group knows about.
func forceRevokeSecurityGroupRules(ctx context.Context, conn *ec2.Client, id string, searchAll bool) error {
	conns.GlobalMutexKV.Lock(id)
	defer conns.GlobalMutexKV.Unlock(id)

	rules, err := rulesInSGsTouchingThis(ctx, conn, id, searchAll)
	if err != nil {
		return fmt.Errorf("describing security group rules: %s", err)
	}

	for _, rule := range rules {
		var err error

		if rule.IsEgress == nil || !aws.ToBool(rule.IsEgress) {
			input := &ec2.RevokeSecurityGroupIngressInput{
				SecurityGroupRuleIds: []string{aws.ToString(rule.SecurityGroupRuleId)},
			}

			if rule.GroupId != nil {
				input.GroupId = rule.GroupId
			} else {
				// If this rule isn't "owned" by this group, this will be wrong.
				// However, ec2.SecurityGroupRule doesn't include name so can't
				// be used. If it affects anything, this would affect default
				// VPCs.
				sg, err := findSecurityGroupByID(ctx, conn, id)
				if err != nil {
					return fmt.Errorf("reading Security Group (%s): %w", id, err)
				}

				input.GroupName = sg.GroupName
			}

			_, err = conn.RevokeSecurityGroupIngress(ctx, input)
		} else {
			input := &ec2.RevokeSecurityGroupEgressInput{
				GroupId:              rule.GroupId,
				SecurityGroupRuleIds: []string{aws.ToString(rule.SecurityGroupRuleId)},
			}

			_, err = conn.RevokeSecurityGroupEgress(ctx, input)
		}

		if tfawserr.ErrCodeEquals(err, errCodeInvalidSecurityGroupRuleIdNotFound) {
			continue
		}

		if tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("revoking Security Group (%s) Rule (%s): %w", id, aws.ToString(rule.SecurityGroupRuleId), err)
		}
	}

	return nil
}

// rulesInSGsTouchingThis finds all rules related to this group even if they live in
// other groups. If searchAll = true, this could take a while as it looks through every
// security group accessible from the account. This should only be used for troublesome
// DependencyViolations.
func rulesInSGsTouchingThis(ctx context.Context, conn *ec2.Client, id string, searchAll bool) ([]awstypes.SecurityGroupRule, error) {
	var input *ec2.DescribeSecurityGroupRulesInput

	if searchAll {
		input = &ec2.DescribeSecurityGroupRulesInput{}
	} else {
		sgs, err := relatedSGs(ctx, conn, id)
		if err != nil {
			return nil, fmt.Errorf("describing security group rules: %s", err)
		}

		input = &ec2.DescribeSecurityGroupRulesInput{
			Filters: []awstypes.Filter{
				{
					Name:   aws.String("group-id"),
					Values: sgs,
				},
			},
		}
	}

	rules := []awstypes.SecurityGroupRule{}

	pages := ec2.NewDescribeSecurityGroupRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("reading Security Group rules: %w", err)
		}

		for _, rule := range page.SecurityGroupRules {
			if rule.GroupId == nil {
				continue
			}

			if aws.ToString(rule.GroupId) == id {
				rules = append(rules, rule)
				continue
			}

			if rule.ReferencedGroupInfo != nil && rule.ReferencedGroupInfo.GroupId != nil && aws.ToString(rule.ReferencedGroupInfo.GroupId) == id {
				rules = append(rules, rule)
				continue
			}
		}
	}

	return rules, nil
}

// relatedSGs returns security group IDs of any other security group that is related
// to this one through a rule that this group knows about. This group can still have
// dependent rules beyond those in these groups. However, the majority of the time,
// revoking related rules should allow the group to be deleted.
func relatedSGs(ctx context.Context, conn *ec2.Client, id string) ([]string, error) {
	relatedSGs := []string{id}

	sg, err := findSecurityGroupByID(ctx, conn, id)
	if err != nil {
		return nil, fmt.Errorf("reading Security Group (%s): %w", id, err)
	}

	if len(sg.IpPermissions) > 0 {
		for _, v := range sg.IpPermissions {
			for _, v := range v.UserIdGroupPairs {
				if v.GroupId != nil && aws.ToString(v.GroupId) != id {
					relatedSGs = append(relatedSGs, aws.ToString(v.GroupId))
				}
			}
		}
	}

	if len(sg.IpPermissionsEgress) > 0 {
		for _, v := range sg.IpPermissionsEgress {
			for _, v := range v.UserIdGroupPairs {
				if v.GroupId != nil && aws.ToString(v.GroupId) != id {
					relatedSGs = append(relatedSGs, aws.ToString(v.GroupId))
				}
			}
		}
	}

	return relatedSGs, nil
}

func securityGroupRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["from_port"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["to_port"].(int)))
	p := protocolForValue(m[names.AttrProtocol].(string))
	buf.WriteString(fmt.Sprintf("%s-", p))
	buf.WriteString(fmt.Sprintf("%t-", m["self"].(bool)))

	// We need to make sure to sort the strings below so that we always
	// generate the same hash code no matter what is in the set.
	if v, ok := m["cidr_blocks"]; ok {
		vs := v.([]interface{})
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if v, ok := m["ipv6_cidr_blocks"]; ok {
		vs := v.([]interface{})
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if v, ok := m["prefix_list_ids"]; ok {
		vs := v.([]interface{})
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if v, ok := m[names.AttrSecurityGroups]; ok {
		vs := v.(*schema.Set).List()
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if m[names.AttrDescription].(string) != "" {
		buf.WriteString(fmt.Sprintf("%s-", m[names.AttrDescription].(string)))
	}

	return create.StringHashcode(buf.String())
}

func securityGroupIPPermGather(groupId string, permissions []awstypes.IpPermission, ownerId *string) []map[string]interface{} {
	ruleMap := make(map[string]map[string]interface{})
	for _, perm := range permissions {
		if len(perm.IpRanges) > 0 {
			for _, ip := range perm.IpRanges {
				desc := aws.ToString(ip.Description)

				rule := initSecurityGroupRule(ruleMap, perm, desc)

				raw, ok := rule["cidr_blocks"]
				if !ok {
					raw = make([]string, 0)
				}
				list := raw.([]string)

				rule["cidr_blocks"] = append(list, *ip.CidrIp)
			}
		}

		if len(perm.Ipv6Ranges) > 0 {
			for _, ip := range perm.Ipv6Ranges {
				desc := aws.ToString(ip.Description)

				rule := initSecurityGroupRule(ruleMap, perm, desc)

				raw, ok := rule["ipv6_cidr_blocks"]
				if !ok {
					raw = make([]string, 0)
				}
				list := raw.([]string)

				rule["ipv6_cidr_blocks"] = append(list, *ip.CidrIpv6)
			}
		}

		if len(perm.PrefixListIds) > 0 {
			for _, pl := range perm.PrefixListIds {
				desc := aws.ToString(pl.Description)

				rule := initSecurityGroupRule(ruleMap, perm, desc)

				raw, ok := rule["prefix_list_ids"]
				if !ok {
					raw = make([]string, 0)
				}
				list := raw.([]string)

				rule["prefix_list_ids"] = append(list, *pl.PrefixListId)
			}
		}

		groups := flattenSecurityGroups(perm.UserIdGroupPairs, ownerId)
		if len(groups) > 0 {
			for _, g := range groups {
				desc := aws.ToString(g.Description)

				rule := initSecurityGroupRule(ruleMap, perm, desc)

				if aws.ToString(g.GroupId) == groupId {
					rule["self"] = true
					continue
				}

				raw, ok := rule[names.AttrSecurityGroups]
				if !ok {
					raw = schema.NewSet(schema.HashString, nil)
				}
				list := raw.(*schema.Set)

				if g.GroupName != nil {
					list.Add(*g.GroupName)
				} else {
					list.Add(*g.GroupId)
				}
				rule[names.AttrSecurityGroups] = list
			}
		}
	}

	rules := make([]map[string]interface{}, 0, len(ruleMap))
	for _, m := range ruleMap {
		rules = append(rules, m)
	}

	return rules
}

func updateSecurityGroupRules(ctx context.Context, conn *ec2.Client, d *schema.ResourceData, ruleType string, group *awstypes.SecurityGroup) error {
	if !d.HasChange(ruleType) {
		return nil
	}

	o, n := d.GetChange(ruleType)
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}

	os := securityGroupExpandRules(o.(*schema.Set))
	ns := securityGroupExpandRules(n.(*schema.Set))

	del, err := expandIPPerms(group, securityGroupCollapseRules(ruleType, os.Difference(ns).List()))

	if err != nil {
		return fmt.Errorf("updating rules: %w", err)
	}

	add, err := expandIPPerms(group, securityGroupCollapseRules(ruleType, ns.Difference(os).List()))

	if err != nil {
		return fmt.Errorf("updating rules: %w", err)
	}

	// TODO: We need to handle partial state better in the in-between
	// in this update.

	// TODO: It'd be nicer to authorize before removing, but then we have
	// to deal with complicated unrolling to get individual CIDR blocks
	// to avoid authorizing already authorized sources. Removing before
	// adding is easier here, and Terraform should be fast enough to
	// not have service issues.

	if len(del) > 0 {
		if ruleType == "egress" {
			input := &ec2.RevokeSecurityGroupEgressInput{
				GroupId:       group.GroupId,
				IpPermissions: del,
			}

			_, err = conn.RevokeSecurityGroupEgress(ctx, input)
		} else {
			input := &ec2.RevokeSecurityGroupIngressInput{
				GroupId:       group.GroupId,
				IpPermissions: del,
			}

			_, err = conn.RevokeSecurityGroupIngress(ctx, input)
		}

		if err != nil {
			return fmt.Errorf("revoking Security Group (%s) rules: %w", ruleType, err)
		}
	}

	if len(add) > 0 {
		if ruleType == "egress" {
			input := &ec2.AuthorizeSecurityGroupEgressInput{
				GroupId:       group.GroupId,
				IpPermissions: add,
			}

			_, err = conn.AuthorizeSecurityGroupEgress(ctx, input)
		} else {
			input := &ec2.AuthorizeSecurityGroupIngressInput{
				GroupId:       group.GroupId,
				IpPermissions: add,
			}

			_, err = conn.AuthorizeSecurityGroupIngress(ctx, input)
		}

		if err != nil {
			return fmt.Errorf("authorizing Security Group (%s) rules: %w", ruleType, err)
		}
	}

	return nil
}

// Takes the result of flatmap.Expand for an array of ingress/egress security
// group rules and returns EC2 API compatible objects. This function will error
// if it finds invalid permissions input, namely a protocol of "-1" with either
// to_port or from_port set to a non-zero value.
func expandIPPerms(group *awstypes.SecurityGroup, configured []interface{}) ([]awstypes.IpPermission, error) {
	perms := make([]awstypes.IpPermission, len(configured))
	for i, mRaw := range configured {
		var perm awstypes.IpPermission
		m := mRaw.(map[string]interface{})

		perm.IpProtocol = aws.String(protocolForValue(m[names.AttrProtocol].(string)))

		if protocol, fromPort, toPort := aws.ToString(perm.IpProtocol), m["from_port"].(int), m["to_port"].(int); protocol != "-1" {
			perm.FromPort = aws.Int32(int32(fromPort))
			perm.ToPort = aws.Int32(int32(toPort))
		} else if fromPort != 0 || toPort != 0 {
			// When protocol is "-1", AWS won't store any ports for the
			// rule, but also won't error if the user specifies ports other
			// than '0'. Force the user to make a deliberate '0' port
			// choice when specifying a "-1" protocol, and tell them about
			// AWS's behavior in the error message.
			return nil, fmt.Errorf(
				"from_port (%d) and to_port (%d) must both be 0 to use the 'ALL' \"-1\" protocol!",
				fromPort, toPort)
		}

		var groups []string
		if raw, ok := m[names.AttrSecurityGroups]; ok {
			list := raw.(*schema.Set).List()
			for _, v := range list {
				groups = append(groups, v.(string))
			}
		}
		if v, ok := m["self"]; ok && v.(bool) {
			groups = append(groups, aws.ToString(group.GroupId))
		}

		if len(groups) > 0 {
			perm.UserIdGroupPairs = make([]awstypes.UserIdGroupPair, len(groups))
			for i, name := range groups {
				ownerId, id := "", name
				if items := strings.Split(id, "/"); len(items) > 1 {
					ownerId, id = items[0], items[1]
				}

				perm.UserIdGroupPairs[i] = awstypes.UserIdGroupPair{
					GroupId: aws.String(id),
				}

				if ownerId != "" {
					perm.UserIdGroupPairs[i].UserId = aws.String(ownerId)
				}
			}
		}

		if raw, ok := m["cidr_blocks"]; ok {
			list := raw.([]interface{})
			for _, v := range list {
				perm.IpRanges = append(perm.IpRanges, awstypes.IpRange{CidrIp: aws.String(v.(string))})
			}
		}
		if raw, ok := m["ipv6_cidr_blocks"]; ok {
			list := raw.([]interface{})
			for _, v := range list {
				perm.Ipv6Ranges = append(perm.Ipv6Ranges, awstypes.Ipv6Range{CidrIpv6: aws.String(v.(string))})
			}
		}

		if raw, ok := m["prefix_list_ids"]; ok {
			list := raw.([]interface{})
			for _, v := range list {
				perm.PrefixListIds = append(perm.PrefixListIds, awstypes.PrefixListId{PrefixListId: aws.String(v.(string))})
			}
		}

		if raw, ok := m[names.AttrDescription]; ok {
			description := raw.(string)
			if description != "" {
				for i := range perm.IpRanges {
					perm.IpRanges[i].Description = aws.String(description)
				}
				for i := range perm.Ipv6Ranges {
					perm.Ipv6Ranges[i].Description = aws.String(description)
				}
				for i := range perm.PrefixListIds {
					perm.PrefixListIds[i].Description = aws.String(description)
				}
				for i := range perm.UserIdGroupPairs {
					perm.UserIdGroupPairs[i].Description = aws.String(description)
				}
			}
		}

		perms[i] = perm
	}

	return perms, nil
}

// Like ec2.GroupIdentifier but with additional rule description.
type groupIdentifier struct {
	// The ID of the security group.
	GroupId *string

	// The name of the security group.
	GroupName *string

	Description *string
}

// Flattens an array of UserSecurityGroups into a []*GroupIdentifier
func flattenSecurityGroups(list []awstypes.UserIdGroupPair, ownerId *string) []*groupIdentifier {
	result := make([]*groupIdentifier, 0, len(list))

	for _, g := range list {
		var userID string
		if aws.ToString(g.UserId) != "" && (ownerId == nil || aws.ToString(ownerId) != aws.ToString(g.UserId)) {
			userID = aws.ToString(g.UserId)
		}
		// userid nil here for same vpc groups

		id := aws.ToString(g.GroupId)
		if userID != "" {
			id = userID + "/" + id
		}

		result = append(result, &groupIdentifier{
			GroupId:     aws.String(id),
			Description: g.Description,
		})
	}

	return result
}

// matchRules receives the group id, type of rules, and the local / remote maps
// of rules. We iterate through the local set of rules trying to find a matching
// remote rule, which may be structured differently because of how AWS
// aggregates the rules under the to, from, and type.
//
// Matching rules are written to state, with their elements removed from the
// remote set
//
// If no match is found, we'll write the remote rule to state and let the graph
// sort things out
func matchRules(rType string, local []interface{}, remote []map[string]interface{}) []map[string]interface{} {
	// For each local ip or security_group, we need to match against the remote
	// ruleSet until all ips or security_groups are found

	// saves represents the rules that have been identified to be saved to state,
	// in the appropriate d.Set("{ingress,egress}") call.
	var saves []map[string]interface{}
	for _, raw := range local {
		l := raw.(map[string]interface{})

		var selfVal bool
		if v, ok := l["self"]; ok {
			selfVal = v.(bool)
		}

		// matching against self is required to detect rules that only include self
		// as the rule. SecurityGroupIPPermGather parses the group out
		// and replaces it with self if it's ID is found
		localHash := idHash(rType, l[names.AttrProtocol].(string), int32(l["to_port"].(int)), int32(l["from_port"].(int)), selfVal)

		// loop remote rules, looking for a matching hash
		for _, r := range remote {
			var remoteSelfVal bool
			if v, ok := r["self"]; ok {
				remoteSelfVal = v.(bool)
			}

			// hash this remote rule and compare it for a match consideration with the
			// local rule we're examining
			rHash := idHash(rType, r[names.AttrProtocol].(string), r["to_port"].(int32), r["from_port"].(int32), remoteSelfVal)
			if rHash == localHash {
				var numExpectedCidrs, numExpectedIpv6Cidrs, numExpectedPrefixLists, numExpectedSGs, numRemoteCidrs, numRemoteIpv6Cidrs, numRemotePrefixLists, numRemoteSGs int
				var matchingCidrs []string
				var matchingIpv6Cidrs []string
				var matchingSGs []string
				var matchingPrefixLists []string

				// grab the local/remote cidr and sg groups, capturing the expected and
				// actual counts
				lcRaw, ok := l["cidr_blocks"]
				if ok {
					numExpectedCidrs = len(l["cidr_blocks"].([]interface{}))
				}
				liRaw, ok := l["ipv6_cidr_blocks"]
				if ok {
					numExpectedIpv6Cidrs = len(l["ipv6_cidr_blocks"].([]interface{}))
				}
				lpRaw, ok := l["prefix_list_ids"]
				if ok {
					numExpectedPrefixLists = len(l["prefix_list_ids"].([]interface{}))
				}
				lsRaw, ok := l[names.AttrSecurityGroups]
				if ok {
					numExpectedSGs = len(l[names.AttrSecurityGroups].(*schema.Set).List())
				}

				rcRaw, ok := r["cidr_blocks"]
				if ok {
					numRemoteCidrs = len(r["cidr_blocks"].([]string))
				}
				riRaw, ok := r["ipv6_cidr_blocks"]
				if ok {
					numRemoteIpv6Cidrs = len(r["ipv6_cidr_blocks"].([]string))
				}
				rpRaw, ok := r["prefix_list_ids"]
				if ok {
					numRemotePrefixLists = len(r["prefix_list_ids"].([]string))
				}

				rsRaw, ok := r[names.AttrSecurityGroups]
				if ok {
					numRemoteSGs = len(r[names.AttrSecurityGroups].(*schema.Set).List())
				}

				// check some early failures
				if numExpectedCidrs > numRemoteCidrs {
					log.Printf("[DEBUG] Local rule has more CIDR blocks, continuing (%d/%d)", numExpectedCidrs, numRemoteCidrs)
					continue
				}
				if numExpectedIpv6Cidrs > numRemoteIpv6Cidrs {
					log.Printf("[DEBUG] Local rule has more IPV6 CIDR blocks, continuing (%d/%d)", numExpectedIpv6Cidrs, numRemoteIpv6Cidrs)
					continue
				}
				if numExpectedPrefixLists > numRemotePrefixLists {
					log.Printf("[DEBUG] Local rule has more prefix lists, continuing (%d/%d)", numExpectedPrefixLists, numRemotePrefixLists)
					continue
				}
				if numExpectedSGs > numRemoteSGs {
					log.Printf("[DEBUG] Local rule has more Security Groups, continuing (%d/%d)", numExpectedSGs, numRemoteSGs)
					continue
				}

				// match CIDRs by converting both to sets, and using Set methods
				var localCidrs []interface{}
				if lcRaw != nil {
					localCidrs = lcRaw.([]interface{})
				}
				localCidrSet := schema.NewSet(schema.HashString, localCidrs)

				// remote cidrs are presented as a slice of strings, so we need to
				// reformat them into a slice of interfaces to be used in creating the
				// remote cidr set
				var remoteCidrs []string
				if rcRaw != nil {
					remoteCidrs = rcRaw.([]string)
				}
				// convert remote cidrs to a set, for easy comparisons
				var list []interface{}
				for _, s := range remoteCidrs {
					list = append(list, s)
				}
				remoteCidrSet := schema.NewSet(schema.HashString, list)

				// Build up a list of local cidrs that are found in the remote set
				for _, s := range localCidrSet.List() {
					if remoteCidrSet.Contains(s) {
						matchingCidrs = append(matchingCidrs, s.(string))
					}
				}

				//IPV6 CIDRs
				var localIpv6Cidrs []interface{}
				if liRaw != nil {
					localIpv6Cidrs = liRaw.([]interface{})
				}
				localIpv6CidrSet := schema.NewSet(schema.HashString, localIpv6Cidrs)

				var remoteIpv6Cidrs []string
				if riRaw != nil {
					remoteIpv6Cidrs = riRaw.([]string)
				}
				var listIpv6 []interface{}
				for _, s := range remoteIpv6Cidrs {
					listIpv6 = append(listIpv6, s)
				}
				remoteIpv6CidrSet := schema.NewSet(schema.HashString, listIpv6)

				for _, s := range localIpv6CidrSet.List() {
					if remoteIpv6CidrSet.Contains(s) {
						matchingIpv6Cidrs = append(matchingIpv6Cidrs, s.(string))
					}
				}

				// match prefix lists by converting both to sets, and using Set methods
				var localPrefixLists []interface{}
				if lpRaw != nil {
					localPrefixLists = lpRaw.([]interface{})
				}
				localPrefixListsSet := schema.NewSet(schema.HashString, localPrefixLists)

				// remote prefix lists are presented as a slice of strings, so we need to
				// reformat them into a slice of interfaces to be used in creating the
				// remote prefix list set
				var remotePrefixLists []string
				if rpRaw != nil {
					remotePrefixLists = rpRaw.([]string)
				}
				// convert remote prefix lists to a set, for easy comparison
				list = nil
				for _, s := range remotePrefixLists {
					list = append(list, s)
				}
				remotePrefixListsSet := schema.NewSet(schema.HashString, list)

				// Build up a list of local prefix lists that are found in the remote set
				for _, s := range localPrefixListsSet.List() {
					if remotePrefixListsSet.Contains(s) {
						matchingPrefixLists = append(matchingPrefixLists, s.(string))
					}
				}

				// match SGs. Both local and remote are already sets
				var localSGSet *schema.Set
				if lsRaw == nil {
					localSGSet = schema.NewSet(schema.HashString, nil)
				} else {
					localSGSet = lsRaw.(*schema.Set)
				}

				var remoteSGSet *schema.Set
				if rsRaw == nil {
					remoteSGSet = schema.NewSet(schema.HashString, nil)
				} else {
					remoteSGSet = rsRaw.(*schema.Set)
				}

				// Build up a list of local security groups that are found in the remote set
				for _, s := range localSGSet.List() {
					if remoteSGSet.Contains(s) {
						matchingSGs = append(matchingSGs, s.(string))
					}
				}

				// compare equalities for matches.
				// If we found the number of cidrs and number of sgs, we declare a
				// match, and then remove those elements from the remote rule, so that
				// this remote rule can still be considered by other local rules
				if numExpectedCidrs == len(matchingCidrs) {
					if numExpectedIpv6Cidrs == len(matchingIpv6Cidrs) {
						if numExpectedPrefixLists == len(matchingPrefixLists) {
							if numExpectedSGs == len(matchingSGs) {
								// confirm that self references match
								var lSelf bool
								var rSelf bool
								if _, ok := l["self"]; ok {
									lSelf = l["self"].(bool)
								}
								if _, ok := r["self"]; ok {
									rSelf = r["self"].(bool)
								}
								if rSelf == lSelf {
									delete(r, "self")
									// pop local cidrs from remote
									diffCidr := remoteCidrSet.Difference(localCidrSet)
									var newCidr []string
									for _, cRaw := range diffCidr.List() {
										newCidr = append(newCidr, cRaw.(string))
									}

									// reassigning
									if len(newCidr) > 0 {
										r["cidr_blocks"] = newCidr
									} else {
										delete(r, "cidr_blocks")
									}

									//// IPV6
									//// Comparison
									diffIpv6Cidr := remoteIpv6CidrSet.Difference(localIpv6CidrSet)
									var newIpv6Cidr []string
									for _, cRaw := range diffIpv6Cidr.List() {
										newIpv6Cidr = append(newIpv6Cidr, cRaw.(string))
									}

									// reassigning
									if len(newIpv6Cidr) > 0 {
										r["ipv6_cidr_blocks"] = newIpv6Cidr
									} else {
										delete(r, "ipv6_cidr_blocks")
									}

									// pop local prefix lists from remote
									diffPrefixLists := remotePrefixListsSet.Difference(localPrefixListsSet)
									var newPrefixLists []string
									for _, pRaw := range diffPrefixLists.List() {
										newPrefixLists = append(newPrefixLists, pRaw.(string))
									}

									// reassigning
									if len(newPrefixLists) > 0 {
										r["prefix_list_ids"] = newPrefixLists
									} else {
										delete(r, "prefix_list_ids")
									}

									// pop local sgs from remote
									diffSGs := remoteSGSet.Difference(localSGSet)
									if len(diffSGs.List()) > 0 {
										r[names.AttrSecurityGroups] = diffSGs
									} else {
										delete(r, names.AttrSecurityGroups)
									}

									// copy over any remote rule description
									if _, ok := r[names.AttrDescription]; ok {
										l[names.AttrDescription] = r[names.AttrDescription]
									}

									saves = append(saves, l)
								}
							}
						}
					}
				}
			}
		}
	}
	// Here we catch any remote rules that have not been stripped of all self,
	// cidrs, and security groups. We'll add remote rules here that have not been
	// matched locally, and let the graph sort things out. This will happen when
	// rules are added externally to Terraform
	for _, r := range remote {
		var lenCidr, lenIpv6Cidr, lenPrefixLists, lenSGs int
		if rCidrs, ok := r["cidr_blocks"]; ok {
			lenCidr = len(rCidrs.([]string))
		}
		if rIpv6Cidrs, ok := r["ipv6_cidr_blocks"]; ok {
			lenIpv6Cidr = len(rIpv6Cidrs.([]string))
		}
		if rPrefixLists, ok := r["prefix_list_ids"]; ok {
			lenPrefixLists = len(rPrefixLists.([]string))
		}
		if rawSGs, ok := r[names.AttrSecurityGroups]; ok {
			lenSGs = len(rawSGs.(*schema.Set).List())
		}

		if _, ok := r["self"]; ok {
			if r["self"].(bool) {
				lenSGs++
			}
		}

		if lenSGs+lenCidr+lenIpv6Cidr+lenPrefixLists > 0 {
			log.Printf("[DEBUG] Found a remote Rule that wasn't empty: (%#v)", r)
			saves = append(saves, r)
		}
	}

	return saves
}

// Duplicate ingress/egress block structure and fill out all
// the required fields
func resourceSecurityGroupCopyRule(src map[string]interface{}, self bool, k string, v interface{}) map[string]interface{} {
	var keys_to_copy = []string{names.AttrDescription, "from_port", "to_port", names.AttrProtocol}

	dst := make(map[string]interface{})
	for _, key := range keys_to_copy {
		if val, ok := src[key]; ok {
			dst[key] = val
		}
	}
	if k != "" {
		dst[k] = v
	}
	if _, ok := src["self"]; ok {
		dst["self"] = self
	}
	return dst
}

// Given a set of SG rules (ingress/egress blocks), this function
// will group the rules by from_port/to_port/protocol/description
// tuples. This is inverse operation of
// securityGroupExpandRules()
//
// For more detail, see comments for
// securityGroupExpandRules()
func securityGroupCollapseRules(ruleset string, rules []interface{}) []interface{} {
	var keys_to_collapse = []string{"cidr_blocks", "ipv6_cidr_blocks", "prefix_list_ids", names.AttrSecurityGroups}

	collapsed := make(map[string]map[string]interface{})

	for _, rule := range rules {
		r := rule.(map[string]interface{})

		ruleHash := idCollapseHash(ruleset, r[names.AttrProtocol].(string), int32(r["to_port"].(int)), int32(r["from_port"].(int)), r[names.AttrDescription].(string))

		if _, ok := collapsed[ruleHash]; ok {
			if v, ok := r["self"]; ok && v.(bool) {
				collapsed[ruleHash]["self"] = r["self"]
			}
		} else {
			collapsed[ruleHash] = r
			continue
		}

		for _, key := range keys_to_collapse {
			if _, ok := r[key]; ok {
				if _, ok := collapsed[ruleHash][key]; ok {
					if key == names.AttrSecurityGroups {
						collapsed[ruleHash][key] = collapsed[ruleHash][key].(*schema.Set).Union(r[key].(*schema.Set))
					} else {
						collapsed[ruleHash][key] = append(collapsed[ruleHash][key].([]interface{}), r[key].([]interface{})...)
					}
				} else {
					collapsed[ruleHash][key] = r[key]
				}
			}
		}
	}

	values := make([]interface{}, 0, len(collapsed))
	for _, val := range collapsed {
		values = append(values, val)
	}
	return values
}

// securityGroupExpandRules works in pair with
// securityGroupCollapseRules and is used as a
// workaround for the problem explained in
// https://github.com/hashicorp/terraform-provider-aws/pull/4726
//
// This function converts every ingress/egress block that
// contains multiple rules to multiple blocks with only one
// rule. Doing a Difference operation on such a normalized
// set helps to avoid unnecessary removal of unchanged
// rules during the Apply step.
//
// For example, in terraform syntax, the following block:
//
//	ingress {
//	  from_port = 80
//	  to_port = 80
//	  protocol = "tcp"
//	  cidr_blocks = [
//	    "192.168.0.1/32",
//	    "192.168.0.2/32",
//	  ]
//	}
//
// will be converted to the two blocks below:
//
//	ingress {
//	  from_port = 80
//	  to_port = 80
//	  protocol = "tcp"
//	  cidr_blocks = [ "192.168.0.1/32" ]
//	}
//
//	ingress {
//	  from_port = 80
//	  to_port = 80
//	  protocol = "tcp"
//	  cidr_blocks = [ "192.168.0.2/32" ]
//	}
//
// Then the Difference operation is executed on the new set
// to find which rules got modified, and the resulting set
// is then passed to securityGroupCollapseRules
// to convert the "diff" back to a more compact form for
// execution. Such compact form helps reduce the number of
// API calls.
func securityGroupExpandRules(rules *schema.Set) *schema.Set {
	var keys_to_expand = []string{"cidr_blocks", "ipv6_cidr_blocks", "prefix_list_ids", names.AttrSecurityGroups}

	normalized := schema.NewSet(securityGroupRuleHash, nil)

	for _, rawRule := range rules.List() {
		rule := rawRule.(map[string]interface{})

		if v, ok := rule["self"]; ok && v.(bool) {
			new_rule := resourceSecurityGroupCopyRule(rule, true, "", nil)
			normalized.Add(new_rule)
		}
		for _, key := range keys_to_expand {
			item, exists := rule[key]
			if exists {
				var list []interface{}
				if key == names.AttrSecurityGroups {
					list = item.(*schema.Set).List()
				} else {
					list = item.([]interface{})
				}
				for _, v := range list {
					var new_rule map[string]interface{}
					if key == names.AttrSecurityGroups {
						new_v := schema.NewSet(schema.HashString, nil)
						new_v.Add(v)
						new_rule = resourceSecurityGroupCopyRule(rule, false, key, new_v)
					} else {
						new_v := make([]interface{}, 0)
						new_v = append(new_v, v)
						new_rule = resourceSecurityGroupCopyRule(rule, false, key, new_v)
					}
					normalized.Add(new_rule)
				}
			}
		}
	}

	return normalized
}

// Convert type-to_port-from_port-protocol-description tuple
// to a hash to use as a key in Set.
func idCollapseHash(rType, protocol string, toPort, fromPort int32, description string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", rType))
	buf.WriteString(fmt.Sprintf("%d-", toPort))
	buf.WriteString(fmt.Sprintf("%d-", fromPort))
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(protocol)))
	buf.WriteString(fmt.Sprintf("%s-", description))

	return fmt.Sprintf("rule-%d", create.StringHashcode(buf.String()))
}

// Creates a unique hash for the type, ports, and protocol, used as a key in
// maps
func idHash(rType, protocol string, toPort, fromPort int32, self bool) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", rType))
	buf.WriteString(fmt.Sprintf("%d-", toPort))
	buf.WriteString(fmt.Sprintf("%d-", fromPort))
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(protocol)))
	buf.WriteString(fmt.Sprintf("%t-", self))

	return fmt.Sprintf("rule-%d", create.StringHashcode(buf.String()))
}

// protocolStateFunc ensures we only store a string in any protocol field
func protocolStateFunc(v interface{}) string {
	switch v := v.(type) {
	case string:
		p := protocolForValue(v)
		return p
	default:
		log.Printf("[WARN] Non String value given for Protocol: %#v", v)
		return ""
	}
}

// protocolForValue converts a valid Internet Protocol number into it's name
// representation. If a name is given, it validates that it's a proper protocol
// name. Names/numbers are as defined at
// https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml
func protocolForValue(v string) string {
	// special case -1
	protocol := strings.ToLower(v)
	if protocol == "-1" || protocol == "all" {
		return "-1"
	}
	// if it's a name like tcp, return that
	if _, ok := securityGroupProtocolIntegers[protocol]; ok {
		return protocol
	}
	// convert to int, look for that value
	p, err := strconv.Atoi(protocol)
	if err != nil {
		// we were unable to convert to int, suggesting a string name, but it wasn't
		// found above
		log.Printf("[WARN] Unable to determine valid protocol: %s", err)
		return protocol
	}

	for k, v := range securityGroupProtocolIntegers {
		if p == v {
			// guard against protocolIntegers sometime in the future not having lower
			// case ids in the map
			return strings.ToLower(k)
		}
	}

	// fall through
	log.Printf("[WARN] Unable to determine valid protocol: no matching protocols found")
	return protocol
}

// a map of protocol names and their codes, defined at
// https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml,
// documented to be supported by AWS Security Groups
// http://docs.aws.amazon.com/fr_fr/AWSEC2/latest/APIReference/API_IpPermission.html
// Similar to protocolIntegers() used by Network ACLs, but explicitly only
// supports "tcp", "udp", "icmp", "icmpv6", and "all"
var securityGroupProtocolIntegers = map[string]int{
	"icmpv6": 58,
	"udp":    17,
	"tcp":    6,
	"icmp":   1,
	"all":    -1,
}

func initSecurityGroupRule(ruleMap map[string]map[string]interface{}, perm awstypes.IpPermission, desc string) map[string]interface{} {
	var fromPort, toPort int32
	if v := perm.FromPort; v != nil {
		fromPort = aws.ToInt32(v)
	}
	if v := perm.ToPort; v != nil {
		toPort = aws.ToInt32(v)
	}
	k := fmt.Sprintf("%s-%d-%d-%s", aws.ToString(perm.IpProtocol), fromPort, toPort, desc)
	rule, ok := ruleMap[k]
	if !ok {
		rule = make(map[string]interface{})
		ruleMap[k] = rule
	}
	rule[names.AttrProtocol] = aws.ToString(perm.IpProtocol)
	rule["from_port"] = fromPort
	rule["to_port"] = toPort
	if desc != "" {
		rule[names.AttrDescription] = desc
	}

	return rule
}
