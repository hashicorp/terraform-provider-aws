// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackage_test

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/mediapackage/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediapackage"
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
	tfmediapackage "github.com/hashicorp/terraform-provider-aws/internal/service/mediapackage"
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
func TestOriginEndpointExampleUnitTest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Input    string
		Expected string
		Error    bool
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: "",
			Error:    true,
		},
		{
			TestName: "descriptive name",
			Input:    "some input",
			Expected: "some output",
			Error:    false,
		},
		{
			TestName: "another descriptive name",
			Input:    "more input",
			Expected: "more output",
			Error:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()
			got, err := tfmediapackage.FunctionFromResource(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (%s) and no error, expected error", got)
			}

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.
func TestAccMediaPackageOriginEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var originendpoint mediapackage.DescribeOriginEndpointResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediapackage_origin_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaPackageOriginEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginEndpointConfig_basic(rName, "1.0.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOriginEndpointExists(ctx, resourceName, &originEndpoint),
					resource.TestCheckResourceAttr(resourceName, "origin_endpoint_name", rName),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMediaPackage"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mediapackage.t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "authentication_strategy", "simple"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "efs"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "true"),
					resource.TestCheckResourceAttr(resourceName, "user.0.username", "Test"),
					resource.TestCheckResourceAttr(resourceName, "user.0.password", "TestTest1234"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "mediapackage", regexache.MustCompile(`originendpoint:.+$`)),
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

func TestAccMediaPackageOriginEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var originendpoint mediapackage.DescribeOriginEndpointResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediapackage_origin_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaPackageEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaPackageServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginEndpointConfig_basic(rName, testAccOriginEndpointVersionNewer),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOriginEndpointExists(ctx, resourceName, &originendpoint),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmediapackage.ResourceOriginEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckOriginEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mediapackage_origin_endpoint" {
				continue
			}

			// TIP: ==== FINDERS ====
			// The find function should be exported. Since it won't be used outside of the package, it can be exported
			// in the `exports_test.go` file.
			_, err := tfmediapackage.FindOriginEndpointByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.MediaPackage, create.ErrActionCheckingDestroyed, tfmediapackage.ResNameOriginEndpoint, rs.Primary.ID, err)
			}

			return create.Error(names.MediaPackage, create.ErrActionCheckingDestroyed, tfmediapackage.ResNameOriginEndpoint, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckOriginEndpointExists(ctx context.Context, name string, originendpoint *mediapackage.DescribeOriginEndpointResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MediaPackage, create.ErrActionCheckingExistence, tfmediapackage.ResNameOriginEndpoint, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MediaPackage, create.ErrActionCheckingExistence, tfmediapackage.ResNameOriginEndpoint, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageClient(ctx)

		resp, err := tfmediapackage.FindOriginEndpointByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.MediaPackage, create.ErrActionCheckingExistence, tfmediapackage.ResNameOriginEndpoint, rs.Primary.ID, err)
		}

		*originendpoint = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageClient(ctx)

	input := &mediapackage.ListOriginEndpointsInput{}

	_, err := conn.ListOriginEndpoints(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckOriginEndpointNotRecreated(before, after *mediapackage.DescribeOriginEndpointResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.OriginEndpointId), aws.ToString(after.OriginEndpointId); before != after {
			return create.Error(names.MediaPackage, create.ErrActionCheckingNotRecreated, tfmediapackage.ResNameOriginEndpoint, aws.ToString(before.OriginEndpointId), errors.New("recreated"))
		}

		return nil
	}
}

func testAccOriginEndpointConfig_basic(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mediapackage_origin_endpoint" "test" {
  origin_endpoint_name             = %[1]q
  engine_type             = "ActiveMediaPackage"
  engine_version          = %[2]q
  host_instance_type      = "mediapackage.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version)
}
