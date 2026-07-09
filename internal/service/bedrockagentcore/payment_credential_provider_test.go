// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCorePaymentCredentialProvider_stripePrivy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_payment_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPaymentCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPaymentCredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPaymentCredentialProviderConfig_stripePrivy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPaymentCredentialProviderExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_vendor"), knownvalue.StringExact("StripePrivy")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("credential_provider_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`.+`))),
					// The service creates a managed secret and returns its ARN; assert it is surfaced.
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provider_configuration").AtSliceIndex(0).AtMapKey("stripe_privy_configuration").AtSliceIndex(0).AtMapKey("app_secret_arn").AtSliceIndex(0).AtMapKey("secret_arn"), knownvalue.NotNull()),
					// Regression: tags must round-trip (Get does not return them; the tagging
					// interceptor reads them via ListTags). A perpetual diff here means Read clobbered them.
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll).AtMapKey("Name"), knownvalue.StringExact(rName)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				// The provider configuration holds write-only secrets that the API does
				// not return, so the block cannot be reconstructed on import.
				ImportStateVerifyIgnore: []string{
					"provider_configuration",
				},
			},
		},
	})
}

func TestAccBedrockAgentCorePaymentCredentialProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_payment_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPaymentCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPaymentCredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPaymentCredentialProviderConfig_stripePrivy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPaymentCredentialProviderExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourcePaymentCredentialProvider, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBedrockAgentCorePaymentCredentialProvider_nameValidation(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPaymentCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPaymentCredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Regression: the name schema validator must reject values with
				// disallowed characters at plan time, before any API call.
				Config:      testAccPaymentCredentialProviderConfig_stripePrivy("invalid name!"),
				ExpectError: regexache.MustCompile(`must contain only letters, numbers, hyphens, and underscores`),
			},
		},
	})
}

func TestAccBedrockAgentCorePaymentCredentialProvider_requiredFields(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPaymentCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPaymentCredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Regression: app_id is SDK-required within stripe_privy_configuration,
				// so omitting it must fail at plan time rather than at apply.
				Config:      testAccPaymentCredentialProviderConfig_stripePrivyMissingAppID(rName),
				ExpectError: regexache.MustCompile(`The argument "app_id" is required`),
			},
		},
	})
}

func TestAccBedrockAgentCorePaymentCredentialProvider_externalSecretRequiresConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPaymentCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPaymentCredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Regression: *_secret_source=EXTERNAL requires the matching
				// *_secret_config, enforced at plan time by ValidateConfig.
				Config:      testAccPaymentCredentialProviderConfig_stripePrivyExternalNoConfig(rName),
				ExpectError: regexache.MustCompile(`must be configured`),
			},
		},
	})
}

func TestAccBedrockAgentCorePaymentCredentialProvider_secretSourceForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_payment_credential_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPaymentCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPaymentCredentialProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPaymentCredentialProviderConfig_stripePrivy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPaymentCredentialProviderExists(ctx, t, resourceName),
				),
			},
			{
				// Regression: the service rejects changing a secret source between
				// MANAGED and EXTERNAL in place, so *_secret_source must force
				// replacement rather than plan a doomed in-place update.
				Config: testAccPaymentCredentialProviderConfig_stripePrivyExternal(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPaymentCredentialProviderExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccCheckPaymentCredentialProviderDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_payment_credential_provider" {
				continue
			}

			_, err := tfbedrockagentcore.FindPaymentCredentialProviderByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Payment Credential Provider %s still exists", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckPaymentCredentialProviderExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		_, err := tfbedrockagentcore.FindPaymentCredentialProviderByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
		return err
	}
}

func testAccPreCheckPaymentCredentialProviders(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)
	input := bedrockagentcorecontrol.ListPaymentCredentialProvidersInput{}

	_, err := conn.ListPaymentCredentialProviders(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPaymentCredentialProviderConfig_stripePrivy(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_payment_credential_provider" "test" {
  name                       = %[1]q
  credential_provider_vendor = "StripePrivy"

  tags = {
    Name = %[1]q
  }

  provider_configuration {
    stripe_privy_configuration {
      app_id                    = "app_test_id"
      app_secret                = "sk_test_secret"
      authorization_id          = "auth_test_id"
      authorization_private_key = "MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgNlzLhFfPe/14eR6GlFuWOzYTHgfXgyKs1yHwtpFISo6hRANCAAQPrRtegKcGCGBALTzewz0OnIpa9AeOe5BpcT0OS+Ej7odZ7fsTN8YgZzq5kBAY3u2UcZNHn6YJC70Z4bgpiuKI"
    }
  }
}
`, rName)
}

func testAccPaymentCredentialProviderConfig_stripePrivyExternal(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ secret = "sk_test_secret" })
}

resource "aws_bedrockagentcore_payment_credential_provider" "test" {
  name                       = %[1]q
  credential_provider_vendor = "StripePrivy"

  tags = {
    Name = %[1]q
  }

  provider_configuration {
    stripe_privy_configuration {
      app_id            = "app_test_id"
      app_secret_source = "EXTERNAL"

      app_secret_config {
        json_key  = "secret"
        secret_id = aws_secretsmanager_secret_version.test.secret_id
      }

      authorization_id          = "auth_test_id"
      authorization_private_key = "MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgNlzLhFfPe/14eR6GlFuWOzYTHgfXgyKs1yHwtpFISo6hRANCAAQPrRtegKcGCGBALTzewz0OnIpa9AeOe5BpcT0OS+Ej7odZ7fsTN8YgZzq5kBAY3u2UcZNHn6YJC70Z4bgpiuKI"
    }
  }
}
`, rName)
}

func testAccPaymentCredentialProviderConfig_stripePrivyExternalNoConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_payment_credential_provider" "test" {
  name                       = %[1]q
  credential_provider_vendor = "StripePrivy"

  provider_configuration {
    stripe_privy_configuration {
      app_id            = "app_test_id"
      authorization_id  = "auth_test_id"
      app_secret_source = "EXTERNAL"
    }
  }
}
`, rName)
}

func testAccPaymentCredentialProviderConfig_stripePrivyMissingAppID(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_payment_credential_provider" "test" {
  name                       = %[1]q
  credential_provider_vendor = "StripePrivy"

  provider_configuration {
    stripe_privy_configuration {
      app_secret                = "sk_test_secret"
      authorization_id          = "auth_test_id"
      authorization_private_key = "MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgNlzLhFfPe/14eR6GlFuWOzYTHgfXgyKs1yHwtpFISo6hRANCAAQPrRtegKcGCGBALTzewz0OnIpa9AeOe5BpcT0OS+Ej7odZ7fsTN8YgZzq5kBAY3u2UcZNHn6YJC70Z4bgpiuKI"
    }
  }
}
`, rName)
}
