// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"encoding/base64"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_ecr_authorization_token")
func DataSourceAuthorizationToken() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAuthorizationTokenRead,

		Schema: map[string]*schema.Schema{
			"registry_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"authorization_token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"proxy_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expires_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceAuthorizationTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRConn(ctx)
	params := &ecr.GetAuthorizationTokenInput{}
	if v, ok := d.GetOk("registry_id"); ok {
		params.RegistryIds = []*string{aws.String(v.(string))}
	}
	log.Printf("[DEBUG] Getting ECR authorization token")
	out, err := conn.GetAuthorizationTokenWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting ECR authorization token: %s", err)
	}
	log.Printf("[DEBUG] Received ECR AuthorizationData %v", out.AuthorizationData)
	authorizationData := out.AuthorizationData[0]
	authorizationToken := aws.StringValue(authorizationData.AuthorizationToken)
	expiresAt := aws.TimeValue(authorizationData.ExpiresAt).Format(time.RFC3339)
	proxyEndpoint := aws.StringValue(authorizationData.ProxyEndpoint)
	authBytes, err := base64.URLEncoding.DecodeString(authorizationToken)
	if err != nil {
		d.SetId("")
		return sdkdiag.AppendErrorf(diags, "decoding ECR authorization token: %s", err)
	}
	basicAuthorization := strings.Split(string(authBytes), ":")
	if len(basicAuthorization) != 2 {
		return sdkdiag.AppendErrorf(diags, "unknown ECR authorization token format")
	}
	userName := basicAuthorization[0]
	password := basicAuthorization[1]
	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("authorization_token", authorizationToken)
	d.Set("proxy_endpoint", proxyEndpoint)
	d.Set("expires_at", expiresAt)
	d.Set("user_name", userName)
	d.Set("password", password)
	return diags
}
