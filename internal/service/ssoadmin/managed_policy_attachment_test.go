// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminManagedPolicyAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPolicyAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPolicyAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(ctx, t, resourceName),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "managed_policy_arn", regexache.MustCompile(`policy/AlexaForBusinessDeviceSetup`)),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_name", "AlexaForBusinessDeviceSetup"),
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

func TestAccSSOAdminManagedPolicyAttachment_forceNew(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPolicyAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPolicyAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccManagedPolicyAttachmentConfig_forceNew(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(ctx, t, resourceName),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "managed_policy_arn", regexache.MustCompile(`policy/AmazonCognitoReadOnly`)),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_name", "AmazonCognitoReadOnly"),
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

func TestAccSSOAdminManagedPolicyAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPolicyAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPolicyAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfssoadmin.ResourceManagedPolicyAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminManagedPolicyAttachment_Disappears_permissionSet(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPolicyAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPolicyAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfssoadmin.ResourcePermissionSet(), permissionSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminManagedPolicyAttachment_multipleManagedPolicies(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	otherResourceName := "aws_ssoadmin_managed_policy_attachment.other"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPolicyAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPolicyAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccManagedPolicyAttachmentConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(ctx, t, resourceName),
					testAccCheckManagedPolicyAttachmentExists(ctx, t, otherResourceName),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(otherResourceName, "managed_policy_arn", regexache.MustCompile(`policy/AmazonDynamoDBReadOnlyAccess`)),
					resource.TestCheckResourceAttr(otherResourceName, "managed_policy_name", "AmazonDynamoDBReadOnlyAccess"),
					resource.TestCheckResourceAttrPair(otherResourceName, "instance_arn", permissionSetResourceName, "instance_arn"),
					resource.TestCheckResourceAttrPair(otherResourceName, "permission_set_arn", permissionSetResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      otherResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckManagedPolicyAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_managed_policy_attachment" {
				continue
			}

			managedPolicyARN, permissionSetARN, instanceARN, err := tfssoadmin.ParseManagedPolicyAttachmentID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfssoadmin.FindManagedPolicyByThreePartKey(ctx, conn, managedPolicyARN, permissionSetARN, instanceARN)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSO Managed Policy Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckManagedPolicyAttachmentExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		managedPolicyARN, permissionSetARN, instanceARN, err := tfssoadmin.ParseManagedPolicyAttachmentID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx)

		_, err = tfssoadmin.FindManagedPolicyByThreePartKey(ctx, conn, managedPolicyARN, permissionSetARN, instanceARN)

		return err
	}
}

func testAccManagedPolicyAttachmentConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccManagedPolicyAttachmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccManagedPolicyAttachmentConfig_base(rName), `
resource "aws_ssoadmin_managed_policy_attachment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AlexaForBusinessDeviceSetup"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`)
}

func testAccManagedPolicyAttachmentConfig_forceNew(rName string) string {
	return acctest.ConfigCompose(testAccManagedPolicyAttachmentConfig_base(rName), `
resource "aws_ssoadmin_managed_policy_attachment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonCognitoReadOnly"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`)
}

func testAccManagedPolicyAttachmentConfig_multiple(rName string) string {
	return acctest.ConfigCompose(testAccManagedPolicyAttachmentConfig_basic(rName), `
resource "aws_ssoadmin_managed_policy_attachment" "other" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonDynamoDBReadOnlyAccess"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`)
}
