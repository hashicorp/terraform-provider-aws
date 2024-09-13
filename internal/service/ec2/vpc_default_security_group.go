// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_default_security_group", name="Security Group")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceDefaultSecurityGroup() *schema.Resource {
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
		MigrateState:  securityGroupMigrateState,

		// Keep in sync with aws_security_group's schema with the following changes:
		//   - description is Computed-only
		//   - name is Computed-only
		//   - name_prefix is Computed-only
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"egress":  securityGroupRuleSetNestedBlock,
			"ingress": securityGroupRuleSetNestedBlock,
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrNamePrefix: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			// Not used.
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

func resourceDefaultSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.semgrep.tags.calling-UpdateTags-in-resource-create
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"group-name": defaultSecurityGroupName,
			},
		),
	}

	if v, ok := d.GetOk(names.AttrVPCID); ok {
		input.Filters = append(input.Filters, newAttributeFilterList(
			map[string]string{
				"vpc-id": v.(string),
			},
		)...)
	} else {
		input.Filters = append(input.Filters, newAttributeFilterList(
			map[string]string{
				names.AttrDescription: "default group",
			},
		)...)
	}

	sg, err := findSecurityGroup(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Default Security Group: %s", err)
	}

	d.SetId(aws.ToString(sg.GroupId))

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	newTags := keyValueTags(ctx, getTagsIn(ctx))
	oldTags := keyValueTags(ctx, sg.Tags).IgnoreSystem(names.EC2).IgnoreConfig(ignoreTagsConfig)

	if !newTags.Equal(oldTags) {
		if err := updateTags(ctx, conn, d.Id(), oldTags, newTags); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Default Security Group (%s) tags: %s", d.Id(), err)
		}
	}

	if err := forceRevokeSecurityGroupRules(ctx, conn, d.Id(), false); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceSecurityGroupUpdate(ctx, d, meta)...)
}
