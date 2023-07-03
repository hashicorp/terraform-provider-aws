// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMServiceSetting_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var setting ssm.ServiceSetting
	resourceName := "aws_ssm_service_setting.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceSettingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSettingConfig_basic("false"),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceSettingExists(ctx, resourceName, &setting),
					resource.TestCheckResourceAttr(resourceName, "setting_value", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceSettingConfig_basic("true"),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceSettingExists(ctx, resourceName, &setting),
					resource.TestCheckResourceAttr(resourceName, "setting_value", "true"),
				),
			},
		},
	})
}

func testAccCheckServiceSettingDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_service_setting" {
				continue
			}

			output, err := tfssm.FindServiceSettingByID(ctx, conn, rs.Primary.Attributes["setting_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if aws.StringValue(output.Status) == "Default" {
				continue
			}

			return create.Error(names.SSM, create.ErrActionCheckingDestroyed, tfssm.ResNameServiceSetting, rs.Primary.Attributes["setting_id"], err)
		}

		return nil
	}
}

func testAccServiceSettingExists(ctx context.Context, n string, v *ssm.ServiceSetting) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Service Setting ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn(ctx)

		output, err := tfssm.FindServiceSettingByID(ctx, conn, rs.Primary.Attributes["setting_id"])

		if err != nil {
			return create.Error(names.SSM, create.ErrActionReading, tfssm.ResNameServiceSetting, rs.Primary.Attributes["setting_id"], err)
		}

		*v = *output

		return nil
	}
}

func testAccServiceSettingConfig_basic(settingValue string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_ssm_service_setting" "test" {
  setting_id    = "arn:${data.aws_partition.current.partition}:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:servicesetting/ssm/parameter-store/high-throughput-enabled"
  setting_value = %[1]q
}
`, settingValue)
}
