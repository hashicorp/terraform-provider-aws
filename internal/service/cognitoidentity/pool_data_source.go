// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidentity

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cognito_identity_pool", name="Pool")
// @Tags(identifierAttribute="arn")
func dataSourcePool() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePoolRead,

		Schema: map[string]*schema.Schema{
			"allow_classic_flow": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"allow_unauthenticated_identities": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cognito_identity_providers": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrClientID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrProviderName: {
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
			"identity_pool_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validIdentityPoolName,
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
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNamePool         = "Pool Data Source"
	ListPoolMaxResults = 20
)

func dataSourcePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIdentityClient(ctx)

	name := d.Get("identity_pool_name").(string)

	ip, err := findPoolByName(ctx, conn, name)
	if err != nil {
		return create.AppendDiagError(diags, names.CognitoIdentity, create.ErrActionReading, DSNamePool, name, err)
	}

	d.SetId(aws.ToString(ip.IdentityPoolId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "cognito-identity",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("identitypool/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())
	d.Set("identity_pool_name", ip.IdentityPoolName)
	d.Set("allow_unauthenticated_identities", ip.AllowUnauthenticatedIdentities)
	d.Set("allow_classic_flow", ip.AllowClassicFlow)
	d.Set("developer_provider_name", ip.DeveloperProviderName)

	setTagsOut(ctx, ip.IdentityPoolTags)

	if err := d.Set("cognito_identity_providers", flattenIdentityProviders(ip.CognitoIdentityProviders)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cognito_identity_providers error: %s", err)
	}

	if err := d.Set("openid_connect_provider_arns", ip.OpenIdConnectProviderARNs); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting openid_connect_provider_arns error: %s", err)
	}

	if err := d.Set("saml_provider_arns", ip.SamlProviderARNs); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting saml_provider_arns error: %s", err)
	}

	if err := d.Set("supported_login_providers", ip.SupportedLoginProviders); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting supported_login_providers error: %s", err)
	}

	return diags
}

func findPoolByName(ctx context.Context, conn *cognitoidentity.Client, name string) (*cognitoidentity.DescribeIdentityPoolOutput, error) {
	var poolID string
	input := &cognitoidentity.ListIdentityPoolsInput{
		MaxResults: aws.Int32(ListPoolMaxResults),
	}

	p := cognitoidentity.NewListIdentityPoolsPaginator(conn, input)
	for p.HasMorePages() {
		pools, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, pool := range pools.IdentityPools {
			if aws.ToString(pool.IdentityPoolName) == name {
				poolID = aws.ToString(pool.IdentityPoolId)
				break
			}
		}
	}

	if poolID == "" {
		return nil, fmt.Errorf("no identity pool found with name %q", name)
	}

	pool, err := conn.DescribeIdentityPool(ctx, &cognitoidentity.DescribeIdentityPoolInput{
		IdentityPoolId: aws.String(poolID),
	})

	if err != nil {
		return nil, err
	}

	return pool, nil
}
