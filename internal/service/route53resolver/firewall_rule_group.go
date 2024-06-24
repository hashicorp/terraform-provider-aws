// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_firewall_rule_group", name="Firewall Rule Group")
// @Tags(identifierAttribute="arn")
func ResourceFirewallRuleGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFirewallRuleGroupCreate,
		ReadWithoutTimeout:   resourceFirewallRuleGroupRead,
		UpdateWithoutTimeout: resourceFirewallRuleGroupUpdate,
		DeleteWithoutTimeout: resourceFirewallRuleGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validResolverName,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"share_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFirewallRuleGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &route53resolver.CreateFirewallRuleGroupInput{
		CreatorRequestId: aws.String(id.PrefixedUniqueId("tf-r53-resolver-firewall-rule-group-")),
		Name:             aws.String(name),
		Tags:             getTagsIn(ctx),
	}

	output, err := conn.CreateFirewallRuleGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Firewall Rule Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.FirewallRuleGroup.Id))

	return append(diags, resourceFirewallRuleGroupRead(ctx, d, meta)...)
}

func resourceFirewallRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	ruleGroup, err := FindFirewallRuleGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Firewall Rule Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Firewall Rule Group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(ruleGroup.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrName, ruleGroup.Name)
	d.Set(names.AttrOwnerID, ruleGroup.OwnerId)
	d.Set("share_status", ruleGroup.ShareStatus)

	return diags
}

func resourceFirewallRuleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceFirewallRuleGroupRead(ctx, d, meta)
}

func resourceFirewallRuleGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	log.Printf("[DEBUG] Deleting Route53 Resolver Firewall Rule Group: %s", d.Id())
	_, err := conn.DeleteFirewallRuleGroupWithContext(ctx, &route53resolver.DeleteFirewallRuleGroupInput{
		FirewallRuleGroupId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Resolver Firewall Rule Group (%s): %s", d.Id(), err)
	}

	return diags
}

func FindFirewallRuleGroupByID(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.FirewallRuleGroup, error) {
	input := &route53resolver.GetFirewallRuleGroupInput{
		FirewallRuleGroupId: aws.String(id),
	}

	output, err := conn.GetFirewallRuleGroupWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FirewallRuleGroup == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.FirewallRuleGroup, nil
}
