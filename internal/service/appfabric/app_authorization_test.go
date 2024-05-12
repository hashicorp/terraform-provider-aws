// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAppAuthorization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_app_authorization.test"
	var appauthorization types.AppAuthorization

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credential"},
			},
		},
	})
}

func testAccAppAuthorization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var appauthorization types.AppAuthorization
	resourceName := "aws_appfabric_app_authorization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappfabric.ResourceAppAuthorization, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppAuthorizationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appfabric_app_authorization" {
				continue
			}

			_, err := tfappfabric.FindAppAuthorizationByTwoPartKey(ctx, conn, rs.Primary.Attributes["arn"], rs.Primary.Attributes["app_bundle_identifier"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Fabric App Authorization %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppAuthorizationExists(ctx context.Context, n string, v *types.AppAuthorization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		output, err := tfappfabric.FindAppAuthorizationByTwoPartKey(ctx, conn, rs.Primary.Attributes["arn"], rs.Primary.Attributes["app_bundle_identifier"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAppAuthorizationConfig_basic() string {
	return `

resource "aws_appfabric_app_authorization" "test" {
  app_bundle_identifier   = "arn:aws:appfabric:eu-west-1:926562225508:appbundle/7546871d-363f-4374-a90e-464d8b69e3e7"
  app             		  = "TERRAFORMCLOUD"
  auth_type 			  = "apiKey"
  credential {
	api_key_credential {
		api_key = "testkey"
	}
  }
  tenant {
	tenant_display_name = "mkandylis-org-aws"
	tenant_identifier   = "mkandylis-org-aws"
  }
}
`
}
