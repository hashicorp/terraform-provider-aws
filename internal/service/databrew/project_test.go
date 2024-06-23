// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package databrew_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databrew"
	"github.com/aws/aws-sdk-go-v2/service/databrew/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfdatabrew "github.com/hashicorp/terraform-provider-aws/internal/service/databrew"
)

// func TestProjectExampleUnitTest(t *testing.T) {
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
// 			got, err := tfdatabrew.FunctionFromResource(testCase.Input)

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

// Acceptance test access AWS and cost money to run.
func TestAccDataBrewProject_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var project databrew.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_databrew_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataBrewServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataBrewServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "databrew", regexache.MustCompile(`project:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccDataBrewProject_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var project databrew.DescribeProjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_databrew_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataBrewServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataBrewServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(ctx, resourceName, &project),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceProject = newResourceProject
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatabrew.ResourceDomain, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProjectDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataBrewClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_databrew_project" {
				continue
			}

			_, err := conn.DescribeProject(ctx, &databrew.DescribeProjectInput{
				Name: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DataBrew, create.ErrActionCheckingDestroyed, tfdatabrew.ResNameProject, rs.Primary.ID, err)
			}

			return create.Error(names.DataBrew, create.ErrActionCheckingDestroyed, tfdatabrew.ResNameProject, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckProjectExists(ctx context.Context, name string, project *databrew.DescribeProjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataBrew, create.ErrActionCheckingExistence, tfdatabrew.ResNameProject, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataBrew, create.ErrActionCheckingExistence, tfdatabrew.ResNameProject, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataBrewClient(ctx)
		resp, err := conn.DescribeProject(ctx, &databrew.DescribeProjectInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.DataBrew, create.ErrActionCheckingExistence, tfdatabrew.ResNameProject, rs.Primary.ID, err)
		}

		*project = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataBrewClient(ctx)

	input := &databrew.ListProjectsInput{}
	_, err := conn.ListProjects(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckProjectNotRecreated(before, after *databrew.DescribeProjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Name), aws.ToString(after.Name); before != after {
			return create.Error(names.DataBrew, create.ErrActionCheckingNotRecreated, tfdatabrew.ResNameProject, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccProjectConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_databrew_project" "test" {
  project_name = %[1]q
  dataset_name = "test"
  recipe_name  = "test"
  role_arn     = "test"
}
`, rName)
}
