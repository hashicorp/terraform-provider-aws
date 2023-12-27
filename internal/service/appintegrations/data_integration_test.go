// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appintegrations_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appintegrationsservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAppIntegrationsDataIntegration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dataIntegration appintegrationsservice.GetDataIntegrationOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "example description"
	firstExecutionFrom := "1439788442681"

	resourceName := "aws_appintegrations_data_integration.test"

	key := "DATA_INTEGRATION_SOURCE_URI"
	sourceUri := os.Getenv(key)
	if sourceUri == "" {
		t.Skip("Environment variable DATA_INTEGRATION_SOURCE_URI is not set")
		// sourceUri of the form Salesforce://AppFlow/<NameOfSalesforceConnectorProfile>
		// sourceUri = "Salesforce://AppFlow/test"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appintegrationsservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataIntegrationConfig_basic(rName, description, sourceUri, firstExecutionFrom),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataIntegrationExists(ctx, resourceName, &dataIntegration),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key", "aws_kms_key.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
		},
	})
}

func TestAccAppIntegrationsDataIntegration_updateDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var dataIntegration appintegrationsservice.GetDataIntegrationOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalDescription := "original description"
	updatedDescription := "updated description"
	firstExecutionFrom := "1439788442681"

	resourceName := "aws_appintegrations_data_integration.test"

	key := "DATA_INTEGRATION_SOURCE_URI"
	sourceUri := os.Getenv(key)
	if sourceUri == "" {
		t.Skip("Environment variable DATA_INTEGRATION_SOURCE_URI is not set")
		// sourceUri of the form Salesforce://AppFlow/<NameOfSalesforceConnectorProfile>
		// sourceUri = "Salesforce://AppFlow/test"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appintegrationsservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataIntegrationConfig_basic(rName, originalDescription, sourceUri, firstExecutionFrom),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataIntegrationExists(ctx, resourceName, &dataIntegration),
					resource.TestCheckResourceAttr(resourceName, "description", originalDescription),
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
					testAccCheckDataIntegrationExists(ctx, resourceName, &dataIntegration),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
				),
			},
		},
	})
}

func TestAccAppIntegrationsDataIntegration_updateName(t *testing.T) {
	ctx := acctest.Context(t)
	var dataIntegration appintegrationsservice.GetDataIntegrationOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "example description"
	firstExecutionFrom := "1439788442681"

	resourceName := "aws_appintegrations_data_integration.test"

	key := "DATA_INTEGRATION_SOURCE_URI"
	sourceUri := os.Getenv(key)
	if sourceUri == "" {
		t.Skip("Environment variable DATA_INTEGRATION_SOURCE_URI is not set")
		// sourceUri of the form Salesforce://AppFlow/<NameOfSalesforceConnectorProfile>
		// sourceUri = "Salesforce://AppFlow/test"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appintegrationsservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataIntegrationConfig_basic(rName, description, sourceUri, firstExecutionFrom),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataIntegrationExists(ctx, resourceName, &dataIntegration),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataIntegrationConfig_basic(rName2, description, sourceUri, firstExecutionFrom),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataIntegrationExists(ctx, resourceName, &dataIntegration),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccAppIntegrationsDataIntegration_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var dataIntegration appintegrationsservice.GetDataIntegrationOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "example description"
	firstExecutionFrom := "1439788442681"

	resourceName := "aws_appintegrations_data_integration.test"

	key := "DATA_INTEGRATION_SOURCE_URI"
	sourceUri := os.Getenv(key)
	if sourceUri == "" {
		t.Skip("Environment variable DATA_INTEGRATION_SOURCE_URI is not set")
		// sourceUri of the form Salesforce://AppFlow/<NameOfSalesforceConnectorProfile>
		// sourceUri = "Salesforce://AppFlow/test"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appintegrationsservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataIntegrationConfig_basic(rName, description, sourceUri, firstExecutionFrom),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataIntegrationExists(ctx, resourceName, &dataIntegration),
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
					testAccCheckDataIntegrationExists(ctx, resourceName, &dataIntegration),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccDataIntegrationConfig_tagsUpdated(rName, description, sourceUri, firstExecutionFrom),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataIntegrationExists(ctx, resourceName, &dataIntegration),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func testAccCheckDataIntegrationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppIntegrationsConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appintegrations_data_integration" {
				continue
			}

			input := &appintegrationsservice.GetDataIntegrationInput{
				Identifier: aws.String(rs.Primary.ID),
			}

			resp, err := conn.GetDataIntegrationWithContext(ctx, input)

			if err == nil {
				if aws.StringValue(resp.Id) == rs.Primary.ID {
					return fmt.Errorf("Data Integration '%s' was not deleted properly", rs.Primary.ID)
				}
			}
		}

		return nil
	}
}

func testAccCheckDataIntegrationExists(ctx context.Context, name string, dataIntegration *appintegrationsservice.GetDataIntegrationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppIntegrationsConn(ctx)
		input := &appintegrationsservice.GetDataIntegrationInput{
			Identifier: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetDataIntegrationWithContext(ctx, input)

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
