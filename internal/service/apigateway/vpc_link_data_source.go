// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_api_gateway_vpc_link", name="VPC Link")
// @Tags
func dataSourceVPCLink() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVPCLinkRead,

		Schema: map[string]*schema.Schema{
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatusMessage: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"target_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceVPCLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	name := d.Get(names.AttrName)
	input := &apigateway.GetVpcLinksInput{}

	match, err := findVPCLink(ctx, conn, input, func(v *types.VpcLink) bool {
		return aws.ToString(v.Name) == name
	})

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("API Gateway VPC Link", err))
	}

	d.SetId(aws.ToString(match.Id))
	d.Set(names.AttrDescription, match.Description)
	d.Set(names.AttrName, match.Name)
	d.Set(names.AttrStatus, match.Status)
	d.Set(names.AttrStatusMessage, match.StatusMessage)
	d.Set("target_arns", match.TargetArns)

	setTagsOut(ctx, match.Tags)

	return diags
}

func findVPCLink(ctx context.Context, conn *apigateway.Client, input *apigateway.GetVpcLinksInput, filter tfslices.Predicate[*types.VpcLink]) (*types.VpcLink, error) {
	output, err := findVPCLinks(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCLinks(ctx context.Context, conn *apigateway.Client, input *apigateway.GetVpcLinksInput, filter tfslices.Predicate[*types.VpcLink]) ([]types.VpcLink, error) {
	var output []types.VpcLink

	pages := apigateway.NewGetVpcLinksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
