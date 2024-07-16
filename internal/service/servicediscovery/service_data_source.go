// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_service_discovery_service", name="Service")
func dataSourceService() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_records": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ttl": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrType: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"namespace_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"routing_policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"health_check_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"failure_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"resource_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"health_check_custom_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"failure_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags: tftags.TagsSchema(),
			names.AttrTagsAll: {
				Type:       schema.TypeMap,
				Optional:   true,
				Computed:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Deprecated: `this attribute has been deprecated`,
			},
		},
	}
}

func dataSourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get(names.AttrName).(string)
	serviceSummary, err := findServiceByNameAndNamespaceID(ctx, conn, name, d.Get("namespace_id").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Discovery Service (%s): %s", name, err)
	}

	serviceID := aws.ToString(serviceSummary.Id)
	service, err := findServiceByID(ctx, conn, serviceID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Discovery Service (%s): %s", serviceID, err)
	}

	d.SetId(serviceID)
	arn := aws.ToString(service.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, service.Description)
	if tfMap := flattenDNSConfig(service.DnsConfig); len(tfMap) > 0 {
		if err := d.Set("dns_config", []interface{}{tfMap}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting dns_config: %s", err)
		}
	} else {
		d.Set("dns_config", nil)
	}
	if tfMap := flattenHealthCheckConfig(service.HealthCheckConfig); len(tfMap) > 0 {
		if err := d.Set("health_check_config", []interface{}{tfMap}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting health_check_config: %s", err)
		}
	} else {
		d.Set("health_check_config", nil)
	}
	if tfMap := flattenHealthCheckCustomConfig(service.HealthCheckCustomConfig); len(tfMap) > 0 {
		if err := d.Set("health_check_custom_config", []interface{}{tfMap}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting health_check_custom_config: %s", err)
		}
	} else {
		d.Set("health_check_custom_config", nil)
	}
	d.Set(names.AttrName, service.Name)
	d.Set("namespace_id", service.NamespaceId)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Service Discovery Service (%s): %s", arn, err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
