// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_apigatewayv2_api", name="API")
// @Tags
func dataSourceAPI() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAPIRead,

		Schema: map[string]*schema.Schema{
			"api_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"api_key_selection_expression": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cors_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_credentials": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"allow_headers": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allow_methods": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allow_origins": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"expose_headers": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"max_age": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disable_execute_api_endpoint": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"route_selection_expression": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAPIRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	apiID := d.Get("api_id").(string)
	api, err := findAPIByID(ctx, conn, apiID)

	if tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "no API Gateway v2 API matched; change the search criteria and try again")
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 API (%s): %s", apiID, err)
	}

	d.SetId(apiID)
	d.Set("api_endpoint", api.ApiEndpoint)
	d.Set("api_key_selection_expression", api.ApiKeySelectionExpression)
	d.Set(names.AttrARN, apiARN(meta.(*conns.AWSClient), d.Id()))
	if err := d.Set("cors_configuration", flattenCORSConfiguration(api.CorsConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cors_configuration: %s", err)
	}
	d.Set(names.AttrDescription, api.Description)
	d.Set("disable_execute_api_endpoint", api.DisableExecuteApiEndpoint)
	d.Set("execution_arn", apiInvokeARN(meta.(*conns.AWSClient), d.Id()))
	d.Set(names.AttrName, api.Name)
	d.Set("protocol_type", api.ProtocolType)
	d.Set("route_selection_expression", api.RouteSelectionExpression)
	d.Set(names.AttrVersion, api.Version)

	setTagsOut(ctx, api.Tags)

	return diags
}
