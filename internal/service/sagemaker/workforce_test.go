package sagemaker_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccWorkforce_cognitoConfig(t *testing.T) {
	var workforce sagemaker.Workforce
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workforce.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkforceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkforceCognitoConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkforceExists(resourceName, &workforce),
					resource.TestCheckResourceAttr(resourceName, "workforce_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workforce/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cognito_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cognito_config.0.client_id", "aws_cognito_user_pool_client.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "cognito_config.0.user_pool", "aws_cognito_user_pool.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.0.cidrs.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "subdomain"),
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

func testAccWorkforce_oidcConfig(t *testing.T) {
	var workforce sagemaker.Workforce
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workforce.test"
	endpoint1 := "https://example.com"
	endpoint2 := "https://test.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkforceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkforceOIDCConfig(rName, endpoint1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkforceExists(resourceName, &workforce),
					resource.TestCheckResourceAttr(resourceName, "workforce_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workforce/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cognito_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.authorization_endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.client_id", rName),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.client_secret", rName),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.issuer", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.jwks_uri", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.logout_endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.token_endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.user_info_endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.0.cidrs.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "subdomain"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oidc_config.0.client_secret"},
			},
			{
				Config: testAccWorkforceOIDCConfig(rName, endpoint2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkforceExists(resourceName, &workforce),
					resource.TestCheckResourceAttr(resourceName, "workforce_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`workforce/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cognito_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.authorization_endpoint", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.client_id", rName),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.client_secret", rName),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.issuer", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.jwks_uri", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.logout_endpoint", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.token_endpoint", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "oidc_config.0.user_info_endpoint", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.0.cidrs.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "subdomain"),
				),
			},
		},
	})
}
func testAccWorkforce_sourceIPConfig(t *testing.T) {
	var workforce sagemaker.Workforce
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workforce.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkforceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkforceSourceIP1Config(rName, "1.1.1.1/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkforceExists(resourceName, &workforce),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.0.cidrs.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "source_ip_config.0.cidrs.*", "1.1.1.1/32"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkforceSourceIP2Config(rName, "2.2.2.2/32", "3.3.3.3/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkforceExists(resourceName, &workforce),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "source_ip_config.0.cidrs.*", "2.2.2.2/32"),
					resource.TestCheckTypeSetElemAttr(resourceName, "source_ip_config.0.cidrs.*", "3.3.3.3/32"),
				),
			},
			{
				Config: testAccWorkforceSourceIP1Config(rName, "2.2.2.2/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkforceExists(resourceName, &workforce),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_ip_config.0.cidrs.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "source_ip_config.0.cidrs.*", "2.2.2.2/32"),
				),
			},
		},
	})
}

func testAccWorkforce_disappears(t *testing.T) {
	var workforce sagemaker.Workforce
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_workforce.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkforceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkforceCognitoConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkforceExists(resourceName, &workforce),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceWorkforce(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckWorkforceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_workforce" {
			continue
		}

		_, err := tfsagemaker.FindWorkforceByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SageMaker Workforce %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckWorkforceExists(n string, workforce *sagemaker.Workforce) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SageMaker Workforce ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

		output, err := tfsagemaker.FindWorkforceByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*workforce = *output

		return nil
	}
}

func testAccWorkforceBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_client" "test" {
  name            = %[1]q
  generate_secret = true
  user_pool_id    = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}
`, rName)
}

func testAccWorkforceCognitoConfig(rName string) string {
	return testAccWorkforceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_workforce" "test" {
  workforce_name = %[1]q

  cognito_config {
    client_id = aws_cognito_user_pool_client.test.id
    user_pool = aws_cognito_user_pool_domain.test.user_pool_id
  }
}
`, rName)
}

func testAccWorkforceSourceIP1Config(rName, cidr1 string) string {
	return testAccWorkforceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_workforce" "test" {
  workforce_name = %[1]q

  cognito_config {
    client_id = aws_cognito_user_pool_client.test.id
    user_pool = aws_cognito_user_pool_domain.test.user_pool_id
  }

  source_ip_config {
    cidrs = [%[2]q]
  }
}
`, rName, cidr1)
}

func testAccWorkforceSourceIP2Config(rName, cidr1, cidr2 string) string {
	return testAccWorkforceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_workforce" "test" {
  workforce_name = %[1]q

  cognito_config {
    client_id = aws_cognito_user_pool_client.test.id
    user_pool = aws_cognito_user_pool_domain.test.user_pool_id
  }

  source_ip_config {
    cidrs = [%[2]q, %[3]q]
  }
}
`, rName, cidr1, cidr2)
}

func testAccWorkforceOIDCConfig(rName, endpoint string) string {
	return testAccWorkforceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_workforce" "test" {
  workforce_name = %[1]q

  oidc_config {
    authorization_endpoint = %[2]q
    client_id              = %[1]q
    client_secret          = %[1]q
    issuer                 = %[2]q
    jwks_uri               = %[2]q
    logout_endpoint        = %[2]q
    token_endpoint         = %[2]q
    user_info_endpoint     = %[2]q
  }
}
`, rName, endpoint)
}
