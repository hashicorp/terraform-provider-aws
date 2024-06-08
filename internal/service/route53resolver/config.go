// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_config")
func ResourceConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigCreate,
		ReadWithoutTimeout:   resourceConfigRead,
		UpdateWithoutTimeout: resourceConfigUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"autodefined_reverse_flag": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(autodefinedReverseFlag_Values(), false),
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	autodefinedReverseFlag := d.Get("autodefined_reverse_flag").(string)
	input := &route53resolver.UpdateResolverConfigInput{
		AutodefinedReverseFlag: aws.String(autodefinedReverseFlag),
		ResourceId:             aws.String(d.Get(names.AttrResourceID).(string)),
	}

	output, err := conn.UpdateResolverConfigWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Config: %s", err)
	}

	d.SetId(aws.StringValue(output.ResolverConfig.Id))

	if _, err = waitAutodefinedReverseUpdated(ctx, conn, d.Id(), autodefinedReverseFlag); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Config (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceConfigRead(ctx, d, meta)...)
}

func resourceConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	resolverConfig, err := FindResolverConfigByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Config (%s): %s", d.Id(), err)
	}

	var autodefinedReverseFlag string
	if aws.StringValue(resolverConfig.AutodefinedReverse) == route53resolver.ResolverAutodefinedReverseStatusEnabled {
		autodefinedReverseFlag = autodefinedReverseFlagEnable
	} else {
		autodefinedReverseFlag = autodefinedReverseFlagDisable
	}
	d.Set("autodefined_reverse_flag", autodefinedReverseFlag)
	d.Set(names.AttrOwnerID, resolverConfig.OwnerId)
	d.Set(names.AttrResourceID, resolverConfig.ResourceId)

	return diags
}

func resourceConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	autodefinedReverseFlag := d.Get("autodefined_reverse_flag").(string)
	input := &route53resolver.UpdateResolverConfigInput{
		AutodefinedReverseFlag: aws.String(autodefinedReverseFlag),
		ResourceId:             aws.String(d.Get(names.AttrResourceID).(string)),
	}

	_, err := conn.UpdateResolverConfigWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Resolver Config: %s", err)
	}

	if _, err = waitAutodefinedReverseUpdated(ctx, conn, d.Id(), autodefinedReverseFlag); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Config (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceConfigRead(ctx, d, meta)...)
}

func FindResolverConfigByID(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverConfig, error) {
	input := &route53resolver.ListResolverConfigsInput{}
	var output *route53resolver.ResolverConfig

	// GetResolverConfig does not support query by ID.
	err := conn.ListResolverConfigsPagesWithContext(ctx, input, func(page *route53resolver.ListResolverConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResolverConfigs {
			if aws.StringValue(v.Id) == id {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &retry.NotFoundError{LastRequest: input}
	}

	return output, nil
}

const (
	// https://docs.aws.amazon.com/Route53/latest/APIReference/API_route53resolver_UpdateResolverConfig.html#API_route53resolver_UpdateResolverConfig_RequestSyntax
	autodefinedReverseFlagDisable = "DISABLE"
	autodefinedReverseFlagEnable  = "ENABLE"
)

func autodefinedReverseFlag_Values() []string {
	return []string{
		autodefinedReverseFlagDisable,
		autodefinedReverseFlagEnable,
	}
}

func statusAutodefinedReverse(ctx context.Context, conn *route53resolver.Route53Resolver, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindResolverConfigByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.AutodefinedReverse), nil
	}
}

const (
	autodefinedReverseUpdatedTimeout = 10 * time.Minute
)

func waitAutodefinedReverseUpdated(ctx context.Context, conn *route53resolver.Route53Resolver, id, autodefinedReverseFlag string) (*route53resolver.ResolverConfig, error) {
	if autodefinedReverseFlag == autodefinedReverseFlagDisable {
		return waitAutodefinedReverseDisabled(ctx, conn, id)
	} else {
		return waitAutodefinedReverseEnabled(ctx, conn, id)
	}
}

func waitAutodefinedReverseEnabled(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{route53resolver.ResolverAutodefinedReverseStatusEnabling},
		Target:  []string{route53resolver.ResolverAutodefinedReverseStatusEnabled},
		Refresh: statusAutodefinedReverse(ctx, conn, id),
		Timeout: autodefinedReverseUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53resolver.ResolverConfig); ok {
		return output, err
	}

	return nil, err
}

func waitAutodefinedReverseDisabled(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{route53resolver.ResolverAutodefinedReverseStatusDisabling},
		Target:  []string{route53resolver.ResolverAutodefinedReverseStatusDisabled},
		Refresh: statusAutodefinedReverse(ctx, conn, id),
		Timeout: autodefinedReverseUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53resolver.ResolverConfig); ok {
		return output, err
	}

	return nil, err
}
