// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppFabricAppBundle_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var appbundle awstypes.AppBundle
	resourceName := "aws_appfabric_app_bundle.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppFabricEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppBundleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppBundleConfig_basic,
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

			_, err := tfappfabric.FindAppBundleByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppFabric App Bundle %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppBundleExists(ctx context.Context, n string, v *awstypes.AppBundle) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		output, err := tfappfabric.FindAppBundleByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

const testAccAppBundleConfig_basic = `
resource "aws_appfabric_app_bundle" "test" {}
`

func testAccAppBundleConfig_customer_managed_key() string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
	customer_managed_key = "arn:aws:kms:us-east-1:732859338261:key/c67081be-29a0-4049-a821-1436d27bde94"
}
`)
}

func testAccAppBundleConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccAppBundleConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
	tags = {
		%[1]q = %[2]q
		%[3]q = %[4]q
	}
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAppBundleConfig_full(rName string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
	customer_managed_key = "arn:aws:kms:us-east-1:732859338261:key/c67081be-29a0-4049-a821-1436d27bde94"
	tags = {
		Name = "AppFabricTesting"
	}
}
`)
}
