// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRRegistryPolicy_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccRegistryPolicy_basic,
		acctest.CtDisappears: testAccRegistryPolicy_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccRegistryPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v ecr.GetRegistryPolicyOutput
	resourceName := "aws_ecr_registry_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryPolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryPolicyExists(ctx, resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"ecr:ReplicateImage".+`)),
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
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"ecr:ReplicateImage".+`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"ecr:CreateRepository".+`)),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecr_registry_policy" {
				continue
			}

			_, err := tfecr.FindRegistryPolicy(ctx, conn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECR Registry Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRegistryPolicyExists(ctx context.Context, n string, v *ecr.GetRegistryPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRClient(ctx)

		output, err := tfecr.FindRegistryPolicy(ctx, conn)

		if err != nil {
			return err
		}

		*v = *output

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
