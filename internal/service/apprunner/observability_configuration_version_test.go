// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapprunner "github.com/hashicorp/terraform-provider-aws/internal/service/apprunner"
)

func TestAccAppRunnerObservabilityConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_observability_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObservabilityConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObservabilityConfigurationExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`observabilityconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "observability_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "observability_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "status", tfapprunner.ObservabilityConfigurationStatusActive),
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

func TestAccAppRunnerObservabilityConfiguration_traceConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_observability_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObservabilityConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityConfigurationConfig_traceConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObservabilityConfigurationExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`observabilityconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "observability_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "observability_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, "status", tfapprunner.ObservabilityConfigurationStatusActive),
					resource.TestCheckResourceAttr(resourceName, "trace_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "trace_configuration.0.vendor", "AWSXRAY"),
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

func TestAccAppRunnerObservabilityConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_observability_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObservabilityConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObservabilityConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapprunner.ResourceObservabilityConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppRunnerObservabilityConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_observability_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObservabilityConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityConfigurationConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObservabilityConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccObservabilityConfigurationConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObservabilityConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccObservabilityConfigurationConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObservabilityConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckObservabilityConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apprunner_observability_configuration" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerConn(ctx)

			input := &apprunner.DescribeObservabilityConfigurationInput{
				ObservabilityConfigurationArn: aws.String(rs.Primary.ID),
			}

			output, err := conn.DescribeObservabilityConfigurationWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}

			if output != nil && output.ObservabilityConfiguration != nil && aws.StringValue(output.ObservabilityConfiguration.Status) != apprunner.ObservabilityConfigurationStatusInactive {
				return fmt.Errorf("App Runner Observability Configuration (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckObservabilityConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Runner Service ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerConn(ctx)

		input := &apprunner.DescribeObservabilityConfigurationInput{
			ObservabilityConfigurationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeObservabilityConfigurationWithContext(ctx, input)

		if err != nil {
			return err
		}

		if output == nil || output.ObservabilityConfiguration == nil {
			return fmt.Errorf("App Runner Observability Configuration (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccObservabilityConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_observability_configuration" "test" {
  observability_configuration_name = %[1]q
}
`, rName)
}

func testAccObservabilityConfigurationConfig_traceConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_observability_configuration" "test" {
  observability_configuration_name = %[1]q
  trace_configuration {
    vendor = "AWSXRAY"
  }
}
`, rName)
}

func testAccObservabilityConfigurationConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_observability_configuration" "test" {
  observability_configuration_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccObservabilityConfigurationConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_observability_configuration" "test" {
  observability_configuration_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
