package ses_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
)

func TestAccSESConfigurationSet_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ses", fmt.Sprintf("configuration-set/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sending_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "reputation_metrics_enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "last_fresh_start"),
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

func TestAccSESConfigurationSet_sendingEnabled(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetSendingConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sending_enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "last_fresh_start"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetSendingConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sending_enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "last_fresh_start"),
				),
			},
			{
				Config: testAccConfigurationSetSendingConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sending_enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "last_fresh_start"),
				),
			},
		},
	})
}

func TestAccSESConfigurationSet_reputationMetricsEnabled(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetReputationMetricsConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "reputation_metrics_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetReputationMetricsConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "reputation_metrics_enabled", "true"),
				),
			},
			{
				Config: testAccConfigurationSetReputationMetricsConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "reputation_metrics_enabled", "false"),
				),
			},
		},
	})
}

func TestAccSESConfigurationSet_deliveryOptions(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetDeliveryOptionsConfig(rName, ses.TlsPolicyRequire),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", ses.TlsPolicyRequire),
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

func TestAccSESConfigurationSet_Update_deliveryOptions(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
				),
			},
			{
				Config: testAccConfigurationSetDeliveryOptionsConfig(rName, ses.TlsPolicyRequire),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", ses.TlsPolicyRequire),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetDeliveryOptionsConfig(rName, ses.TlsPolicyOptional),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", ses.TlsPolicyOptional),
				),
			},
			{
				Config: testAccConfigurationSetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "0"),
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

func TestAccSESConfigurationSet_emptyDeliveryOptions(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEmptyDeliveryOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", ses.TlsPolicyOptional),
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

func TestAccSESConfigurationSet_Update_emptyDeliveryOptions(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "0"),
				),
			},
			{
				Config: testAccConfigurationSetEmptyDeliveryOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.0.tls_policy", ses.TlsPolicyOptional),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "delivery_options.#", "0"),
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

func TestAccSESConfigurationSet_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_configuration_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ses.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfses.ResourceConfigurationSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConfigurationSetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES configuration set not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES configuration set ID not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

		response, err := conn.DescribeConfigurationSet(&ses.DescribeConfigurationSetInput{
			ConfigurationSetName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if aws.StringValue(response.ConfigurationSet.Name) != rs.Primary.ID {
			return fmt.Errorf("The configuration set was not created")
		}
		return nil

	}
}

func testAccCheckConfigurationSetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_configuration_set" {
			continue
		}

		_, err := conn.DescribeConfigurationSet(&ses.DescribeConfigurationSetInput{
			ConfigurationSetName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, ses.ErrCodeConfigurationSetDoesNotExistException) {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccConfigurationSetBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name = %[1]q
}
`, rName)
}

func testAccConfigurationSetSendingConfig(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name            = %[1]q
  sending_enabled = %[2]t
}
`, rName, enabled)
}

func testAccConfigurationSetReputationMetricsConfig(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name                       = %[1]q
  reputation_metrics_enabled = %[2]t
}
`, rName, enabled)
}

func testAccConfigurationSetDeliveryOptionsConfig(rName, tlsPolicy string) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name = %[1]q

  delivery_options {
    tls_policy = %[2]q
  }
}
`, rName, tlsPolicy)
}

func testAccConfigurationSetEmptyDeliveryOptionsConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name = %[1]q

  delivery_options {}
}
`, rName)
}
