// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_firewall_domain_list", name="Firewall Domain List")
// @Tags(identifierAttribute="arn")
func ResourceFirewallDomainList() *schema.Resource {
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFirewallDomainListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &route53resolver.CreateFirewallDomainListInput{
		CreatorRequestId: aws.String(id.PrefixedUniqueId("tf-r53-resolver-firewall-domain-list-")),
		Name:             aws.String(name),
		Tags:             getTagsIn(ctx),
	}

	output, err := conn.CreateFirewallDomainListWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Firewall Domain List (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.FirewallDomainList.Id))

	if v, ok := d.GetOk("domains"); ok && v.(*schema.Set).Len() > 0 {
		_, err := conn.UpdateFirewallDomainsWithContext(ctx, &route53resolver.UpdateFirewallDomainsInput{
			FirewallDomainListId: aws.String(d.Id()),
			Domains:              flex.ExpandStringSet(v.(*schema.Set)),
			Operation:            aws.String(route53resolver.FirewallDomainUpdateOperationAdd),
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

func resourceFirewallDomainListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	firewallDomainList, err := FindFirewallDomainListByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Firewall Domain List (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Firewall Domain List (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(firewallDomainList.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrName, firewallDomainList.Name)

	input := &route53resolver.ListFirewallDomainsInput{
		FirewallDomainListId: aws.String(d.Id()),
	}
	var output []*string

	err = conn.ListFirewallDomainsPagesWithContext(ctx, input, func(page *route53resolver.ListFirewallDomainsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Domains...)

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Route53 Resolver Firewall Domain List (%s) domains: %s", d.Id(), err)
	}

	d.Set("domains", aws.StringValueSlice(output))

	return diags
}

func resourceFirewallDomainListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

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
		operation := route53resolver.FirewallDomainUpdateOperationReplace

		if domains.Len() == 0 {
			domains = os
			operation = route53resolver.FirewallDomainUpdateOperationRemove
		}

		_, err := conn.UpdateFirewallDomainsWithContext(ctx, &route53resolver.UpdateFirewallDomainsInput{
			FirewallDomainListId: aws.String(d.Id()),
			Domains:              flex.ExpandStringSet(domains),
			Operation:            aws.String(operation),
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

func resourceFirewallDomainListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	log.Printf("[DEBUG] Deleting Route53 Resolver Firewall Domain List: %s", d.Id())
	_, err := conn.DeleteFirewallDomainListWithContext(ctx, &route53resolver.DeleteFirewallDomainListInput{
		FirewallDomainListId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
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

func FindFirewallDomainListByID(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.FirewallDomainList, error) {
	input := &route53resolver.GetFirewallDomainListInput{
		FirewallDomainListId: aws.String(id),
	}

	output, err := conn.GetFirewallDomainListWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
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

func statusFirewallDomainList(ctx context.Context, conn *route53resolver.Route53Resolver, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFirewallDomainListByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

const (
	firewallDomainListUpdatedTimeout = 5 * time.Minute
	firewallDomainListDeletedTimeout = 5 * time.Minute
)

func waitFirewallDomainListUpdated(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.FirewallDomainList, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{route53resolver.FirewallDomainListStatusUpdating, route53resolver.FirewallDomainListStatusImporting},
		Target: []string{route53resolver.FirewallDomainListStatusComplete,
			route53resolver.FirewallDomainListStatusCompleteImportFailed,
		},
		Refresh: statusFirewallDomainList(ctx, conn, id),
		Timeout: firewallDomainListUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53resolver.FirewallDomainList); ok {
		if status := aws.StringValue(output.Status); status == route53resolver.FirewallDomainListStatusCompleteImportFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitFirewallDomainListDeleted(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.FirewallDomainList, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{route53resolver.FirewallDomainListStatusDeleting},
		Target:  []string{},
		Refresh: statusFirewallDomainList(ctx, conn, id),
		Timeout: firewallDomainListDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53resolver.FirewallDomainList); ok {
		return output, err
	}

	return nil, err
}
