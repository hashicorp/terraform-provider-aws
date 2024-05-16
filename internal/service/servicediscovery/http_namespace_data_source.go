// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_service_discovery_http_namespace")
func DataSourceHTTPNamespace() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceHTTPNamespaceRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"http_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validNamespaceName,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceHTTPNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get(names.AttrName).(string)
	nsSummary, err := findNamespaceByNameAndType(ctx, conn, name, servicediscovery.NamespaceTypeHttp)

	if err != nil {
		return diag.Errorf("reading Service Discovery HTTP Namespace (%s): %s", name, err)
	}

	namespaceID := aws.StringValue(nsSummary.Id)

	ns, err := FindNamespaceByID(ctx, conn, namespaceID)

	if err != nil {
		return diag.Errorf("reading Service Discovery HTTP Namespace (%s): %s", namespaceID, err)
	}

	d.SetId(namespaceID)
	arn := aws.StringValue(ns.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, ns.Description)
	if ns.Properties != nil && ns.Properties.HttpProperties != nil {
		d.Set("http_name", ns.Properties.HttpProperties.HttpName)
	} else {
		d.Set("http_name", nil)
	}
	d.Set(names.AttrName, ns.Name)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Service Discovery HTTP Namespace (%s): %s", arn, err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
