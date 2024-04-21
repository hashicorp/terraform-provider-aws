// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_cognito_user_pool_clients", name="User Pool Clients")
func dataSourceUserPoolClients() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceuserPoolClientsRead,
		Schema: map[string]*schema.Schema{
			"client_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"client_names": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceuserPoolClientsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID := d.Get("user_pool_id").(string)
	input := &cognitoidentityprovider.ListUserPoolClientsInput{
		UserPoolId: aws.String(userPoolID),
	}

	var clientIDs []string
	var clientNames []string

	pages := cognitoidentityprovider.NewListUserPoolClientsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "getting user pool clients: %s", err)
		}

		for _, v := range page.UserPoolClients {
			clientNames = append(clientNames, aws.ToString(v.ClientName))
			clientIDs = append(clientIDs, aws.ToString(v.ClientId))
		}

	}

	d.SetId(userPoolID)
	d.Set("client_ids", clientIDs)
	d.Set("client_names", clientNames)

	return diags
}
