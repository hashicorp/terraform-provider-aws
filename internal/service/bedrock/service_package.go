// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (p *servicePackage) withExtraOptions(ctx context.Context, config map[string]any) []func(*bedrock.Options) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return []func(*bedrock.Options){
		func(o *bedrock.Options) {
			// bedrock.New prefers AWS_BEARER_TOKEN_BEDROCK over the provider-configured credentials.
			if os.Getenv("AWS_BEARER_TOKEN_BEDROCK") != "" {
				tflog.Info(ctx, "ignoring AWS_BEARER_TOKEN_BEDROCK in favor of provider-configured credentials", map[string]any{
					"service": p.ServicePackageName(),
				})
				o.AuthSchemePreference = cfg.AuthSchemePreference
				o.BearerAuthTokenProvider = cfg.BearerAuthTokenProvider
			}
		},
	}
}
