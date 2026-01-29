// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminCustomerManagedPolicyAttachmentsExclusive_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_customer_managed_policy_attachments_exclusive.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerManagedPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "customer_managed_policy_reference.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "customer_managed_policy_reference.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "customer_managed_policy_reference.0.path", "/"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "instance_arn", "permission_set_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "instance_arn",
			},
		},
	})
}

func TestAccSSOAdminCustomerManagedPolicyAttachmentsExclusive_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_customer_managed_policy_attachments_exclusive.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerManagedPolicyAttachmentsExclusiveConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "customer_managed_policy_reference.#", "2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "instance_arn", "permission_set_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "instance_arn",
			},
		},
	})
}

func TestAccSSOAdminCustomerManagedPolicyAttachmentsExclusive_empty(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_customer_managed_policy_attachments_exclusive.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerManagedPolicyAttachmentsExclusiveConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "customer_managed_policy_reference.#", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "instance_arn", "permission_set_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "instance_arn",
			},
		},
	})
}

func TestAccSSOAdminCustomerManagedPolicyAttachmentsExclusive_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_customer_managed_policy_attachments_exclusive.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerManagedPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					testAccCheckCustomerManagedPolicyDetach(ctx, resourceName, rName, "/"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccCustomerManagedPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "customer_managed_policy_reference.#", "1"),
				),
			},
		},
	})
}

func TestAccSSOAdminCustomerManagedPolicyAttachmentsExclusive_outOfBandAddition(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_customer_managed_policy_attachments_exclusive.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerManagedPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					testAccCheckCustomerManagedPolicyAttach(ctx, resourceName, fmt.Sprintf("%s-2", rName), "/"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccCustomerManagedPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "customer_managed_policy_reference.#", "1"),
				),
			},
		},
	})
}

func testAccCheckCustomerManagedPolicyAttachmentsExclusiveExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		_, err := tfssoadmin.FindCustomerManagedPolicyAttachmentsByTwoPartKey(ctx, conn, rs.Primary.Attributes["permission_set_arn"], rs.Primary.Attributes["instance_arn"])

		return err
	}
}

func testAccCheckCustomerManagedPolicyDetach(ctx context.Context, n, policyName, policyPath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		_, err := conn.DetachCustomerManagedPolicyReferenceFromPermissionSet(ctx, &ssoadmin.DetachCustomerManagedPolicyReferenceFromPermissionSetInput{
			InstanceArn:      aws.String(rs.Primary.Attributes["instance_arn"]),
			PermissionSetArn: aws.String(rs.Primary.Attributes["permission_set_arn"]),
			CustomerManagedPolicyReference: &awstypes.CustomerManagedPolicyReference{
				Name: aws.String(policyName),
				Path: aws.String(policyPath),
			},
		})

		return err
	}
}

func testAccCheckCustomerManagedPolicyAttach(ctx context.Context, n, policyName, policyPath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		_, err := conn.AttachCustomerManagedPolicyReferenceToPermissionSet(ctx, &ssoadmin.AttachCustomerManagedPolicyReferenceToPermissionSetInput{
			InstanceArn:      aws.String(rs.Primary.Attributes["instance_arn"]),
			PermissionSetArn: aws.String(rs.Primary.Attributes["permission_set_arn"]),
			CustomerManagedPolicyReference: &awstypes.CustomerManagedPolicyReference{
				Name: aws.String(policyName),
				Path: aws.String(policyPath),
			},
		})

		return err
	}
}

func testAccCustomerManagedPolicyAttachmentsExclusiveConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_iam_policy" "test" {
  name = %[1]q

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = "s3:ListBucket"
      Effect   = "Allow"
      Resource = "*"
    }]
  })
}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_customer_managed_policy_attachments_exclusive" "test" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  permission_set_arn = aws_ssoadmin_permission_set.test.arn

  customer_managed_policy_reference {
    name = aws_iam_policy.test.name
  }
}
`, rName)
}

func testAccCustomerManagedPolicyAttachmentsExclusiveConfig_multiple(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_iam_policy" "test" {
  name = %[1]q

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = "s3:ListBucket"
      Effect   = "Allow"
      Resource = "*"
    }]
  })
}

resource "aws_iam_policy" "test2" {
  name = "%[1]s-2"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = "s3:GetObject"
      Effect   = "Allow"
      Resource = "*"
    }]
  })
}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_customer_managed_policy_attachments_exclusive" "test" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  permission_set_arn = aws_ssoadmin_permission_set.test.arn

  customer_managed_policy_reference {
    name = aws_iam_policy.test.name
  }

  customer_managed_policy_reference {
    name = aws_iam_policy.test2.name
  }
}
`, rName)
}

func testAccCustomerManagedPolicyAttachmentsExclusiveConfig_empty(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_customer_managed_policy_attachments_exclusive" "test" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`, rName)
}
