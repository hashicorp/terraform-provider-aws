// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/oam"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_oam_links")
func DataSourceLinks() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLinksRead,

		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

const (
	DSNameLinks = "Links Data Source"
)

func dataSourceLinksRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	listLinksInput := &oam.ListLinksInput{}
	paginator := oam.NewListLinksPaginator(conn, listLinksInput)
	var arns []string

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionReading, DSNameLinks, "", err)
		}

		for _, listLinksItem := range page.Items {
			arns = append(arns, aws.StringValue(listLinksItem.Arn))
		}
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrARNs, arns)

	return nil
}
