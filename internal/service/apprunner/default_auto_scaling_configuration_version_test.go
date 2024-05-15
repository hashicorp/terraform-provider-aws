// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapprunner "github.com/hashicorp/terraform-provider-aws/internal/service/apprunner"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppRunnerDefaultAutoScalingConfigurationVersion_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccDefaultAutoScalingConfigurationVersion_basic,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccDefaultAutoScalingConfigurationVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_default_auto_scaling_configuration_version.test"
	var priorDefault *string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerClient(ctx)

					output, err := tfapprunner.FindDefaultAutoScalingConfigurationSummary(ctx, conn)

					if err == nil {
						priorDefault = output.AutoScalingConfigurationArn
					}
				},
				Config: testAccDefaultAutoScalingConfigurationVersionConfig_basic(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_configuration_arn", "aws_apprunner_auto_scaling_configuration_version.test.0", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDefaultAutoScalingConfigurationVersionConfig_basic(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_configuration_arn", "aws_apprunner_auto_scaling_configuration_version.test.1", names.AttrARN),
				),
			},
			// Restore the prior default, else "InvalidRequestException: You can't delete a reserved auto scaling configuration".
			{
				Config: testAccDefaultAutoScalingConfigurationVersionConfig_basic(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultAutoScalingConfigurationVersionRestore(ctx, &priorDefault),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDefaultAutoScalingConfigurationVersionRestore(ctx context.Context, priorDefault **string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerClient(ctx)

		return tfapprunner.PutDefaultAutoScalingConfiguration(ctx, conn, aws.ToString(*priorDefault))
	}
}

func testAccDefaultAutoScalingConfigurationVersionConfig_basic(rName string, idx int) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  count = 2

  auto_scaling_configuration_name = format("%%s-%%d", substr(%[1]q, 0, 26), count.index)
}

resource "aws_apprunner_default_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_arn = aws_apprunner_auto_scaling_configuration_version.test[%[2]d].arn
}
`, rName, idx)
}
