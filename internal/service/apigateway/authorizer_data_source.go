// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_api_gateway_authorizer", name="Authorizer")
func dataSourceAuthorizer() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAuthorizerRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorizer_credentials": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorizer_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"authorizer_result_ttl_in_seconds": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"authorizer_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_validation_expression": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAuthorizerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	authorizerID := d.Get("authorizer_id").(string)
	apiID := d.Get("rest_api_id").(string)
	authorizer, err := findAuthorizerByTwoPartKey(ctx, conn, authorizerID, apiID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Authorizer (%s): %s", authorizerID, err)
	}

	d.SetId(authorizerID)
	d.Set(names.AttrARN, authorizerARN(meta.(*conns.AWSClient), apiID, d.Id()))
	d.Set("authorizer_credentials", authorizer.AuthorizerCredentials)
	if authorizer.AuthorizerResultTtlInSeconds != nil { // nosemgrep:ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("authorizer_result_ttl_in_seconds", authorizer.AuthorizerResultTtlInSeconds)
	} else {
		d.Set("authorizer_result_ttl_in_seconds", defaultAuthorizerTTL)
	}
	d.Set("authorizer_uri", authorizer.AuthorizerUri)
	d.Set("identity_source", authorizer.IdentitySource)
	d.Set("identity_validation_expression", authorizer.IdentityValidationExpression)
	d.Set(names.AttrName, authorizer.Name)
	d.Set("provider_arns", authorizer.ProviderARNs)
	d.Set(names.AttrType, authorizer.Type)

	return diags
}
