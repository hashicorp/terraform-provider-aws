// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_api_gateway_resource", name="Resource")
func dataSourceResource() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResourceRead,

		Schema: map[string]*schema.Schema{
			"parent_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Required: true,
			},
			"path_part": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	path := d.Get(names.AttrPath).(string)
	input := &apigateway.GetResourcesInput{
		RestApiId: aws.String(d.Get("rest_api_id").(string)),
	}

	match, err := findResource(ctx, conn, input, func(v *types.Resource) bool {
		return aws.ToString(v.Path) == path
	})

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("API Gateway Resource", err))
	}

	d.SetId(aws.ToString(match.Id))
	d.Set("parent_id", match.ParentId)
	d.Set("path_part", match.PathPart)

	return diags
}

func findResource(ctx context.Context, conn *apigateway.Client, input *apigateway.GetResourcesInput, filter tfslices.Predicate[*types.Resource]) (*types.Resource, error) {
	output, err := findResources(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findResources(ctx context.Context, conn *apigateway.Client, input *apigateway.GetResourcesInput, filter tfslices.Predicate[*types.Resource]) ([]types.Resource, error) {
	var output []types.Resource

	pages := apigateway.NewGetResourcesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

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
