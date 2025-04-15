// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
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

// @SDKResource("aws_route53_resolver_dnssec_config", name="DNSSEC Config")
func resourceDNSSECConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDNSSECConfigCreate,
		ReadWithoutTimeout:   resourceDNSSECConfigRead,
		DeleteWithoutTimeout: resourceDNSSECConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
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
			"validation_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDNSSECConfigCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	input := &route53resolver.UpdateResolverDnssecConfigInput{
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
		Validation: awstypes.ValidationEnable,
	}

	output, err := conn.UpdateResolverDnssecConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver DNSSEC Config: %s", err)
	}

	d.SetId(aws.ToString(output.ResolverDNSSECConfig.Id))

	if _, err := waitDNSSECConfigCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver DNSSEC Config (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceDNSSECConfigRead(ctx, d, meta)...)
}

func resourceDNSSECConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	dnssecConfig, err := findResolverDNSSECConfigByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver DNSSEC Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver DNSSEC Config (%s): %s", d.Id(), err)
	}

	ownerID := aws.ToString(dnssecConfig.OwnerId)
	resourceID := aws.ToString(dnssecConfig.ResourceId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "route53resolver",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: ownerID,
		Resource:  fmt.Sprintf("resolver-dnssec-config/%s", resourceID),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrOwnerID, ownerID)
	d.Set(names.AttrResourceID, resourceID)
	d.Set("validation_status", dnssecConfig.ValidationStatus)

	return diags
}

func resourceDNSSECConfigDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	log.Printf("[DEBUG] Deleting Route53 Resolver DNSSEC Config: %s", d.Id())
	_, err := conn.UpdateResolverDnssecConfig(ctx, &route53resolver.UpdateResolverDnssecConfigInput{
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
		Validation: awstypes.ValidationDisable,
	})

	if errs.IsA[*awstypes.AccessDeniedException](err) {
		// VPC doesn't exist.
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Resolver DNSSEC Config (%s): %s", d.Id(), err)
	}

	if _, err = waitDNSSECConfigDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver DNSSEC Config (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findResolverDNSSECConfigByID(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.ResolverDnssecConfig, error) {
	input := &route53resolver.ListResolverDnssecConfigsInput{}

	// GetResolverDnssecConfig does not support query by ID.
	pages := route53resolver.NewListResolverDnssecConfigsPaginator(conn, input)
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

		for _, v := range page.ResolverDnssecConfigs {
			if aws.ToString(v.Id) == id {
				if validationStatus := v.ValidationStatus; validationStatus == awstypes.ResolverDNSSECValidationStatusDisabled {
					return nil, &retry.NotFoundError{
						Message:     string(validationStatus),
						LastRequest: input,
					}
				}
				return &v, nil
			}
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func statusDNSSECConfig(ctx context.Context, conn *route53resolver.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findResolverDNSSECConfigByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ValidationStatus), nil
	}
}

const (
	dnssecConfigCreatedTimeout = 10 * time.Minute
	dnssecConfigDeletedTimeout = 10 * time.Minute
)

func waitDNSSECConfigCreated(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.ResolverDnssecConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResolverDNSSECValidationStatusEnabling),
		Target:  enum.Slice(awstypes.ResolverDNSSECValidationStatusEnabled),
		Refresh: statusDNSSECConfig(ctx, conn, id),
		Timeout: dnssecConfigCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResolverDnssecConfig); ok {
		return output, err
	}

	return nil, err
}

func waitDNSSECConfigDeleted(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.ResolverDnssecConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResolverDNSSECValidationStatusDisabling),
		Target:  []string{},
		Refresh: statusDNSSECConfig(ctx, conn, id),
		Timeout: dnssecConfigDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResolverDnssecConfig); ok {
		return output, err
	}

	return nil, err
}
