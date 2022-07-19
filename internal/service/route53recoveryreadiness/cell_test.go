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

func TestAccRoute53RecoveryReadinessCell_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_cell.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCellConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(resourceName),
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

func TestAccRoute53RecoveryReadinessCell_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_cell.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCellConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53recoveryreadiness.ResourceCell(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53RecoveryReadinessCell_nestedCell(t *testing.T) {
	rNameParent := sdkacctest.RandomWithPrefix("tf-acc-test-parent")
	rNameChild := sdkacctest.RandomWithPrefix("tf-acc-test-child")
	resourceNameParent := "aws_route53recoveryreadiness_cell.test_parent"
	resourceNameChild := "aws_route53recoveryreadiness_cell.test_child"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCellConfig_child(rNameChild),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(resourceNameChild),
					acctest.MatchResourceAttrGlobalARN(resourceNameChild, "arn", "route53-recovery-readiness", regexp.MustCompile(`cell/.+`)),
				),
			},
			{
				Config: testAccCellConfig_parent(rNameChild, rNameParent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(resourceNameParent),
					acctest.MatchResourceAttrGlobalARN(resourceNameParent, "arn", "route53-recovery-readiness", regexp.MustCompile(`cell/.+`)),
					resource.TestCheckResourceAttr(resourceNameParent, "cells.#", "1"),
					resource.TestCheckResourceAttr(resourceNameParent, "parent_readiness_scopes.#", "0"),
					testAccCheckCellExists(resourceNameChild),
					acctest.MatchResourceAttrGlobalARN(resourceNameChild, "arn", "route53-recovery-readiness", regexp.MustCompile(`cell/.+`)),
					resource.TestCheckResourceAttr(resourceNameChild, "cells.#", "0"),
				),
			},
			{
				Config: testAccCellConfig_parent(rNameChild, rNameParent),
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

func TestAccRoute53RecoveryReadinessCell_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_cell.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCellConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(resourceName),
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
				Config: testAccCellConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCellConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRoute53RecoveryReadinessCell_timeout(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_cell.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCellConfig_timeout(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCellExists(resourceName),
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

func testAccCheckCellDestroy(s *terraform.State) error {
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

func testAccCheckCellExists(name string) resource.TestCheckFunc {
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

func testAccPreCheck(t *testing.T) {
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

func testAccCellConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = %q
}
`, rName)
}

func testAccCellConfig_child(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test_child" {
  cell_name = %q
}
`, rName)
}

func testAccCellConfig_parent(rName, rName2 string) string {
	return acctest.ConfigCompose(testAccCellConfig_child(rName), fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test_parent" {
  cell_name = %q
  cells     = [aws_route53recoveryreadiness_cell.test_child.arn]
}
`, rName2))
}

func testAccCellConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccCellConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccCellConfig_timeout(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_cell" "test" {
  cell_name = %q

  timeouts {
    delete = "10m"
  }
}
`, rName)
}
