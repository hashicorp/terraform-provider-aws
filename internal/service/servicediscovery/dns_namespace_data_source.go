// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_service_discovery_dns_namespace")
func DataSourceDNSNamespace() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDNSNamespaceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Required: true,
				// HTTP namespaces are handled via the aws_service_discovery_http_namespace data source.
				ValidateFunc: validation.StringInSlice([]string{
					servicediscovery.NamespaceTypeDnsPublic,
					servicediscovery.NamespaceTypeDnsPrivate,
				}, false),
			},
		},
	}
}

func dataSourceDNSNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	nsType := d.Get("type").(string)
	nsSummary, err := findNamespaceByNameAndType(ctx, conn, name, nsType)

	if err != nil {
		return diag.Errorf("reading Service Discovery %s Namespace (%s): %s", name, nsType, err)
	}

	namespaceID := aws.StringValue(nsSummary.Id)

	ns, err := FindNamespaceByID(ctx, conn, namespaceID)

	if err != nil {
		return diag.Errorf("reading Service Discovery %s Namespace (%s): %s", nsType, namespaceID, err)
	}

	d.SetId(namespaceID)
	arn := aws.StringValue(ns.Arn)
	d.Set("arn", arn)
	d.Set("description", ns.Description)
	if ns.Properties != nil && ns.Properties.DnsProperties != nil {
		d.Set("hosted_zone", ns.Properties.DnsProperties.HostedZoneId)
	} else {
		d.Set("hosted_zone", nil)
	}
	d.Set("name", ns.Name)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Service Discovery %s Namespace (%s): %s", nsType, arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
