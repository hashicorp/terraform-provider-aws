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
)

// @SDKDataSource("aws_service_discovery_http_namespace")
func DataSourceHTTPNamespace() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceHTTPNamespaceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"http_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validNamespaceName,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceHTTPNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
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
	d.Set("arn", arn)
	d.Set("description", ns.Description)
	if ns.Properties != nil && ns.Properties.HttpProperties != nil {
		d.Set("http_name", ns.Properties.HttpProperties.HttpName)
	} else {
		d.Set("http_name", nil)
	}
	d.Set("name", ns.Name)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Service Discovery HTTP Namespace (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
