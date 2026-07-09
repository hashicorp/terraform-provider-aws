// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// TestFlattenImportedProviderConfigurationSecretARNs proves that on import — when
// the write-only provider_configuration block is absent from state — the computed
// managed secret ARNs returned by the API are hydrated into a reconstructed block.
// Without this the ARNs plan as null on the documented reconcile apply and it
// fails with "inconsistent result after apply".
func TestFlattenImportedProviderConfigurationSecretARNs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	const appSecretARN = "arn:aws:secretsmanager:us-west-2:123456789012:secret:appsecret-abc"
	const authKeyARN = "arn:aws:secretsmanager:us-west-2:123456789012:secret:authprivkey-def"

	// Import scenario: provider_configuration is null in state.
	data := &paymentCredentialProviderResourceModel{
		ProviderConfiguration: fwtypes.NewListNestedObjectValueOfNull[paymentProviderConfigurationModel](ctx),
	}

	apiObject := &awstypes.PaymentProviderConfigurationOutputMemberStripePrivyConfiguration{
		Value: awstypes.StripePrivyConfigurationOutput{
			AppSecretArn:               &awstypes.Secret{SecretArn: aws.String(appSecretARN)},
			AuthorizationPrivateKeyArn: &awstypes.Secret{SecretArn: aws.String(authKeyARN)},
		},
	}

	if diags := flattenPaymentProviderConfigurationSecretARNs(ctx, apiObject, data); diags.HasError() {
		t.Fatalf("flatten: %v", diags)
	}

	if data.ProviderConfiguration.IsNull() {
		t.Fatalf("expected provider_configuration to be reconstructed on import, got null")
	}

	config, diags := data.ProviderConfiguration.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("ToPtr config: %v", diags)
	}
	stripe, diags := config.StripePrivyConfiguration.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("ToPtr stripe: %v", diags)
	}
	if stripe == nil {
		t.Fatalf("expected stripe_privy_configuration to be reconstructed")
	}

	// The computed ARNs must be populated so the reconcile apply plans them as known.
	appARN, diags := stripe.AppSecretARN.ToPtr(ctx)
	if diags.HasError() || appARN == nil {
		t.Fatalf("expected app_secret_arn to be populated: %v", diags)
	}
	if got := appARN.SecretARN.ValueString(); got != appSecretARN {
		t.Errorf("app_secret_arn = %q, want %q", got, appSecretARN)
	}

	authARN, diags := stripe.AuthorizationPrivateKeyARN.ToPtr(ctx)
	if diags.HasError() || authARN == nil {
		t.Fatalf("expected authorization_private_key_arn to be populated: %v", diags)
	}
	if got := authARN.SecretARN.ValueString(); got != authKeyARN {
		t.Errorf("authorization_private_key_arn = %q, want %q", got, authKeyARN)
	}

	// The write-only secret value stays null on import (re-supplied on the next apply).
	if !stripe.AppSecret.IsNull() {
		t.Errorf("expected app_secret to remain null on import, got %q", stripe.AppSecret.ValueString())
	}
}
