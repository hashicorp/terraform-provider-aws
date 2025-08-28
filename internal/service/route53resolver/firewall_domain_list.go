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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_firewall_domain_list", name="Firewall Domain List")
// @Tags(identifierAttribute="arn")
func resourceFirewallDomainList() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFirewallDomainListCreate,
		ReadWithoutTimeout:   resourceFirewallDomainListRead,
		UpdateWithoutTimeout: resourceFirewallDomainListUpdate,
		DeleteWithoutTimeout: resourceFirewallDomainListDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domains": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validResolverName,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceFirewallDomainListCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &route53resolver.CreateFirewallDomainListInput{
		CreatorRequestId: aws.String(id.PrefixedUniqueId("tf-r53-resolver-firewall-domain-list-")),
		Name:             aws.String(name),
		Tags:             getTagsIn(ctx),
	}

	output, err := conn.CreateFirewallDomainList(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Firewall Domain List (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.FirewallDomainList.Id))

	if v, ok := d.GetOk("domains"); ok && v.(*schema.Set).Len() > 0 {
		_, err := conn.UpdateFirewallDomains(ctx, &route53resolver.UpdateFirewallDomainsInput{
			FirewallDomainListId: aws.String(d.Id()),
			Domains:              flex.ExpandStringValueSet(v.(*schema.Set)),
			Operation:            awstypes.FirewallDomainUpdateOperationAdd,
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Resolver Firewall Domain List (%s) domains: %s", d.Id(), err)
		}

		if _, err = waitFirewallDomainListUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Firewall Domain List (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFirewallDomainListRead(ctx, d, meta)...)
}

func resourceFirewallDomainListRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	firewallDomainList, err := findFirewallDomainListByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Firewall Domain List (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Firewall Domain List (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, firewallDomainList.Arn)
	d.Set(names.AttrName, firewallDomainList.Name)

	input := &route53resolver.ListFirewallDomainsInput{
		FirewallDomainListId: aws.String(d.Id()),
	}
	var output []string

	pages := route53resolver.NewListFirewallDomainsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Route53 Resolver Firewall Domain List (%s) domains: %s", d.Id(), err)
		}

		output = append(output, page.Domains...)
	}

	d.Set("domains", output)

	return diags
}

func resourceFirewallDomainListUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	if d.HasChange("domains") {
		o, n := d.GetChange("domains")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		domains := ns
		operation := awstypes.FirewallDomainUpdateOperationReplace

		if domains.Len() == 0 {
			domains = os
			operation = awstypes.FirewallDomainUpdateOperationRemove
		}

		_, err := conn.UpdateFirewallDomains(ctx, &route53resolver.UpdateFirewallDomainsInput{
			FirewallDomainListId: aws.String(d.Id()),
			Domains:              flex.ExpandStringValueSet(domains),
			Operation:            operation,
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Resolver Firewall Domain List (%s) domains: %s", d.Id(), err)
		}

		if _, err = waitFirewallDomainListUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Firewall Domain List (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFirewallDomainListRead(ctx, d, meta)...)
}

func resourceFirewallDomainListDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	log.Printf("[DEBUG] Deleting Route53 Resolver Firewall Domain List: %s", d.Id())
	_, err := conn.DeleteFirewallDomainList(ctx, &route53resolver.DeleteFirewallDomainListInput{
		FirewallDomainListId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Resolver Firewall Domain List (%s): %s", d.Id(), err)
	}

	if _, err = waitFirewallDomainListDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Firewall Domain List (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findFirewallDomainListByID(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.FirewallDomainList, error) {
	input := &route53resolver.GetFirewallDomainListInput{
		FirewallDomainListId: aws.String(id),
	}

	output, err := conn.GetFirewallDomainList(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FirewallDomainList == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.FirewallDomainList, nil
}

func statusFirewallDomainList(ctx context.Context, conn *route53resolver.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findFirewallDomainListByID(ctx, conn, id)

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
	firewallDomainListUpdatedTimeout = 5 * time.Minute
	firewallDomainListDeletedTimeout = 5 * time.Minute
)

func waitFirewallDomainListUpdated(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.FirewallDomainList, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FirewallDomainListStatusUpdating, awstypes.FirewallDomainListStatusImporting),
		Target:  enum.Slice(awstypes.FirewallDomainListStatusComplete, awstypes.FirewallDomainListStatusCompleteImportFailed),
		Refresh: statusFirewallDomainList(ctx, conn, id),
		Timeout: firewallDomainListUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FirewallDomainList); ok {
		if status := output.Status; status == awstypes.FirewallDomainListStatusCompleteImportFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitFirewallDomainListDeleted(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.FirewallDomainList, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FirewallDomainListStatusDeleting),
		Target:  []string{},
		Refresh: statusFirewallDomainList(ctx, conn, id),
		Timeout: firewallDomainListDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FirewallDomainList); ok {
		return output, err
	}

	return nil, err
}
