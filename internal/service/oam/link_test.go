// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oam_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/oam"
	"github.com/aws/aws-sdk-go-v2/service/oam/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfoam "github.com/hashicorp/terraform-provider-aws/internal/service/oam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccObservabilityAccessManagerLink_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var link oam.GetLinkOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_oam_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckLinkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLinkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, resourceName, &link),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "oam", regexp.MustCompile(`link/+.`)),
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

func TestAccObservabilityAccessManagerLink_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var link oam.GetLinkOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_oam_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckLinkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLinkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, resourceName, &link),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfoam.ResourceLink(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccObservabilityAccessManagerLink_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var link oam.GetLinkOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_oam_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckLinkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLinkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, resourceName, &link),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "oam", regexp.MustCompile(`link/+.`)),
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
					testAccCheckLinkExists(ctx, resourceName, &link),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "oam", regexp.MustCompile(`link/+.`)),
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

func TestAccObservabilityAccessManagerLink_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var link oam.GetLinkOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_oam_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckLinkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLinkConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, resourceName, &link),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccLinkConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, resourceName, &link),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLinkConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinkExists(ctx, resourceName, &link),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func testAccCheckLinkDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_oam_link" {
				continue
			}

			input := &oam.GetLinkInput{
				Identifier: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetLink(ctx, input)
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.ObservabilityAccessManager, create.ErrActionCheckingDestroyed, tfoam.ResNameLink, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckLinkExists(ctx context.Context, name string, link *oam.GetLinkOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ObservabilityAccessManager, create.ErrActionCheckingExistence, tfoam.ResNameLink, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ObservabilityAccessManager, create.ErrActionCheckingExistence, tfoam.ResNameLink, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

		resp, err := conn.GetLink(ctx, &oam.GetLinkInput{
			Identifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.ObservabilityAccessManager, create.ErrActionCheckingExistence, tfoam.ResNameLink, rs.Primary.ID, err)
		}

		*link = *resp

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

  sink_identifier = aws_oam_sink.test.id
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
  sink_identifier = aws_oam_sink.test.id
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

  sink_identifier = aws_oam_sink.test.id
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
  sink_identifier = aws_oam_sink.test.id
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

  sink_identifier = aws_oam_sink.test.id
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
  sink_identifier = aws_oam_sink.test.id
  tags = {
    %[2]q = %[3]q
  }
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

  sink_identifier = aws_oam_sink.test.id
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
  sink_identifier = aws_oam_sink.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}
