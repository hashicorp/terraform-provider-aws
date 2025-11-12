// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminRenameApplicationAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	newName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRenameApplicationActionConfig_basic(rName, newName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRenameApplicationActionExecuted(ctx, "aws_ssoadmin_application.test", newName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRenameApplicationActionExecuted(ctx context.Context, resourceName, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		output, err := conn.DescribeApplication(ctx, &ssoadmin.DescribeApplicationInput{
			ApplicationArn: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("error describing SSO Admin Application (%s): %w", rs.Primary.ID, err)
		}

		if aws.ToString(output.Name) != expectedName {
			return fmt.Errorf("SSO Admin Application name was not renamed. Expected: %s, Got: %s", expectedName, aws.ToString(output.Name))
		}

		return nil
	}
}

func testAccRenameApplicationActionConfig_basic(rName, newName string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_basic(rName, testAccApplicationProviderARN), fmt.Sprintf(`
action "aws_ssoadmin_rename_application" "test" {
  config {
    application_arn = aws_ssoadmin_application.test.application_arn
    new_name        = %[1]q
  }
}

resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_ssoadmin_rename_application.test]
    }
  }

  depends_on = [
    aws_ssoadmin_application.test,
  ]
}
`, newName))
}
