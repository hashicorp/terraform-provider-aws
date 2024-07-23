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

func TestAccAthenaNamedQuery_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_athena_named_query.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamedQueryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamedQueryConfig_basic(sdkacctest.RandInt(), sdkacctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamedQueryExists(ctx, resourceName),
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

func TestAccAthenaNamedQuery_withWorkGroup(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_athena_named_query.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamedQueryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamedQueryConfig_workGroup(sdkacctest.RandInt(), sdkacctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamedQueryExists(ctx, resourceName),
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

func testAccCheckNamedQueryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_athena_named_query" {
				continue
			}

			_, err := tfathena.FindNamedQueryByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Athena Named Query %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckNamedQueryExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)

		_, err := tfathena.FindNamedQueryByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccNamedQueryConfig_base(rInt int, rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%[3]s-%[1]s-%[2]d"
  force_destroy = true
}

resource "aws_athena_database" "test" {
  name   = %[1]q
  bucket = aws_s3_bucket.test.bucket
}
`, rName, rInt, acctest.ResourcePrefix)
}

func testAccNamedQueryConfig_basic(rInt int, rName string) string {
	return acctest.ConfigCompose(testAccNamedQueryConfig_base(rInt, rName), fmt.Sprintf(`
resource "aws_athena_named_query" "test" {
  name        = "%[2]s-%[1]s"
  database    = aws_athena_database.test.name
  query       = "SELECT * FROM ${aws_athena_database.test.name} limit 10;"
  description = "tf test"
}
`, rName, acctest.ResourcePrefix))
}

func testAccNamedQueryConfig_workGroup(rInt int, rName string) string {
	return acctest.ConfigCompose(testAccNamedQueryConfig_base(rInt, rName), fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = "%[3]s-%[1]s-%[2]d"
}

resource "aws_athena_named_query" "test" {
  name        = "%[3]s-%[1]s"
  workgroup   = aws_athena_workgroup.test.id
  database    = aws_athena_database.test.name
  query       = "SELECT * FROM ${aws_athena_database.test.name} limit 10;"
  description = "tf test"
}
`, rName, rInt, acctest.ResourcePrefix))
}
