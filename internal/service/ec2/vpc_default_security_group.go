package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDefaultSecurityGroup() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceDefaultSecurityGroupCreate,
		Read:   resourceSecurityGroupRead,
		Update: resourceDefaultSecurityGroupUpdate,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1, // Keep in sync with aws_security_group's schema version.
		MigrateState:  SecurityGroupMigrateState,

		// Keep in sync with aws_security_group's schema with the following changes:
		//   - description is Computed-only
		//   - name is Computed-only
		//   - name_prefix is Computed-only
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"egress":  securityGroupRuleSetNestedBlock,
			"ingress": securityGroupRuleSetNestedBlock,
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// Not used.
			"revoke_rules_on_delete": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDefaultSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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

	oTagsAll := KeyValueTags(g.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	nTagsAll := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if !nTagsAll.Equal(oTagsAll) {
		if err := UpdateTags(conn, d.Id(), oTagsAll.Map(), nTagsAll.Map()); err != nil {
			return fmt.Errorf("updating Default Security Group (%s) tags: %w", d.Id(), err)
		}
	}

	if err := revokeDefaultSecurityGroupRules(meta, g); err != nil {
		return err
	}

	return resourceDefaultSecurityGroupUpdate(d, meta)
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

	return resourceSecurityGroupRead(d, meta)
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
