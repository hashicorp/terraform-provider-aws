// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_config", name="Config")
func resourceConfig() *schema.Resource {
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AutodefinedReverseFlag](),
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

func resourceConfigCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	autodefinedReverseFlag := awstypes.AutodefinedReverseFlag(d.Get("autodefined_reverse_flag").(string))
	input := &route53resolver.UpdateResolverConfigInput{
		AutodefinedReverseFlag: autodefinedReverseFlag,
		ResourceId:             aws.String(d.Get(names.AttrResourceID).(string)),
	}

	output, err := conn.UpdateResolverConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Config: %s", err)
	}

	d.SetId(aws.ToString(output.ResolverConfig.Id))

	if _, err = waitAutodefinedReverseUpdated(ctx, conn, d.Id(), autodefinedReverseFlag); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Config (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceConfigRead(ctx, d, meta)...)
}

func resourceConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	resolverConfig, err := findResolverConfigByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Config (%s): %s", d.Id(), err)
	}

	var autodefinedReverseFlag awstypes.AutodefinedReverseFlag
	if resolverConfig.AutodefinedReverse == awstypes.ResolverAutodefinedReverseStatusEnabled {
		autodefinedReverseFlag = awstypes.AutodefinedReverseFlagEnable
	} else {
		autodefinedReverseFlag = awstypes.AutodefinedReverseFlagDisable
	}
	d.Set("autodefined_reverse_flag", autodefinedReverseFlag)
	d.Set(names.AttrOwnerID, resolverConfig.OwnerId)
	d.Set(names.AttrResourceID, resolverConfig.ResourceId)

	return diags
}

func resourceConfigUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	autodefinedReverseFlag := awstypes.AutodefinedReverseFlag(d.Get("autodefined_reverse_flag").(string))
	input := &route53resolver.UpdateResolverConfigInput{
		AutodefinedReverseFlag: autodefinedReverseFlag,
		ResourceId:             aws.String(d.Get(names.AttrResourceID).(string)),
	}

	_, err := conn.UpdateResolverConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Resolver Config: %s", err)
	}

	if _, err = waitAutodefinedReverseUpdated(ctx, conn, d.Id(), autodefinedReverseFlag); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Config (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceConfigRead(ctx, d, meta)...)
}

func findResolverConfigByID(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.ResolverConfig, error) {
	input := &route53resolver.ListResolverConfigsInput{}

	// GetResolverConfig does not support query by ID.
	pages := route53resolver.NewListResolverConfigsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ResolverConfigs {
			if aws.ToString(v.Id) == id {
				return &v, nil
			}
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func statusAutodefinedReverse(ctx context.Context, conn *route53resolver.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findResolverConfigByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.AutodefinedReverse), nil
	}
}

const (
	autodefinedReverseUpdatedTimeout = 10 * time.Minute
)

func waitAutodefinedReverseUpdated(ctx context.Context, conn *route53resolver.Client, id string, autodefinedReverseFlag awstypes.AutodefinedReverseFlag) (*awstypes.ResolverConfig, error) {
	if autodefinedReverseFlag == awstypes.AutodefinedReverseFlagDisable {
		return waitAutodefinedReverseDisabled(ctx, conn, id)
	} else {
		return waitAutodefinedReverseEnabled(ctx, conn, id)
	}
}

func waitAutodefinedReverseEnabled(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.ResolverConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResolverAutodefinedReverseStatusEnabling),
		Target:  enum.Slice(awstypes.ResolverAutodefinedReverseStatusEnabled),
		Refresh: statusAutodefinedReverse(ctx, conn, id),
		Timeout: autodefinedReverseUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResolverConfig); ok {
		return output, err
	}

	return nil, err
}

func waitAutodefinedReverseDisabled(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.ResolverConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResolverAutodefinedReverseStatusDisabling),
		Target:  enum.Slice(awstypes.ResolverAutodefinedReverseStatusDisabled),
		Refresh: statusAutodefinedReverse(ctx, conn, id),
		Timeout: autodefinedReverseUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResolverConfig); ok {
		return output, err
	}

	return nil, err
}
