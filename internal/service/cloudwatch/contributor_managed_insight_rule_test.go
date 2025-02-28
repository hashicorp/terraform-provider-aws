// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudWatchContributorManagedInsightRule_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var contributormanagedinsightrule types.ManagedRuleDescription
	rName := sdkacctest.RandomWithPrefix("tfacctest")
	resourceName := "aws_cloudwatch_contributor_managed_insight_rule.test"
	templateName := "VpcEndpointService-NewConnectionsByEndpointId-v1"
	vpcEndpointService := "aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContributorManagedInsightRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContributorManagedInsightRuleConfig_basic(rName, templateName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorManagedInsightRuleExists(ctx, resourceName, &contributormanagedinsightrule),
					resource.TestCheckResourceAttr(resourceName, "template_name", templateName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, vpcEndpointService, names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrResourceARN,
				ImportStateIdFunc:                    testAccContributorManagedInsightRuleImportStateIDFunc(resourceName),
				ImportStateVerifyIgnore:              []string{"rule_name", names.AttrState},
			},
		},
	})
}

func TestAccCloudWatchContributorManagedInsightRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var v types.ManagedRuleDescription
	rName := sdkacctest.RandomWithPrefix("tfacctest")
	resourceName := "aws_cloudwatch_contributor_managed_insight_rule.test"
	templateName := "VpcEndpointService-NewConnectionsByEndpointId-v1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContributorManagedInsightRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContributorManagedInsightRuleConfig_basic(rName, templateName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorManagedInsightRuleExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcloudwatch.ResourceContributorManagedInsightRule, resourceName),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccCloudWatchContributorManagedInsightRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.ManagedRuleDescription
	rName := sdkacctest.RandomWithPrefix("tfacctest")
	resourceName := "aws_cloudwatch_contributor_managed_insight_rule.test"
	templateName := "VpcEndpointService-NewConnectionsByEndpointId-v1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContributorManagedInsightRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContributorManagedInsightRuleConfig_tags1(rName, templateName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContributorManagedInsightRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrResourceARN,
				ImportStateIdFunc:                    testAccContributorManagedInsightRuleImportStateIDFunc(resourceName),
				ImportStateVerifyIgnore:              []string{"rule_name", names.AttrState},
			},
		},
	})
}

func testAccCheckContributorManagedInsightRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_contributor_managed_insight_rule" {
				continue
			}

			rule, err := tfcloudwatch.FindContributorManagedInsightRuleDescriptionByTemplateName(
				ctx,
				conn,
				rs.Primary.Attributes[names.AttrResourceARN],
				rs.Primary.Attributes["template_name"],
			)

			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			// Consider empty state as "not found" as even if the rule is deleted, the rule will still show in the list, but with an empty state
			if rule.RuleState == nil || rule.RuleState.RuleName == nil {
				return nil
			}

			return fmt.Errorf("CloudWatch Contributor Managed Insight Rule still exists: %s", rs.Primary.Attributes[names.AttrResourceARN])
		}

		return nil
	}
}

func testAccCheckContributorManagedInsightRuleExists(ctx context.Context, name string, contributormanagedinsightrule *types.ManagedRuleDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CloudWatch, create.ErrActionCheckingExistence, tfcloudwatch.ResNameContributorManagedInsightRule, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchClient(ctx)

		output, err := tfcloudwatch.FindContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN], rs.Primary.Attributes["template_name"])

		if err != nil {
			return err
		}

		*contributormanagedinsightrule = *output

		return nil
	}
}

func testAccContributorManagedInsightRuleImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes[names.AttrResourceARN], rs.Primary.Attributes["template_name"]), nil
	}
}

func testAccContributorManagedInsightRuleConfig_basic(rName, templateName string) string {
	return acctest.ConfigCompose(testAccContributorManagedInsightRuleConfig_baseNetworkLoadBalancer(rName, 2), fmt.Sprintf(`
resource "aws_cloudwatch_contributor_managed_insight_rule" "test" {
  resource_arn  = aws_vpc_endpoint_service.test.arn
  template_name = %[2]q
  state         = "ENABLED"
}
`, rName, templateName))
}

func testAccContributorManagedInsightRuleConfig_tags1(rName, template_name, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccContributorManagedInsightRuleConfig_baseNetworkLoadBalancer(rName, 1), fmt.Sprintf(`
resource "aws_cloudwatch_contributor_managed_insight_rule" "test" {
  resource_arn  = aws_vpc_endpoint_service.test.arn
  template_name = %[2]q
  state         = "ENABLED"

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, template_name, tagKey1, tagValue1))
}

func testAccContributorManagedInsightRuleConfig_baseNetworkLoadBalancer(rName string, count int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = %[2]d

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  count = %[2]d

  load_balancer_type = "network"
  name               = "%[1]s-${count.index}"

  subnets = aws_subnet.test[*].id

  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
}
`, rName, count)
}
