package appintegrations_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appintegrationsservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccDataIntegration_basic(t *testing.T) {
	var dataIntegration appintegrationsservice.GetDataIntegrationOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalDescription := "original description"
	updatedDescription := "updated description"
	firstExecutionFrom := "1439788442681"

	resourceName := "aws_appintegrations_data_integration.test"

	key := "DATA_INTEGRATION_SOURCE_URI"
	var sourceUri string
	sourceUri = os.Getenv(key)
	if sourceUri == "" {
		sourceUri = "Salesforce://AppFlow/test"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appintegrationsservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDataIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataIntegrationConfig_basic(rName, originalDescription, sourceUri, firstExecutionFrom),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataIntegrationExists(resourceName, &dataIntegration),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", originalDescription),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key", "aws_kms_key.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_uri", sourceUri),
					resource.TestCheckResourceAttr(resourceName, "schedule_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule_config.0.first_execution_from", firstExecutionFrom),
					resource.TestCheckResourceAttr(resourceName, "schedule_config.0.object", "Account"),
					resource.TestCheckResourceAttr(resourceName, "schedule_config.0.schedule_expression", "rate(1 hour)"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataIntegrationConfig_basic(rName, updatedDescription, sourceUri, firstExecutionFrom),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataIntegrationExists(resourceName, &dataIntegration),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key", "aws_kms_key.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "source_uri", sourceUri),
					resource.TestCheckResourceAttr(resourceName, "schedule_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule_config.0.first_execution_from", firstExecutionFrom),
					resource.TestCheckResourceAttr(resourceName, "schedule_config.0.object", "Account"),
					resource.TestCheckResourceAttr(resourceName, "schedule_config.0.schedule_expression", "rate(1 hour)"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
				),
			},
		},
	})
}

func TestAccDataIntegration_updateTags(t *testing.T) {
	var dataIntegration appintegrationsservice.GetDataIntegrationOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "example description"
	firstExecutionFrom := "1439788442681"

	resourceName := "aws_appintegrations_data_integration.test"

	key := "DATA_INTEGRATION_SOURCE_URI"
	var sourceUri string
	sourceUri = os.Getenv(key)
	if sourceUri == "" {
		sourceUri = "Salesforce://AppFlow/test"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appintegrationsservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDataIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataIntegrationConfig_basic(rName, description, sourceUri, firstExecutionFrom),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataIntegrationExists(resourceName, &dataIntegration),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataIntegrationConfig_tags(rName, description, sourceUri, firstExecutionFrom),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataIntegrationExists(resourceName, &dataIntegration),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccDataIntegrationConfig_tagsUpdated(rName, description, sourceUri, firstExecutionFrom),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataIntegrationExists(resourceName, &dataIntegration),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func testAccCheckDataIntegrationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppIntegrationsConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appintegrations_data_integration" {
			continue
		}

		input := &appintegrationsservice.GetDataIntegrationInput{
			Identifier: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetDataIntegration(input)

		if err == nil {
			if aws.StringValue(resp.Id) == rs.Primary.ID {
				return fmt.Errorf("Data Integration '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckDataIntegrationExists(name string, dataIntegration *appintegrationsservice.GetDataIntegrationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppIntegrationsConn
		input := &appintegrationsservice.GetDataIntegrationInput{
			Identifier: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetDataIntegration(input)

		if err != nil {
			return err
		}

		*dataIntegration = *resp

		return nil
	}
}

func testAccDataIntegrationBaseConfig() string {
	return `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}
`
}

func testAccDataIntegrationConfig_basic(rName, description, sourceUri, firstExecutionFrom string) string {
	return acctest.ConfigCompose(
		testAccDataIntegrationBaseConfig(),
		fmt.Sprintf(`
resource "aws_appintegrations_data_integration" "test" {
  name        = %[1]q
  description = %[2]q
  kms_key     = aws_kms_key.test.arn
  source_uri  = %[3]q

  schedule_config {
    first_execution_from = %[4]q
    object               = "Account"
    schedule_expression  = "rate(1 hour)"
  }

  tags = {
    "Key1" = "Value1"
  }
}
`, rName, description, sourceUri, firstExecutionFrom))
}

func testAccDataIntegrationConfig_tags(rName, description, sourceUri, firstExecutionFrom string) string {
	return acctest.ConfigCompose(
		testAccDataIntegrationBaseConfig(),
		fmt.Sprintf(`
resource "aws_appintegrations_data_integration" "test" {
  name        = %[1]q
  description = %[2]q
  kms_key     = aws_kms_key.test.arn
  source_uri  = %[3]q

  schedule_config {
    first_execution_from = %[4]q
    object               = "Account"
    schedule_expression  = "rate(1 hour)"
  }

  tags = {
    "Key1" = "Value1"
    "Key2" = "Value2a"
  }
}
`, rName, description, sourceUri, firstExecutionFrom))
}

func testAccDataIntegrationConfig_tagsUpdated(rName, description, sourceUri, firstExecutionFrom string) string {
	return acctest.ConfigCompose(
		testAccDataIntegrationBaseConfig(),
		fmt.Sprintf(`
resource "aws_appintegrations_data_integration" "test" {
  name        = %[1]q
  description = %[2]q
  kms_key     = aws_kms_key.test.arn
  source_uri  = %[3]q

  schedule_config {
    first_execution_from = %[4]q
    object               = "Account"
    schedule_expression  = "rate(1 hour)"
  }

  tags = {
    "Key1" = "Value1"
    "Key2" = "Value2b"
    "Key3" = "Value3"
  }
}
`, rName, description, sourceUri, firstExecutionFrom))
}
