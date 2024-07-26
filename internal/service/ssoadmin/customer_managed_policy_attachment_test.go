// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminCustomerManagedPolicyAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_customer_managed_policy_attachment.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNamePolicy1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNamePolicy2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerManagedPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerManagedPolicyAttachmentConfig_basic(rName, rNamePolicy1, rNamePolicy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "customer_managed_policy_reference.0.name", rNamePolicy1),
					resource.TestCheckResourceAttrPair(resourceName, "instance_arn", permissionSetResourceName, "instance_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "permission_set_arn", permissionSetResourceName, names.AttrARN),
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

func TestAccSSOAdminCustomerManagedPolicyAttachment_forceNew(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_customer_managed_policy_attachment.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNamePolicy1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNamePolicy2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerManagedPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerManagedPolicyAttachmentConfig_basic(rName, rNamePolicy1, rNamePolicy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentExists(ctx, resourceName),
				),
			},
			{
				Config: testAccCustomerManagedPolicyAttachmentConfig_forceNew(rName, rNamePolicy1, rNamePolicy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "customer_managed_policy_reference.0.name", rNamePolicy2),
					resource.TestCheckResourceAttrPair(resourceName, "instance_arn", permissionSetResourceName, "instance_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "permission_set_arn", permissionSetResourceName, names.AttrARN),
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

func TestAccSSOAdminCustomerManagedPolicyAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_customer_managed_policy_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNamePolicy1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNamePolicy2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerManagedPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerManagedPolicyAttachmentConfig_basic(rName, rNamePolicy1, rNamePolicy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssoadmin.ResourceCustomerManagedPolicyAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminCustomerManagedPolicyAttachment_Disappears_permissionSet(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_customer_managed_policy_attachment.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNamePolicy1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNamePolicy2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerManagedPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerManagedPolicyAttachmentConfig_basic(rName, rNamePolicy1, rNamePolicy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssoadmin.ResourcePermissionSet(), permissionSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminCustomerManagedPolicyAttachment_multipleManagedPolicies(t *testing.T) {
	ctx := acctest.Context(t)
	resource1Name := "aws_ssoadmin_customer_managed_policy_attachment.test"
	resource2Name := "aws_ssoadmin_customer_managed_policy_attachment.test2"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNamePolicy1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNamePolicy2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNamePolicy3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerManagedPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerManagedPolicyAttachmentConfig_basic(rName, rNamePolicy1, rNamePolicy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentExists(ctx, resource1Name),
				),
			},
			{
				Config: testAccCustomerManagedPolicyAttachmentConfig_multiple(rName, rNamePolicy1, rNamePolicy2, rNamePolicy3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerManagedPolicyAttachmentExists(ctx, resource1Name),
					testAccCheckCustomerManagedPolicyAttachmentExists(ctx, resource2Name),
					resource.TestCheckResourceAttr(resource2Name, "customer_managed_policy_reference.0.name", rNamePolicy3),
					resource.TestCheckResourceAttrPair(resource2Name, "instance_arn", permissionSetResourceName, "instance_arn"),
					resource.TestCheckResourceAttrPair(resource2Name, "permission_set_arn", permissionSetResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resource2Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckCustomerManagedPolicyAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_customer_managed_policy_attachment" {
				continue
			}

			policyName, policyPath, permissionSetARN, instanceARN, err := tfssoadmin.CustomerManagedPolicyAttachmentParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfssoadmin.FindCustomerManagedPolicy(ctx, conn, policyName, policyPath, permissionSetARN, instanceARN)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSO Customer Managed Policy Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCustomerManagedPolicyAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSO Customer Managed Policy Attachment ID is set")
		}

		policyName, policyPath, permissionSetARN, instanceARN, err := tfssoadmin.CustomerManagedPolicyAttachmentParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		_, err = tfssoadmin.FindCustomerManagedPolicy(ctx, conn, policyName, policyPath, permissionSetARN, instanceARN)

		return err
	}
}

func testAccCustomerManagedPolicyAttachmentConfig_base(rName, rNamePolicy1, rNamePolicy2 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_iam_policy" "test1" {
  name        = %[2]q
  path        = "/"
  description = "My test policy"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ec2:Describe*",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}

resource "aws_iam_policy" "test2" {
  name        = %[3]q
  path        = "/"
  description = "My test policy"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ec2:Describe*",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}
`, rName, rNamePolicy1, rNamePolicy2)
}

func testAccCustomerManagedPolicyAttachmentConfig_basic(rName, rNamePolicy1, rNamePolicy2 string) string {
	return acctest.ConfigCompose(testAccCustomerManagedPolicyAttachmentConfig_base(rName, rNamePolicy1, rNamePolicy2), `
resource "aws_ssoadmin_customer_managed_policy_attachment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn

  customer_managed_policy_reference {
    name = aws_iam_policy.test1.name
    path = "/"
  }
}
`)
}

func testAccCustomerManagedPolicyAttachmentConfig_forceNew(rName, rNamePolicy1, rNamePolicy2 string) string {
	return acctest.ConfigCompose(testAccCustomerManagedPolicyAttachmentConfig_base(rName, rNamePolicy1, rNamePolicy2), `
resource "aws_ssoadmin_customer_managed_policy_attachment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn

  customer_managed_policy_reference {
    name = aws_iam_policy.test2.name
    path = "/"
  }
}
`)
}

func testAccCustomerManagedPolicyAttachmentConfig_multiple(rName, rNamePolicy1, rNamePolicy2, rNamePolicy3 string) string {
	return acctest.ConfigCompose(testAccCustomerManagedPolicyAttachmentConfig_basic(rName, rNamePolicy1, rNamePolicy2), fmt.Sprintf(`
resource "aws_ssoadmin_customer_managed_policy_attachment" "test2" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
  customer_managed_policy_reference {
    name = aws_iam_policy.test3.name
    path = "/"
  }
}

resource "aws_iam_policy" "test3" {
  name        = %[1]q
  path        = "/"
  description = "My other test policy"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ec2:Describe*",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}
`, rNamePolicy3))
}
