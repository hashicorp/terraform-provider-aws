// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package oam_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/oam"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfoam "github.com/hashicorp/terraform-provider-aws/internal/service/oam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccObservabilityAccessManagerSinkPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var sinkPolicy oam.GetSinkPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_oam_sink_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSinkPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSinkPolicyConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkPolicyExists(ctx, t, resourceName, &sinkPolicy),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, "aws_oam_sink.test", names.AttrARN),
					resource.TestCheckResourceAttrWith(resourceName, names.AttrPolicy, func(value string) error {
						_, err := awspolicy.PoliciesAreEquivalent(value, fmt.Sprintf(`
{
	"Version": "2012-10-17",
	"Statement": [{
		"Action": ["oam:CreateLink", "oam:UpdateLink"],
		"Effect": "Allow",
		"Resource": "*",
		"Principal": { "AWS": "arn:%s:iam::%s:root" },
		"Condition": {
			"ForAllValues:StringEquals": {
				"oam:ResourceTypes": [
					"AWS::CloudWatch::Metric",
					"AWS::Logs::LogGroup"
				]
			}
		}
  }]
}
							`, acctest.Partition(), acctest.AccountID(ctx)))
						return err
					}),
					resource.TestCheckResourceAttrSet(resourceName, "sink_id"),
					resource.TestCheckResourceAttrPair(resourceName, "sink_identifier", "aws_oam_sink.test", names.AttrID),
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

func testAccObservabilityAccessManagerSinkPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var sinkPolicy oam.GetSinkPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_oam_sink_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSinkPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSinkPolicyConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkPolicyExists(ctx, t, resourceName, &sinkPolicy),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "oam", regexache.MustCompile(`sink/.+$`)),
					resource.TestCheckResourceAttrWith(resourceName, names.AttrPolicy, func(value string) error {
						_, err := awspolicy.PoliciesAreEquivalent(value, fmt.Sprintf(`
{
	"Version": "2012-10-17",
	"Statement": [{
		"Action": ["oam:CreateLink", "oam:UpdateLink"],
		"Effect": "Allow",
		"Resource": "*",
		"Principal": { "AWS": "arn:%s:iam::%s:root" },
		"Condition": {
			"ForAllValues:StringEquals": {
				"oam:ResourceTypes": [
					"AWS::CloudWatch::Metric",
					"AWS::Logs::LogGroup"
				]
			}
		}
  }]
}
							`, acctest.Partition(), acctest.AccountID(ctx)))
						return err
					}),
					resource.TestCheckResourceAttrSet(resourceName, "sink_id"),
					resource.TestCheckResourceAttrPair(resourceName, "sink_identifier", "aws_oam_sink.test", names.AttrID),
				),
			},
			{
				Config: testAccSinkPolicyConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkPolicyExists(ctx, t, resourceName, &sinkPolicy),
					resource.TestCheckResourceAttrPair(resourceName, "sink_identifier", "aws_oam_sink.test", names.AttrID),
					resource.TestCheckResourceAttrWith(resourceName, names.AttrPolicy, func(value string) error {
						_, err := awspolicy.PoliciesAreEquivalent(value, fmt.Sprintf(`
{
	"Version": "2012-10-17",
	"Statement": [{
		"Action": ["oam:CreateLink", "oam:UpdateLink"],
		"Effect": "Allow",
		"Resource": "*",
		"Principal": { "AWS": "arn:%s:iam::%s:root" },
		"Condition": {
			"ForAllValues:StringEquals": {
				"oam:ResourceTypes": "AWS::CloudWatch::Metric"
			}
		}
  }]
}
							`, acctest.Partition(), acctest.AccountID(ctx)))
						return err
					}),
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

func testAccCheckSinkPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ObservabilityAccessManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_oam_sink_policy" {
				continue
			}

			_, err := tfoam.FindSinkPolicyByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ObservabilityAccessManager Sink %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSinkPolicyExists(ctx context.Context, t *testing.T, n string, v *oam.GetSinkPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ObservabilityAccessManagerClient(ctx)

		output, err := tfoam.FindSinkPolicyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSinkPolicyConfigBasic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_oam_sink" "test" {
  name = %[1]q
}

resource "aws_oam_sink_policy" "test" {
  sink_identifier = aws_oam_sink.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["oam:CreateLink", "oam:UpdateLink"]
        Effect   = "Allow"
        Resource = "*"
        Principal = {
          "AWS" = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Condition = {
          "ForAllValues:StringEquals" = {
            "oam:ResourceTypes" = ["AWS::CloudWatch::Metric", "AWS::Logs::LogGroup"]
          }
        }
      }
    ]
  })
}
`, rName)
}

func testAccSinkPolicyConfigUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_oam_sink" "test" {
  name = %[1]q
}

resource "aws_oam_sink_policy" "test" {
  sink_identifier = aws_oam_sink.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["oam:CreateLink", "oam:UpdateLink"]
        Effect   = "Allow"
        Resource = "*"
        Principal = {
          "AWS" = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Condition = {
          "ForAllValues:StringEquals" = {
            "oam:ResourceTypes" = "AWS::CloudWatch::Metric"
          }
        }
      }
    ]
  })
}
`, rName)
}
