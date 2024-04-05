// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_api_gateway_authorizers")
func DataSourceAuthorizers() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAuthorizersRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAuthorizersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	apiID := d.Get("rest_api_id").(string)
	input := &apigateway.GetAuthorizersInput{
		RestApiId: aws.String(apiID),
	}
	var ids []*string

	err := getAuthorizersPages(ctx, conn, input, func(page *apigateway.GetAuthorizersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			if v == nil {
				continue
			}

			ids = append(ids, v.Id)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Authorizers (%s): %s", apiID, err)
	}

	d.SetId(apiID)
	d.Set("ids", aws.StringValueSlice(ids))

	return diags
}
