package cognitoidp_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCognitoIDPRiskConfiguration_exception(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_risk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRiskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRiskConfigurationConfig_riskException(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRiskConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", "aws_cognito_user_pool.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.*", "10.10.10.10/32"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.0.skipped_ip_range_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "account_takeover_risk_configuration.#", "0"),
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
					testAccCheckRiskConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", "aws_cognito_user_pool.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.*", "10.10.10.10/32"),
					resource.TestCheckTypeSetElemAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.*", "10.10.10.11/32"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.0.skipped_ip_range_list.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "risk_exception_configuration.0.skipped_ip_range_list.*", "10.10.10.12/32"),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "account_takeover_risk_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccCognitoIDPRiskConfiguration_client(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_risk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRiskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRiskConfigurationConfig_riskExceptionClient(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRiskConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", "aws_cognito_user_pool.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "client_id", "aws_cognito_user_pool_client.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "risk_exception_configuration.0.blocked_ip_range_list.*", "10.10.10.10/32"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.0.skipped_ip_range_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "account_takeover_risk_configuration.#", "0"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_risk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRiskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRiskConfigurationConfig_compromised(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRiskConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "user_pool_id", "aws_cognito_user_pool.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "risk_exception_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.0.event_filter.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compromised_credentials_risk_configuration.0.event_filter.*", "SIGN_IN"),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.0.actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compromised_credentials_risk_configuration.0.actions.0.event_action", "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "account_takeover_risk_configuration.#", "0"),
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

func TestAccCognitoIDPRiskConfiguration_disappears_userPool(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_risk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRiskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRiskConfigurationConfig_riskException(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRiskConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcognitoidp.ResourceUserPool(), "aws_cognito_user_pool.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRiskConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_risk_configuration" {
			continue
		}

		_, err := tfcognitoidp.FindRiskConfigurationById(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckRiskConfigurationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito Risk Configuration ID set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

		_, err := tfcognitoidp.FindRiskConfigurationById(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccRiskConfigurationConfig_riskException(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_risk_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id

  risk_exception_configuration {
    blocked_ip_range_list = ["10.10.10.10/32"]
  }
}
`, rName)
}

func testAccRiskConfigurationConfig_riskExceptionUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_risk_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id

  risk_exception_configuration {
    blocked_ip_range_list = ["10.10.10.10/32", "10.10.10.11/32"]
    skipped_ip_range_list = ["10.10.10.12/32"]
  }
}
`, rName)
}

func testAccRiskConfigurationConfig_compromised(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_risk_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id

  compromised_credentials_risk_configuration {
    event_filter = ["SIGN_IN"]
    actions {
      event_action = "BLOCK"
    }
  }
}
`, rName)
}

func testAccRiskConfigurationConfig_riskExceptionClient(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_client" "test" {
  name                = %[1]q
  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}

resource "aws_cognito_risk_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
  client_id    = aws_cognito_user_pool_client.test.id

  risk_exception_configuration {
    blocked_ip_range_list = ["10.10.10.10/32"]
  }
}
`, rName)
}
