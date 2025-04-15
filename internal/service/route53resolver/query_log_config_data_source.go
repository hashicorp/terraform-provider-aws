// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/namevaluesfilters"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route53_resolver_query_log_config", name="Query Log Config")
func dataSourceQueryLogConfig() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceQueryLogConfigRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDestinationARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFilter: namevaluesfilters.Schema(),
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validResolverName,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resolver_query_log_config_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"share_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameQueryLogConfig = "Query Log Config Data Source"
)

func dataSourceQueryLogConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	configID := d.Get("resolver_query_log_config_id").(string)

	input := &route53resolver.ListResolverQueryLogConfigsInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok && v.(*schema.Set).Len() > 0 {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).Route53ResolverFilters()
	}

	var configs []awstypes.ResolverQueryLogConfig

	pages := route53resolver.NewListResolverQueryLogConfigsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Route53 resolver Query Logging Configurations: %s", err)
		}

		for _, v := range page.ResolverQueryLogConfigs {
			if configID != "" {
				if aws.ToString(v.Id) == configID {
					configs = append(configs, v)
				}
			} else {
				configs = append(configs, v)
			}
		}
	}

	if n := len(configs); n == 0 {
		return create.AppendDiagError(diags, names.Route53Resolver, create.ErrActionReading, DSNameQueryLogConfig, configID, errors.New("your query returned no results, "+
			"please change your search criteria and try again"))
	} else if n > 1 {
		return create.AppendDiagError(diags, names.Route53Resolver, create.ErrActionReading, DSNameQueryLogConfig, configID, errors.New("your query returned more than one result, "+
			"please try more specific search criteria"))
	}

	config := configs[0]

	d.SetId(aws.ToString(config.Id))
	arn := aws.ToString(config.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDestinationARN, config.DestinationArn)
	d.Set(names.AttrName, config.Name)
	d.Set(names.AttrOwnerID, config.OwnerId)
	d.Set("resolver_query_log_config_id", config.Id)

	shareStatus := config.ShareStatus
	d.Set("share_status", shareStatus)

	if shareStatus != awstypes.ShareStatusSharedWithMe {
		tags, err := listTags(ctx, conn, arn)

		if err != nil {
			return create.AppendDiagError(diags, names.AppConfig, create.ErrActionReading, DSNameQueryLogConfig, configID, err)
		}

		ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)
		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		if err := d.Set(names.AttrTags, tags.Map()); err != nil {
			return create.AppendDiagError(diags, names.AppConfig, create.ErrActionSetting, DSNameQueryLogConfig, configID, err)
		}
	}

	return diags
}
