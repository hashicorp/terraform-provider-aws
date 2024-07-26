// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route53_resolver_query_log_config")
func DataSourceQueryLogConfig() *schema.Resource {
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

func dataSourceQueryLogConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	configID := d.Get("resolver_query_log_config_id").(string)

	input := &route53resolver.ListResolverQueryLogConfigsInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok && v.(*schema.Set).Len() > 0 {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).Route53resolverFilters()
	}

	var configs []*route53resolver.ResolverQueryLogConfig

	err := conn.ListResolverQueryLogConfigsPagesWithContext(ctx, input, func(page *route53resolver.ListResolverQueryLogConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResolverQueryLogConfigs {
			if configID != "" {
				if aws.StringValue(v.Id) == configID {
					configs = append(configs, v)
				}
			} else {
				configs = append(configs, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Route53 resolver Query Logging Configurations: %s", err)
	}

	if n := len(configs); n == 0 {
		return create.AppendDiagError(diags, names.Route53Resolver, create.ErrActionReading, DSNameQueryLogConfig, configID, errors.New("your query returned no results, "+
			"please change your search criteria and try again"))
	} else if n > 1 {
		return create.AppendDiagError(diags, names.Route53Resolver, create.ErrActionReading, DSNameQueryLogConfig, configID, errors.New("your query returned more than one result, "+
			"please try more specific search criteria"))
	}

	config := configs[0]

	d.SetId(aws.StringValue(config.Id))
	arn := aws.StringValue(config.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDestinationARN, config.DestinationArn)
	d.Set(names.AttrName, config.Name)
	d.Set(names.AttrOwnerID, config.OwnerId)
	d.Set("resolver_query_log_config_id", config.Id)

	shareStatus := aws.StringValue(config.ShareStatus)
	d.Set("share_status", shareStatus)

	if shareStatus != route53resolver.ShareStatusSharedWithMe {
		tags, err := listTags(ctx, conn, arn)

		if err != nil {
			return create.AppendDiagError(diags, names.AppConfig, create.ErrActionReading, DSNameQueryLogConfig, configID, err)
		}

		ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set(names.AttrTags, tags.Map()); err != nil {
			return create.AppendDiagError(diags, names.AppConfig, create.ErrActionSetting, DSNameQueryLogConfig, configID, err)
		}
	}

	return diags
}
