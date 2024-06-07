// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_query_log_config_association")
func ResourceQueryLogConfigAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQueryLogConfigAssociationCreate,
		ReadWithoutTimeout:   resourceQueryLogConfigAssociationRead,
		DeleteWithoutTimeout: resourceQueryLogConfigAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"resolver_query_log_config_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceQueryLogConfigAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	input := &route53resolver.AssociateResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: aws.String(d.Get("resolver_query_log_config_id").(string)),
		ResourceId:               aws.String(d.Get(names.AttrResourceID).(string)),
	}

	output, err := conn.AssociateResolverQueryLogConfigWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Query Log Config Association: %s", err)
	}

	d.SetId(aws.StringValue(output.ResolverQueryLogConfigAssociation.Id))

	if _, err := waitQueryLogConfigAssociationCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Query Log Config Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceQueryLogConfigAssociationRead(ctx, d, meta)...)
}

func resourceQueryLogConfigAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	queryLogConfigAssociation, err := FindResolverQueryLogConfigAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Query Log Config Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Query Log Config Association (%s): %s", d.Id(), err)
	}

	d.Set("resolver_query_log_config_id", queryLogConfigAssociation.ResolverQueryLogConfigId)
	d.Set(names.AttrResourceID, queryLogConfigAssociation.ResourceId)

	return diags
}

func resourceQueryLogConfigAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	log.Printf("[DEBUG] Deleting Route53 Resolver Query Log Config Association: %s", d.Id())
	_, err := conn.DisassociateResolverQueryLogConfigWithContext(ctx, &route53resolver.DisassociateResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: aws.String(d.Get("resolver_query_log_config_id").(string)),
		ResourceId:               aws.String(d.Get(names.AttrResourceID).(string)),
	})

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Resolver Query Log Config Association (%s): %s", d.Id(), err)
	}

	if _, err := waitQueryLogConfigAssociationDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Query Log Config Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindResolverQueryLogConfigAssociationByID(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	input := &route53resolver.GetResolverQueryLogConfigAssociationInput{
		ResolverQueryLogConfigAssociationId: aws.String(id),
	}

	output, err := conn.GetResolverQueryLogConfigAssociationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResolverQueryLogConfigAssociation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResolverQueryLogConfigAssociation, nil
}

func statusQueryLogConfigAssociation(ctx context.Context, conn *route53resolver.Route53Resolver, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindResolverQueryLogConfigAssociationByID(ctx, conn, id)

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
	queryLogConfigAssociationCreatedTimeout = 5 * time.Minute
	queryLogConfigAssociationDeletedTimeout = 5 * time.Minute
)

func waitQueryLogConfigAssociationCreated(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigAssociationStatusCreating},
		Target:  []string{route53resolver.ResolverQueryLogConfigAssociationStatusActive},
		Refresh: statusQueryLogConfigAssociation(ctx, conn, id),
		Timeout: queryLogConfigAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53resolver.ResolverQueryLogConfigAssociation); ok {
		if status := aws.StringValue(output.Status); status == route53resolver.ResolverQueryLogConfigAssociationStatusFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.Error), aws.StringValue(output.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitQueryLogConfigAssociationDeleted(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigAssociationStatusDeleting},
		Target:  []string{},
		Refresh: statusQueryLogConfigAssociation(ctx, conn, id),
		Timeout: queryLogConfigAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53resolver.ResolverQueryLogConfigAssociation); ok {
		if status := aws.StringValue(output.Status); status == route53resolver.ResolverQueryLogConfigAssociationStatusFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.Error), aws.StringValue(output.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}
