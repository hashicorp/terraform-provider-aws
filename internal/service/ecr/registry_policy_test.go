// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
)

func TestAccECRRegistryPolicy_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"basic":      testAccRegistryPolicy_basic,
		"disappears": testAccRegistryPolicy_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccRegistryPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v ecr.GetRegistryPolicyOutput
	resourceName := "aws_ecr_registry_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryPolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryPolicyExists(ctx, resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"ecr:ReplicateImage".+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRegistryPolicyConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryPolicyExists(ctx, resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"ecr:ReplicateImage".+`)),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"ecr:CreateRepository".+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
				),
			},
		},
	})
}

func testAccRegistryPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v ecr.GetRegistryPolicyOutput
	resourceName := "aws_ecr_registry_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryPolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryPolicyExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfecr.ResourceRegistryPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRegistryPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecr_registry_policy" {
				continue
			}

			_, err := conn.GetRegistryPolicyWithContext(ctx, &ecr.GetRegistryPolicyInput{})
			if err != nil {
				if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
					return nil
				}
				return err
			}
		}

		return nil
	}
}

func testAccCheckRegistryPolicyExists(ctx context.Context, name string, res *ecr.GetRegistryPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR registry policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn(ctx)

		output, err := conn.GetRegistryPolicyWithContext(ctx, &ecr.GetRegistryPolicyInput{})
		if err != nil {
			if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
				return fmt.Errorf("ECR repository %s not found", rs.Primary.ID)
			}
			return err
		}

		*res = *output

		return nil
	}
}

func testAccRegistryPolicyConfig_basic() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_ecr_registry_policy" "test" {
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "testpolicy",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action" : "ecr:ReplicateImage",
        "Resource" : "arn:${data.aws_partition.current.partition}:ecr:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:repository/*",
      }
    ]
  })
}
`
}

func testAccRegistryPolicyConfig_updated() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_ecr_registry_policy" "test" {
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "testpolicy",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action" : [
          "ecr:ReplicateImage",
          "ecr:CreateRepository"
        ],
        "Resource" : [
          "arn:${data.aws_partition.current.partition}:ecr:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:repository/*"
        ]
      }
    ]
  })
}
`
}
