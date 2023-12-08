// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminApplicationAccessScope_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_access_scope.test"
	applicationResourceName := "aws_ssoadmin_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAccessScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAccessScopeConfig_basic(rName, "sso:account:access"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAccessScopeExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "application_arn", applicationResourceName, "application_arn"),
					resource.TestCheckResourceAttr(resourceName, "scope", "sso:account:access"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccApplicationAccessScopeImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSOAdminApplicationAccessScope_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_access_scope.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAccessScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAccessScopeConfig_basic(rName, "sso:account:access"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAccessScopeExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, ssoadmin.ResourceApplicationAccessScope(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckApplicationAccessScopeExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		applicationARN, scope, err := ssoadmin.ApplicationAccessScopeParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = ssoadmin.FindApplicationAccessScopeByScopeAndApplicationARN(ctx, conn, applicationARN, scope)

		return err
	}
}

func testAccCheckApplicationAccessScopeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_application_access_scope" {
				continue
			}

			var applicationARN, scope, err = ssoadmin.ApplicationAccessScopeParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = ssoadmin.FindApplicationAccessScopeByScopeAndApplicationARN(ctx, conn, applicationARN, scope)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSO Application Access Scope %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccApplicationAccessScopeImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["application_arn"], rs.Primary.Attributes["scope"]), nil
	}
}

func testAccApplicationAccessScopeConfig_basic(rName, scope string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_application" "test" {
  name                     = %[1]q
  application_provider_arn = %[2]q
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_application_access_scope" "test" {
  application_arn    = aws_ssoadmin_application.test.application_arn
  authorized_targets = [aws_ssoadmin_application.test.application_arn]
  scope              = %[3]q
}
`, rName, testAccApplicationProviderARN, scope)
}
