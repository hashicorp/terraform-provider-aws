// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_firewall_rule_group_association", name="Firewall Rule Group Association")
// @Tags(identifierAttribute="arn")
func resourceFirewallRuleGroupAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFirewallRuleGroupAssociationCreate,
		ReadWithoutTimeout:   resourceFirewallRuleGroupAssociationRead,
		UpdateWithoutTimeout: resourceFirewallRuleGroupAssociationUpdate,
		DeleteWithoutTimeout: resourceFirewallRuleGroupAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"firewall_rule_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mutation_protection": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.MutationProtectionStatus](),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validResolverName,
			},
			names.AttrPriority: {
				Type:     schema.TypeInt,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFirewallRuleGroupAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &route53resolver.AssociateFirewallRuleGroupInput{
		CreatorRequestId:    aws.String(id.PrefixedUniqueId("tf-r53-rslvr-frgassoc-")),
		FirewallRuleGroupId: aws.String(d.Get("firewall_rule_group_id").(string)),
		Name:                aws.String(name),
		Priority:            aws.Int32(int32(d.Get(names.AttrPriority).(int))),
		Tags:                getTagsIn(ctx),
		VpcId:               aws.String(d.Get(names.AttrVPCID).(string)),
	}

	if v, ok := d.GetOk("mutation_protection"); ok {
		input.MutationProtection = awstypes.MutationProtectionStatus(v.(string))
	}

	output, err := conn.AssociateFirewallRuleGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Firewall Rule Group Association (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.FirewallRuleGroupAssociation.Id))

	if _, err := waitFirewallRuleGroupAssociationCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Firewall Rule Group Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceFirewallRuleGroupAssociationRead(ctx, d, meta)...)
}

func resourceFirewallRuleGroupAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	ruleGroupAssociation, err := findFirewallRuleGroupAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Firewall Rule Group Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Firewall Rule Group Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, ruleGroupAssociation.Arn)
	d.Set(names.AttrName, ruleGroupAssociation.Name)
	d.Set("firewall_rule_group_id", ruleGroupAssociation.FirewallRuleGroupId)
	d.Set("mutation_protection", ruleGroupAssociation.MutationProtection)
	d.Set(names.AttrPriority, ruleGroupAssociation.Priority)
	d.Set(names.AttrVPCID, ruleGroupAssociation.VpcId)

	return diags
}

func resourceFirewallRuleGroupAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	if d.HasChanges(names.AttrName, "mutation_protection", names.AttrPriority) {
		input := &route53resolver.UpdateFirewallRuleGroupAssociationInput{
			FirewallRuleGroupAssociationId: aws.String(d.Id()),
			Name:                           aws.String(d.Get(names.AttrName).(string)),
			Priority:                       aws.Int32(int32(d.Get(names.AttrPriority).(int))),
		}

		if v, ok := d.GetOk("mutation_protection"); ok {
			input.MutationProtection = awstypes.MutationProtectionStatus(v.(string))
		}

		_, err := conn.UpdateFirewallRuleGroupAssociation(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Resolver Firewall Rule Group Association (%s): %s", d.Id(), err)
		}

		if _, err := waitFirewallRuleGroupAssociationUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Firewall Rule Group Association (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFirewallRuleGroupAssociationRead(ctx, d, meta)...)
}

func resourceFirewallRuleGroupAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	log.Printf("[DEBUG] Deleting Route53 Resolver Firewall Rule Group Association: %s", d.Id())
	_, err := conn.DisassociateFirewallRuleGroup(ctx, &route53resolver.DisassociateFirewallRuleGroupInput{
		FirewallRuleGroupAssociationId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Resolver Firewall Rule Group Association (%s): %s", d.Id(), err)
	}

	if _, err := waitFirewallRuleGroupAssociationDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Firewall Rule Group Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findFirewallRuleGroupAssociationByID(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.FirewallRuleGroupAssociation, error) {
	input := &route53resolver.GetFirewallRuleGroupAssociationInput{
		FirewallRuleGroupAssociationId: aws.String(id),
	}

	output, err := conn.GetFirewallRuleGroupAssociation(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FirewallRuleGroupAssociation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.FirewallRuleGroupAssociation, nil
}

func statusFirewallRuleGroupAssociation(ctx context.Context, conn *route53resolver.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findFirewallRuleGroupAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

const (
	firewallRuleGroupAssociationCreatedTimeout = 5 * time.Minute
	firewallRuleGroupAssociationUpdatedTimeout = 5 * time.Minute
	firewallRuleGroupAssociationDeletedTimeout = 5 * time.Minute
)

func waitFirewallRuleGroupAssociationCreated(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.FirewallRuleGroupAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FirewallRuleGroupAssociationStatusUpdating),
		Target:  enum.Slice(awstypes.FirewallRuleGroupAssociationStatusComplete),
		Refresh: statusFirewallRuleGroupAssociation(ctx, conn, id),
		Timeout: firewallRuleGroupAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FirewallRuleGroupAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitFirewallRuleGroupAssociationUpdated(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.FirewallRuleGroupAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FirewallRuleGroupAssociationStatusUpdating),
		Target:  enum.Slice(awstypes.FirewallRuleGroupAssociationStatusComplete),
		Refresh: statusFirewallRuleGroupAssociation(ctx, conn, id),
		Timeout: firewallRuleGroupAssociationUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FirewallRuleGroupAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitFirewallRuleGroupAssociationDeleted(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.FirewallRuleGroupAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FirewallRuleGroupAssociationStatusDeleting),
		Target:  []string{},
		Refresh: statusFirewallRuleGroupAssociation(ctx, conn, id),
		Timeout: firewallRuleGroupAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FirewallRuleGroupAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}
