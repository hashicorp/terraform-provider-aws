// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_basic(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ec2", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ecr", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccOrganizationConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_basic(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfinspector2.ResourceOrganizationConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationConfiguration_ec2ECR(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_basic(true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ec2", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ecr", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccOrganizationConfiguration_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_lambda(false, false, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ec2", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ecr", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.lambda", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.lambda_code", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccOrganizationConfiguration_lambdaCode(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_lambda(false, false, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ec2", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ecr", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.lambda", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.lambda_code", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckOrganizationConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector2_organization_configuration" {
				continue
			}

			_, err := tfinspector2.FindOrganizationConfiguration(ctx, conn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Inspector2 Organization Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOrganizationConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client(ctx)

		_, err := tfinspector2.FindOrganizationConfiguration(ctx, conn)

		return err
	}
}

func testAccOrganizationConfigurationConfig_basic(ec2, ecr bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_inspector2_delegated_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}

resource "aws_inspector2_organization_configuration" "test" {
  auto_enable {
    ec2 = %[1]t
    ecr = %[2]t
  }

  depends_on = [aws_inspector2_delegated_admin_account.test]
}
`, ec2, ecr)
}

func testAccOrganizationConfigurationConfig_lambda(ec2, ecr, lambda, lambda_code bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
resource "aws_inspector2_delegated_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
resource "aws_inspector2_organization_configuration" "test" {
  auto_enable {
    ec2         = %[1]t
    ecr         = %[2]t
    lambda      = %[3]t
    lambda_code = %[4]t
  }
  depends_on = [aws_inspector2_delegated_admin_account.test]
}
`, ec2, ecr, lambda, lambda_code)
}
