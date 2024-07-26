// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var configuration1 kafka.DescribeConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, resourceName, &configuration1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "kafka", regexache.MustCompile(`configuration/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_versions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "latest_revision", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestMatchResourceAttr(resourceName, "server_properties", regexache.MustCompile(`auto.create.topics.enable = true`)),
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

func TestAccKafkaConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var configuration1 kafka.DescribeConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, resourceName, &configuration1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkafka.ResourceConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaConfiguration_description(t *testing.T) {
	ctx := acctest.Context(t)
	var configuration1, configuration2 kafka.DescribeConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, resourceName, &configuration1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
					resource.TestCheckResourceAttr(resourceName, "latest_revision", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccKafkaConfiguration_kafkaVersions(t *testing.T) {
	ctx := acctest.Context(t)
	var configuration1 kafka.DescribeConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_versions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, resourceName, &configuration1),
					resource.TestCheckResourceAttr(resourceName, "kafka_versions.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "kafka_versions.*", "2.6.0"),
					resource.TestCheckTypeSetElemAttr(resourceName, "kafka_versions.*", "2.7.0"),
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

func TestAccKafkaConfiguration_serverProperties(t *testing.T) {
	ctx := acctest.Context(t)
	var configuration1, configuration2 kafka.DescribeConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_configuration.test"
	serverProperty1 := "auto.create.topics.enable = false"
	serverProperty2 := "auto.create.topics.enable = true"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationConfig_serverProperties(rName, serverProperty1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, resourceName, &configuration1),
					resource.TestMatchResourceAttr(resourceName, "server_properties", regexache.MustCompile(serverProperty1)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationConfig_serverProperties(rName, serverProperty2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationExists(ctx, resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, "latest_revision", acctest.Ct2),
					resource.TestMatchResourceAttr(resourceName, "server_properties", regexache.MustCompile(serverProperty2)),
				),
			},
		},
	})
}

func testAccCheckConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_configuration" {
				continue
			}

			_, err := tfkafka.FindConfigurationByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MSK Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckConfigurationExists(ctx context.Context, n string, v *kafka.DescribeConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

		output, err := tfkafka.FindConfigurationByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  name = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
delete.topic.enable = true
PROPERTIES
}
`, rName)
}

func testAccConfigurationConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  description = %[2]q
  name        = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
PROPERTIES
}
`, rName, description)
}

func testAccConfigurationConfig_versions(rName string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  kafka_versions = ["2.6.0", "2.7.0"]
  name           = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
PROPERTIES
}
`, rName)
}

func testAccConfigurationConfig_serverProperties(rName string, serverProperty string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  name = %[1]q

  server_properties = <<PROPERTIES
%[2]s
PROPERTIES
}
`, rName, serverProperty)
}
