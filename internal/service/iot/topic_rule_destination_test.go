package iot_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIoTTopicRuleDestination_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_topic_rule_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleDestinationExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "iot", regexp.MustCompile(`ruledestination/vpc/.+`)),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_configuration.0.role_arn"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.0.subnet_ids.#", "2"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_topic_rule_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleDestinationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfiot.ResourceTopicRuleDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTTopicRuleDestination_enabled(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_topic_rule_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicRuleDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicRuleDestinationConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleDestinationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
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
					testAccCheckTopicRuleDestinationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccTopicRuleDestinationConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicRuleDestinationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			// Delete everything but the IAM Role assumed by the IoT service.
			{
				Config: testAccTopicRuleConfig_destinationRole(rName),
			},
		},
	})
}

func testAccCheckTopicRuleDestinationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_topic_rule_destination" {
			continue
		}

		_, err := tfiot.FindTopicRuleDestinationByARN(context.TODO(), conn, rs.Primary.ID)

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

func testAccCheckTopicRuleDestinationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Topic Rule Destination ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

		_, err := tfiot.FindTopicRuleDestinationByARN(context.TODO(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
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
