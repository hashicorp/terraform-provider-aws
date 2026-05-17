// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package oam_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/oam"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfoam "github.com/hashicorp/terraform-provider-aws/internal/service/oam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccObservabilityAccessManagerLink_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var link oam.GetLinkOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_oam_link.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckLinkDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLinkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, t, resourceName, &link),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "oam", "link/{link_id}"),
					resource.TestCheckResourceAttrSet(resourceName, "label"),
					resource.TestCheckResourceAttr(resourceName, "label_template", "$AccountName"),
					resource.TestCheckResourceAttrSet(resourceName, "link_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.0", "AWS::CloudWatch::Metric"),
					resource.TestCheckResourceAttrSet(resourceName, "sink_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "sink_identifier"),
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

func testAccObservabilityAccessManagerLink_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var link oam.GetLinkOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_oam_link.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckLinkDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLinkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, t, resourceName, &link),
					acctest.CheckSDKResourceDisappears(ctx, t, tfoam.ResourceLink(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccObservabilityAccessManagerLink_update(t *testing.T) {
	ctx := acctest.Context(t)
	var link oam.GetLinkOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_oam_link.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckLinkDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLinkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, t, resourceName, &link),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "oam", regexache.MustCompile(`link/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "label"),
					resource.TestCheckResourceAttr(resourceName, "label_template", "$AccountName"),
					resource.TestCheckResourceAttrSet(resourceName, "link_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.0", "AWS::CloudWatch::Metric"),
					resource.TestCheckResourceAttrSet(resourceName, "sink_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "sink_identifier"),
				),
			},
			{
				Config: testAccLinkConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, t, resourceName, &link),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "oam", regexache.MustCompile(`link/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "label"),
					resource.TestCheckResourceAttr(resourceName, "label_template", "$AccountName"),
					resource.TestCheckResourceAttrSet(resourceName, "link_id"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.0", "AWS::CloudWatch::Metric"),
					resource.TestCheckResourceAttr(resourceName, "resource_types.1", "AWS::Logs::LogGroup"),
					resource.TestCheckResourceAttrSet(resourceName, "sink_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "sink_identifier"),
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

func testAccObservabilityAccessManagerLink_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var link oam.GetLinkOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_oam_link.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckLinkDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLinkConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, t, resourceName, &link),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccLinkConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, t, resourceName, &link),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLinkConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, t, resourceName, &link),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
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

func testAccObservabilityAccessManagerLink_logGroupConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var link oam.GetLinkOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_oam_link.test"
	filter1 := "LogGroupName LIKE 'aws/lambda/%' OR LogGroupName LIKE 'AWSLogs%'"
	filter2 := "LogGroupName NOT IN ('Private-Log-Group', 'Private-Log-Group-2')"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckLinkDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLinkConfig_logGroupConfiguration(rName, filter1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, t, resourceName, &link),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.0.log_group_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.0.metric_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.0.log_group_configuration.0.filter", filter1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLinkConfig_logGroupConfiguration(rName, filter2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, t, resourceName, &link),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.0.log_group_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.0.metric_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.0.log_group_configuration.0.filter", filter2),
				),
			},
		},
	})
}

func testAccObservabilityAccessManagerLink_metricConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var link oam.GetLinkOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_oam_link.test"
	filter1 := "Namespace IN ('AWS/EC2', 'AWS/ELB', 'AWS/S3')"
	filter2 := "Namespace NOT LIKE 'AWS/%'"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckLinkDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLinkConfig_metricConfiguration(rName, filter1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, t, resourceName, &link),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.0.log_group_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.0.metric_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.0.metric_configuration.0.filter", filter1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLinkConfig_metricConfiguration(rName, filter2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, t, resourceName, &link),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.0.log_group_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.0.metric_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "link_configuration.0.metric_configuration.0.filter", filter2),
				),
			},
		},
	})
}

func testAccCheckLinkDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ObservabilityAccessManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_oam_link" {
				continue
			}

			_, err := tfoam.FindLinkByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ObservabilityAccessManager Link %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLinkExists(ctx context.Context, t *testing.T, n string, v *oam.GetLinkOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ObservabilityAccessManagerClient(ctx)

		output, err := tfoam.FindLinkByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLinkConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_caller_identity" "source" {}
data "aws_partition" "source" {}

data "aws_caller_identity" "monitoring" {
  provider = "awsalternate"
}
data "aws_partition" "monitoring" {
  provider = "awsalternate"
}

resource "aws_oam_sink" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_oam_sink_policy" "test" {
  provider = "awsalternate"

  sink_identifier = aws_oam_sink.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["oam:CreateLink", "oam:UpdateLink"]
        Effect   = "Allow"
        Resource = "*"
        Principal = {
          "AWS" = "arn:${data.aws_partition.source.partition}:iam::${data.aws_caller_identity.source.account_id}:root"
        }
        Condition = {
          "ForAnyValue:StringEquals" = {
            "oam:ResourceTypes" = ["AWS::CloudWatch::Metric", "AWS::Logs::LogGroup"]
          }
        }
      }
    ]
  })
}

resource "aws_oam_link" "test" {
  label_template  = "$AccountName"
  resource_types  = ["AWS::CloudWatch::Metric"]
  sink_identifier = aws_oam_sink.test.arn

  depends_on = [
    aws_oam_sink_policy.test
  ]
}
`, rName))
}

func testAccLinkConfig_update(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_caller_identity" "source" {}
data "aws_partition" "source" {}

data "aws_caller_identity" "monitoring" {
  provider = "awsalternate"
}
data "aws_partition" "monitoring" {
  provider = "awsalternate"
}

resource "aws_oam_sink" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_oam_sink_policy" "test" {
  provider = "awsalternate"

  sink_identifier = aws_oam_sink.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["oam:CreateLink", "oam:UpdateLink"]
        Effect   = "Allow"
        Resource = "*"
        Principal = {
          "AWS" = "arn:${data.aws_partition.source.partition}:iam::${data.aws_caller_identity.source.account_id}:root"
        }
        Condition = {
          "ForAnyValue:StringEquals" = {
            "oam:ResourceTypes" = ["AWS::CloudWatch::Metric", "AWS::Logs::LogGroup"]
          }
        }
      }
    ]
  })
}

resource "aws_oam_link" "test" {
  label_template  = "$AccountName"
  resource_types  = ["AWS::CloudWatch::Metric", "AWS::Logs::LogGroup"]
  sink_identifier = aws_oam_sink.test.arn

  depends_on = [
    aws_oam_sink_policy.test
  ]
}
`, rName))
}

func testAccLinkConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_caller_identity" "source" {}
data "aws_partition" "source" {}

data "aws_caller_identity" "monitoring" {
  provider = "awsalternate"
}
data "aws_partition" "monitoring" {
  provider = "awsalternate"
}

resource "aws_oam_sink" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_oam_sink_policy" "test" {
  provider = "awsalternate"

  sink_identifier = aws_oam_sink.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["oam:CreateLink", "oam:UpdateLink"]
        Effect   = "Allow"
        Resource = "*"
        Principal = {
          "AWS" = "arn:${data.aws_partition.source.partition}:iam::${data.aws_caller_identity.source.account_id}:root"
        }
        Condition = {
          "ForAnyValue:StringEquals" = {
            "oam:ResourceTypes" = ["AWS::CloudWatch::Metric", "AWS::Logs::LogGroup"]
          }
        }
      }
    ]
  })
}

