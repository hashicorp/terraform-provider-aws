// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_service_discovery_http_namespace", name="HTTP Namespace")
func dataSourceHTTPNamespace() *schema.Resource {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get(names.AttrName).(string)
	nsSummary, err := findNamespaceByNameAndType(ctx, conn, name, awstypes.NamespaceTypeHttp)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Discovery HTTP Namespace (%s): %s", name, err)
	}

	namespaceID := aws.ToString(nsSummary.Id)
	ns, err := findNamespaceByID(ctx, conn, namespaceID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Discovery HTTP Namespace (%s): %s", namespaceID, err)
	}

	d.SetId(namespaceID)
	arn := aws.ToString(ns.Arn)
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
		return sdkdiag.AppendErrorf(diags, "listing tags for Service Discovery HTTP Namespace (%s): %s", arn, err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
