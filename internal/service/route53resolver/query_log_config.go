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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_query_log_config", name="Query Log Config")
// @Tags(identifierAttribute="arn")
func resourceQueryLogConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQueryLogConfigCreate,
		ReadWithoutTimeout:   resourceQueryLogConfigRead,
		UpdateWithoutTimeout: resourceQueryLogConfigUpdate,
		DeleteWithoutTimeout: resourceQueryLogConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDestinationARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
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

func resourceQueryLogConfigCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &route53resolver.CreateResolverQueryLogConfigInput{
		CreatorRequestId: aws.String(id.PrefixedUniqueId("tf-r53-resolver-query-log-config-")),
		DestinationArn:   aws.String(d.Get(names.AttrDestinationARN).(string)),
		Name:             aws.String(name),
		Tags:             getTagsIn(ctx),
	}

	output, err := conn.CreateResolverQueryLogConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Query Log Config (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ResolverQueryLogConfig.Id))

	if _, err := waitQueryLogConfigCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Query Log Config (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceQueryLogConfigRead(ctx, d, meta)...)
}

func resourceQueryLogConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	queryLogConfig, err := findResolverQueryLogConfigByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Query Log Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Query Log Config (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, queryLogConfig.Arn)
	d.Set(names.AttrDestinationARN, queryLogConfig.DestinationArn)
	d.Set(names.AttrName, queryLogConfig.Name)
	d.Set(names.AttrOwnerID, queryLogConfig.OwnerId)
	d.Set("share_status", queryLogConfig.ShareStatus)

	return diags
}

func resourceQueryLogConfigUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceQueryLogConfigRead(ctx, d, meta)
}

func resourceQueryLogConfigDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	log.Printf("[DEBUG] Deleting Route53 Resolver Query Log Config: %s", d.Id())
	_, err := conn.DeleteResolverQueryLogConfig(ctx, &route53resolver.DeleteResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Resolver Query Log Config (%s): %s", d.Id(), err)
	}

	if _, err := waitQueryLogConfigDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Query Log Config (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findResolverQueryLogConfigByID(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.ResolverQueryLogConfig, error) {
	input := &route53resolver.GetResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: aws.String(id),
	}

	output, err := conn.GetResolverQueryLogConfig(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResolverQueryLogConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResolverQueryLogConfig, nil
}

func statusQueryLogConfig(ctx context.Context, conn *route53resolver.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findResolverQueryLogConfigByID(ctx, conn, id)

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
	queryLogConfigCreatedTimeout = 5 * time.Minute
	queryLogConfigDeletedTimeout = 5 * time.Minute
)

func waitQueryLogConfigCreated(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.ResolverQueryLogConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResolverQueryLogConfigStatusCreating),
		Target:  enum.Slice(awstypes.ResolverQueryLogConfigStatusCreated),
		Refresh: statusQueryLogConfig(ctx, conn, id),
		Timeout: queryLogConfigCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResolverQueryLogConfig); ok {
		return output, err
	}

	return nil, err
}

func waitQueryLogConfigDeleted(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.ResolverQueryLogConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResolverQueryLogConfigStatusDeleting),
		Target:  []string{},
		Refresh: statusQueryLogConfig(ctx, conn, id),
		Timeout: queryLogConfigDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResolverQueryLogConfig); ok {
		return output, err
	}

	return nil, err
}