resource "aws_oam_link" "test" {
  label_template  = "$AccountName"
  resource_types  = ["AWS::CloudWatch::Metric"]
  sink_identifier = aws_oam_sink.test.arn
  tags = {
    %[2]q = %[3]q
  }

  depends_on = [
    aws_oam_sink_policy.test
  ]
}
`, rName, tag1Key, tag1Value))
}

func testAccLinkConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_caller_identity" "source" {}
data "aws_partition" "source" {}

data "aws_caller_identity" "monitoring" {
  provider = "awsalternate"
}
data "aws_partition" "monitoring" {
  provider = "awsalternate"
}

resource "aws_oam_sink" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_oam_sink_policy" "test" {
  provider = "awsalternate"

  sink_identifier = aws_oam_sink.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["oam:CreateLink", "oam:UpdateLink"]
        Effect   = "Allow"
        Resource = "*"
        Principal = {
          "AWS" = "arn:${data.aws_partition.source.partition}:iam::${data.aws_caller_identity.source.account_id}:root"
        }
        Condition = {
          "ForAnyValue:StringEquals" = {
            "oam:ResourceTypes" = ["AWS::CloudWatch::Metric", "AWS::Logs::LogGroup"]
          }
        }
      }
    ]
  })
}

resource "aws_oam_link" "test" {
  label_template  = "$AccountName"
  resource_types  = ["AWS::CloudWatch::Metric"]
  sink_identifier = aws_oam_sink.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [
    aws_oam_sink_policy.test
  ]
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccLinkConfig_logGroupConfiguration(rName, filter string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_caller_identity" "source" {}
data "aws_partition" "source" {}

data "aws_caller_identity" "monitoring" {
  provider = "awsalternate"
}
data "aws_partition" "monitoring" {
  provider = "awsalternate"
}

resource "aws_oam_sink" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_oam_sink_policy" "test" {
  provider = "awsalternate"

  sink_identifier = aws_oam_sink.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["oam:CreateLink", "oam:UpdateLink"]
        Effect   = "Allow"
        Resource = "*"
        Principal = {
          "AWS" = "arn:${data.aws_partition.source.partition}:iam::${data.aws_caller_identity.source.account_id}:root"
        }
        Condition = {
          "ForAnyValue:StringEquals" = {
            "oam:ResourceTypes" = ["AWS::CloudWatch::Metric", "AWS::Logs::LogGroup"]
          }
        }
      }
    ]
  })
}

resource "aws_oam_link" "test" {
  label_template = "$AccountName"
  link_configuration {
    log_group_configuration {
      filter = %[2]q
    }
  }
  resource_types  = ["AWS::Logs::LogGroup"]
  sink_identifier = aws_oam_sink.test.arn

  depends_on = [
    aws_oam_sink_policy.test
  ]
}
`, rName, filter))
}

func testAccLinkConfig_metricConfiguration(rName, filter string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_caller_identity" "source" {}
data "aws_partition" "source" {}

data "aws_caller_identity" "monitoring" {
  provider = "awsalternate"
}
data "aws_partition" "monitoring" {
  provider = "awsalternate"
}

resource "aws_oam_sink" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_oam_sink_policy" "test" {
  provider = "awsalternate"

  sink_identifier = aws_oam_sink.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["oam:CreateLink", "oam:UpdateLink"]
        Effect   = "Allow"
        Resource = "*"
        Principal = {
          "AWS" = "arn:${data.aws_partition.source.partition}:iam::${data.aws_caller_identity.source.account_id}:root"
        }
        Condition = {
          "ForAnyValue:StringEquals" = {
            "oam:ResourceTypes" = ["AWS::CloudWatch::Metric", "AWS::Logs::LogGroup"]
          }
        }
      }
    ]
  })
}

resource "aws_oam_link" "test" {
  label_template = "$AccountName"
  link_configuration {
    metric_configuration {
      filter = %[2]q
    }
  }
  resource_types  = ["AWS::CloudWatch::Metric"]
  sink_identifier = aws_oam_sink.test.arn

  depends_on = [
    aws_oam_sink_policy.test
  ]
}
`, rName, filter))
}
