package ec2

import (
	"bytes"
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDefaultSecurityGroup() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceDefaultSecurityGroupCreate,
		Read:   resourceDefaultSecurityGroupRead,
		Update: resourceDefaultSecurityGroupUpdate,
		Delete: resourceDefaultSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,
		MigrateState:  DefaultSecurityGroupMigrateState,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"ingress": {
				Type:       schema.TypeSet,
				Optional:   true,
				Computed:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr_blocks": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidCIDRNetworkAddress,
							},
						},
						"description": {
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
								ValidateFunc: verify.ValidCIDRNetworkAddress,
							},
						},
						"prefix_list_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"protocol": {
							Type:      schema.TypeString,
							Required:  true,
							StateFunc: ProtocolStateFunc,
						},
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
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
				},
				Set: resourceDefaultSecurityGroupRuleHash,
			},
			"egress": {
				Type:       schema.TypeSet,
				Optional:   true,
				Computed:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr_blocks": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidCIDRNetworkAddress,
							},
						},
						"description": {
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
								ValidateFunc: verify.ValidCIDRNetworkAddress,
							},
						},
						"prefix_list_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"protocol": {
							Type:      schema.TypeString,
							Required:  true,
							StateFunc: ProtocolStateFunc,
						},
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
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
				},
				Set: resourceDefaultSecurityGroupRuleHash,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			// This is not implemented. Added to prevent breaking changes.
			"revoke_rules_on_delete": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDefaultSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	securityGroupOpts := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-name"),
				Values: aws.StringSlice([]string{DefaultSecurityGroupName}),
			},
		},
	}

	var vpcID string
	if v, ok := d.GetOk("vpc_id"); ok {
		vpcID = v.(string)
		securityGroupOpts.Filters = append(securityGroupOpts.Filters, &ec2.Filter{
			Name:   aws.String("vpc-id"),
			Values: aws.StringSlice([]string{vpcID}),
		})
	}

	var err error
	log.Printf("[DEBUG] Commandeer Default Security Group: %s", securityGroupOpts)
	resp, err := conn.DescribeSecurityGroups(securityGroupOpts)
	if err != nil {
		return fmt.Errorf("Error creating Default Security Group: %w", err)
	}

	var g *ec2.SecurityGroup
	if vpcID != "" {
		// if vpcId contains a value, then we expect just a single Security Group
		// returned, as default is a protected name for each VPC, and for each
		// Region on EC2 Classic
		if len(resp.SecurityGroups) != 1 {
			return fmt.Errorf("Error finding default security group; found (%d) groups: %s", len(resp.SecurityGroups), resp)
		}
		g = resp.SecurityGroups[0]
	} else {
		// we need to filter through any returned security groups for the group
		// named "default", and does not belong to a VPC
		for _, sg := range resp.SecurityGroups {
			if sg.VpcId == nil && aws.StringValue(sg.GroupName) == DefaultSecurityGroupName {
				g = sg
				break
			}
		}
	}

	if g == nil {
		return fmt.Errorf("Error finding default security group: no matching group found")
	}

	d.SetId(aws.StringValue(g.GroupId))

	log.Printf("[INFO] Default Security Group ID: %s", d.Id())

	if len(tags) > 0 {
		if err := CreateTags(conn, d.Id(), tags); err != nil {
			return fmt.Errorf("error adding EC2 Default Security Group (%s) tags: %w", d.Id(), err)
		}
	}

	if err := revokeDefaultSecurityGroupRules(meta, g); err != nil {
		return err
	}

	return resourceDefaultSecurityGroupUpdate(d, meta)
}

func resourceDefaultSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	group, err := FindSecurityGroupByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return err
	}

	remoteIngressRules := SecurityGroupIPPermGather(d.Id(), group.IpPermissions, group.OwnerId)
	remoteEgressRules := SecurityGroupIPPermGather(d.Id(), group.IpPermissionsEgress, group.OwnerId)

	localIngressRules := d.Get("ingress").(*schema.Set).List()
	localEgressRules := d.Get("egress").(*schema.Set).List()

	// Loop through the local state of rules, doing a match against the remote
	// ruleSet we built above.
	ingressRules := MatchRules("ingress", localIngressRules, remoteIngressRules)
	egressRules := MatchRules("egress", localEgressRules, remoteEgressRules)

	sgArn := arn.ARN{
		AccountID: aws.StringValue(group.OwnerId),
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("security-group/%s", aws.StringValue(group.GroupId)),
		Service:   ec2.ServiceName,
	}

	d.Set("arn", sgArn.String())
	d.Set("description", group.Description)
	d.Set("name", group.GroupName)
	d.Set("owner_id", group.OwnerId)
	d.Set("vpc_id", group.VpcId)

	if err := d.Set("ingress", ingressRules); err != nil {
		return fmt.Errorf("error setting Ingress rule set for (%s): %w", d.Id(), err)
	}

	if err := d.Set("egress", egressRules); err != nil {
		return fmt.Errorf("error setting Egress rule set for (%s): %w", d.Id(), err)
	}

	tags := KeyValueTags(group.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceDefaultSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	group, err := FindSecurityGroupByID(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error updating Default Security Group (%s): %w", d.Id(), err)
	}

	err = resourceSecurityGroupUpdateRules(d, "ingress", meta, group)
	if err != nil {
		return fmt.Errorf("error updating Default Security Group (%s): %w", d.Id(), err)
	}

	if d.Get("vpc_id") != nil {
		err = resourceSecurityGroupUpdateRules(d, "egress", meta, group)
		if err != nil {
			return fmt.Errorf("error updating Default Security Group (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") && !d.IsNewResource() {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Default Security Group (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceDefaultSecurityGroupRead(d, meta)
}

func resourceDefaultSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy Default Security Group. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}

func revokeDefaultSecurityGroupRules(meta interface{}, g *ec2.SecurityGroup) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	groupID := aws.StringValue(g.GroupId)
	log.Printf("[WARN] Removing all ingress and egress rules found on Default Security Group (%s)", groupID)
	if len(g.IpPermissionsEgress) > 0 {
		req := &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       g.GroupId,
			IpPermissions: g.IpPermissionsEgress,
		}

		log.Printf("[DEBUG] Revoking default egress rules for Default Security Group for %s", groupID)
		if _, err := conn.RevokeSecurityGroupEgress(req); err != nil {
			return fmt.Errorf("error revoking default egress rules for Default Security Group (%s): %w", groupID, err)
		}
	}
	if len(g.IpPermissions) > 0 {
		// a limitation in EC2 Classic is that a call to RevokeSecurityGroupIngress
		// cannot contain both the GroupName and the GroupId
		for _, p := range g.IpPermissions {
			for _, uigp := range p.UserIdGroupPairs {
				if uigp.GroupId != nil && uigp.GroupName != nil {
					uigp.GroupName = nil
				}
			}
		}
		req := &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       g.GroupId,
			IpPermissions: g.IpPermissions,
		}

		log.Printf("[DEBUG] Revoking default ingress rules for Default Security Group for (%s): %s", groupID, req)
		if _, err := conn.RevokeSecurityGroupIngress(req); err != nil {
			return fmt.Errorf("Error revoking default ingress rules for Default Security Group (%s): %w", groupID, err)
		}
	}

	return nil
}

func resourceDefaultSecurityGroupRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["from_port"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["to_port"].(int)))
	p := ProtocolForValue(m["protocol"].(string))
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
	if v, ok := m["security_groups"]; ok {
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
	if m["description"].(string) != "" {
		buf.WriteString(fmt.Sprintf("%s-", m["description"].(string)))
	}

	return create.StringHashcode(buf.String())
}

func DefaultSecurityGroupMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found AWS Default Security Group state v0; migrating to v1")
		return migrateDefaultSecurityGroupStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateDefaultSecurityGroupStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)

	// set default for revoke_rules_on_delete
	is.Attributes["revoke_rules_on_delete"] = "false"

	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil
}
