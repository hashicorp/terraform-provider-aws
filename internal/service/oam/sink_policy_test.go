// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oam_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/oam"
	"github.com/aws/aws-sdk-go-v2/service/oam/types"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfoam "github.com/hashicorp/terraform-provider-aws/internal/service/oam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccObservabilityAccessManagerSinkPolicy_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var sinkPolicy oam.GetSinkPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_oam_sink_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSinkPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSinkPolicyConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkPolicyExists(ctx, resourceName, &sinkPolicy),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "oam", regexache.MustCompile(`sink/+.`)),
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
					`, acctest.Partition(), acctest.AccountID()))
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var sinkPolicy oam.GetSinkPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_oam_sink_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSinkPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSinkPolicyConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkPolicyExists(ctx, resourceName, &sinkPolicy),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "oam", regexache.MustCompile(`sink/+.`)),
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
					`, acctest.Partition(), acctest.AccountID()))
						return err
					}),
					resource.TestCheckResourceAttrSet(resourceName, "sink_id"),
					resource.TestCheckResourceAttrPair(resourceName, "sink_identifier", "aws_oam_sink.test", names.AttrID),
				),
			},
			{
				Config: testAccSinkPolicyConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkPolicyExists(ctx, resourceName, &sinkPolicy),
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
					`, acctest.Partition(), acctest.AccountID()))
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

func testAccCheckSinkPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_oam_sink_policy" {
				continue
			}

			input := &oam.GetSinkPolicyInput{
				SinkIdentifier: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetSinkPolicy(ctx, input)
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.ObservabilityAccessManager, create.ErrActionCheckingDestroyed, tfoam.ResNameSinkPolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckSinkPolicyExists(ctx context.Context, name string, sinkPolicy *oam.GetSinkPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ObservabilityAccessManager, create.ErrActionCheckingExistence, tfoam.ResNameSinkPolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ObservabilityAccessManager, create.ErrActionCheckingExistence, tfoam.ResNameSinkPolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

		resp, err := conn.GetSinkPolicy(ctx, &oam.GetSinkPolicyInput{
			SinkIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.ObservabilityAccessManager, create.ErrActionCheckingExistence, tfoam.ResNameSinkPolicy, rs.Primary.ID, err)
		}

		*sinkPolicy = *resp

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
  sink_identifier = aws_oam_sink.test.id
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
  sink_identifier = aws_oam_sink.test.id
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
