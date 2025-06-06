// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/aws/aws-sdk-go-v2/service/controltower/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfcontroltower "github.com/hashicorp/terraform-provider-aws/internal/service/controltower"
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
// func TestBaselineExampleUnitTest(t *testing.T) {
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
// 			got, err := tfcontroltower.FunctionFromResource(testCase.Input)

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

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.
func TestAccControlTowerBaseline_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var baseline types.EnabledBaselineDetails
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_controltower_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ControlTowerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBaselineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaselineExists(ctx, resourceName, &baseline),
					resource.TestCheckResourceAttr(resourceName, "baseline_version", "4.0"),
					resource.TestCheckResourceAttrSet(resourceName, "baseline_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "arn", "controltower", regexache.MustCompile(`enabledbaseline:+.`)),
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

func TestAccControlTowerBaseline_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var baseline types.EnabledBaselineDetails
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_controltower_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ControlTowerEndpointID)
			testAccEnabledBaselinesPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBaselineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaselineExists(ctx, resourceName, &baseline),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcontroltower.ResourceBaseline, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBaselineDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ControlTowerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_controltower_baseline" {
				continue
			}

			input := &controltower.GetEnabledBaselineInput{
				EnabledBaselineIdentifier: aws.String(rs.Primary.Attributes[names.AttrARN]),
			}
			_, err := conn.GetEnabledBaseline(ctx, input)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ControlTower, create.ErrActionCheckingDestroyed, tfcontroltower.ResNameBaseline, rs.Primary.ID, err)
			}

			return create.Error(names.ControlTower, create.ErrActionCheckingDestroyed, tfcontroltower.ResNameBaseline, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckBaselineExists(ctx context.Context, name string, baseline *types.EnabledBaselineDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ControlTower, create.ErrActionCheckingExistence, tfcontroltower.ResNameBaseline, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ControlTower, create.ErrActionCheckingExistence, tfcontroltower.ResNameBaseline, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ControlTowerClient(ctx)
		resp, err := conn.GetEnabledBaseline(ctx, &controltower.GetEnabledBaselineInput{
			EnabledBaselineIdentifier: aws.String(rs.Primary.Attributes[names.AttrARN]),
		})

		if err != nil {
			return create.Error(names.ControlTower, create.ErrActionCheckingExistence, tfcontroltower.ResNameBaseline, rs.Primary.ID, err)
		}

		*baseline = *resp.EnabledBaselineDetails

		return nil
	}
}

func testAccEnabledBaselinesPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ControlTowerClient(ctx)

	input := &controltower.ListEnabledBaselinesInput{}
	_, err := conn.ListEnabledBaselines(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccBaselineConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

resource "aws_controltower_baseline" "test" {
  baseline_identifier             = "arn:aws:controltower:us-east-1::baseline/17BSJV3IGJ2QSGA2"
  baseline_version                = "4.0"
  target_identifier               = aws_organizations_organizational_unit.test.arn
  parameters {
  key = "IdentityCenterEnabledBaselineArn"
  value = "arn:aws:controltower:us-east-1:664418989480:enabledbaseline/XALULM96QHI525UOC"
  }
}
`, rName)
}
