package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAwsRoute53RecoveryReadinessRecoveryGroup_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_recovery_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessRecoveryGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupExists(resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "route53-recovery-readiness", regexp.MustCompile(`recovery-group/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cells.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAwsRoute53RecoveryReadinessRecoveryGroup_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_recovery_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessRecoveryGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, resourceAwsRoute53RecoveryReadinessRecoveryGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsRoute53RecoveryReadinessRecoveryGroup_nestedCell(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rNameCell := sdkacctest.RandomWithPrefix("tf-acc-test-cell")
	resourceName := "aws_route53recoveryreadiness_recovery_group.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessRecoveryGroupAndCellConfig(rName, rNameCell),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupExists(resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "route53-recovery-readiness", regexp.MustCompile(`recovery-group/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cells.#", "1"),
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

func TestAccAwsRoute53RecoveryReadinessRecoveryGroup_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_recovery_group.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessRecoveryGroupConfig_Tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupExists(resourceName),
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
				Config: testAccAwsRoute53RecoveryReadinessRecoveryGroupConfig_Tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsRoute53RecoveryReadinessRecoveryGroupConfig_Tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAwsRoute53RecoveryReadinessRecoveryGroup_timeout(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_recovery_group.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessRecoveryGroupConfig_Timeout(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupExists(resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "route53-recovery-readiness", regexp.MustCompile(`recovery-group/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cells.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*AWSClient).route53recoveryreadinessconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53recoveryreadiness_recovery_group" {
			continue
		}

		input := &route53recoveryreadiness.GetRecoveryGroupInput{
			RecoveryGroupName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetRecoveryGroup(input)
		if err == nil {
			return fmt.Errorf("Route53RecoveryReadiness Recovery Group (%s) not deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsRoute53RecoveryReadinessRecoveryGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*AWSClient).route53recoveryreadinessconn

		input := &route53recoveryreadiness.GetRecoveryGroupInput{
			RecoveryGroupName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetRecoveryGroup(input)

		return err
	}
}

func testAccAwsRoute53RecoveryReadinessRecoveryGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_recovery_group" "test" {
  recovery_group_name = %q
}
`, rName)
}

func testAccAwsRoute53RecoveryReadinessRecoveryGroupAndCellConfig(rName, rNameCell string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = %[2]q
}

resource "aws_route53recoveryreadiness_recovery_group" "test" {
  recovery_group_name = %[1]q
  cells               = [aws_route53recoveryreadiness_cell.test.arn]
}
`, rName, rNameCell)
}

func testAccAwsRoute53RecoveryReadinessRecoveryGroupConfig_Tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_recovery_group" "test" {
  recovery_group_name = %[1]q
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAwsRoute53RecoveryReadinessRecoveryGroupConfig_Tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_recovery_group" "test" {
  recovery_group_name = %[1]q
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAwsRoute53RecoveryReadinessRecoveryGroupConfig_Timeout(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_recovery_group" "test" {
  recovery_group_name = %q

  timeouts {
    delete = "10m"
  }
}
`, rName)
}
