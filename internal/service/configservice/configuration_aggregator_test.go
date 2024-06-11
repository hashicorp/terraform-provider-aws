// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccConfigServiceConfigurationAggregator_account(t *testing.T) {
	ctx := acctest.Context(t)
	var ca types.ConfigurationAggregator
	//Name is upper case on purpose to test https://github.com/hashicorp/terraform-provider-aws/issues/8432
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationAggregatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(ctx, resourceName, &ca),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "config", regexache.MustCompile(`config-aggregator/config-aggregator-.+`)),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.0.account_ids.#", acctest.Ct1),
					acctest.CheckResourceAttrAccountID(resourceName, "account_aggregation_source.0.account_ids.0"),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.0.regions.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "account_aggregation_source.0.regions.0", "data.aws_region.current", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationAggregatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_organization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(ctx, resourceName, &ca),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", acctest.Ct1),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationAggregatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", acctest.Ct0),
				),
			},
			{
				Config: testAccConfigurationAggregatorConfig_organization(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccConfigServiceConfigurationAggregator_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var ca types.ConfigurationAggregator
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationAggregatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(ctx, resourceName, &ca),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccConfigurationAggregatorConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(ctx, resourceName, &ca),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationAggregatorConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(ctx, resourceName, &ca),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccConfigServiceConfigurationAggregator_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ca types.ConfigurationAggregator
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationAggregatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(ctx, resourceName, &ca),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconfig.ResourceConfigurationAggregator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConfigurationAggregatorExists(ctx context.Context, n string, v *types.ConfigurationAggregator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		output, err := tfconfig.FindConfigurationAggregatorByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConfigurationAggregatorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_config_configuration_aggregator" {
				continue
			}

			_, err := tfconfig.FindConfigurationAggregatorByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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
    regions     = [data.aws_region.current.name]
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

func testAccConfigurationAggregatorConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_config_configuration_aggregator" "test" {
  name = %[1]q

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccConfigurationAggregatorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_config_configuration_aggregator" "test" {
  name = %[1]q

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
