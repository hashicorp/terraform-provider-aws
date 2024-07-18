// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTTopicRuleDestination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_topic_rule_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleDestinationExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "iot", regexache.MustCompile(`ruledestination/vpc/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_configuration.0.role_arn"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.0.security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_configuration.0.vpc_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete everything but the IAM Role assumed by the IoT service.
			{
				Config: testAccTopicRuleConfig_destinationRole(rName),
			},
		},
	})
}

func TestAccIoTTopicRuleDestination_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_topic_rule_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleDestinationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiot.ResourceTopicRuleDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTTopicRuleDestination_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_topic_rule_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicRuleDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleDestinationConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicRuleDestinationConfig_enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				Config: testAccTopicRuleDestinationConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			// Delete everything but the IAM Role assumed by the IoT service.
			{
				Config: testAccTopicRuleConfig_destinationRole(rName),
			},
		},
	})
}

func testAccCheckTopicRuleDestinationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_topic_rule_destination" {
				continue
			}

			_, err := tfiot.FindTopicRuleDestinationByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Topic Rule Destination %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTopicRuleDestinationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Topic Rule Destination ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		_, err := tfiot.FindTopicRuleDestinationByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccTopicRuleDestinationBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		testAccTopicRuleConfig_destinationRole(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTopicRuleDestinationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTopicRuleDestinationBaseConfig(rName), `
resource "aws_iot_topic_rule_destination" "test" {
  vpc_configuration {
    role_arn        = aws_iam_role.test.arn
    security_groups = [aws_security_group.test.id]
    subnet_ids      = aws_subnet.test[*].id
    vpc_id          = aws_vpc.test.id
  }
}
`)
}

func testAccTopicRuleDestinationConfig_enabled(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccTopicRuleDestinationBaseConfig(rName), fmt.Sprintf(`
resource "aws_iot_topic_rule_destination" "test" {
  enabled = %[1]t

  vpc_configuration {
    role_arn        = aws_iam_role.test.arn
    security_groups = [aws_security_group.test.id]
    subnet_ids      = aws_subnet.test[*].id
    vpc_id          = aws_vpc.test.id
  }
}
`, enabled))
}
