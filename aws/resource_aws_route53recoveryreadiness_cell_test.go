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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAwsRoute53RecoveryReadinessCell_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_cell.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessCellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessCellConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessCellExists(resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "route53-recovery-readiness", regexp.MustCompile(`cell/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cells.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parent_readiness_scopes.#", "0"),
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

func TestAccAwsRoute53RecoveryReadinessCell_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_cell.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessCellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessCellConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessCellExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, resourceAwsRoute53RecoveryReadinessCell(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsRoute53RecoveryReadinessCell_nestedCell(t *testing.T) {
	rNameParent := sdkacctest.RandomWithPrefix("tf-acc-test-parent")
	rNameChild := sdkacctest.RandomWithPrefix("tf-acc-test-child")
	resourceNameParent := "aws_route53recoveryreadiness_cell.test_parent"
	resourceNameChild := "aws_route53recoveryreadiness_cell.test_child"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessCellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessCellChildConfig(rNameChild),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessCellExists(resourceNameChild),
					acctest.MatchResourceAttrGlobalARN(resourceNameChild, "arn", "route53-recovery-readiness", regexp.MustCompile(`cell/.+`)),
				),
			},
			{
				Config: testAccAwsRoute53RecoveryReadinessCellParentConfig(rNameChild, rNameParent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessCellExists(resourceNameParent),
					acctest.MatchResourceAttrGlobalARN(resourceNameParent, "arn", "route53-recovery-readiness", regexp.MustCompile(`cell/.+`)),
					resource.TestCheckResourceAttr(resourceNameParent, "cells.#", "1"),
					resource.TestCheckResourceAttr(resourceNameParent, "parent_readiness_scopes.#", "0"),
					testAccCheckAwsRoute53RecoveryReadinessCellExists(resourceNameChild),
					acctest.MatchResourceAttrGlobalARN(resourceNameChild, "arn", "route53-recovery-readiness", regexp.MustCompile(`cell/.+`)),
					resource.TestCheckResourceAttr(resourceNameChild, "cells.#", "0"),
				),
			},
			{
				Config: testAccAwsRoute53RecoveryReadinessCellParentConfig(rNameChild, rNameParent),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceNameChild, "parent_readiness_scopes.#", "1"),
				),
			},
			{
				ResourceName:      resourceNameParent,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceNameChild,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsRoute53RecoveryReadinessCell_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_cell.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessCellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessCellConfig_Tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessCellExists(resourceName),
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
				Config: testAccAwsRoute53RecoveryReadinessCellConfig_Tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessCellExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsRoute53RecoveryReadinessCellConfig_Tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessCellExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAwsRoute53RecoveryReadinessCell_timeout(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_cell.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessCellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessCellConfig_Timeout(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessCellExists(resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "route53-recovery-readiness", regexp.MustCompile(`cell/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cells.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parent_readiness_scopes.#", "0"),
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

func testAccCheckAwsRoute53RecoveryReadinessCellDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryReadinessConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53recoveryreadiness_cell" {
			continue
		}

		input := &route53recoveryreadiness.GetCellInput{
			CellName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetCell(input)
		if err == nil {
			return fmt.Errorf("Route53RecoveryReadiness Channel (%s) not deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsRoute53RecoveryReadinessCellExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryReadinessConn

		input := &route53recoveryreadiness.GetCellInput{
			CellName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetCell(input)

		return err
	}
}

func testAccPreCheckAwsRoute53RecoveryReadiness(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryReadinessConn

	input := &route53recoveryreadiness.ListCellsInput{}

	_, err := conn.ListCells(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAwsRoute53RecoveryReadinessCellConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = %q
}
`, rName)
}

func testAccAwsRoute53RecoveryReadinessCellChildConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test_child" {
  cell_name = %q
}
`, rName)
}

func testAccAwsRoute53RecoveryReadinessCellParentConfig(rName, rName2 string) string {
	return acctest.ConfigCompose(testAccAwsRoute53RecoveryReadinessCellChildConfig(rName), fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test_parent" {
  cell_name = %q
  cells     = [aws_route53recoveryreadiness_cell.test_child.arn]
}
`, rName2))
}

func testAccAwsRoute53RecoveryReadinessCellConfig_Tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAwsRoute53RecoveryReadinessCellConfig_Tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAwsRoute53RecoveryReadinessCellConfig_Timeout(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = %q

  timeouts {
    delete = "10m"
  }
}
`, rName)
}
