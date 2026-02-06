// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminManagedPolicyAttachmentsExclusive_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_managed_policy_attachments_exclusive.test"
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
				Config: testAccManagedPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "managed_policy_arns.*", fmt.Sprintf("arn:%s:iam::aws:policy/ReadOnlyAccess", acctest.Partition())),
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

func TestAccSSOAdminManagedPolicyAttachmentsExclusive_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_managed_policy_attachments_exclusive.test"
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
				Config: testAccManagedPolicyAttachmentsExclusiveConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "managed_policy_arns.*", fmt.Sprintf("arn:%s:iam::aws:policy/ReadOnlyAccess", acctest.Partition())),
					resource.TestCheckTypeSetElemAttr(resourceName, "managed_policy_arns.*", fmt.Sprintf("arn:%s:iam::aws:policy/PowerUserAccess", acctest.Partition())),
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

func TestAccSSOAdminManagedPolicyAttachmentsExclusive_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_managed_policy_attachments_exclusive.test"
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
				Config: testAccManagedPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
				),
			},
			{
				Config: testAccManagedPolicyAttachmentsExclusiveConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "2"),
				),
			},
			{
				Config: testAccManagedPolicyAttachmentsExclusiveConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "0"),
				),
			},
		},
	})
}

func TestAccSSOAdminManagedPolicyAttachmentsExclusive_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_managed_policy_attachments_exclusive.test"
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
				Config: testAccManagedPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					testAccCheckManagedPolicyDetach(ctx, resourceName, "ReadOnlyAccess"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccManagedPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
				),
			},
		},
	})
}

func TestAccSSOAdminManagedPolicyAttachmentsExclusive_outOfBandAddition(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_managed_policy_attachments_exclusive.test"
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
				Config: testAccManagedPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					testAccCheckManagedPolicyAttach(ctx, resourceName, "PowerUserAccess"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccManagedPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_arns.#", "1"),
				),
			},
		},
	})
}

func testAccCheckManagedPolicyAttachmentsExclusiveExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		_, err := tfssoadmin.FindManagedPolicyAttachmentsByTwoPartKey(ctx, conn, rs.Primary.Attributes["permission_set_arn"], rs.Primary.Attributes["instance_arn"])

		return err
	}
}

func testAccCheckManagedPolicyDetach(ctx context.Context, n, policyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)
		policyARN := fmt.Sprintf("arn:%s:iam::aws:policy/%s", acctest.Partition(), policyName)

		_, err := conn.DetachManagedPolicyFromPermissionSet(ctx, &ssoadmin.DetachManagedPolicyFromPermissionSetInput{
			InstanceArn:      aws.String(rs.Primary.Attributes["instance_arn"]),
			PermissionSetArn: aws.String(rs.Primary.Attributes["permission_set_arn"]),
			ManagedPolicyArn: aws.String(policyARN),
		})

		return err
	}
}

func testAccCheckManagedPolicyAttach(ctx context.Context, n, policyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)
		policyARN := fmt.Sprintf("arn:%s:iam::aws:policy/%s", acctest.Partition(), policyName)

		_, err := conn.AttachManagedPolicyToPermissionSet(ctx, &ssoadmin.AttachManagedPolicyToPermissionSetInput{
			InstanceArn:      aws.String(rs.Primary.Attributes["instance_arn"]),
			PermissionSetArn: aws.String(rs.Primary.Attributes["permission_set_arn"]),
			ManagedPolicyArn: aws.String(policyARN),
		})

		return err
	}
}

func testAccManagedPolicyAttachmentsExclusiveConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_managed_policy_attachments_exclusive" "test" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  permission_set_arn = aws_ssoadmin_permission_set.test.arn

  managed_policy_arns = [
    "arn:${data.aws_partition.current.partition}:iam::aws:policy/ReadOnlyAccess",
  ]
}
`, rName)
}

func testAccManagedPolicyAttachmentsExclusiveConfig_multiple(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_managed_policy_attachments_exclusive" "test" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  permission_set_arn = aws_ssoadmin_permission_set.test.arn

  managed_policy_arns = [
    "arn:${data.aws_partition.current.partition}:iam::aws:policy/ReadOnlyAccess",
    "arn:${data.aws_partition.current.partition}:iam::aws:policy/PowerUserAccess",
  ]
}
`, rName)
}

func testAccManagedPolicyAttachmentsExclusiveConfig_empty(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_managed_policy_attachments_exclusive" "test" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  permission_set_arn = aws_ssoadmin_permission_set.test.arn

  managed_policy_arns = []
}
`, rName)
}
