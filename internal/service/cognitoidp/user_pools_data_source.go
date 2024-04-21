// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/aws/aws-sdk-go/aws/arn"
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
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceUserPoolsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	output, err := findUserPoolDescriptionTypes(ctx, conn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cognito User Pools: %s", err)
	}

	name := d.Get("name").(string)
	var arns, userPoolIDs []string

	for _, v := range output {
		if name != aws.ToString(v.Name) {
			continue
		}

		userPoolID := aws.ToString(v.Id)
		arn := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   names.CognitoIDPEndpointID,
			Region:    meta.(*conns.AWSClient).Region,
			AccountID: meta.(*conns.AWSClient).AccountID,
			Resource:  fmt.Sprintf("userpool/%s", userPoolID),
		}.String()

		userPoolIDs = append(userPoolIDs, userPoolID)
		arns = append(arns, arn)
	}

	d.SetId(name)
	d.Set("ids", userPoolIDs)
	d.Set("arns", arns)

	return diags
}

func findUserPoolDescriptionTypes(ctx context.Context, conn *cognitoidentityprovider.Client) ([]awstypes.UserPoolDescriptionType, error) {
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

		output = append(output, page.UserPools...)
	}

	return output, nil
}
