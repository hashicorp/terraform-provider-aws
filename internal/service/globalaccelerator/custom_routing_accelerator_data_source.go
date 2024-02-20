// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_globalaccelerator_custom_routing_accelerator")
func DataSourceCustomRoutingAccelerator() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCustomRoutingAcceleratorRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"attributes": {
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
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_address_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_sets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_addresses": {
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
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceCustomRoutingAcceleratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var results []*globalaccelerator.CustomRoutingAccelerator

	err := conn.ListCustomRoutingAcceleratorsPagesWithContext(ctx, &globalaccelerator.ListCustomRoutingAcceleratorsInput{}, func(page *globalaccelerator.ListCustomRoutingAcceleratorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, l := range page.Accelerators {
			if l == nil {
				continue
			}

			if v, ok := d.GetOk("arn"); ok && v.(string) != aws.StringValue(l.AcceleratorArn) {
				continue
			}

			if v, ok := d.GetOk("name"); ok && v.(string) != aws.StringValue(l.Name) {
				continue
			}

			results = append(results, l)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Global Accelerator Custom Routing Accelerators: %s", err)
	}

	if count := len(results); count != 1 {
		return sdkdiag.AppendErrorf(diags, "search returned %d results, please revise so only one is returned", count)
	}

	accelerator := results[0]
	d.SetId(aws.StringValue(accelerator.AcceleratorArn))
	d.Set("arn", accelerator.AcceleratorArn)
	d.Set("dns_name", accelerator.DnsName)
	d.Set("enabled", accelerator.Enabled)
	d.Set("hosted_zone_id", meta.(*conns.AWSClient).GlobalAcceleratorHostedZoneID())
	d.Set("ip_address_type", accelerator.IpAddressType)
	if err := d.Set("ip_sets", flattenIPSets(accelerator.IpSets)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ip_sets: %s", err)
	}
	d.Set("name", accelerator.Name)

	acceleratorAttributes, err := FindCustomRoutingAcceleratorAttributesByARN(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global Accelerator Custom Routing Accelerator (%s) attributes: %s", d.Id(), err)
	}

	if err := d.Set("attributes", []interface{}{flattenCustomRoutingAcceleratorAttributes(acceleratorAttributes)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting attributes: %s", err)
	}

	tags, err := listTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Global Accelerator Custom Routing Accelerator (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
