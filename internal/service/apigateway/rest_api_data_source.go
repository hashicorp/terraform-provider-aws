// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"strconv"

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
	"github.com/hashicorp/terraform-provider-aws/internal/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_api_gateway_rest_api", name="REST API")
// @Tags
func dataSourceRestAPI() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRestAPIRead,

		Schema: map[string]*schema.Schema{
			"api_key_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"binary_media_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"types": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_endpoint_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"minimum_compression_size": {
				Type:     nullable.TypeNullableInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceRestAPIRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	name := d.Get("name")
	input := &apigateway.GetRestApisInput{}

	match, err := findRestAPI(ctx, conn, input, func(v *types.RestApi) bool {
		return aws.ToString(v.Name) == name
	})

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("API Gateway REST API", err))
	}

	d.SetId(aws.ToString(match.Id))
	d.Set("api_key_source", match.ApiKeySource)
	d.Set("arn", apiARN(meta.(*conns.AWSClient), d.Id()))
	d.Set("binary_media_types", match.BinaryMediaTypes)
	d.Set("description", match.Description)
	if err := d.Set("endpoint_configuration", flattenEndpointConfiguration(match.EndpointConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint_configuration: %s", err)
	}
	d.Set("execution_arn", apiInvokeARN(meta.(*conns.AWSClient), d.Id()))
	if match.MinimumCompressionSize == nil {
		d.Set("minimum_compression_size", nil)
	} else {
		d.Set("minimum_compression_size", strconv.FormatInt(int64(aws.ToInt32(match.MinimumCompressionSize)), 10))
	}
	d.Set("policy", match.Policy)
	d.Set("root_resource_id", match.RootResourceId)

	setTagsOut(ctx, match.Tags)

	return diags
}

func findRestAPI(ctx context.Context, conn *apigateway.Client, input *apigateway.GetRestApisInput, filter tfslices.Predicate[*types.RestApi]) (*types.RestApi, error) {
	output, err := findRestAPIs(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findRestAPIs(ctx context.Context, conn *apigateway.Client, input *apigateway.GetRestApisInput, filter tfslices.Predicate[*types.RestApi]) ([]types.RestApi, error) {
	var output []types.RestApi

	pages := apigateway.NewGetRestApisPaginator(conn, input)
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
