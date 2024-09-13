// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecrpublic

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ecrpublic_authorization_token")
func DataSourceAuthorizationToken() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAuthorizationTokenRead,

		Schema: map[string]*schema.Schema{
			"authorization_token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"expires_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPassword: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			names.AttrUserName: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAuthorizationTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRPublicClient(ctx)
	params := &ecrpublic.GetAuthorizationTokenInput{}

	out, err := conn.GetAuthorizationToken(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting ECR Public authorization token: %s", err)
	}

	authorizationData := out.AuthorizationData
	authorizationToken := aws.ToString(authorizationData.AuthorizationToken)
	expiresAt := aws.ToTime(authorizationData.ExpiresAt).Format(time.RFC3339)
	authBytes, err := itypes.Base64Decode(authorizationToken)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding ECR Public authorization token: %s", err)
	}

	basicAuthorization := strings.Split(string(authBytes), ":")
	if len(basicAuthorization) != 2 {
		return sdkdiag.AppendErrorf(diags, "unknown ECR Public authorization token format")
	}

	userName := basicAuthorization[0]
	password := basicAuthorization[1]
	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("authorization_token", authorizationToken)
	d.Set("expires_at", expiresAt)
	d.Set(names.AttrUserName, userName)
	d.Set(names.AttrPassword, password)

	return diags
}
