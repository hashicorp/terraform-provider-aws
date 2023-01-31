package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDefaultSecurityGroup() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceDefaultSecurityGroupCreate,
		ReadWithoutTimeout:   resourceSecurityGroupRead,
		UpdateWithoutTimeout: resourceSecurityGroupUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceDefaultSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()
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

	sg, err := FindSecurityGroup(ctx, conn, input)

	if err != nil {
		return diag.Errorf("reading Default Security Group: %s", err)
	}

	d.SetId(aws.StringValue(sg.GroupId))

	oTagsAll := KeyValueTags(sg.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	nTagsAll := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if !nTagsAll.Equal(oTagsAll) {
		if err := UpdateTags(ctx, conn, d.Id(), oTagsAll.Map(), nTagsAll.Map()); err != nil {
			return diag.Errorf("updating Default Security Group (%s) tags: %s", d.Id(), err)
		}
	}

	if err := forceRevokeSecurityGroupRules(ctx, conn, d.Id(), false); err != nil {
		return diag.FromErr(err)
	}

	return resourceSecurityGroupUpdate(ctx, d, meta)
}
