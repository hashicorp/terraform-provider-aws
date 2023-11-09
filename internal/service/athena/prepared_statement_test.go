// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfathena "github.com/hashicorp/terraform-provider-aws/internal/service/athena"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAthenaPreparedStatement_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	resourceName := "aws_athena_prepared_statement.test"
	condition := "x = ?"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AthenaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreparedStatementDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPreparedStatementConfig_basic(rName, condition),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPreparedStatementExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrHasSuffix(resourceName, "query_statement", condition),
					resource.TestCheckResourceAttrSet(resourceName, "workgroup"),
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

func TestAccAthenaPreparedStatement_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	resourceName := "aws_athena_prepared_statement.test"
	condition := "x = ?"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AthenaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreparedStatementDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPreparedStatementConfig_basic(rName, condition),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreparedStatementExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfathena.ResourcePreparedStatement(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAthenaPreparedStatement_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	resourceName := "aws_athena_prepared_statement.test"
	condition := "x = ?"
	updatedCondition := "y = ?"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AthenaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreparedStatementDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPreparedStatementConfig_update(rName, condition, "desc1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreparedStatementExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "desc1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrHasSuffix(resourceName, "query_statement", condition),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPreparedStatementConfig_update(rName, updatedCondition, "desc2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreparedStatementExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "desc2"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrHasSuffix(resourceName, "query_statement", updatedCondition),
				),
			},
		},
	})
}

func testAccCheckPreparedStatementDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_athena_prepared_statement" {
				continue
			}

			_, err := tfathena.FindPreparedStatementByTwoPartKey(ctx, conn, rs.Primary.Attributes["workgroup"], rs.Primary.Attributes["name"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Athena Prepared Statement %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPreparedStatementExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)

		_, err := tfathena.FindPreparedStatementByTwoPartKey(ctx, conn, rs.Primary.Attributes["workgroup"], rs.Primary.Attributes["name"])

		return err
	}
}

func testAccPreparedStatementConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%[2]s-%[1]s"
  force_destroy = true
}

resource "aws_athena_workgroup" "test" {
  name = "%[2]s-%[1]s"
}

resource "aws_athena_database" "test" {
  name   = %[1]q
  bucket = aws_s3_bucket.test.bucket
}
`, rName, acctest.ResourcePrefix)
}

func testAccPreparedStatementConfig_basic(rName, condition string) string {
	return acctest.ConfigCompose(testAccPreparedStatementConfig_base(rName), fmt.Sprintf(`
resource "aws_athena_prepared_statement" "test" {
  name            = %[1]q
  query_statement = "SELECT * FROM ${aws_athena_database.test.name} WHERE %[2]s"
  workgroup       = aws_athena_workgroup.test.name
}
`, rName, condition))
}

func testAccPreparedStatementConfig_update(rName, condition, description string) string {
	return acctest.ConfigCompose(testAccPreparedStatementConfig_base(rName), fmt.Sprintf(`
resource "aws_athena_prepared_statement" "test" {
  name            = %[1]q
  description     = %[3]q
  query_statement = "SELECT * FROM ${aws_athena_database.test.name} WHERE %[2]s"
  workgroup       = aws_athena_workgroup.test.name
}
`, rName, condition, description))
}
