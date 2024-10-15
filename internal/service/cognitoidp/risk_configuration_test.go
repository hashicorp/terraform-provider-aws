// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPRiskConfiguration_exception(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_risk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRiskConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRiskConfigurationConfig_riskException(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRiskConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, "aws_cognito_user_pool.test", names.AttrID),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrClientID),
					resource.TestCheckResourceAttr(resourceName, "account_takeover_risk_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.*", "10.10.10.10/32"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.0.skipped_ip_range_list.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRiskConfigurationConfig_riskExceptionUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRiskConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, "aws_cognito_user_pool.test", names.AttrID),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrClientID),
					resource.TestCheckResourceAttr(resourceName, "account_takeover_risk_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.*", "10.10.10.10/32"),
					resource.TestCheckTypeSetElemAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.*", "10.10.10.11/32"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.0.skipped_ip_range_list.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "risk_exception_configuration.0.skipped_ip_range_list.*", "10.10.10.12/32"),
				),
			},
		},
	})
}

func TestAccCognitoIDPRiskConfiguration_client(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_risk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRiskConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRiskConfigurationConfig_riskExceptionClient(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRiskConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, "aws_cognito_user_pool.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClientID, "aws_cognito_user_pool_client.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "account_takeover_risk_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.*", "10.10.10.10/32"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.0.skipped_ip_range_list.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPRiskConfiguration_compromised(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_risk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRiskConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRiskConfigurationConfig_compromised(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRiskConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, "aws_cognito_user_pool.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "account_takeover_risk_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.0.event_filter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compromised_credentials_risk_configuration.0.event_filter.*", "SIGN_IN"),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.0.actions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.0.actions.0.event_action", "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPRiskConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_risk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRiskConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRiskConfigurationConfig_riskException(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRiskConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceRiskConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPRiskConfiguration_disappears_userPool(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_risk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRiskConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRiskConfigurationConfig_riskException(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRiskConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceUserPool(), "aws_cognito_user_pool.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPRiskConfiguration_empty(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRiskConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRiskConfigurationConfig_empty(rName),
				ExpectError: acctest.ExpectErrorAttrAtLeastOneOf(
					"account_takeover_risk_configuration",
					"compromised_credentials_risk_configuration",
					"risk_exception_configuration",
				),
			},
		},
	})
}

func TestAccCognitoIDPRiskConfiguration_nullRiskException(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRiskConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRiskConfigurationConfig_nullRiskException(rName),
				ExpectError: acctest.ExpectErrorAttrAtLeastOneOf(
					"risk_exception_configuration.0.blocked_ip_range_list",
					"risk_exception_configuration.0.skipped_ip_range_list",
				),
			},
		},
	})
}

func TestAccCognitoIDPRiskConfiguration_emptyRiskException(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRiskConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccRiskConfigurationConfig_emptyRiskException(rName),
				ExpectError: acctest.ExpectErrorAttrMinItems("risk_exception_configuration.0.blocked_ip_range_list", 1, 0),
			},
		},
	})
}

func testAccCheckRiskConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_risk_configuration" {
				continue
			}

			_, err := tfcognitoidp.FindRiskConfigurationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrClientID])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito Risk Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRiskConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		_, err := tfcognitoidp.FindRiskConfigurationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrClientID])

		return err
	}
}

func testAccRiskConfigurationConfig_riskException(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_risk_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id

  risk_exception_configuration {
    blocked_ip_range_list = ["10.10.10.10/32"]
  }
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRiskConfigurationConfig_riskExceptionUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_risk_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id

  risk_exception_configuration {
    blocked_ip_range_list = ["10.10.10.10/32", "10.10.10.11/32"]
    skipped_ip_range_list = ["10.10.10.12/32"]
  }
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRiskConfigurationConfig_compromised(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_risk_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id

  compromised_credentials_risk_configuration {
    event_filter = ["SIGN_IN"]
    actions {
      event_action = "BLOCK"
    }
  }
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRiskConfigurationConfig_riskExceptionClient(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_risk_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  client_id    = aws_cognito_user_pool_client.test.id

  risk_exception_configuration {
    blocked_ip_range_list = ["10.10.10.10/32"]
  }
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_client" "test" {
  name                = %[1]q
  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}
`, rName)
}

func testAccRiskConfigurationConfig_empty(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_risk_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRiskConfigurationConfig_nullRiskException(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_risk_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id

  risk_exception_configuration {
    blocked_ip_range_list = null
    skipped_ip_range_list = null
  }
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRiskConfigurationConfig_emptyRiskException(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_risk_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id

  risk_exception_configuration {
    blocked_ip_range_list = []
    skipped_ip_range_list = []
  }
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName)
}
