// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cognito_user_pools", name="User Pools")
func dataSourceUserPools() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserPoolsRead,

		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceUserPoolsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	name := d.Get(names.AttrName).(string)
	output, err := findUserPoolDescriptionTypesByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cognito User Pools: %s", err)
	}

	var arns, userPoolIDs []string

	for _, v := range output {
		userPoolID := aws.ToString(v.Id)
		arn := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   "cognito-idp",
			Region:    meta.(*conns.AWSClient).Region,
			AccountID: meta.(*conns.AWSClient).AccountID,
			Resource:  "userpool/" + userPoolID,
		}.String()

		userPoolIDs = append(userPoolIDs, userPoolID)
		arns = append(arns, arn)
	}

	d.SetId(name)
	d.Set(names.AttrIDs, userPoolIDs)
	d.Set(names.AttrARNs, arns)

	return diags
}

func findUserPoolDescriptionTypesByName(ctx context.Context, conn *cognitoidentityprovider.Client, name string) ([]awstypes.UserPoolDescriptionType, error) {
	input := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int32(60),
	}
	var output []awstypes.UserPoolDescriptionType

	pages := cognitoidentityprovider.NewListUserPoolsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.UserPools {
			if aws.ToString(v.Name) == name {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
