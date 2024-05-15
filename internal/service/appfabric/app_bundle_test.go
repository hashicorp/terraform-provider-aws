// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
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

	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
)

func TestAccAppFabricAppBundle_basic(t *testing.T) {
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

func testAccAppBundleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
	tags = {
		Name = "AppFabricTesting"
	}
}
`)
}
