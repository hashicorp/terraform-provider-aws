// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestOAuth2ProviderConfigExpandClientCredentials(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		clientID       types.String
		clientSecret   types.String
		clientIDWO     types.String
		clientSecretWO types.String
		external       bool
		wantID         string
		wantSecret     string
	}{
		"write-only ID with external secret": {
			clientIDWO: types.StringValue("synthetic-client-id"),
			external:   true,
			wantID:     "synthetic-client-id",
		},
		"both write-only": {
			clientIDWO:     types.StringValue("synthetic-client-id"),
			clientSecretWO: types.StringValue("synthetic-client-secret"),
			wantID:         "synthetic-client-id",
			wantSecret:     "synthetic-client-secret",
		},
		"both persistent": {
			clientID:     types.StringValue("synthetic-client-id"),
			clientSecret: types.StringValue("synthetic-client-secret"),
			wantID:       "synthetic-client-id",
			wantSecret:   "synthetic-client-secret",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			credentials := oauth2ProviderClientCredentialsModel{
				ClientID:           testCase.clientID,
				ClientSecret:       testCase.clientSecret,
				ClientIDWO:         testCase.clientIDWO,
				ClientSecretWO:     testCase.clientSecretWO,
				ClientSecretConfig: fwtypes.NewListNestedObjectValueOfNull[secretReferenceModel](ctx),
			}
			const secretARN = "arn:aws:secretsmanager:eu-west-2:000000000000:secret:synthetic-abcdef"
			if testCase.external {
				credentials.ClientSecretSource = fwtypes.StringEnumValue(awstypes.SecretSourceTypeExternal)
				reference, diags := fwtypes.NewListNestedObjectValueOfPtr(ctx, &secretReferenceModel{
					SecretID: types.StringValue(secretARN),
					JSONKey:  types.StringValue("SYNTHETIC_CLIENT_SECRET"),
				})
				if diags.HasError() {
					t.Fatalf("building secret reference: %v", diags)
				}
				credentials.ClientSecretConfig = reference
			}
			ctx = oauth2ClientCredentialsCtxKey.NewContext(ctx, oauth2CredentialProviderWriteOnlyCredentialsModel{
				oauth2ProviderClientCredentialsModel: credentials,
			})
			custom, diags := fwtypes.NewListNestedObjectValueOfPtr(ctx, &customOAuth2ProviderConfigModel{
				basicOAuth2ProviderConfigModel: basicOAuth2ProviderConfigModel{
					oauth2ProviderClientCredentialsModel: credentials,
					OAuthDiscovery:                       fwtypes.NewListNestedObjectValueOfNull[oauth2DiscoveryModel](ctx),
				},
				OnBehalfOfTokenExchangeConfig: fwtypes.NewListNestedObjectValueOfNull[onBehalfOfTokenExchangeConfigTypeModel](ctx),
				PrivateEndpoint:               fwtypes.NewListNestedObjectValueOfNull[privateEndpointModel](ctx),
				PrivateEndpointOverrides:      fwtypes.NewListNestedObjectValueOfNull[privateEndpointOverrideModel](ctx),
				PrivateKeyJWTConfig:           fwtypes.NewListNestedObjectValueOfNull[privateKeyJWTConfigModel](ctx),
			})
			if diags.HasError() {
				t.Fatalf("building provider config: %v", diags)
			}
			model := oauth2ProviderConfigModel{
				AtlassianOAuth2ProviderConfig: fwtypes.NewListNestedObjectValueOfNull[atlassianOAuth2ProviderConfigModel](ctx),
				CustomOAuth2ProviderConfig:    custom,
			}
			output, diags := model.Expand(ctx)
			if diags.HasError() {
				t.Fatalf("expanding provider config: %v", diags)
			}
			config, ok := output.(*awstypes.Oauth2ProviderConfigInputMemberCustomOauth2ProviderConfig)
			if !ok {
				t.Fatalf("unexpected output type: %T", output)
			}
			if got := aws.ToString(config.Value.ClientId); got != testCase.wantID {
				t.Errorf("ClientId = %q, want %q", got, testCase.wantID)
			}
			if got := aws.ToString(config.Value.ClientSecret); got != testCase.wantSecret {
				t.Errorf("ClientSecret = %q, want %q", got, testCase.wantSecret)
			}
			if testCase.external {
				if config.Value.ClientSecret != nil {
					t.Error("external secret must not send an inline client secret")
				}
				if config.Value.ClientSecretSource != awstypes.SecretSourceTypeExternal {
					t.Error("external secret source was not preserved")
				}
				if reference := config.Value.ClientSecretConfig; reference == nil ||
					aws.ToString(reference.SecretId) != secretARN ||
					aws.ToString(reference.JsonKey) != "SYNTHETIC_CLIENT_SECRET" {
					t.Errorf("external secret reference was not preserved: %#v", reference)
				}
			}
		})
	}
}
