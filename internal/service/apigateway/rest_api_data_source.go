// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types/nullable"
)

// @SDKDataSource("aws_api_gateway_rest_api")
func DataSourceRestAPI() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRestAPIRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"root_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_key_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"minimum_compression_size": {
				Type:     nullable.TypeNullableInt,
				Computed: true,
			},
			"binary_media_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceRestAPIRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	params := &apigateway.GetRestApisInput{}

	target := d.Get("name")
	var matchedApis []awstypes.RestApi
	log.Printf("[DEBUG] Reading API Gateway REST APIs: %+v", params)
	pages := apigateway.NewGetRestApisPaginator(conn, params)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "describing API Gateway REST APIs: %s", err)
		}

		for _, api := range page.Items {
			if aws.ToString(api.Name) == target {
				matchedApis = append(matchedApis, api)
			}
		}
	}

	match, err := tfresource.AssertSingleValueResult(matchedApis)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Exactly one API Gateway REST API not found with name: %s", target)
	}

	d.SetId(aws.ToString(match.Id))

	restApiArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/restapis/%s", d.Id()),
	}.String()
	d.Set("arn", restApiArn)
	d.Set("description", match.Description)
	d.Set("policy", match.Policy)
	d.Set("api_key_source", match.ApiKeySource)
	d.Set("binary_media_types", match.BinaryMediaTypes)

	if match.MinimumCompressionSize == nil {
		d.Set("minimum_compression_size", nil)
	} else {
		d.Set("minimum_compression_size", strconv.FormatInt(int64(aws.ToInt32(match.MinimumCompressionSize)), 10))
	}

	if err := d.Set("endpoint_configuration", flattenEndpointConfiguration(match.EndpointConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint_configuration: %s", err)
	}

	if err := d.Set("tags", KeyValueTags(ctx, match.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	executionArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "execute-api",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  d.Id(),
	}.String()
	d.Set("execution_arn", executionArn)
	d.Set("root_resource_id", match.RootResourceId)
	return diags
}
