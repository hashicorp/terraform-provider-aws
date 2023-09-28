// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

//import (
//	"context"
//	"fmt"
//	"testing"
//
//	"github.com/YakDriver/regexache"
//
//	"github.com/aws/aws-sdk-go/aws"
//	"github.com/aws/aws-sdk-go/service/dynamodb"
//	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
//	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
//	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
//	"github.com/hashicorp/terraform-plugin-testing/terraform"
//	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
//	"github.com/hashicorp/terraform-provider-aws/internal/conns"
//	"github.com/hashicorp/terraform-provider-aws/internal/create"
//	"github.com/hashicorp/terraform-provider-aws/internal/errs"
//	"github.com/hashicorp/terraform-provider-aws/names"
//	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
//)
//
//func TestImportTableExampleUnitTest(t *testing.T) {
//	t.Parallel()
//
//	testCases := []struct {
//		TestName string
//		Input    string
//		Expected string
//		Error    bool
//	}{
//		{
//			TestName: "empty",
//			Input:    "",
//			Expected: "",
//			Error:    true,
//		},
//		{
//			TestName: "descriptive name",
//			Input:    "some input",
//			Expected: "some output",
//			Error:    false,
//		},
//		{
//			TestName: "another descriptive name",
//			Input:    "more input",
//			Expected: "more output",
//			Error:    false,
//		},
//	}
//
//	for _, testCase := range testCases {
//		testCase := testCase
//		t.Run(testCase.TestName, func(t *testing.T) {
//			t.Parallel()
//			got, err := tfdynamodb.FunctionFromResource(testCase.Input)
//
//			if err != nil && !testCase.Error {
//				t.Errorf("got error (%s), expected no error", err)
//			}
//
//			if err == nil && testCase.Error {
//				t.Errorf("got (%s) and no error, expected error", got)
//			}
//
//			if got != testCase.Expected {
//				t.Errorf("got %s, expected %s", got, testCase.Expected)
//			}
//		})
//	}
//}
//
//func TestAccDynamoDBImportTable_basic(t *testing.T) {
//	ctx := acctest.Context(t)
//	if testing.Short() {
//		t.Skip("skipping long-running test in short mode")
//	}
//
//	var importtable dynamodb.DescribeImportTableResponse
//	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//	resourceName := "aws_dynamodb_import_table.test"
//
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck: func() {
//			acctest.PreCheck(ctx, t)
//			acctest.PreCheckPartitionHasService(t, dynamodb.EndpointsID)
//			testAccPreCheck(ctx, t)
//		},
//		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
//		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
//		CheckDestroy:             testAccCheckImportTableDestroy(ctx),
//		Steps: []resource.TestStep{
//			{
//				Config: testAccImportTableConfig_basic(rName),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckImportTableExists(ctx, resourceName, &importtable),
//					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
//					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
//					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
//						"console_access": "false",
//						"groups.#":       "0",
//						"username":       "Test",
//						"password":       "TestTest1234",
//					}),
//					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "dynamodb", regexache.MustCompile(`importtable:+.`)),
//				),
//			},
//			{
//				ResourceName:            resourceName,
//				ImportState:             true,
//				ImportStateVerify:       true,
//				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
//			},
//		},
//	})
//}
//
//func TestAccDynamoDBImportTable_disappears(t *testing.T) {
//	ctx := acctest.Context(t)
//	if testing.Short() {
//		t.Skip("skipping long-running test in short mode")
//	}
//
//	var importtable dynamodb.DescribeImportTableResponse
//	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//	resourceName := "aws_dynamodb_import_table.test"
//
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck: func() {
//			acctest.PreCheck(ctx, t)
//			acctest.PreCheckPartitionHasService(t, dynamodb.EndpointsID)
//			testAccPreCheck(t)
//		},
//		ErrorCheck:               acctest.ErrorCheck(t, dynamodb.EndpointsID),
//		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
//		CheckDestroy:             testAccCheckImportTableDestroy(ctx),
//		Steps: []resource.TestStep{
//			{
//				Config: testAccImportTableConfig_basic(rName, testAccImportTableVersionNewer),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckImportTableExists(ctx, resourceName, &importtable),
//					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdynamodb.ResourceImportTable, resourceName),
//				),
//				ExpectNonEmptyPlan: true,
//			},
//		},
//	})
//}
//
//func testAccCheckImportTableDestroy(ctx context.Context) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn(ctx)
//
//		for _, rs := range s.RootModule().Resources {
//			if rs.Type != "aws_dynamodb_import_table" {
//				continue
//			}
//
//			input := &dynamodb.DescribeImportTableInput{
//				ImportTableId: aws.String(rs.Primary.ID),
//			}
//			_, err := conn.DescribeImportTableWithContext(ctx, &dynamodb.DescribeImportTableInput{
//				ImportTableId: aws.String(rs.Primary.ID),
//			})
//			if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeNotFoundException) {
//				return nil
//			}
//			if err != nil {
//				return nil
//			}
//
//			return create.Error(names.DynamoDB, create.ErrActionCheckingDestroyed, tfdynamodb.ResNameImportTable, rs.Primary.ID, errors.New("not destroyed"))
//		}
//
//		return nil
//	}
//}
//
//func testAccCheckImportTableExists(ctx context.Context, name string, importtable *dynamodb.DescribeImportTableResponse) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		rs, ok := s.RootModule().Resources[name]
//		if !ok {
//			return create.Error(names.DynamoDB, create.ErrActionCheckingExistence, tfdynamodb.ResNameImportTable, name, errors.New("not found"))
//		}
//
//		if rs.Primary.ID == "" {
//			return create.Error(names.DynamoDB, create.ErrActionCheckingExistence, tfdynamodb.ResNameImportTable, name, errors.New("not set"))
//		}
//
//		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn(ctx)
//		resp, err := conn.DescribeImportTableWithContext(ctx, &dynamodb.DescribeImportTableInput{
//			ImportTableId: aws.String(rs.Primary.ID),
//		})
//
//		if err != nil {
//			return create.Error(names.DynamoDB, create.ErrActionCheckingExistence, tfdynamodb.ResNameImportTable, rs.Primary.ID, err)
//		}
//
//		*importtable = *resp
//
//		return nil
//	}
//}
//
//func testAccPreCheck(ctx context.Context, t *testing.T) {
//	conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn(ctx)
//
//	input := &dynamodb.ListImportTablesInput{}
//	_, err := conn.ListImportTablesWithContext(ctx, input)
//
//	if acctest.PreCheckSkipError(err) {
//		t.Skipf("skipping acceptance testing: %s", err)
//	}
//	if err != nil {
//		t.Fatalf("unexpected PreCheck error: %s", err)
//	}
//}
//
//func testAccCheckImportTableNotRecreated(before, after *dynamodb.DescribeImportTableResponse) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		if before, after := aws.StringValue(before.ImportTableId), aws.StringValue(after.ImportTableId); before != after {
//			return create.Error(names.DynamoDB, create.ErrActionCheckingNotRecreated, tfdynamodb.ResNameImportTable, aws.StringValue(before.ImportTableId), errors.New("recreated"))
//		}
//
//		return nil
//	}
//}
//
//func testAccImportTableConfig_basic(rName, version string) string {
//	return fmt.Sprintf(`
//resource "aws_security_group" "test" {
//  name = %[1]q
//}
//
//resource "aws_dynamodb_import_table" "test" {
//  import_table_name             = %[1]q
//  engine_type             = "ActiveDynamoDB"
//  engine_version          = %[2]q
//  host_instance_type      = "dynamodb.t2.micro"
//  security_groups         = [aws_security_group.test.id]
//  authentication_strategy = "simple"
//  storage_type            = "efs"
//
//  logs {
//    general = true
//  }
//
//  user {
//    username = "Test"
//    password = "TestTest1234"
//  }
//}
//`, rName, version)
//}
