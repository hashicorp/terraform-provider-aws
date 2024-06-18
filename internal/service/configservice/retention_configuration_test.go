// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRetentionConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rc types.RetentionConfiguration
	resourceName := "aws_config_retention_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRetentionConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRetentionConfigurationConfig_basic(90),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRetentionConfigurationExists(ctx, resourceName, &rc),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "default"),
					resource.TestCheckResourceAttr(resourceName, "retention_period_in_days", "90"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRetentionConfigurationConfig_basic(180),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRetentionConfigurationExists(ctx, resourceName, &rc),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "default"),
					resource.TestCheckResourceAttr(resourceName, "retention_period_in_days", "180"),
				),
			},
		},
	})
}

func testAccRetentionConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var rc types.RetentionConfiguration
	resourceName := "aws_config_retention_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRetentionConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRetentionConfigurationConfig_basic(90),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRetentionConfigurationExists(ctx, resourceName, &rc),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfconfig.ResourceRetentionConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRetentionConfigurationExists(ctx context.Context, n string, v *types.RetentionConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		output, err := tfconfig.FindRetentionConfigurationByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRetentionConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_config_retention_configuration" {
				continue
			}

			_, err := tfconfig.FindRetentionConfigurationByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ConfigService Retention Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccRetentionConfigurationConfig_basic(days int) string {
	return fmt.Sprintf(`
resource "aws_config_retention_configuration" "test" {
  retention_period_in_days = %[1]d
}
`, days)
}
