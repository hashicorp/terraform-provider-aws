// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_ram_resource_share")
func DataSourceResourceShare() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResourceShareRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"resource_owner": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ram.ResourceOwner_Values(), false),
			},

			"resource_share_status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ram.ResourceShareStatus_Values(), false),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owning_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tftags.TagsSchemaComputed(),

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceResourceShareRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	owner := d.Get("resource_owner").(string)

	filters, filtersOk := d.GetOk("filter")

	params := &ram.GetResourceSharesInput{
		Name:          aws.String(name),
		ResourceOwner: aws.String(owner),
	}

	if v, ok := d.GetOk("resource_share_status"); ok {
		params.ResourceShareStatus = aws.String(v.(string))
	}

	if filtersOk {
		params.TagFilters = buildTagFilters(filters.(*schema.Set))
	}

	for {
		resp, err := conn.GetResourceSharesWithContext(ctx, params)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "retrieving resource share: empty response for: %s", params)
		}

		if len(resp.ResourceShares) > 1 {
			return sdkdiag.AppendErrorf(diags, "Multiple resource shares found for: %s", name)
		}

		if resp == nil || len(resp.ResourceShares) == 0 {
			return sdkdiag.AppendErrorf(diags, "No matching resource found: %s", err)
		}

		for _, r := range resp.ResourceShares {
			if aws.StringValue(r.Name) == name {
				d.SetId(aws.StringValue(r.ResourceShareArn))
				d.Set("arn", r.ResourceShareArn)
				d.Set("owning_account_id", r.OwningAccountId)
				d.Set("status", r.Status)

				if err := d.Set("tags", KeyValueTags(ctx, r.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
					return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
				}

				break
			}
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	listInput := &ram.ListResourcesInput{
		ResourceOwner:     aws.String(d.Get("resource_owner").(string)),
		ResourceShareArns: aws.StringSlice([]string{d.Get("arn").(string)}),
	}

	var resourceARNs []*string
	err := conn.ListResourcesPages(listInput, func(page *ram.ListResourcesOutput, lastPage bool) bool {
		for _, resource := range page.Resources {
			resourceARNs = append(resourceARNs, resource.Arn)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM resource share (%s) resources: %s", d.Id(), err)
	}

	if err := d.Set("resources", flex.FlattenStringList(resourceARNs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting resources: %s", err)
	}

	return diags
}

func buildTagFilters(set *schema.Set) []*ram.TagFilter {
	var filters []*ram.TagFilter

	for _, v := range set.List() {
		m := v.(map[string]interface{})
		var filterValues []*string
		for _, e := range m["values"].([]interface{}) {
			filterValues = append(filterValues, aws.String(e.(string)))
		}
		filters = append(filters, &ram.TagFilter{
			TagKey:    aws.String(m["name"].(string)),
			TagValues: filterValues,
		})
	}

	return filters
}
