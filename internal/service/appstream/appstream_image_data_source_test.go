// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream_test

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
	// using the services/appstream/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.

	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this data source's maintainability by following this
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
// In designing a data source's implementation, isolate complex bits from AWS bits
// so that they can be tested through a unit test. We encourage more unit tests
// in the provider.
//
// Cut and dry functions using well-used patterns, like typical flatteners and
// expanders, don't need unit testing. However, if they are complex or
// intricate, they should be unit tested.
func TestAccAppStreamAppstreamImageDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	//var describeimages appstream.DescribeDescribeImagesResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appstream_image_builder"
	dataSourceName := "data.aws_appstream_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppStreamEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppstreamImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppstreamImageDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(

					resource.TestCheckResourceAttrPair(resourceName, "applications", dataSourceName, "applications"), // this is a list of lists
					resource.TestCheckResourceAttrPair(resourceName, "app_stream_agent_version", dataSourceName, "app_stream_agent_version"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "base_image_arn", dataSourceName, "base_image_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_time", dataSourceName, "created_time"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "display_name", dataSourceName, "display_name"),
					resource.TestCheckResourceAttrPair(resourceName, "image_builder_name", dataSourceName, "image_builder_name"),
					resource.TestCheckResourceAttrPair(resourceName, "image_builder_supported", dataSourceName, "image_builder_support"),
					resource.TestCheckResourceAttrPair(resourceName, "image_errors", dataSourceName, "image_errors"),
					resource.TestCheckResourceAttrPair(resourceName, "image_permissions.#", dataSourceName, "image_permissions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "platform", dataSourceName, "platform"),
					resource.TestCheckResourceAttrPair(resourceName, "public_base_image_released_date", dataSourceName, "public_base_image_released_date"),
					resource.TestCheckResourceAttrPair(resourceName, "state_change_reason", dataSourceName, "state_change_reason"),
					resource.TestCheckResourceAttrPair(resourceName, "visibility", dataSourceName, "visibility"),
				),
			},
		},
	})
}

func testAccAppstreamImageDataSourceConfig_basic(rName string) string { // , version string
	return fmt.Sprintf(`
	resource "aws_appstream_image_builder" "test" {
	image_name    = "AppStream-WinServer2022-03-24-2024"
	instance_type = %[1]q
	name          = %[2]q
	image_arn     = "arn:${data.aws_partition.current.partition}:appstream:%[1]s::image/%[2]s"
	}


	data "aws_appstream_image" "test" {
		arns = "????"
		max_results = 1
		names = %[2]q
		type = "PRIVATE"
	}

`, rName)
}
