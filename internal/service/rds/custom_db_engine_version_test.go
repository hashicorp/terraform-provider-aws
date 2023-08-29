// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations

// TIP: ==== UNIT TESTS ====
// This is an example of a unit test. Its name is not prefixed with
// "TestAcc" like an acceptance test.
//
// Unlike acceptance tests, unit tests do not access AWS and are focused on a
// function (or method). Because of this, they are quick and cheap to run.
//
// In designing a resource's implementation, isolate complex bits from AWS bits
// so that they can be tested through a unit test. We encourage more unit tests
// in the provider.
//
// Cut and dry functions using well-used patterns, like typical flatteners and
// expanders, don't need unit testing. However, if they are complex or
// intricate, they should be unit tested.
// func TestCustomDBEngineVersionExampleUnitTest(t *testing.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		TestName string
// 		Input    string
// 		Expected string
// 		Error    bool
// 	}{
// 		{
// 			TestName: "empty",
// 			Input:    "",
// 			Expected: "",
// 			Error:    true,
// 		},
// 		{
// 			TestName: "descriptive name",
// 			Input:    "some input",
// 			Expected: "some output",
// 			Error:    false,
// 		},
// 		{
// 			TestName: "another descriptive name",
// 			Input:    "more input",
// 			Expected: "more output",
// 			Error:    false,
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase
// 		t.Run(testCase.TestName, func(t *testing.T) {
// 			t.Parallel()
// 			got, err := tfrds.FunctionFromResource(testCase.Input)

// 			if err != nil && !testCase.Error {
// 				t.Errorf("got error (%s), expected no error", err)
// 			}

// 			if err == nil && testCase.Error {
// 				t.Errorf("got (%s) and no error, expected error", got)
// 			}

// 			if got != testCase.Expected {
// 				t.Errorf("got %s, expected %s", got, testCase.Expected)
// 			}
// 		})
// 	}
// }

func TestAccRDSCustomDBEngineVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var customdbengineversion rds.DBEngineVersion
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_custom_db_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, rds.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDBEngineVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDBEngineVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDBEngineVersionExists(ctx, resourceName, &customdbengineversion),
					resource.TestCheckResourceAttr(resourceName, "engine_version", rName),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`customdbengineversion:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "manifest_hash"},
			},
		},
	})
}

func TestAccRDSCustomDBEngineVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var customdbengineversion rds.DBEngineVersion
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_custom_db_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, rds.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDBEngineVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDBEngineVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDBEngineVersionExists(ctx, resourceName, &customdbengineversion),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceCustomDBEngineVersion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCustomDBEngineVersionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_custom_db_engine_version" {
				continue
			}

			_, err := tfrds.FindCustomDBEngineVersionByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.RDS, create.ErrActionCheckingDestroyed, tfrds.ResNameCustomDBEngineVersion, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckCustomDBEngineVersionExists(ctx context.Context, name string, customdbengineversion *rds.DBEngineVersion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.RDS, create.ErrActionCheckingExistence, tfrds.ResNameCustomDBEngineVersion, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.RDS, create.ErrActionCheckingExistence, tfrds.ResNameCustomDBEngineVersion, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		output, err := tfrds.FindCustomDBEngineVersionByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.RDS, create.ErrActionCheckingExistence, tfrds.ResNameCustomDBEngineVersion, rs.Primary.ID, err)
		}

		*customdbengineversion = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DescribeDBEngineVersionsInput{}
	_, err := conn.DescribeDBEngineVersionsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCustomDBEngineVersionConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["Windows_Server-2022-English-Full-SQL_2019_Standard-2023.08.10"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "block-device-mapping.volume-type"
    values = ["gp2"]
  }
}

resource "aws_kms_key" "rdscfo_kms_key" {
  description = "KMS symmetric key for RDS Custom for Oracle"
}

resource "aws_rds_custom_db_engine_version" "test" {
  engine         = "custom-sqlserver-se"
  engine_version = %[1]q
  image-id       = aws_ami.test.id
  kms_key_id     = aws_kms_key.rdscfo_kms_key.key_id
}
`, rName)
}
