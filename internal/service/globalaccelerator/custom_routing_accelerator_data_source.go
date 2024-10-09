// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"
	awstypes "github.com/aws/aws-sdk-go-v2/service/globalaccelerator/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_globalaccelerator_custom_routing_accelerator", name="Custom Routing Accelerator")
func dataSourceCustomRoutingAccelerator() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCustomRoutingAcceleratorRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrAttributes: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"flow_logs_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"flow_logs_s3_bucket": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"flow_logs_s3_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrIPAddressType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_sets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrIPAddresses: {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ip_family": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceCustomRoutingAcceleratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var results []awstypes.CustomRoutingAccelerator
	pages := globalaccelerator.NewListCustomRoutingAcceleratorsPaginator(conn, &globalaccelerator.ListCustomRoutingAcceleratorsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Global Accelerator Custom Routing Accelerators: %s", err)
		}

		for _, accelerator := range page.Accelerators {
			if v, ok := d.GetOk(names.AttrARN); ok && v.(string) != aws.ToString(accelerator.AcceleratorArn) {
				continue
			}

			if v, ok := d.GetOk(names.AttrName); ok && v.(string) != aws.ToString(accelerator.Name) {
				continue
			}

			results = append(results, accelerator)
		}
	}

	if count := len(results); count != 1 {
		return sdkdiag.AppendErrorf(diags, "search returned %d results, please revise so only one is returned", count)
	}

	accelerator := results[0]
	d.SetId(aws.ToString(accelerator.AcceleratorArn))
	d.Set(names.AttrARN, accelerator.AcceleratorArn)
	d.Set(names.AttrDNSName, accelerator.DnsName)
	d.Set(names.AttrEnabled, accelerator.Enabled)
	d.Set(names.AttrHostedZoneID, meta.(*conns.AWSClient).GlobalAcceleratorHostedZoneID(ctx))
	d.Set(names.AttrIPAddressType, accelerator.IpAddressType)
	if err := d.Set("ip_sets", flattenIPSets(accelerator.IpSets)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ip_sets: %s", err)
	}
	d.Set(names.AttrName, accelerator.Name)

	acceleratorAttributes, err := findCustomRoutingAcceleratorAttributesByARN(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global Accelerator Custom Routing Accelerator (%s) attributes: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrAttributes, []interface{}{flattenCustomRoutingAcceleratorAttributes(acceleratorAttributes)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting attributes: %s", err)
	}

	tags, err := listTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Global Accelerator Custom Routing Accelerator (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
