package ec2

import (
	"fmt"

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
		Update: resourceSecurityGroupUpdate,
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

	input := &ec2.DescribeSecurityGroupsInput{
		Filters: BuildAttributeFilterList(
			map[string]string{
				"group-name": DefaultSecurityGroupName,
			},
		),
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		input.Filters = append(input.Filters, BuildAttributeFilterList(
			map[string]string{
				"vpc-id": v.(string),
			},
		)...)
	} else {
		input.Filters = append(input.Filters, BuildAttributeFilterList(
			map[string]string{
				"description": "default group",
			},
		)...)
	}

	sg, err := FindSecurityGroup(conn, input)

	if err != nil {
		return fmt.Errorf("reading Default Security Group: %w", err)
	}

	d.SetId(aws.StringValue(sg.GroupId))

	oTagsAll := KeyValueTags(sg.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	nTagsAll := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if !nTagsAll.Equal(oTagsAll) {
		if err := UpdateTags(conn, d.Id(), oTagsAll.Map(), nTagsAll.Map()); err != nil {
			return fmt.Errorf("updating Default Security Group (%s) tags: %w", d.Id(), err)
		}
	}

	if err := forceRevokeSecurityGroupRules(conn, d.Id(), false); err != nil {
		return err
	}

	return resourceSecurityGroupUpdate(d, meta)
}
