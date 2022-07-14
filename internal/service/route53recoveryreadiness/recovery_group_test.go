package route53recoveryreadiness_test

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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53recoveryreadiness "github.com/hashicorp/terraform-provider-aws/internal/service/route53recoveryreadiness"
)

func TestAccRoute53RecoveryReadinessRecoveryGroup_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_recovery_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRecoveryGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRecoveryGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecoveryGroupExists(resourceName),
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

func TestAccRoute53RecoveryReadinessRecoveryGroup_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_recovery_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRecoveryGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRecoveryGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecoveryGroupExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53recoveryreadiness.ResourceRecoveryGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53RecoveryReadinessRecoveryGroup_nestedCell(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameCell := sdkacctest.RandomWithPrefix("tf-acc-test-cell")
	resourceName := "aws_route53recoveryreadiness_recovery_group.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRecoveryGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRecoveryGroupConfig_andCell(rName, rNameCell),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecoveryGroupExists(resourceName),
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

func TestAccRoute53RecoveryReadinessRecoveryGroup_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_recovery_group.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRecoveryGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRecoveryGroupConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecoveryGroupExists(resourceName),
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
				Config: testAccRecoveryGroupConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecoveryGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRecoveryGroupConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecoveryGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRoute53RecoveryReadinessRecoveryGroup_timeout(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_recovery_group.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRecoveryGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRecoveryGroupConfig_timeout(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecoveryGroupExists(resourceName),
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

func testAccCheckRecoveryGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryReadinessConn

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

func testAccCheckRecoveryGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryReadinessConn

		input := &route53recoveryreadiness.GetRecoveryGroupInput{
			RecoveryGroupName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetRecoveryGroup(input)

		return err
	}
}

func testAccRecoveryGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_recovery_group" "test" {
  recovery_group_name = %q
}
`, rName)
}

func testAccRecoveryGroupConfig_andCell(rName, rNameCell string) string {
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

func testAccRecoveryGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_recovery_group" "test" {
  recovery_group_name = %[1]q
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRecoveryGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccRecoveryGroupConfig_timeout(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_recovery_group" "test" {
  recovery_group_name = %q

  timeouts {
    delete = "10m"
  }
}
`, rName)
}
