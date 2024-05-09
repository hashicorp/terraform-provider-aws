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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_service_discovery_dns_namespace")
func DataSourceDNSNamespace() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDNSNamespaceRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrType: {
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

	name := d.Get(names.AttrName).(string)
	nsType := d.Get(names.AttrType).(string)
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
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, ns.Description)
	if ns.Properties != nil && ns.Properties.DnsProperties != nil {
		d.Set("hosted_zone", ns.Properties.DnsProperties.HostedZoneId)
	} else {
		d.Set("hosted_zone", nil)
	}
	d.Set(names.AttrName, ns.Name)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Service Discovery %s Namespace (%s): %s", nsType, arn, err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
