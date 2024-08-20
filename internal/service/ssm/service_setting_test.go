// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMServiceSetting_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var setting awstypes.ServiceSetting
	resourceName := "aws_ssm_service_setting.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceSettingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSettingConfig_basic(acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceSettingExists(ctx, resourceName, &setting),
					resource.TestCheckResourceAttr(resourceName, "setting_value", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceSettingConfig_basic(acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceSettingExists(ctx, resourceName, &setting),
					resource.TestCheckResourceAttr(resourceName, "setting_value", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckServiceSettingDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_service_setting" {
				continue
			}

			output, err := tfssm.FindServiceSettingByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if aws.ToString(output.Status) == "Default" {
				continue
			}

			return fmt.Errorf("SSM Service Setting %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccServiceSettingExists(ctx context.Context, n string, v *awstypes.ServiceSetting) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		output, err := tfssm.FindServiceSettingByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
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
