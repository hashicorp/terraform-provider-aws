// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_service_discovery_service")
func DataSourceService() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
									"type": {
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
						"type": {
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tftags.TagsSchema(),
			"tags_all": {
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
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	serviceSummary, err := findServiceByNameAndNamespaceID(ctx, conn, name, d.Get("namespace_id").(string))

	if err != nil {
		return diag.Errorf("reading Service Discovery Service (%s): %s", name, err)
	}

	serviceID := aws.StringValue(serviceSummary.Id)

	service, err := FindServiceByID(ctx, conn, serviceID)

	if err != nil {
		return diag.Errorf("reading Service Discovery Service (%s): %s", serviceID, err)
	}

	d.SetId(serviceID)
	arn := aws.StringValue(service.Arn)
	d.Set("arn", arn)
	d.Set("description", service.Description)
	if tfMap := flattenDNSConfig(service.DnsConfig); len(tfMap) > 0 {
		if err := d.Set("dns_config", []interface{}{tfMap}); err != nil {
			return diag.Errorf("setting dns_config: %s", err)
		}
	} else {
		d.Set("dns_config", nil)
	}
	if tfMap := flattenHealthCheckConfig(service.HealthCheckConfig); len(tfMap) > 0 {
		if err := d.Set("health_check_config", []interface{}{tfMap}); err != nil {
			return diag.Errorf("setting health_check_config: %s", err)
		}
	} else {
		d.Set("health_check_config", nil)
	}
	if tfMap := flattenHealthCheckCustomConfig(service.HealthCheckCustomConfig); len(tfMap) > 0 {
		if err := d.Set("health_check_custom_config", []interface{}{tfMap}); err != nil {
			return diag.Errorf("setting health_check_custom_config: %s", err)
		}
	} else {
		d.Set("health_check_custom_config", nil)
	}
	d.Set("name", service.Name)
	d.Set("namespace_id", service.NamespaceId)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Service Discovery Service (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
