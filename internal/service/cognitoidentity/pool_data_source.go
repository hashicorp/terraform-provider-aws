// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidentity

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"

	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"

	// tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"

	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cognito_identity_pool", name="Pool")
func DataSourcePool() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePoolRead,

		Schema: map[string]*schema.Schema{
			"identity_pool_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validIdentityPoolName,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cognito_identity_providers": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"provider_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"server_side_token_check": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"developer_provider_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"allow_unauthenticated_identities": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"allow_classic_flow": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"openid_connect_provider_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"saml_provider_arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"supported_login_providers": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// "tags": tftags.TagsSchemaComputed(), // TIP: Many, but not all, data sources have `tags` attributes.
		},
	}
}

const (
	DSNamePool         = "Pool Data Source"
	ListPoolMaxResults = 20
)

func dataSourcePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIdentityConn(ctx)

	name := d.Get("identity_pool_name").(string)

	ip, err := findPoolByName(ctx, conn, name)
	if err != nil {
		return append(diags, create.DiagError(names.CognitoIdentity, create.ErrActionReading, DSNamePool, name, err)...)
	}

	d.SetId(*ip.IdentityPoolId)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "cognito-identity",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("identitypool/%s", d.Id()),
	}
	d.Set("arn", arn.String())
	d.Set("identity_pool_name", ip.IdentityPoolName)

	return diags
}

func findPoolByName(ctx context.Context, conn *cognitoidentity.CognitoIdentity, name string) (*cognitoidentity.IdentityPool, error) {
	var poolID string

	pools, err := conn.ListIdentityPoolsWithContext(ctx, &cognitoidentity.ListIdentityPoolsInput{
		MaxResults: aws.Int64(ListPoolMaxResults),
	})
	if err != nil {
		return nil, err
	}

	for _, p := range pools.IdentityPools {
		if aws.StringValue(p.IdentityPoolName) == name {
			poolID = aws.StringValue(p.IdentityPoolId)
		} else {
			return nil, fmt.Errorf("no identity pool found with name %q", name)
		}
	}

	pool, err := conn.DescribeIdentityPoolWithContext(ctx, &cognitoidentity.DescribeIdentityPoolInput{
		IdentityPoolId: aws.String(poolID),
	})
	if err != nil {
		return nil, err
	}

	return pool, nil

}
