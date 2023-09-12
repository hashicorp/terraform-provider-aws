// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfathena "github.com/hashicorp/terraform-provider-aws/internal/service/athena"
)

func TestAccAthenaPreparedStatement_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var preparedstatement athena.GetPreparedStatementOutput
	rName := sdkacctest.RandString(5)
	resourceName := "aws_athena_prepared_statement.test"
	condition := "x = ?"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, athena.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, athena.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreparedStatementDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPreparedStatementConfig_basic(rName, condition),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreparedStatementExists(ctx, resourceName, &preparedstatement),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPreparedStatementResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAthenaPreparedStatement_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var preparedstatement1, preparedstatement2 athena.GetPreparedStatementOutput
	rName := sdkacctest.RandString(5)
	resourceName := "aws_athena_prepared_statement.test"
	condition := "x = ?"
	updated_condition := "y = ?"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, athena.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, athena.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreparedStatementDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPreparedStatementConfig_basic(rName, condition),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreparedStatementExists(ctx, resourceName, &preparedstatement1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPreparedStatementResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccPreparedStatementConfig_basic(rName, updated_condition),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreparedStatementExists(ctx, resourceName, &preparedstatement2),
					testAccCheckPreparedStatementNotRecreated(&preparedstatement1, &preparedstatement2),
				),
			},
		},
	})
}

func TestAccAthenaPreparedStatement_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var preparedstatement athena.GetPreparedStatementOutput
	rName := sdkacctest.RandString(5)
	resourceName := "aws_athena_prepared_statement.test"
	condition := "x = ?"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, athena.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, athena.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreparedStatementDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPreparedStatementConfig_basic(rName, condition),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreparedStatementExists(ctx, resourceName, &preparedstatement),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfathena.ResourcePreparedStatement(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPreparedStatementDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_athena_prepared_statement" {
				continue
			}

			input := &athena.GetPreparedStatementInput{
				StatementName: aws.String(rs.Primary.Attributes["name"]),
				WorkGroup:     aws.String(rs.Primary.Attributes["workgroup"]),
			}
			_, err := conn.GetPreparedStatementWithContext(ctx, input)
			if tfawserr.ErrCodeEquals(err, athena.ErrCodeResourceNotFoundException) {
				return nil
			}
			if err != nil {
				return nil
			}

			return create.Error(names.Athena, create.ErrActionCheckingDestroyed, tfathena.ResNamePreparedStatement, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckPreparedStatementExists(ctx context.Context, name string, preparedstatement *athena.GetPreparedStatementOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Athena, create.ErrActionCheckingExistence, tfathena.ResNamePreparedStatement, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Athena, create.ErrActionCheckingExistence, tfathena.ResNamePreparedStatement, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn(ctx)
		resp, err := conn.GetPreparedStatementWithContext(ctx, &athena.GetPreparedStatementInput{
			StatementName: aws.String(rs.Primary.Attributes["name"]),
			WorkGroup:     aws.String(rs.Primary.Attributes["workgroup"]),
		})

		if err != nil {
			return create.Error(names.Athena, create.ErrActionCheckingExistence, tfathena.ResNamePreparedStatement, rs.Primary.ID, err)
		}

		*preparedstatement = *resp

		return nil
	}
}

func testAccPreparedStatementResourceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["workgroup"], rs.Primary.ID), nil
	}
}

func testAccCheckPreparedStatementNotRecreated(before, after *athena.GetPreparedStatementOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.PreparedStatement.StatementName), aws.StringValue(after.PreparedStatement.StatementName); before != after {
			return create.Error(names.Athena, create.ErrActionCheckingNotRecreated, tfathena.ResNamePreparedStatement, before, errors.New("recreated"))
		}
		if before, after := aws.StringValue(before.PreparedStatement.WorkGroupName), aws.StringValue(after.PreparedStatement.WorkGroupName); before != after {
			return create.Error(names.Athena, create.ErrActionCheckingNotRecreated, tfathena.ResNamePreparedStatement, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccPreparedStatementConfig_basic(rName, rCondition string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "tf-test-athena-bucket-%s"
  force_destroy = true
}

resource "aws_athena_workgroup" "test" {
  name = "tf-athena-workgroup-%s"
}

resource "aws_athena_database" "test" {
  name   = "%s"
  bucket = aws_s3_bucket.test.bucket
}

resource "aws_athena_prepared_statement" "test" {
  name            = "tf_athena_prepared_statement_%s"
  query_statement = "SELECT * FROM ${aws_athena_database.test.name} WHERE %s" 
  workgroup       = aws_athena_workgroup.test.name
}
`, rName, rName, rName, rName, rCondition)
}
