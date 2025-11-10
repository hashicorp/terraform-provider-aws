// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_firewall_rule_group", name="Firewall Rule Group")
// @Tags(identifierAttribute="arn")
func resourceFirewallRuleGroup() *schema.Resource {
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
	}
}

func resourceFirewallRuleGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &route53resolver.CreateFirewallRuleGroupInput{
		CreatorRequestId: aws.String(id.PrefixedUniqueId("tf-r53-resolver-firewall-rule-group-")),
		Name:             aws.String(name),
		Tags:             getTagsIn(ctx),
	}

	output, err := conn.CreateFirewallRuleGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Firewall Rule Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.FirewallRuleGroup.Id))

	return append(diags, resourceFirewallRuleGroupRead(ctx, d, meta)...)
}

func resourceFirewallRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	ruleGroup, err := findFirewallRuleGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Firewall Rule Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Firewall Rule Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, ruleGroup.Arn)
	d.Set(names.AttrName, ruleGroup.Name)
	d.Set(names.AttrOwnerID, ruleGroup.OwnerId)
	d.Set("share_status", ruleGroup.ShareStatus)

	return diags
}

func resourceFirewallRuleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceFirewallRuleGroupRead(ctx, d, meta)
}

func resourceFirewallRuleGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	log.Printf("[DEBUG] Deleting Route53 Resolver Firewall Rule Group: %s", d.Id())
	_, err := conn.DeleteFirewallRuleGroup(ctx, &route53resolver.DeleteFirewallRuleGroupInput{
		FirewallRuleGroupId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Resolver Firewall Rule Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findFirewallRuleGroupByID(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.FirewallRuleGroup, error) {
	input := &route53resolver.GetFirewallRuleGroupInput{
		FirewallRuleGroupId: aws.String(id),
	}

	output, err := conn.GetFirewallRuleGroup(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
