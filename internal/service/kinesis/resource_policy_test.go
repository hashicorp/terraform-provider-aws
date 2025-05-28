// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkinesis "github.com/hashicorp/terraform-provider-aws/internal/service/kinesis"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_resource_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPolicy}, // TODO terraform-plugin-testing
			},
		},
	})
}

func TestAccKinesisResourcePolicy_Identity_Basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_resource_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrResourceARN), tfknownvalue.RegionalARNExact("kinesis", ("stream/"+rName))),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrResourceARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPolicy},
			},
		},
	})
}

func TestAccKinesisResourcePolicy_Identity_RegionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_resource_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_regionOverride(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrResourceARN), tfknownvalue.RegionalARNAlternateRegionExact("kinesis", ("stream/"+rName))),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrResourceARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       acctest.CrossRegionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPolicy},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPolicy},
			},
		},
	})
}

func TestAccKinesisResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_resource_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfkinesis.ResourceResourcePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourcePolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisClient(ctx)

		_, err := tfkinesis.FindResourcePolicyByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckResourcePolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_resource_policy" {
				continue
			}

			_, err := tfkinesis.FindResourcePolicyByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Kinesis Resource Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccResourcePolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "target" {
  provider = "awsalternate"
}

resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2
}

resource "aws_kinesis_resource_policy" "test" {
  resource_arn = aws_kinesis_stream.test.arn

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "writePolicy",
  "Statement": [{
    "Sid": "writestatement",
    "Effect": "Allow",
    "Principal": {
      "AWS": "${data.aws_caller_identity.target.account_id}"
    },
    "Action": [
      "kinesis:DescribeStreamSummary",
      "kinesis:ListShards",
      "kinesis:PutRecord",
      "kinesis:PutRecords"
    ],
    "Resource": "${aws_kinesis_stream.test.arn}"
  }]
}
EOF
}
`, rName))
}

func testAccResourcePolicyConfig_regionOverride(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "target" {
  provider = "awsalternate"
}

resource "aws_kinesis_stream" "test" {
  region = %[2]q

  name        = %[1]q
  shard_count = 2
}

resource "aws_kinesis_resource_policy" "test" {
  region = %[2]q

  resource_arn = aws_kinesis_stream.test.arn

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "writePolicy",
  "Statement": [{
    "Sid": "writestatement",
    "Effect": "Allow",
    "Principal": {
      "AWS": "${data.aws_caller_identity.target.account_id}"
    },
    "Action": [
      "kinesis:DescribeStreamSummary",
      "kinesis:ListShards",
      "kinesis:PutRecord",
      "kinesis:PutRecords"
    ],
    "Resource": "${aws_kinesis_stream.test.arn}"
  }]
}
EOF
}
`, rName, acctest.AlternateRegion()))
}
