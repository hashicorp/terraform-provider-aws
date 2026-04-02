// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccConfigServiceConfigurationAggregator_account(t *testing.T) {
	ctx := acctest.Context(t)
	var ca types.ConfigurationAggregator
	//Name is upper case on purpose to test https://github.com/hashicorp/terraform-provider-aws/issues/8432
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationAggregatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(ctx, t, resourceName, &ca),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "config", regexache.MustCompile(`config-aggregator/config-aggregator-.+`)),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.0.account_ids.#", "1"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "account_aggregation_source.0.account_ids.0"),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.0.regions.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "account_aggregation_source.0.regions.0", "data.aws_region.current", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccConfigServiceConfigurationAggregator_organization(t *testing.T) {
	ctx := acctest.Context(t)
	var ca types.ConfigurationAggregator
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationAggregatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_organization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(ctx, t, resourceName, &ca),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_aggregation_source.0.role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.0.all_regions", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccConfigServiceConfigurationAggregator_switch(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationAggregatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", "0"),
				),
			},
			{
				Config: testAccConfigurationAggregatorConfig_organization(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", "1"),
				),
			},
		},
	})
}

func TestAccConfigServiceConfigurationAggregator_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ca types.ConfigurationAggregator
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationAggregatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(ctx, t, resourceName, &ca),
					acctest.CheckSDKResourceDisappears(ctx, t, tfconfig.ResourceConfigurationAggregator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConfigurationAggregatorExists(ctx context.Context, t *testing.T, n string, v *types.ConfigurationAggregator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ConfigServiceClient(ctx)

		output, err := tfconfig.FindConfigurationAggregatorByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConfigurationAggregatorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ConfigServiceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_config_configuration_aggregator" {
				continue
			}

			_, err := tfconfig.FindConfigurationAggregatorByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ConfigService Configuration Aggregator %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccConfigurationAggregatorConfig_account(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_config_configuration_aggregator" "test" {
  name = %[1]q

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.region]
  }
}
`, rName)
}

func testAccConfigurationAggregatorConfig_organization(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config.${data.aws_partition.current.dns_suffix}"]
}

resource "aws_config_configuration_aggregator" "test" {
  depends_on = [aws_iam_role_policy_attachment.test, aws_organizations_organization.test]

  name = %[1]q

  organization_aggregation_source {
    all_regions = true
    role_arn    = aws_iam_role.test.arn
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "config.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRoleForOrganizations"
}
`, rName)
}
