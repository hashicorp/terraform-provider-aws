// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

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
	// using the services/appfabric/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appfabric"
	"github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
)

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.
func TestAccAppFabricAppBundle_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var appbundle appfabric.GetAppBundleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appfabric_app_bundle.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppBundleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppBundleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppBundleExists(ctx, resourceName, &appbundle),
					//resource.TestCheckResourceAttrSet(resourceName, "customer_managed_key_identifier"),
					// Do we add client token here?
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAppBundleImportStateIDFunc(ctx, resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppFabricAppBundle_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var appbundle appfabric.GetAppBundleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appfabric_app_bundle.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppBundleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppBundleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppBundleExists(ctx, resourceName, &appbundle),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceAppBundle = newResourceAppBundle
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappfabric.ResourceAppBundle, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppBundleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appfabric_app_bundle" {
				continue
			}

			_, err := conn.GetAppBundle(ctx, &appfabric.GetAppBundleInput{
				AppBundleIdentifier: aws.String(rs.Primary.Attributes["arn"]),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.AppFabric, create.ErrActionCheckingDestroyed, tfappfabric.ResNameAppBundle, rs.Primary.ID, err)
			}

			return create.Error(names.AppFabric, create.ErrActionCheckingDestroyed, tfappfabric.ResNameAppBundle, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAppBundleExists(ctx context.Context, name string, appbundle *appfabric.GetAppBundleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.AppFabric, create.ErrActionCheckingExistence, tfappfabric.ResNameAppBundle, name, errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return create.Error(names.AppFabric, create.ErrActionCheckingExistence, tfappfabric.ResNameAppBundle, name, errors.New("not set"))
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)
		resp, err := conn.GetAppBundle(ctx, &appfabric.GetAppBundleInput{
			AppBundleIdentifier: aws.String(rs.Primary.Attributes["arn"]),
		})
		if err != nil {
			return create.Error(names.AppFabric, create.ErrActionCheckingExistence, tfappfabric.ResNameAppBundle, rs.Primary.ID, err)
		}
		*appbundle = *resp
		return nil
	}
}

// leave default
func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)
	input := &appfabric.ListAppBundlesInput{}
	_, err := conn.ListAppBundles(ctx, input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
func testAccCheckAppBundleNotRecreated(before, after *appfabric.GetAppBundleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.AppBundle.Arn), aws.ToString(after.AppBundle.Arn); before != after {
			return create.Error(names.AppFabric, create.ErrActionCheckingNotRecreated, tfappfabric.ResNameAppBundle, before, errors.New("recreated"))
		}
		return nil
	}
}

func testAccAppBundleImportStateIDFunc(ctx context.Context, resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return "", errors.New("No AppBundle ID set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)
		appBundleARN := rs.Primary.Attributes["arn"]

		_, err := conn.GetAppBundle(ctx, &appfabric.GetAppBundleInput{
			AppBundleIdentifier: aws.String(appBundleARN),
		})

		if err != nil {
			return "", err
		}

		return appBundleARN, nil
	}
}

// might need to change arn to app bundle identifier... not too sure here
// need to change to an actual arn
// not sure if needing customer managed key arn...
func testAccAppBundleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
	tags = {
		Name = "AppFabricTesting"
	}
}
`)
}
